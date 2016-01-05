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
