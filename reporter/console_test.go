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
	"bytes"
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

func (s *MySuite) TestStepStartAndStepEnd(c *C) {
	Verbose = true
	cw := newConsole(true)
	b := &bytes.Buffer{}
	cw.writer.Out = b

	input := "* Say hello to all"
	cw.StepStart(input)

	expectedStepStartOutput := spaces(cw.indentation) + "* Say hello to all\n"
	c.Assert(b.String(), Equals, expectedStepStartOutput)
	b.Reset()

	cw.StepEnd(true)

	if util.IsWindows() {
		expectedStepEndOutput := strings.Repeat(cursorLeftWindows+eraseCharWindows, len(expectedStepStartOutput)) + spaces(stepIndentation) + "* Say hello to all\t ...[FAIL]\n"
		c.Assert(b.String(), Equals, expectedStepEndOutput)
	} else {
		expectedStepEndOutput := cursorUpUnix + eraseLineUnix + spaces(stepIndentation) + "* Say hello to all\t ...[FAIL]\n"
		c.Assert(b.String(), Equals, expectedStepEndOutput)
	}
}

func (s *MySuite) TestScenarioStartAndScenarioEndInColoredDebugMode(c *C) {
	Verbose = true
	cw := newConsole(true)
	b := &bytes.Buffer{}
	cw.writer.Out = b

	cw.ScenarioStart("First Scenario")
	c.Assert(b.String(), Equals, spaces(scenarioIndentation)+"## First Scenario\n")
	b.Reset()

	input := "* Say hello to all"
	cw.StepStart(input)

	twoLevelIndentation := spaces(scenarioIndentation) + spaces(stepIndentation)
	expectedStepStartOutput := twoLevelIndentation + input + newline
	c.Assert(b.String(), Equals, expectedStepStartOutput)
	b.Reset()

	cw.StepEnd(false)

	if util.IsWindows() {
		c.Assert(b.String(), Equals, strings.Repeat(cursorLeftWindows+eraseCharWindows, len(expectedStepStartOutput))+twoLevelIndentation+"* Say hello to all\t ...[PASS]\n")
	} else {
		c.Assert(b.String(), Equals, cursorUpUnix+eraseLineUnix+twoLevelIndentation+"* Say hello to all\t ...[PASS]\n")
	}
	cw.ScenarioEnd(false)
	c.Assert(cw.headingText.String(), Equals, "")
	c.Assert(cw.buffer.String(), Equals, "")

}

func (s *MySuite) TestStacktraceConsoleFormat(c *C) {
	Verbose = true
	b := &bytes.Buffer{}
	cw := newConsole(true)
	cw.writer.Out = b
	stacktrace := "Stacktrace: [StepImplementation.fail(StepImplementation.java:21)\n" +
		"sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)\n" +
		"com.thoughtworks.gauge.execution.HookExecutionStage.execute(HookExecutionStage.java:42)\n" +
		"com.thoughtworks.gauge.execution.ExecutionPipeline.start(ExecutionPipeline.java:31)\n" +
		"com.thoughtworks.gauge.processor.ExecuteStepProcessor.process(ExecuteStepProcessor.java:37)\n" +
		"com.thoughtworks.gauge.connection.MessageDispatcher.dispatchMessages(MessageDispatcher.java:72)\n" +
		"com.thoughtworks.gauge.GaugeRuntime.main(GaugeRuntime.java:37)\n" +
		"]          "

	fmt.Fprint(cw, stacktrace)

	formattedStacktrace := spaces(sysoutIndentation) + "Stacktrace: [StepImplementation.fail(StepImplementation.java:21)\n" +
		spaces(sysoutIndentation) + "sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.execution.HookExecutionStage.execute(HookExecutionStage.java:42)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.execution.ExecutionPipeline.start(ExecutionPipeline.java:31)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.processor.ExecuteStepProcessor.process(ExecuteStepProcessor.java:37)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.connection.MessageDispatcher.dispatchMessages(MessageDispatcher.java:72)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.GaugeRuntime.main(GaugeRuntime.java:37)\n" +
		spaces(sysoutIndentation) + "]\n"
	c.Assert(b.String(), Equals, formattedStacktrace)
	c.Assert(cw.buffer.String(), Equals, formattedStacktrace)
}

func (s *MySuite) TestConceptStartAndEnd(c *C) {
	Verbose = true
	b := &bytes.Buffer{}
	cw := newConsole(true)
	cw.writer.Out = b
	cw.indentation = noIndentation

	cw.ConceptStart("my concept")
	cw.indentation = stepIndentation

	cw.ConceptStart("my concept1")
	cw.indentation = stepIndentation + stepIndentation

	cw.ConceptEnd(true)
	cw.indentation = stepIndentation

	cw.ConceptEnd(true)
	cw.indentation = noIndentation

}
