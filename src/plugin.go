package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"net"
	"os"
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
	pluginConnectionTimeout = time.Second * 3
)

type pluginDescriptor struct {
	Id          string
	Version     string
	Name        string
	Description string
	Command     struct {
		Windows string
		Linux   string
		Darwin  string
	}
	Scope      []string
	pluginPath string
}

type pluginHandler struct {
	pluginsMap map[string]*plugin
}

type plugin struct {
	connection net.Conn
	process    *os.Process
	descriptor *pluginDescriptor
}

func (plugin *plugin) kill(wg *sync.WaitGroup) error {
	defer wg.Done()
	readyToClose := make(chan bool)
	go plugin.readTillClose(readyToClose)
	fmt.Println(fmt.Sprintf("Waiting for plugin [%s] with pid [%d] to close connection [%s]", plugin.descriptor.Name, plugin.process.Pid, plugin.connection.RemoteAddr()))
	select {
	case <-readyToClose:
		{
			fmt.Println(fmt.Sprintf("Plugin [%s] with pid [%d] has closed the connection [%s]", plugin.descriptor.Name, plugin.process.Pid, plugin.connection.RemoteAddr()))
			break
		}
	case <-time.After(pluginConnectionTimeout):
		{
			fmt.Println(fmt.Sprintf("Plugin [%s] with pid [%d] did not respond after %.2f seconds", plugin.descriptor.Name, plugin.process.Pid, pluginConnectionTimeout.Seconds()))
			break
		}
	}

	err := plugin.connection.Close()
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("Killing plugin [%s] with pid [%d]", plugin.descriptor.Name, plugin.process.Pid))
	err = plugin.process.Kill()
	return err
}
func (plugin *plugin) readTillClose(readyToClose chan bool) {
	data := make([]byte, 1024)
	//sees if plugin has closed the connection
	plugin.connection.Read(data)
	readyToClose <- true
}

func isPluginInstalled(pluginName, pluginVersion string) bool {
	pluginsPath, err := common.GetPluginsPath()
	if err != nil {
		return false
	}

	thisPluginDir := path.Join(pluginsPath, pluginName)
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

func getPluginJsonPath(pluginName, pluginVersion string) (string, error) {
	if !isPluginInstalled(pluginName, pluginVersion) {
		return "", errors.New(fmt.Sprintf("%s %s is not installed", pluginName, pluginVersion))
	}

	pluginsPath, err := common.GetPluginsPath()
	if err != nil {
		return "", err
	}

	thisPluginDir := path.Join(pluginsPath, pluginName)
	if pluginVersion != "" {
		return path.Join(thisPluginDir, pluginVersion, common.PluginJsonFile), nil
	} else {
		pluginJson := ""
		walkFn := func(path string, info os.FileInfo, err error) error {
			if info.Name() == common.PluginJsonFile {
				if pluginJson != "" {
					return errors.New(fmt.Sprintf("Multiple versions of '%s' found. Specify the exact version to be used", pluginName))
				}
				pluginJson = path
			}
			return nil
		}

		err := filepath.Walk(thisPluginDir, walkFn)
		if err != nil {
			return "", err
		}

		return pluginJson, nil
	}
}

func getPluginDescriptor(pluginName, pluginVersion string) (*pluginDescriptor, error) {
	pluginJson, err := getPluginJsonPath(pluginName, pluginVersion)
	if err != nil {
		return nil, err
	}

	pluginJsonContents, err := common.ReadFileContents(pluginJson)
	if err != nil {
		return nil, err
	}
	var pd pluginDescriptor
	if err = json.Unmarshal([]byte(pluginJsonContents), &pd); err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", pluginJson, err.Error()))
	}
	pd.pluginPath = strings.Replace(pluginJson, filepath.Base(pluginJson), "", -1)

	return &pd, nil
}

func startPlugin(pd *pluginDescriptor, action string, wait bool) (*os.Process, error) {
	command := ""
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

	cmd := common.GetExecutableCommand(path.Join(pd.pluginPath, command))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if wait {
		return cmd.Process, cmd.Wait()
	}

	return cmd.Process, nil
}

