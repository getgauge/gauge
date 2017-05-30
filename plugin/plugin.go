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
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
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
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/version"
	"github.com/golang/protobuf/proto"
)

const (
	executionScope          = "execution"
	docScope                = "documentation"
	pluginConnectionPortEnv = "plugin_connection_port"
	debugEnv                = "debugging"
)

type pluginDescriptor struct {
	ID          string
	Version     string
	Name        string
	Description string
	Command     struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	Scope               []string
	GaugeVersionSupport version.VersionSupport
	pluginPath          string
}

type plugin struct {
	mutex      *sync.Mutex
	connection net.Conn
	pluginCmd  *exec.Cmd
	descriptor *pluginDescriptor
}

func (p *plugin) IsProcessRunning() bool {
	p.mutex.Lock()
	ps := p.pluginCmd.ProcessState
	p.mutex.Unlock()
	return ps == nil || !ps.Exited()
}

func (p *plugin) kill(wg *sync.WaitGroup) error {
	defer wg.Done()
	if p.IsProcessRunning() {
		defer p.connection.Close()
		conn.SendProcessKillMessage(p.connection)

		exited := make(chan bool, 1)
		go func() {
			for {
				if p.IsProcessRunning() {
					time.Sleep(100 * time.Millisecond)
				} else {
					exited <- true
					return
				}
			}
		}()
		select {
		case done := <-exited:
			if done {
				logger.Debug("Plugin [%s] with pid [%d] has exited", p.descriptor.Name, p.pluginCmd.Process.Pid)
			}
		case <-time.After(config.PluginKillTimeout()):
			logger.Warning("Plugin [%s] with pid [%d] did not exit after %.2f seconds. Forcefully killing it.", p.descriptor.Name, p.pluginCmd.Process.Pid, config.PluginKillTimeout().Seconds())
			err := p.pluginCmd.Process.Kill()
			if err != nil {
				logger.Warning("Error while killing plugin %s : %s ", p.descriptor.Name, err.Error())
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
		pluginJSON := filepath.Join(thisPluginDir, pluginVersion, common.PluginJSONFile)
		if common.FileExists(pluginJSON) {
			return true
		}
		return false
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

func StartPlugin(pd *pluginDescriptor, action string) (*plugin, error) {
	command := []string{}
	switch runtime.GOOS {
	case "windows":
		command = pd.Command.Windows
		break
	case "darwin":
		command = pd.Command.Darwin
		break
	default:
		command = pd.Command.Linux
		break
	}
	if len(command) == 0 {
		return nil, fmt.Errorf("Platform specific command not specified: %s.", runtime.GOOS)
	}

	cmd, err := common.ExecuteCommand(command, pd.pluginPath, reporter.Current(), reporter.Current())

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

func SetEnvForPlugin(action string, pd *pluginDescriptor, manifest *manifest.Manifest, pluginEnvVars map[string]string) error {
	pluginEnvVars[fmt.Sprintf("%s_action", pd.ID)] = action
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
			warnings = append(warnings, fmt.Sprintf("Unable to start plugin %s. %s. To install, run `gauge --install %s`.", pluginID, err.Error(), pluginID))
			continue
		}
		compatibilityErr := version.CheckCompatibility(version.CurrentGaugeVersion, &pd.GaugeVersionSupport)
		if compatibilityErr != nil {
			warnings = append(warnings, fmt.Sprintf("Compatible %s plugin version to current Gauge version %s not found", pd.Name, version.CurrentGaugeVersion))
			continue
		}
		if isPluginValidFor(pd, executionScope) {
			gaugeConnectionHandler, err := conn.NewGaugeConnectionHandler(0, nil)
			if err != nil {
				warnings = append(warnings, err.Error())
				continue
			}
			envProperties[pluginConnectionPortEnv] = strconv.Itoa(gaugeConnectionHandler.ConnectionPortNumber())
			err = SetEnvForPlugin(executionScope, pd, manifest, envProperties)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error setting environment for plugin %s %s. %s", pd.Name, pd.Version, err.Error()))
				continue
			}

			plugin, err := StartPlugin(pd, executionScope)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. %s", pd.Name, pd.Version, err.Error()))
				continue
			}
			pluginConnection, err := gaugeConnectionHandler.AcceptConnection(config.PluginConnectionTimeout(), make(chan error))
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. Failed to connect to plugin. %s", pd.Name, pd.Version, err.Error()))
				plugin.pluginCmd.Process.Kill()
				continue
			}
			plugin.connection = pluginConnection
			handler.addPlugin(pluginID, plugin)
		}

	}
	return handler, warnings
}

