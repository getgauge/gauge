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
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/version"
	"github.com/golang/protobuf/proto"
	"net"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	executionScope          = "execution"
	pluginConnectionPortEnv = "plugin_connection_port"
)

type pluginDescriptor struct {
	Id          string
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

type PluginHandler struct {
	pluginsMap map[string]*plugin
}

type plugin struct {
	connection net.Conn
	pluginCmd  *exec.Cmd
	descriptor *pluginDescriptor
}

func (plugin *plugin) kill(wg *sync.WaitGroup) error {
	defer wg.Done()
	if plugin.isStillRunning() {

		exited := make(chan bool, 1)
		go func() {
			for {
				if plugin.isStillRunning() {
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
				logger.Log.Debug("Plugin [%s] with pid [%d] has exited", plugin.descriptor.Name, plugin.pluginCmd.Process.Pid)
			}
		case <-time.After(config.PluginConnectionTimeout()):
			logger.Log.Warning("Plugin [%s] with pid [%d] did not exit after %.2f seconds. Forcefully killing it.", plugin.descriptor.Name, plugin.pluginCmd.Process.Pid, config.PluginConnectionTimeout().Seconds())
			return plugin.pluginCmd.Process.Kill()
		}
	}
	return nil
}

func (plugin *plugin) isStillRunning() bool {
	return plugin.pluginCmd.ProcessState == nil || !plugin.pluginCmd.ProcessState.Exited()
}

func IsPluginInstalled(pluginName, pluginVersion string) bool {
	pluginsInstallDir, err := common.GetPluginsInstallDir(pluginName)
	if err != nil {
		return false
	}

	thisPluginDir := path.Join(pluginsInstallDir, pluginName)
	if !common.DirExists(thisPluginDir) {
		return false
	}

	if pluginVersion != "" {
		pluginJson := path.Join(thisPluginDir, pluginVersion, common.PluginJsonFile)
		if common.FileExists(pluginJson) {
			return true
		} else {
			return false
		}
	} else {
		return true
	}
}

func getPluginJsonPath(pluginName, version string) (string, error) {
	if !IsPluginInstalled(pluginName, version) {
		return "", errors.New(fmt.Sprintf("Plugin %s %s is not installed", pluginName, version))
	}

	pluginInstallDir, err := common.GetPluginInstallDir(pluginName, "")
	if err != nil {
		return "", err
	}
	return filepath.Join(pluginInstallDir, common.PluginJsonFile), nil
}

func GetPluginDescriptor(pluginId, pluginVersion string) (*pluginDescriptor, error) {
	pluginJson, err := getPluginJsonPath(pluginId, pluginVersion)
	if err != nil {
		return nil, err
	}
	return GetPluginDescriptorFromJson(pluginJson)
}

func GetPluginDescriptorFromJson(pluginJson string) (*pluginDescriptor, error) {
	pluginJsonContents, err := common.ReadFileContents(pluginJson)
	if err != nil {
		return nil, err
	}
	var pd pluginDescriptor
	if err = json.Unmarshal([]byte(pluginJsonContents), &pd); err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", pluginJson, err.Error()))
	}
	pd.pluginPath = filepath.Dir(pluginJson)

	return &pd, nil
}

func StartPlugin(pd *pluginDescriptor, action string, wait bool) (*exec.Cmd, error) {
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
		return nil, errors.New(fmt.Sprintf("Platform specific command not specified: %s.", runtime.GOOS))
	}

	pluginLogger := &pluginLogger{pluginName: pd.Name}
	cmd, err := common.ExecuteCommand(command, pd.pluginPath, pluginLogger, pluginLogger)

	if err != nil {
		return nil, err
	}

	if wait {
		return cmd, cmd.Wait()
	} else {
		go func() {
			cmd.Wait()
		}()
	}

	return cmd, nil
}

func SetEnvForPlugin(action string, pd *pluginDescriptor, manifest *manifest.Manifest, pluginEnvVars map[string]string) error {
	pluginEnvVars[fmt.Sprintf("%s_action", pd.Id)] = action
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
	for _, pluginId := range manifest.Plugins {
		if pluginId == descriptor.Id {
			return true
		}
	}
	return false
}

