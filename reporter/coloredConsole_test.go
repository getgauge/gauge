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

var (
	eraseLine = "\x1b[2K\r"
	cursorUp  = "\x1b[0A"
)

func setupColoredConsole() (*dummyWriter, *coloredConsole) {
	dw := newDummyWriter()
	cc := newColoredConsole(dw)
	return dw, cc
}

func (s *MySuite) TestSpecStart_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()

	cc.SpecStart("Spec heading")

	c.Assert(dw.output, Equals, "# Spec heading\n")
}

func (s *MySuite) TestSpecEnd_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()

	cc.SpecEnd()

	c.Assert(dw.output, Equals, "\n")
}

func (s *MySuite) TestScenarioStartInVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = false
	cc.indentation = 2

	cc.ScenarioStart("my first scenario")

	c.Assert(dw.output, Equals, "    ## my first scenario\t")
}

func (s *MySuite) TestScenarioStartInNonVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 2

	cc.ScenarioStart("my first scenario")

	c.Assert(dw.output, Equals, "    ## my first scenario\t\n")
}

func (s *MySuite) TestScenarioEndInNonVerbose_ColoredConsole(c *C) {
	_, cc := setupColoredConsole()
	Verbose = false
	cc.indentation = 2
	cc.ScenarioStart("failing step")
	cc.Write([]byte("fail reason: blah"))

	cc.ScenarioEnd(true)

	c.Assert(cc.pluginMessagesBuffer.String(), Equals, "fail reason: blah")
}

func (s *MySuite) TestScenarioStartAndScenarioEnd_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true

	cc.ScenarioStart("First Scenario")
	c.Assert(dw.output, Equals, spaces(scenarioIndentation)+"## First Scenario\t\n")
	dw.output = ""

	input := "* Say hello to all"
	cc.StepStart(input)

	twoLevelIndentation := spaces(scenarioIndentation + stepIndentation)
	expectedStepStartOutput := twoLevelIndentation + input + newline
	c.Assert(dw.output, Equals, expectedStepStartOutput)
	dw.output = ""

	cc.StepEnd(false)

	c.Assert(dw.output, Equals, cursorUp+eraseLine+twoLevelIndentation+"* Say hello to all\t ...[PASS]\n")
	cc.ScenarioEnd(false)
	c.Assert(cc.headingBuffer.String(), Equals, "")
	c.Assert(cc.pluginMessagesBuffer.String(), Equals, "")
}

func (s *MySuite) TestStepStart_Verbose(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 2

	cc.StepStart("* say hello")

	c.Assert(dw.output, Equals, "      * say hello\n")
}

func (s *MySuite) TestFailingStepEndInVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 2
	cc.StepStart("* say hello")
	dw.output = ""

	cc.StepEnd(true)

	c.Assert(dw.output, Equals, cursorUp+eraseLine+"      * say hello\t ...[FAIL]\n")
}

func (s *MySuite) TestFailingStepEnd_NonVerbose(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = false
	cc.indentation = 2
	cc.StepStart("* say hello")
	dw.output = ""

	cc.StepEnd(true)

	c.Assert(dw.output, Equals, getFailureSymbol()+newline)
}

func (s *MySuite) TestPassingStepEndInNonVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = false
	cc.indentation = 2
	cc.StepStart("* say hello")
	dw.output = ""

	cc.StepEnd(false)

	c.Assert(dw.output, Equals, getSuccessSymbol())
}

func (s *MySuite) TestStepStartAndStepEnd_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 2

	input := "* Say hello to all"
	cc.StepStart(input)

	expectedStepStartOutput := spaces(cc.indentation) + "* Say hello to all\n"
	c.Assert(dw.output, Equals, expectedStepStartOutput)
	dw.output = ""

	cc.StepEnd(true)

	expectedStepEndOutput := cursorUp + eraseLine + spaces(6) + "* Say hello to all\t ...[FAIL]\n"
	c.Assert(dw.output, Equals, expectedStepEndOutput)
}

func (s *MySuite) TestConceptStartAndEnd_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 4

	cc.ConceptStart("* my concept")
	c.Assert(dw.output, Equals, "        * my concept\n")
	c.Assert(cc.indentation, Equals, 8)

	dw.output = ""
	cc.ConceptStart("* my concept1")
	c.Assert(dw.output, Equals, "            * my concept1\n")
	c.Assert(cc.indentation, Equals, 12)

	cc.ConceptEnd(true)
	c.Assert(cc.indentation, Equals, 8)

	cc.ConceptEnd(true)
	c.Assert(cc.indentation, Equals, 4)
}

func (s *MySuite) TestDataTable_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 2
	Verbose = true
	table := `|Product|Description                  |
|-------|-----------------------------|
|Gauge  |Test automation with ease    |`

	want := `
|Product|Description                  |
|-------|-----------------------------|
|Gauge  |Test automation with ease    |`

	cc.DataTable(table)

	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestError_ColoredConsole(c *C) {
	_, cc := setupColoredConsole()
	initialIndentation := 6
	cc.indentation = initialIndentation
	Verbose = true

	cc.Error("Failed %s", "network error")

	c.Assert(cc.errorMessagesBuffer.String(), Equals, fmt.Sprintf("%sFailed network error\n", spaces(initialIndentation+errorIndentation)))
}

func (s *MySuite) TestWrite_VerboseColoredConsole(c *C) {
	_, cc := setupColoredConsole()
	cc.indentation = 6
	Verbose = true
	input := "hello, gauge"

	_, err := cc.Write([]byte(input))

	c.Assert(err, Equals, nil)
	c.Assert(cc.pluginMessagesBuffer.String(), Equals, input)
}

func (s *MySuite) TestWrite_ColoredConsole(c *C) {
	_, cc := setupColoredConsole()
	cc.indentation = 6
	Verbose = false
	input := "hello, gauge"

	_, err := cc.Write([]byte(input))

	c.Assert(err, Equals, nil)
	c.Assert(cc.pluginMessagesBuffer.String(), Equals, input)
}
