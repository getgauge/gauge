// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package plugin

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin/pluginInfo"
	"github.com/getgauge/gauge/version"
	"github.com/golang/protobuf/proto"
)

type pluginScope string

const (
	executionScope          pluginScope = "execution"
	docScope                pluginScope = "documentation"
	pluginConnectionPortEnv             = "plugin_connection_port"
)

type plugin struct {
	mutex      *sync.Mutex
	connection net.Conn
	pluginCmd  *exec.Cmd
	descriptor *pluginDescriptor
	killTimer  *time.Timer
}

func isProcessRunning(p *plugin) bool {
	p.mutex.Lock()
	ps := p.pluginCmd.ProcessState
	p.mutex.Unlock()
	return ps == nil || !ps.Exited()
}

func (p *plugin) rejuvenate() error {
	if p.killTimer == nil {
		return fmt.Errorf("timer is uninitialized. Perhaps kill is not yet invoked")
	}
	logger.Debugf(true, "Extending the plugin_kill_timeout for %s", p.descriptor.ID)
	p.killTimer.Reset(config.PluginKillTimeout())
	return nil
}

func (p *plugin) kill(wg *sync.WaitGroup) error {
	defer wg.Done()
	if isProcessRunning(p) {
		defer p.connection.Close()
		p.killTimer = time.NewTimer(config.PluginKillTimeout())
		err := conn.SendProcessKillMessage(p.connection)
		if err != nil {
			logger.Warningf(true, "Error while killing plugin %s : %s ", p.descriptor.Name, err.Error())
		}

		exited := make(chan bool, 1)
		go func() {
			for {
				if isProcessRunning(p) {
					time.Sleep(100 * time.Millisecond)
				} else {
					exited <- true
					return
				}
			}
		}()
		select {
		case <-exited:
			if !p.killTimer.Stop() {
				<-p.killTimer.C
			}
			logger.Debugf(true, "Plugin [%s] with pid [%d] has exited", p.descriptor.Name, p.pluginCmd.Process.Pid)
		case <-p.killTimer.C:
			logger.Warningf(true, "Plugin [%s] with pid [%d] did not exit after %.2f seconds. Forcefully killing it.", p.descriptor.Name, p.pluginCmd.Process.Pid, config.PluginKillTimeout().Seconds())
			err := p.pluginCmd.Process.Kill()
			if err != nil {
				logger.Warningf(true, "Error while killing plugin %s : %s ", p.descriptor.Name, err.Error())
			}
			return err
		}
	}
	return nil
}

func IsPluginInstalled(pluginName, pluginVersion string) bool {
	pluginsInstallDir, err := common.GetPluginsInstallDir(pluginName)
	if err != nil {
		return false
	}

	thisPluginDir := filepath.Join(pluginsInstallDir, pluginName)
	if !common.DirExists(thisPluginDir) {
		return false
	}

	if pluginVersion != "" {
		return common.FileExists(filepath.Join(thisPluginDir, pluginVersion, common.PluginJSONFile))
	}
	return true
}

func getPluginJSONPath(pluginName, pluginVersion string) (string, error) {
	if !IsPluginInstalled(pluginName, pluginVersion) {
		plugin := strings.TrimSpace(fmt.Sprintf("%s %s", pluginName, pluginVersion))
		return "", fmt.Errorf("Plugin %s is not installed", plugin)
	}

	pluginInstallDir, err := GetInstallDir(pluginName, "")
	if err != nil {
		return "", err
	}
	return filepath.Join(pluginInstallDir, common.PluginJSONFile), nil
}

func GetPluginDescriptor(pluginID, pluginVersion string) (*pluginDescriptor, error) {
	pluginJSON, err := getPluginJSONPath(pluginID, pluginVersion)
	if err != nil {
		return nil, err
	}
	return GetPluginDescriptorFromJSON(pluginJSON)
}

func GetPluginDescriptorFromJSON(pluginJSON string) (*pluginDescriptor, error) {
	pluginJSONContents, err := common.ReadFileContents(pluginJSON)
	if err != nil {
		return nil, err
	}
	var pd pluginDescriptor
	if err = json.Unmarshal([]byte(pluginJSONContents), &pd); err != nil {
		return nil, fmt.Errorf("%s: %s", pluginJSON, err.Error())
	}
	pd.pluginPath = filepath.Dir(pluginJSON)

	return &pd, nil
}

