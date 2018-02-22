// Copyright 2018 ThoughtWorks, Inc.

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

package pluginInfo

import (
	"fmt"
	"io/ioutil"
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

func sortPlugins(allPlugins map[string]PluginInfo) []PluginInfo {
	var installedPlugins []PluginInfo
	for _, plugin := range allPlugins {
		installedPlugins = append(installedPlugins, plugin)
	}
	sort.Sort(byPluginName(installedPlugins))
	return installedPlugins
}
