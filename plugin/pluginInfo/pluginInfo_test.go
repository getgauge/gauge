/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package pluginInfo

import (
	"fmt"
	"path/filepath"
	"testing"

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

	latestVersion, err := GetLatestInstalledPlugin(path)

	c.Assert(err, Equals, nil)
	c.Assert(latestVersion.Version.String(), Equals, "1.2.0")
	c.Assert(latestVersion.Name, Equals, "java")
	c.Assert(latestVersion.Path, Equals, filepath.Join(path, "1.2.0"))
}

func (s *MySuite) TestGetLatestPluginPathIfNoPluginsFound(c *C) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)

	_, err := GetLatestInstalledPlugin(path)

	c.Assert(err.Error(), Equals, fmt.Sprintf("No valid versions of plugin %s found in %s", testData, path))
}

func (s *MySuite) TestGetLatestInstalledPlugin(c *C) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "java"))

	latestPlugin, err := GetLatestInstalledPlugin(path)

	c.Assert(err, Equals, nil)
	c.Assert(latestPlugin.Path, Equals, filepath.Join(path, "1.2.0"))
}

func (s *MySuite) TestGetLatestInstalledPluginIfNoPluginsFound(c *C) {
	testData := "_testdata"
	path, _ := filepath.Abs(testData)

	_, err := GetLatestInstalledPlugin(path)

	c.Assert(err.Error(), Equals, fmt.Sprintf("No valid versions of plugin %s found in %s", testData, path))
}

func (s *MySuite) TestLatestVersionWithDifferentYearNightlies(c *C) {
	v, _ := version.ParseVersion("0.2.2")

	pluginInfo1 := PluginInfo{Version: v, Path: "0.2.2.nightly-2015-02-09"}
	pluginInfo2 := PluginInfo{Version: v, Path: "0.2.2.nightly-2016-02-09"}
	pluginInfo3 := PluginInfo{Version: v, Path: "0.2.2.nightly-2017-02-09"}
	plugins := []PluginInfo{pluginInfo1, pluginInfo3, pluginInfo2}
	latestBuild := getLatestOf(plugins, v)

	c.Assert(latestBuild.Path, Equals, pluginInfo3.Path)
	c.Assert(latestBuild.Version, Equals, v)
}

func (s *MySuite) TestLatestVersionWithDifferentMonthNightlies(c *C) {
	v, _ := version.ParseVersion("0.2.2")

	pluginInfo1 := PluginInfo{Version: v, Path: "0.2.2.nightly-2016-03-03"}
	pluginInfo2 := PluginInfo{Version: v, Path: "0.2.2.nightly-2016-02-03"}
	plugins := []PluginInfo{pluginInfo1, pluginInfo2}
	latestBuild := getLatestOf(plugins, v)

	c.Assert(latestBuild.Path, Equals, pluginInfo1.Path)
	c.Assert(latestBuild.Version, Equals, v)
}

func (s *MySuite) TestLatestVersionWithDifferentDaysNightlies(c *C) {
	v, _ := version.ParseVersion("0.2.2")

	pluginInfo1 := PluginInfo{Version: v, Path: "0.2.2.nightly-2016-02-03"}
	pluginInfo2 := PluginInfo{Version: v, Path: "0.2.2.nightly-2016-02-09"}
	plugins := []PluginInfo{pluginInfo1, pluginInfo2}
	latestBuild := getLatestOf(plugins, v)

	c.Assert(latestBuild.Path, Equals, pluginInfo2.Path)
	c.Assert(latestBuild.Version, Equals, v)
}

func (s *MySuite) TestLatestNightlyVersionWithDifferentStableVersion(c *C) {
	v, _ := version.ParseVersion("0.2.2")

	pluginInfo1 := PluginInfo{Version: v, Path: "0.2.2.nightly-2016-02-09"}
	pluginInfo2 := PluginInfo{Version: v, Path: "0.2.3.nightly-2016-02-09"}
	plugins := []PluginInfo{pluginInfo1, pluginInfo2}
	latestBuild := getLatestOf(plugins, v)

	c.Assert(latestBuild.Path, Equals, pluginInfo2.Path)
	c.Assert(latestBuild.Version, Equals, v)
}

func (s *MySuite) TestLatestNightlyVersionWithDifferentDates(c *C) {
	v, _ := version.ParseVersion("0.2.2")

	pluginInfo1 := PluginInfo{Version: v, Path: "2.1.1.nightly-2016-05-02"}
	pluginInfo2 := PluginInfo{Version: v, Path: "2.1.1.nightly-2016-04-27"}
	plugins := []PluginInfo{pluginInfo1, pluginInfo2}
	latestBuild := getLatestOf(plugins, v)

	c.Assert(latestBuild.Path, Equals, pluginInfo1.Path)
	c.Assert(latestBuild.Version, Equals, v)
}

func (s *MySuite) TestLatestVersionWithOnlyStableVersion(c *C) {
	v, _ := version.ParseVersion("0.2.2")

	pluginInfo1 := PluginInfo{Version: v, Path: "0.2.2"}
	plugins := []PluginInfo{pluginInfo1}
	latestBuild := getLatestOf(plugins, v)

	c.Assert(latestBuild.Version, Equals, v)
	c.Assert(latestBuild.Version, Equals, v)
}

func (s *MySuite) TestLatestVersionWithOnlyNightlyVersion(c *C) {
	v, _ := version.ParseVersion("0.2.2")

	pluginInfo1 := PluginInfo{Version: v, Path: "0.2.2.nightly-2016-02-09"}
	plugins := []PluginInfo{pluginInfo1}
	latestBuild := getLatestOf(plugins, v)

	c.Assert(latestBuild.Version, Equals, v)
	c.Assert(latestBuild.Version, Equals, v)
}
