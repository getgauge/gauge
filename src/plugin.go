package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/twist2/common"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
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

	pluginJsonContents := common.ReadFileContents(pluginJson)
	var pd pluginDescriptor
	if err = json.Unmarshal([]byte(pluginJsonContents), &pd); err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", pluginJson, err.Error()))
	}
	pd.pluginPath = strings.Replace(pluginJson, filepath.Base(pluginJson), "", -1)

	return &pd, nil
}

func startPlugin(pd *pluginDescriptor, action string, wait bool) error {
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
		return err
	}

	if wait {
		return cmd.Wait()
	}

	return nil
}

func setEnvForPlugin(action string, pd *pluginDescriptor, manifest *manifest, pluginArgs map[string]string) {
	os.Setenv(fmt.Sprintf("%s_action", pd.Id), action)
	os.Setenv("test_language", manifest.Language)
	for k, v := range pluginArgs {
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
	if err := startPlugin(pd, action, true); err != nil {
		return err
	}

	manifest.Plugins = append(manifest.Plugins, pluginDetails{Id: pd.Id, Version: pd.Version})
	return manifest.save()
}