func startPluginsForExecution(manifest *manifest.Manifest) (*PluginHandler, []string) {
	warnings := make([]string, 0)
	handler := &PluginHandler{}
	envProperties := make(map[string]string)

	for _, pluginId := range manifest.Plugins {
		pd, err := GetPluginDescriptor(pluginId, "")
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Error starting plugin %s. Failed to get plugin.json. %s. To install, run `gauge --install %s`.", pluginId, err.Error(), pluginId))
			continue
		}
		compatibilityErr := version.CheckCompatibility(version.CurrentGaugeVersion, &pd.GaugeVersionSupport)
		if compatibilityErr != nil {
			warnings = append(warnings, fmt.Sprintf("Compatible %s plugin version to current Gauge version %s not found", pd.Name, version.CurrentGaugeVersion))
			continue
		}
		if isExecutionScopePlugin(pd) {
			gaugeConnectionHandler, err := conn.NewGaugeConnectionHandler(0, nil)
			if err != nil {
				warnings = append(warnings, err.Error())
				continue
			}
			envProperties[pluginConnectionPortEnv] = strconv.Itoa(gaugeConnectionHandler.ConnectionPortNumber())
			SetEnvForPlugin(executionScope, pd, manifest, envProperties)

			pluginCmd, err := StartPlugin(pd, executionScope, false)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. %s", pd.Name, pd.Version, err.Error()))
				continue
			}
			pluginConnection, err := gaugeConnectionHandler.AcceptConnection(config.PluginConnectionTimeout(), make(chan error))
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. Failed to connect to plugin. %s", pd.Name, pd.Version, err.Error()))
				pluginCmd.Process.Kill()
				continue
			}
			handler.addPlugin(pluginId, &plugin{connection: pluginConnection, pluginCmd: pluginCmd, descriptor: pd})
		}

	}
	return handler, warnings
}

func isExecutionScopePlugin(pd *pluginDescriptor) bool {
	for _, scope := range pd.Scope {
		if strings.ToLower(scope) == executionScope {
			return true
		}
	}
	return false
}

func (handler *PluginHandler) addPlugin(pluginId string, pluginToAdd *plugin) {
	if handler.pluginsMap == nil {
		handler.pluginsMap = make(map[string]*plugin)
	}
	handler.pluginsMap[pluginId] = pluginToAdd
}

func (handler *PluginHandler) removePlugin(pluginId string) {
	delete(handler.pluginsMap, pluginId)
}

func (handler *PluginHandler) NotifyPlugins(message *gauge_messages.Message) {
	for id, plugin := range handler.pluginsMap {
		err := plugin.sendMessage(message)
		if err != nil {
			logger.Log.Error("Unable to connect to plugin %s %s. %s\n", plugin.descriptor.Name, plugin.descriptor.Version, err.Error())
			handler.killPlugin(id)
		}
	}
}

func (handler *PluginHandler) killPlugin(pluginId string) {
	plugin := handler.pluginsMap[pluginId]
	logger.Log.Debug("Killing Plugin %s %s\n", plugin.descriptor.Name, plugin.descriptor.Version)
	err := plugin.pluginCmd.Process.Kill()
	if err != nil {
		logger.Log.Error("Failed to kill plugin %s %s. %s\n", plugin.descriptor.Name, plugin.descriptor.Version, err.Error())
	}
	handler.removePlugin(pluginId)
}

func (handler *PluginHandler) GracefullyKillPlugins() {
	var wg sync.WaitGroup
	for _, plugin := range handler.pluginsMap {
		wg.Add(1)
		go plugin.kill(&wg)
	}
	wg.Wait()
}

func (plugin *plugin) sendMessage(message *gauge_messages.Message) error {
	messageId := common.GetUniqueId()
	message.MessageId = &messageId
	messageBytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	err = conn.Write(plugin.connection, messageBytes)
	if err != nil {
		return errors.New(fmt.Sprintf("[Warning] Failed to send message to plugin: %d  %s", plugin.descriptor.Id, err.Error()))
	}
	return nil
}

func StartPlugins(manifest *manifest.Manifest) *PluginHandler {
	PluginHandler, warnings := startPluginsForExecution(manifest)
	logger.HandleWarningMessages(warnings)
	return PluginHandler
}