func GenerateDoc(pluginName string, specDirs []string, port int) {
	pd, err := GetPluginDescriptor(pluginName, "")
	if err != nil {
		logger.Fatalf("Error starting plugin %s. Failed to get plugin.json. %s. To install, run `gauge --install %s`.", pluginName, err.Error(), pluginName)
	}
	if err := version.CheckCompatibility(version.CurrentGaugeVersion, &pd.GaugeVersionSupport); err != nil {
		logger.Fatalf("Compatible %s plugin version to current Gauge version %s not found", pd.Name, version.CurrentGaugeVersion)
	}
	if !isPluginValidFor(pd, docScope) {
		logger.Fatalf("Invalid plugin name: %s, this plugin cannot generate documentation.", pd.Name)
	}
	var sources []string
	for _, src := range specDirs {
		path, _ := filepath.Abs(src)
		sources = append(sources, path)
	}
	os.Setenv("GAUGE_SPEC_DIRS", strings.Join(sources, " "))
	os.Setenv("GAUGE_PROJECT_ROOT", config.ProjectRoot)
	os.Setenv(common.APIPortEnvVariableName, strconv.Itoa(port))
	p, err := StartPlugin(pd, docScope)
	if err != nil {
		logger.Fatalf("Error starting plugin %s %s. %s", pd.Name, pd.Version, err.Error())
	}
	for p.IsProcessRunning() {
	}
}

func isPluginValidFor(pd *pluginDescriptor, scope string) bool {
	for _, s := range pd.Scope {
		if strings.ToLower(s) == scope {
			return true
		}
	}
	return false
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
	logger.HandleWarningMessages(warnings)
	return pluginHandler
}

func getLatestInstalledPlugin(pluginDir string) (*PluginInfo, error) {
	files, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("Error listing files in plugin directory %s: %s", pluginDir, err.Error())
	}
	versionToPlugins := make(map[string][]PluginInfo, 0)
	pluginName := filepath.Base(pluginDir)

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		v := file.Name()
		if strings.Contains(file.Name(), "nightly") {
			v = file.Name()[:strings.LastIndex(file.Name(), ".")]
		}
		vp, err := version.ParseVersion(v)
		if err == nil {
			versionToPlugins[v] = append(versionToPlugins[v], PluginInfo{pluginName, vp, filepath.Join(pluginDir, file.Name())})
		}
	}

	if len(versionToPlugins) < 1 {
		return nil, fmt.Errorf("No valid versions of plugin %s found in %s", pluginName, pluginDir)
	}

	var availableVersions []*version.Version
	for k := range versionToPlugins {
		vp, _ := version.ParseVersion(k)
		availableVersions = append(availableVersions, vp)
	}
	latestVersion := version.GetLatestVersion(availableVersions)
	latestBuild := getLatestOf(versionToPlugins[latestVersion.String()], latestVersion)
	return &latestBuild, nil
}

func getLatestOf(plugins []PluginInfo, latestVersion *version.Version) PluginInfo {
	for _, v := range plugins {
		if v.Path == latestVersion.String() {
			return v
		}
	}
	sort.Sort(byPath(plugins))
	return plugins[0]
}

func GetAllInstalledPluginsWithVersion() ([]PluginInfo, error) {
	pluginInstallPrefixes, err := common.GetPluginInstallPrefixes()
	if err != nil {
		return nil, err
	}
	allPlugins := make(map[string]PluginInfo, 0)
	for _, prefix := range pluginInstallPrefixes {
		files, err := ioutil.ReadDir(prefix)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			pluginDir, err := os.Stat(filepath.Join(prefix, file.Name()))
			if err != nil {
				continue
			}

			if !pluginDir.IsDir() {
				continue
			}
			latestPlugin, err := getLatestInstalledPlugin(filepath.Join(prefix, file.Name()))
			if err != nil {
				continue
			}
			allPlugins[file.Name()] = *latestPlugin
		}
	}
	return sortPlugins(allPlugins), nil
}

type PluginInfo struct {
	Name    string
	Version *version.Version
	Path    string
}

type byPluginName []PluginInfo

func (a byPluginName) Len() int      { return len(a) }
func (a byPluginName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byPluginName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

func sortPlugins(allPlugins map[string]PluginInfo) []PluginInfo {
	var installedPlugins []PluginInfo
	for _, plugin := range allPlugins {
		installedPlugins = append(installedPlugins, plugin)
	}
	sort.Sort(byPluginName(installedPlugins))
	return installedPlugins
}

type byPath []PluginInfo

func (a byPath) Len() int      { return len(a) }
func (a byPath) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byPath) Less(i, j int) bool {
	return a[i].Path > a[j].Path
}

func GetPluginsInfo() []PluginInfo {
	allPluginsWithVersion, err := GetAllInstalledPluginsWithVersion()
	if err != nil {
		logger.Info("No plugins found")
		logger.Info("Plugins can be installed with `gauge --install {plugin-name}`")
		os.Exit(0)
	}
	return allPluginsWithVersion
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
		latestPlugin, err := getLatestInstalledPlugin(pluginDir)
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
	pluginInfos, err := GetAllInstalledPluginsWithVersion()
	if err != nil {
		return ""
	}
	var plugins []string
	for _, p := range pluginInfos {
		plugins = append(plugins, p.Name)
	}
	return strings.Join(plugins, ",")
}