func StartPlugin(pd *pluginDescriptor, action pluginScope) (*plugin, error) {
	var command []string
	switch runtime.GOOS {
	case "windows":
		command = pd.Command.Windows
	case "darwin":
		command = pd.Command.Darwin
	default:
		command = pd.Command.Linux
	}
	if len(command) == 0 {
		return nil, fmt.Errorf("Platform specific command not specified: %s.", runtime.GOOS)
	}

	writer := logger.NewLogWriter(pd.ID, true, 0)
	cmd, err := common.ExecuteCommand(command, pd.pluginPath, writer.Stdout, writer.Stderr)

	if err != nil {
		return nil, err
	}
	var mutex = &sync.Mutex{}
	go func() {
		pState, _ := cmd.Process.Wait()
		mutex.Lock()
		cmd.ProcessState = pState
		mutex.Unlock()
	}()
	plugin := &plugin{pluginCmd: cmd, descriptor: pd, mutex: mutex}
	return plugin, nil
}

func SetEnvForPlugin(action pluginScope, pd *pluginDescriptor, manifest *manifest.Manifest, pluginEnvVars map[string]string) error {
	pluginEnvVars[fmt.Sprintf("%s_action", pd.ID)] = string(action)
	pluginEnvVars["test_language"] = manifest.Language
	if err := setEnvironmentProperties(pluginEnvVars); err != nil {
		return err
	}
	return nil
}

func setEnvironmentProperties(properties map[string]string) error {
	for k, v := range properties {
		if err := common.SetEnvVariable(k, v); err != nil {
			return err
		}
	}
	return nil
}

func IsPluginAdded(manifest *manifest.Manifest, descriptor *pluginDescriptor) bool {
	for _, pluginID := range manifest.Plugins {
		if pluginID == descriptor.ID {
			return true
		}
	}
	return false
}

func startPluginsForExecution(manifest *manifest.Manifest) (Handler, []string) {
	var warnings []string
	handler := &GaugePlugins{}
	envProperties := make(map[string]string)

	for _, pluginID := range manifest.Plugins {
		pd, err := GetPluginDescriptor(pluginID, "")
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Unable to start plugin %s. %s. To install, run `gauge install %s`.", pluginID, err.Error(), pluginID))
			continue
		}
		compatibilityErr := version.CheckCompatibility(version.CurrentGaugeVersion, &pd.GaugeVersionSupport)
		if compatibilityErr != nil {
			warnings = append(warnings, fmt.Sprintf("Compatible %s plugin version to current Gauge version %s not found", pd.Name, version.CurrentGaugeVersion))
			continue
		}
		if pd.hasScope(executionScope) {
			gaugeConnectionHandler, err := conn.NewGaugeConnectionHandler(0, &keepAliveHandler{ph: handler})
			if err != nil {
				warnings = append(warnings, err.Error())
				continue
			}
			envProperties[pluginConnectionPortEnv] = strconv.Itoa(gaugeConnectionHandler.ConnectionPortNumber())
			prop, err := common.GetGaugeConfiguration()
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Unable to read Gauge configuration. %s", err.Error()))
				continue
			}
			envProperties["plugin_kill_timeout"] = prop["plugin_kill_timeout"]
			err = SetEnvForPlugin(executionScope, pd, manifest, envProperties)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error setting environment for plugin %s %s. %s", pd.Name, pd.Version, err.Error()))
				continue
			}
			logger.Debugf(true, "Starting %s plugin", pd.Name)
			plugin, err := StartPlugin(pd, executionScope)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. %s", pd.Name, pd.Version, err.Error()))
				continue
			}
			pluginConnection, err := gaugeConnectionHandler.AcceptConnection(config.PluginConnectionTimeout(), make(chan error))
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. Failed to connect to plugin. %s", pd.Name, pd.Version, err.Error()))
				err := plugin.pluginCmd.Process.Kill()
				if err != nil {
					logger.Errorf(false, "unable to kill plugin %s: %s", plugin.descriptor.Name, err.Error())
				}
				continue
			}
			logger.Debugf(true, "Established connection to %s plugin", pd.Name)
			plugin.connection = pluginConnection
			handler.addPlugin(pluginID, plugin)
		}

	}
	return handler, warnings
}

