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

package reporter

import (
	"fmt"

	. "gopkg.in/check.v1"
)

type dummyWriter struct {
	output string
}

func (dw *dummyWriter) Write(b []byte) (int, error) {
	dw.output += string(b)
	return len(b), nil
}

func newDummyWriter() *dummyWriter {
	return &dummyWriter{}
}

func setup() (*dummyWriter, *simpleConsole) {
	dw := newDummyWriter()
	sc := newSimpleConsole(dw)
	return dw, sc
}

func (s *MySuite) TestSpecStart(c *C) {
	dw, sc := setup()
	sc.SpecStart("Specification heading")
	c.Assert(dw.output, Equals, "# Specification heading\n")
}

func (s *MySuite) TestSpecEnd(c *C) {
	dw, sc := setup()
	sc.SpecEnd()
	c.Assert(dw.output, Equals, "\n")
}

func (s *MySuite) TestScenarioStart(c *C) {
	dw, sc := setup()
	sc.ScenarioStart("First Scenario")
	c.Assert(dw.output, Equals, "  ## First Scenario\n")
}

func (s *MySuite) TestScenarioEnd(c *C) {
	_, sc := setup()
	sc.indentation = 2

	sc.ScenarioEnd(true)

	c.Assert(sc.indentation, Equals, 0)
}

func (s *MySuite) TestStepStartInVerboseMode(c *C) {
	dw, sc := setup()
	sc.indentation = 2
	Verbose = true

	sc.StepStart("* Say hello to gauge")

	c.Assert(dw.output, Equals, "      * Say hello to gauge\n")
}

func (s *MySuite) TestStepStartInNonVerboseMode(c *C) {
	dw, sc := setup()
	sc.indentation = 2
	Verbose = false

	sc.StepStart("* Say hello to gauge")

	c.Assert(dw.output, Equals, "")
}

func (s *MySuite) TestStepEnd(c *C) {
	_, sc := setup()
	sc.indentation = 6

	sc.StepEnd(true)

	c.Assert(sc.indentation, Equals, 2)
}

func (s *MySuite) TestSingleConceptStartInVerboseMode(c *C) {
	dw, sc := setup()
	sc.indentation = 2
	Verbose = true

	sc.ConceptStart("* my first concept")

	c.Assert(dw.output, Equals, fmt.Sprintf("%s* my first concept\n", spaces(6)))
}

func (s *MySuite) TestNestedConceptStartInVerboseMode_case1(c *C) {
	dw, sc := setup()
	sc.indentation = 2
	Verbose = true

	sc.ConceptStart("* my first concept")
	dw.output = ""
	sc.ConceptStart("* my second concept")

	c.Assert(dw.output, Equals, fmt.Sprintf("%s* my second concept\n", spaces(10)))
}

func (s *MySuite) TestNestedConceptStartInVerboseMode_case2(c *C) {
	dw, sc := setup()
	sc.indentation = 2
	Verbose = true

	sc.ConceptStart("* my first concept")
	dw.output = ""
	sc.StepStart("* do foo bar")

	c.Assert(dw.output, Equals, fmt.Sprintf("%s* do foo bar\n", spaces(10)))
}

func (s *MySuite) TestNestedConceptStartInVerboseMode_case3(c *C) {
	dw, sc := setup()
	sc.indentation = 2
	Verbose = true

	sc.ConceptStart("* my first concept")
	sc.ConceptStart("* my second concept")
	dw.output = ""
	sc.StepStart("* do foo bar")

	c.Assert(dw.output, Equals, fmt.Sprintf("%s* do foo bar\n", spaces(14)))
}

func (s *MySuite) TestConceptEnd(c *C) {
	_, sc := setup()
	sc.indentation = 6
	Verbose = true

	sc.ConceptEnd(false)

	c.Assert(sc.indentation, Equals, 2)
}

func (s *MySuite) TestDataTable(c *C) {
	dw, sc := setup()
	sc.indentation = 2
	Verbose = true
	table := `|Product|Description                  |
|-------|-----------------------------|
|Gauge  |Test automation with ease    |`

	want := `  |Product|Description                  |
  |-------|-----------------------------|
  |Gauge  |Test automation with ease    |
`

	sc.DataTable(table)

	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestError(c *C) {
	dw, sc := setup()
	sc.indentation = 6
	Verbose = true

	sc.Error("Failed %s", "network error")

	c.Assert(dw.output, Equals, fmt.Sprintf("%sFailed network error\n", spaces(sc.indentation+sysoutIndentation)))
}

func (s *MySuite) TestWrite(c *C) {
	dw, sc := setup()
	sc.indentation = 6
	Verbose = true
	input := "hello, gauge"

	n, err := sc.Write([]byte(input))

	c.Assert(err, Equals, nil)
	c.Assert(n, Equals, len(input+newline)+sc.indentation+sysoutIndentation)
	c.Assert(dw.output, Equals, fmt.Sprintf("%s%s\n", spaces(sc.indentation+sysoutIndentation), input))
}

func (s *MySuite) TestSpecReporting(c *C) {
	dw, sc := setup()
	//	sc.indentation = 6
	Verbose = true

	sc.SpecStart("Specification heading")
	sc.ScenarioStart("My First scenario")
	sc.StepStart("* do foo bar")
	sc.Write([]byte("doing foo bar"))
	sc.StepEnd(false)
	sc.ScenarioEnd(false)
	sc.SpecEnd()

	want := `# Specification heading
  ## My First scenario
      * do foo bar
        doing foo bar

`

	c.Assert(dw.output, Equals, want)
}
