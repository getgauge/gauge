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

package infoGatherer

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/util"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&MySuite{})

var concept1 []byte
var concept2 []byte
var spec1 []byte

type MySuite struct {
	specsDir   string
	projectDir string
}

func (s *MySuite) SetUpTest(c *C) {
	s.projectDir, _ = ioutil.TempDir("", "gaugeTest")
	s.specsDir, _ = util.CreateDirIn(s.projectDir, "specs")
	config.ProjectRoot = s.projectDir

	s.buildTestData()
}

func (s *MySuite) TearDownTest(c *C) {
	os.RemoveAll(s.projectDir)
}

func (s *MySuite) buildTestData() {
	concept1 = make([]byte, 0)
	concept1 = append(concept1, `# foo bar
* first step with "foo"
* say "hello" to me
* a "final" step
`...)

	concept2 = make([]byte, 0)
	concept2 = append(concept2, `# bar
* first step with "foo"
* say "hello" to me
* a "final" step
`...)

	spec1 = make([]byte, 0)
	spec1 = append(spec1, `Specification Heading
=====================
Scenario 1
----------
* say hello
* say "hello" to me
`...)
}

func (s *MySuite) TestGetParsedSpecs(c *C) {
	_, err := util.CreateFileIn(s.specsDir, "spec1.spec", spec1)
	c.Assert(err, Equals, nil)
	specInfoGatherer := new(SpecInfoGatherer)

	specFiles := util.FindSpecFilesIn(s.specsDir)
	specs := specInfoGatherer.getParsedSpecs(specFiles)

	c.Assert(len(specs), Equals, 1)
	c.Assert(specs[0].Heading.Value, Equals, "Specification Heading")
}

func (s *MySuite) TestGetParsedConcepts(c *C) {
	_, err := util.CreateFileIn(s.specsDir, "concept.cpt", concept1)
	c.Assert(err, Equals, nil)
	specInfoGatherer := new(SpecInfoGatherer)

	conceptsMap := specInfoGatherer.getParsedConcepts()

	c.Assert(len(conceptsMap), Equals, 1)
	c.Assert(conceptsMap["foo bar"], NotNil)
	c.Assert(specInfoGatherer.conceptDictionary, NotNil)
}

func (s *MySuite) TestGetParsedStepValues(c *C) {
	steps := []*gauge.Step{
		&gauge.Step{Value: "Step with a {}", LineText: "Step with a <table>", IsConcept: true, HasInlineTable: true},
		&gauge.Step{Value: "A context step", LineText: "A context step", IsConcept: false},
		&gauge.Step{Value: "Say {} to {}", LineText: "Say \"hello\" to \"gauge\"", IsConcept: false,
			Args: []*gauge.StepArg{
				&gauge.StepArg{Name: "first", Value: "hello", ArgType: gauge.Static},
				&gauge.StepArg{Name: "second", Value: "gauge", ArgType: gauge.Static}},
		},
	}

	stepValues := getParsedStepValues(steps)

	c.Assert(len(stepValues), Equals, 2)
	c.Assert(stepValues[0].StepValue, Equals, "A context step")
	c.Assert(stepValues[1].StepValue, Equals, "Say {} to {}")
}

func (s *MySuite) TestInitSpecsCache(c *C) {
	_, err := util.CreateFileIn(s.specsDir, "spec1.spec", spec1)
	c.Assert(err, Equals, nil)
	specInfoGatherer := new(SpecInfoGatherer)
	specInfoGatherer.waitGroup.Add(1)

	specInfoGatherer.initSpecsCache()

	c.Assert(len(specInfoGatherer.specsCache), Equals, 1)
}

func (s *MySuite) TestInitConceptsCache(c *C) {
	_, err := util.CreateFileIn(s.specsDir, "concept1.cpt", concept1)
	c.Assert(err, Equals, nil)
	_, err = util.CreateFileIn(s.specsDir, "concept2.cpt", concept2)
	c.Assert(err, Equals, nil)
	specInfoGatherer := new(SpecInfoGatherer)
	specInfoGatherer.waitGroup.Add(1)

	specInfoGatherer.initConceptsCache()

	c.Assert(len(specInfoGatherer.conceptsCache), Equals, 2)
}
