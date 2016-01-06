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

func (s *MySuite) TestGetPluginLatestVersion(c *C) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "java"))

	latestVersion, err := getPluginLatestVersion(path)

	c.Assert(err, Equals, nil)
	c.Assert(latestVersion.Major, Equals, 1)
	c.Assert(latestVersion.Minor, Equals, 2)
	c.Assert(latestVersion.Patch, Equals, 0)
}

func (s *MySuite) TestGetPluginLatestVersionIfNoPluginsFound(c *C) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)

	_, err := getPluginLatestVersion(path)

	c.Assert(err.Error(), Equals, fmt.Sprintf("No valid versions of plugin %s found in %s", testData, path))
}

func (s *MySuite) TestGetLatestInstalledPluginVersionPath(c *C) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "java"))

	vPath, err := GetLatestInstalledPluginVersionPath(path)

	c.Assert(err, Equals, nil)
	c.Assert(vPath, Equals, filepath.Join(path, "1.2.0"))
}

func (s *MySuite) TestGetLatestInstalledPluginVersionPathIfNoPluginsFound(c *C) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)

	vPath, err := GetLatestInstalledPluginVersionPath(path)

	c.Assert(err.Error(), Equals, fmt.Sprintf("No valid versions of plugin %s found in %s", testData, path))
	c.Assert(vPath, Equals, "")
}

func (s *MySuite) TestGetPluginDescriptorFromJSON(c *C) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)

	pd, err := GetPluginDescriptorFromJSON(filepath.Join(path, "_test.json"))

	c.Assert(err, Equals, nil)
	c.Assert(pd.Id, Equals, "html-report")
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