func setEnvForPlugin(action string, pd *pluginDescriptor, manifest *manifest, pluginArgs map[string]string) {
	os.Setenv(fmt.Sprintf("%s_action", pd.Id), action)
	os.Setenv("test_language", manifest.Language)
	setEnvironmentProperties(pluginArgs)
}

func setEnvironmentProperties(properties map[string]string) {
	for k, v := range properties {
		os.Setenv(k, v)
	}
}

func addPluginToTheProject(pluginName string, pluginArgs map[string]string, manifest *manifest) error {
	pd, err := getPluginDescriptor(pluginName, pluginArgs["version"])
	if err != nil {
		return err
	}

	action := "setup"
	setEnvForPlugin(action, pd, manifest, pluginArgs)
	if _, err := startPlugin(pd, action, true); err != nil {
		return err
	}

	manifest.Plugins = append(manifest.Plugins, pluginDetails{Id: pd.Id, Version: pd.Version})
	return manifest.save()
}

func startPluginsForExecution(manifest *manifest) (*pluginHandler, []string) {
	handler := &pluginHandler{}
	warnings := make([]string, 0)
	envProperties := make(map[string]string)
	pluginListener, err := newGaugeListener("", 0)
	envProperties["plugin_connection_port"] = strconv.Itoa((pluginListener.tcpListener.Addr().(*net.TCPAddr).Port))
	if err != nil {
		warnings = append(warnings, err.Error())
		return nil, warnings
	}

	for _, pluginDetails := range manifest.Plugins {
		pd, err := getPluginDescriptor(pluginDetails.Id, pluginDetails.Version)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. Failed to get plugin.json. %s", pluginDetails.Id, pluginDetails.Version, err.Error()))
			continue
		}
		if isExecutionScopePlugin(pd) {
			setEnvForPlugin(executionScope, pd, manifest, envProperties)

			pluginProcess, err := startPlugin(pd, executionScope, false)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. %s", pd.Name, pluginDetails.Version, err.Error()))
				continue
			}
			pluginConnection, err := pluginListener.acceptConnection(pluginConnectionTimeout)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Error starting plugin %s %s. Failed to connect to plugin. %s", pd.Name, pluginDetails.Version, err.Error()))
				pluginProcess.Kill()
				continue
			}
			handler.addPlugin(pluginDetails.Id, &plugin{connection: pluginConnection, process: pluginProcess, descriptor: pd})
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

func (handler *pluginHandler) addPlugin(pluginId string, pluginToAdd *plugin) {
	if handler.pluginsMap == nil {
		handler.pluginsMap = make(map[string]*plugin)
	}
	handler.pluginsMap[pluginId] = pluginToAdd
}

func (handler *pluginHandler) removePlugin(pluginId string) {
	delete(handler.pluginsMap, pluginId)
}

func (handler *pluginHandler) notifyPlugins(message *Message) {
	for id, plugin := range handler.pluginsMap {
		err := handler.sendMessage(plugin.connection, message)
		if err != nil {
			fmt.Printf("Unable to connect to plugin %s %s. %s\n", plugin.descriptor.Name, plugin.descriptor.Version, err.Error())
			handler.killPlugin(id)
		}
	}
}

func (handler *pluginHandler) sendMessage(conn net.Conn, message *Message) error {
	messageId := common.GetUniqueId()
	message.MessageId = &messageId
	return writeMessage(conn, message)
}

func (handler *pluginHandler) killPlugin(pluginId string) {
	plugin := handler.pluginsMap[pluginId]
	fmt.Printf("Killing Plugin %s %s\n", plugin.descriptor.Name, plugin.descriptor.Version)
	err := plugin.process.Kill()
	if err != nil {
		fmt.Printf("Failed to kill plugin %s %s. %s\n", plugin.descriptor.Name, plugin.descriptor.Version, err.Error())
	}
	handler.removePlugin(pluginId)
}

func (handler *pluginHandler) gracefullyKillPlugins() {
	var wg sync.WaitGroup
	for _, plugin := range handler.pluginsMap {
		wg.Add(1)
		go plugin.kill(&wg)
	}
	wg.Wait()
}
