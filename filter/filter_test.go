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

package filter

import (
	"path/filepath"

	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestAddSpecsToMapPopulatesScenarioInExistingSpec(c *C) {
	specsMap := make(map[string]*gauge.Specification)
	scenario1 := &gauge.Scenario{Heading: &gauge.Heading{Value: "scenario1"}}
	scenario2 := &gauge.Scenario{Heading: &gauge.Heading{Value: "scenario2"}}
	heading := &gauge.Heading{Value: "spec heading"}
	specName := "foo.spec"
	spec1 := &gauge.Specification{Heading: heading, FileName: specName, Scenarios: []*gauge.Scenario{scenario1}, Items: []gauge.Item{heading, scenario1}}
	spec2 := &gauge.Specification{Heading: heading, FileName: specName, Scenarios: []*gauge.Scenario{scenario2}, Items: []gauge.Item{heading, scenario2}}
	specsMap[specName] = spec1
	addSpecsToMap([]*gauge.Specification{spec2}, specsMap)

	c.Assert(len(specsMap), Equals, 1)
	c.Assert(len(specsMap[specName].Scenarios), Equals, 2)
	c.Assert(len(specsMap[specName].Items), Equals, 3)
}

func (s *MySuite) TestSpecsFormArgsForMultipleIndexedArgsForOneSpec(c *C) {
	specs, _ := specsFromArgs(gauge.NewConceptDictionary(), []string{filepath.Join("testdata", "sample.spec:3"), filepath.Join("testdata", "sample.spec:6")})

	c.Assert(len(specs), Equals, 1)
	c.Assert(len(specs[0].Scenarios), Equals, 2)
}
