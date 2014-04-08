package main

import (
	"errors"
	"fmt"
	"github.com/twist2/common"
	"os"
	"path"
	"path/filepath"
)

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

func addPluginToTheProject(pluginName, pluginVersion string) error {
	_, err := getPluginJsonPath(pluginName, pluginVersion)
	if err != nil {
		return err
	}

	return nil
}
