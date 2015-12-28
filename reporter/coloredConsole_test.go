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
	"strings"

	"github.com/getgauge/gauge/util"

	. "gopkg.in/check.v1"
)

var (
	eraseLineUnix = "\x1b[2K\r"
	cursorUpUnix  = "\x1b[0A"

	eraseCharWindows  = "\x1b[2K\r"
	cursorLeftWindows = "\x1b[0A"
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

	c.Assert(dw.output, Equals, "    ## my first scenario    ")
}

func (s *MySuite) TestScenarioStartInNonVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 2

	cc.ScenarioStart("my first scenario")

	c.Assert(dw.output, Equals, "    ## my first scenario\n")
}

func (s *MySuite) TestScenarioEndInNonVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = false
	cc.indentation = 2
	cc.ScenarioStart("failing step")
	cc.Write([]byte("fail reason: blah"))
	dw.output = ""

	cc.ScenarioEnd(true)

	c.Assert(dw.output, Equals, "\n      fail reason: blah\n")
}

func (s *MySuite) TestScenarioStartAndScenarioEnd_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true

	cc.ScenarioStart("First Scenario")
	c.Assert(dw.output, Equals, spaces(scenarioIndentation)+"## First Scenario\n")
	dw.output = ""

	input := "* Say hello to all"
	cc.StepStart(input)

	twoLevelIndentation := spaces(scenarioIndentation + stepIndentation)
	expectedStepStartOutput := twoLevelIndentation + input + newline
	c.Assert(dw.output, Equals, expectedStepStartOutput)
	dw.output = ""

	cc.StepEnd(false)

	if util.IsWindows() {
		c.Assert(dw.output, Equals, strings.Repeat(cursorLeftWindows+eraseCharWindows, len(expectedStepStartOutput))+twoLevelIndentation+"* Say hello to all\t ...[PASS]\n")
	} else {
		c.Assert(dw.output, Equals, cursorUpUnix+eraseLineUnix+twoLevelIndentation+"* Say hello to all\t ...[PASS]\n")
	}
	cc.ScenarioEnd(false)
	c.Assert(cc.headingBuffer.String(), Equals, "")
	c.Assert(cc.pluginMessagesBuffer.String(), Equals, "")
}

func (s *MySuite) TestStepStart_Verbose(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 2

	cc.StepStart("* say hello ")

	c.Assert(dw.output, Equals, "      * say hello\n")
}

func (s *MySuite) TestFailingStepEndInVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 2
	cc.StepStart("* say hello")
	dw.output = ""

	cc.StepEnd(true)

	if util.IsWindows() {
		c.Assert(dw.output, Equals, strings.Repeat(cursorLeftWindows+eraseCharWindows, len(indent("* say hello", cc.indentation+stepIndentation)))+"      * say hello\t ...[FAIL]\n")
	} else {
		c.Assert(dw.output, Equals, cursorUpUnix+eraseLineUnix+"      * say hello\t ...[FAIL]\n")
	}
}

func (s *MySuite) TestFailingStepEnd_NonVerbose(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = false
	cc.indentation = 2
	cc.StepStart("* say hello")
	dw.output = ""

	cc.StepEnd(true)

	c.Assert(dw.output, Equals, getFailureSymbol())
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

	if util.IsWindows() {
		expectedStepEndOutput := strings.Repeat(cursorLeftWindows+eraseCharWindows, len(expectedStepStartOutput)) + spaces(6) + "* Say hello to all\t ...[FAIL]\n"
		c.Assert(dw.output, Equals, expectedStepEndOutput)
	} else {
		expectedStepEndOutput := cursorUpUnix + eraseLineUnix + spaces(6) + "* Say hello to all\t ...[FAIL]\n"
		c.Assert(dw.output, Equals, expectedStepEndOutput)
	}
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

func (s *MySuite) TestStacktraceConsoleFormat(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true

	stacktrace := "Stacktrace: [StepImplementation.fail(StepImplementation.java:21)\n" +
		"sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)\n" +
		"com.thoughtworks.gauge.execution.HookExecutionStage.execute(HookExecutionStage.java:42)\n" +
		"com.thoughtworks.gauge.execution.ExecutionPipeline.start(ExecutionPipeline.java:31)\n" +
		"com.thoughtworks.gauge.processor.ExecuteStepProcessor.process(ExecuteStepProcessor.java:37)\n" +
		"]          "

	fmt.Fprint(cc, stacktrace)

	formattedStacktrace := spaces(sysoutIndentation) + "Stacktrace: [StepImplementation.fail(StepImplementation.java:21)\n" +
		spaces(sysoutIndentation) + "sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.execution.HookExecutionStage.execute(HookExecutionStage.java:42)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.execution.ExecutionPipeline.start(ExecutionPipeline.java:31)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.processor.ExecuteStepProcessor.process(ExecuteStepProcessor.java:37)\n" +
		spaces(sysoutIndentation) + "]\n"
	c.Assert(dw.output, Equals, formattedStacktrace)
	c.Assert(cc.pluginMessagesBuffer.String(), Equals, formattedStacktrace)
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
	dw, cc := setupColoredConsole()
	cc.indentation = 6
	Verbose = true

	cc.Error("Failed %s", "network error")

	c.Assert(dw.output, Equals, fmt.Sprintf("%sFailed network error\n", spaces(cc.indentation+sysoutIndentation)))
}
