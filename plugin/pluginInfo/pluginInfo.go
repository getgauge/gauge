/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package pluginInfo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/version"
)

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

type byPath []PluginInfo

func (a byPath) Len() int      { return len(a) }
func (a byPath) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byPath) Less(i, j int) bool {
	return a[i].Path > a[j].Path
}

func GetPluginsInfo() ([]PluginInfo, error) {
	allPluginsWithVersion, err := GetAllInstalledPluginsWithVersion()
	if err != nil {
		return nil, fmt.Errorf("No plugins found\nPlugins can be installed with `gauge install {plugin-name}`")
	}
	return allPluginsWithVersion, nil
}

// GetAllInstalledPluginsWithVersion Fetches Latest version of all installed plugins.
var GetAllInstalledPluginsWithVersion = func() ([]PluginInfo, error) {
	pluginInstallPrefixes, err := common.GetPluginInstallPrefixes()
	if err != nil {
		return nil, err
	}
	allPlugins := make(map[string]PluginInfo)
	for _, prefix := range pluginInstallPrefixes {
		files, err := os.ReadDir(prefix)
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
			latestPlugin, err := GetLatestInstalledPlugin(filepath.Join(prefix, file.Name()))
			if err != nil {
				continue
			}
			allPlugins[file.Name()] = *latestPlugin
		}
	}
	return sortPlugins(allPlugins), nil
}

func GetLatestInstalledPlugin(pluginDir string) (*PluginInfo, error) {
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("Error listing files in plugin directory %s: %s", pluginDir, err.Error())
	}
	versionToPlugins := make(map[string][]PluginInfo)
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

func sortPlugins(allPlugins map[string]PluginInfo) []PluginInfo {
	var installedPlugins []PluginInfo
	for _, plugin := range allPlugins {
		installedPlugins = append(installedPlugins, plugin)
	}
	sort.Sort(byPluginName(installedPlugins))
	return installedPlugins
}
