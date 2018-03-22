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

package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/plugin/pluginInfo"
	"github.com/getgauge/gauge/version"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

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

func TestGetPluginsWithoutScope(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join("_testdata"))
	os.Setenv(common.GaugeHome, path)

	got := PluginsWithoutScope()

	want := []pluginInfo.PluginInfo{
		{
			Name:    "noscope",
			Version: &version.Version{Major: 1, Minor: 0, Patch: 0},
			Path:    filepath.Join(path, "plugins", "noscope", "1.0.0"),
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Failed GetPluginWithoutScope.\n\tWant: %v\n\tGot: %v", want, got)
	}
}