func GenerateDoc(pluginName string, specDirs []string, port int) {
	pd, err := GetPluginDescriptor(pluginName, "")
	if err != nil {
		logger.Fatalf(true, "Error starting plugin %s. Failed to get plugin.json. %s. To install, run `gauge install %s`.", pluginName, err.Error(), pluginName)
	}
	if err := version.CheckCompatibility(version.CurrentGaugeVersion, &pd.GaugeVersionSupport); err != nil {
		logger.Fatalf(true, "Compatible %s plugin version to current Gauge version %s not found", pd.Name, version.CurrentGaugeVersion)
	}
	if !pd.hasScope(docScope) {
		logger.Fatalf(true, "Invalid plugin name: %s, this plugin cannot generate documentation.", pd.Name)
	}
	var sources []string
	for _, src := range specDirs {
		path, _ := filepath.Abs(src)
		sources = append(sources, path)
	}
	os.Setenv("GAUGE_SPEC_DIRS", strings.Join(sources, "||"))
	os.Setenv("GAUGE_PROJECT_ROOT", config.ProjectRoot)
	os.Setenv(common.APIPortEnvVariableName, strconv.Itoa(port))
	p, err := StartPlugin(pd, docScope)
	if err != nil {
		logger.Fatalf(true, "Error starting plugin %s %s. %s", pd.Name, pd.Version, err.Error())
	}
	for isProcessRunning(p) {
	}
}

func (p *plugin) sendMessage(message *gauge_messages.Message) error {
	messageID := common.GetUniqueID()
	message.MessageId = messageID
	messageBytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	err = conn.Write(p.connection, messageBytes)
	if err != nil {
		return fmt.Errorf("[Warning] Failed to send message to plugin: %s  %s", p.descriptor.ID, err.Error())
	}
	return nil
}

func StartPlugins(manifest *manifest.Manifest) Handler {
	pluginHandler, warnings := startPluginsForExecution(manifest)
	logger.HandleWarningMessages(true, warnings)
	return pluginHandler
}

func PluginsWithoutScope() (infos []pluginInfo.PluginInfo) {
	if plugins, err := pluginInfo.GetAllInstalledPluginsWithVersion(); err == nil {
		for _, p := range plugins {
			pd, err := GetPluginDescriptor(p.Name, p.Version.String())
			if err == nil && !pd.hasAnyScope() {
				infos = append(infos, p)
			}
		}
	}
	return
}

// GetInstallDir returns the install directory of given plugin and a given version.
func GetInstallDir(pluginName, version string) (string, error) {
	allPluginsInstallDir, err := common.GetPluginsInstallDir(pluginName)
	if err != nil {
		return "", err
	}
	pluginDir := filepath.Join(allPluginsInstallDir, pluginName)
	if version != "" {
		pluginDir = filepath.Join(pluginDir, version)
	} else {
		latestPlugin, err := pluginInfo.GetLatestInstalledPlugin(pluginDir)
		if err != nil {
			return "", err
		}
		pluginDir = latestPlugin.Path
	}
	return pluginDir, nil
}

func GetLanguageJSONFilePath(language string) (string, error) {
	languageInstallDir, err := GetInstallDir(language, "")
	if err != nil {
		return "", err
	}
	languageJSON := filepath.Join(languageInstallDir, fmt.Sprintf("%s.json", language))
	if !common.FileExists(languageJSON) {
		return "", fmt.Errorf("Failed to find the implementation for: %s. %s does not exist.", language, languageJSON)
	}

	return languageJSON, nil
}

func IsLanguagePlugin(plugin string) bool {
	if _, err := GetLanguageJSONFilePath(plugin); err != nil {
		return false
	}
	return true
}

func QueryParams() string {
	return fmt.Sprintf("?l=%s&p=%s&o=%s&a=%s", language(), plugins(), runtime.GOOS, runtime.GOARCH)
}

func language() string {
	if config.ProjectRoot == "" {
		return ""
	}
	m, err := manifest.ProjectManifest()
	if err != nil {
		return ""
	}
	return m.Language
}

func plugins() string {
	pluginInfos, err := pluginInfo.GetAllInstalledPluginsWithVersion()
	if err != nil {
		return ""
	}
	var plugins []string
	for _, p := range pluginInfos {
		plugins = append(plugins, p.Name)
	}
	return strings.Join(plugins, ",")
}
