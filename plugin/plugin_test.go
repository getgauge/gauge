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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/version"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestSortingOfPluginInfos(c *C) {
	plugins := make(map[string]PluginInfo)
	plugins["e"] = PluginInfo{Name: "e"}
	plugins["b"] = PluginInfo{Name: "b"}
	plugins["c"] = PluginInfo{Name: "c"}
	plugins["d"] = PluginInfo{Name: "d"}
	plugins["a"] = PluginInfo{Name: "a"}

	actual := sortPlugins(plugins)

	var expected []PluginInfo
	expected = append(expected, PluginInfo{Name: "a"})
	expected = append(expected, PluginInfo{Name: "b"})
	expected = append(expected, PluginInfo{Name: "c"})
	expected = append(expected, PluginInfo{Name: "d"})
	expected = append(expected, PluginInfo{Name: "e"})

	c.Assert(len(expected), Equals, len(plugins))
	for i := range expected {
		c.Assert(expected[i], Equals, actual[i])
	}
}

func (s *MySuite) TestGetLatestPluginPath(c *C) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "java"))

	latestVersion, err := getLatestInstalledPlugin(path)

	c.Assert(err, Equals, nil)
	c.Assert(latestVersion.Version.String(), Equals, "1.2.0")
	c.Assert(latestVersion.Name, Equals, "java")
	c.Assert(latestVersion.Path, Equals, filepath.Join(path, "1.2.0"))
}

func (s *MySuite) TestGetLatestPluginPathIfNoPluginsFound(c *C) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)

	_, err := getLatestInstalledPlugin(path)

	c.Assert(err.Error(), Equals, fmt.Sprintf("No valid versions of plugin %s found in %s", testData, path))
}

func (s *MySuite) TestGetLatestInstalledPlugin(c *C) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "java"))

	latestPlugin, err := getLatestInstalledPlugin(path)

	c.Assert(err, Equals, nil)
	c.Assert(latestPlugin.Path, Equals, filepath.Join(path, "1.2.0"))
}

func (s *MySuite) TestGetLatestInstalledPluginIfNoPluginsFound(c *C) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)

	_, err := getLatestInstalledPlugin(path)

	c.Assert(err.Error(), Equals, fmt.Sprintf("No valid versions of plugin %s found in %s", testData, path))
}

func (s *MySuite) TestGetPluginDescriptorFromJSON(c *C) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)

	pd, err := GetPluginDescriptorFromJSON(filepath.Join(path, "_test.json"))

	c.Assert(err, Equals, nil)
	c.Assert(pd.ID, Equals, "html-report")
	c.Assert(pd.Version, Equals, "1.1.0")
	c.Assert(pd.Name, Equals, "Html Report")
	c.Assert(pd.Description, Equals, "Html reporting plugin")
	c.Assert(pd.pluginPath, Equals, path)
	c.Assert(pd.GaugeVersionSupport.Minimum, Equals, "0.2.0")
	c.Assert(pd.GaugeVersionSupport.Maximum, Equals, "0.4.0")
	c.Assert(pd.Scope, DeepEquals, []string{"Execution"})
	htmlCommand := []string{"bin/html-report"}
	c.Assert(pd.Command.Windows, DeepEquals, htmlCommand)
	c.Assert(pd.Command.Darwin, DeepEquals, htmlCommand)
	c.Assert(pd.Command.Linux, DeepEquals, htmlCommand)
}

func (s *MySuite) TestGetPluginDescriptorFromNonExistingJSON(c *C) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)
	JSONPath := filepath.Join(path, "_test1.json")
	_, err := GetPluginDescriptorFromJSON(JSONPath)

	c.Assert(err, DeepEquals, fmt.Errorf("File %s doesn't exist.", JSONPath))
}

func (s *MySuite) TestGetStablePluginAmongGivenPluginsOfAVersion(c *C) {
	v, _ := version.ParseVersion("0.2.2")

	pluginInfo1 := PluginInfo{Version: v, Path: "0.2.2"}
	plugins := []PluginInfo{pluginInfo1}
	latestBuild := getLatestOf(plugins, v)
	c.Assert(latestBuild.Version, Equals, v)

	pluginInfo2 := PluginInfo{Version: v, Path: "0.2.2.nightly-2016-02-09"}
	plugins = []PluginInfo{pluginInfo2}
	latestBuild = getLatestOf(plugins, v)
	c.Assert(latestBuild.Path, Equals, pluginInfo2.Path)
	c.Assert(latestBuild.Version, Equals, v)

	pluginInfo1.Path = "0.2.2.nightly-2015-02-03"
	pluginInfo2.Path = "0.2.2.nightly-2016-02-09"
	pluginInfo3 := PluginInfo{Version: v, Path: "0.2.2.nightly-2017-02-09"}
	plugins = []PluginInfo{pluginInfo1, pluginInfo3, pluginInfo2}
	latestBuild = getLatestOf(plugins, v)
	c.Assert(latestBuild.Path, Equals, pluginInfo3.Path)
	c.Assert(latestBuild.Version, Equals, v)

	pluginInfo1.Path = "0.2.2.nightly-2015-02-03"
	pluginInfo2.Path = "0.2.2.nightly-2016-02-04"
	plugins = []PluginInfo{pluginInfo1, pluginInfo2}
	latestBuild = getLatestOf(plugins, v)
	c.Assert(latestBuild.Path, Equals, pluginInfo2.Path)
	c.Assert(latestBuild.Version, Equals, v)

	pluginInfo1.Path = "0.2.2.nightly-2015-01-03"
	pluginInfo2.Path = "0.2.2.nightly-2015-02-03"
	plugins = []PluginInfo{pluginInfo1, pluginInfo2}
	latestBuild = getLatestOf(plugins, v)
	c.Assert(latestBuild.Path, Equals, pluginInfo2.Path)

	pluginInfo1.Path = "0.2.2.nightly-2015-01-03"
	pluginInfo2.Path = "0.2.2.nightly-2016-02-03"
	plugins = []PluginInfo{pluginInfo1, pluginInfo2}
	latestBuild = getLatestOf(plugins, v)
	c.Assert(latestBuild.Path, Equals, pluginInfo2.Path)

	pluginInfo1.Path = "0.2.2.nightly-2015-01-03"
	pluginInfo2.Path = "0.2.2.nightly-2017-02-03"
	pluginInfo2.Path = "0.2.2.nightly-2016-02-03"
	plugins = []PluginInfo{pluginInfo1, pluginInfo2}
	latestBuild = getLatestOf(plugins, v)
	c.Assert(latestBuild.Path, Equals, pluginInfo2.Path)

	pluginInfo1.Path = "0.2.2.nightly-2017-01-03"
	pluginInfo2.Path = "0.2.2.nightly-2017-01-05"
	pluginInfo2.Path = "0.2.2.nightly-2017-01-04"
	plugins = []PluginInfo{pluginInfo1, pluginInfo2}
	latestBuild = getLatestOf(plugins, v)
	c.Assert(latestBuild.Path, Equals, pluginInfo2.Path)
}

func (s *MySuite) TestGetLanguageQueryParamWhenProjectRootNotSet(c *C) {
	config.ProjectRoot = ""

	l := language()

	c.Assert(l, Equals, "")
}

func (s *MySuite) TestGetLanguageQueryParam(c *C) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "sample"))
	config.ProjectRoot = path

	l := language()

	c.Assert(l, Equals, "java")
}
