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

package logger

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	. "gopkg.in/check.v1"
)

var (
	eraseLineUnix = "\x1b[2K\r"
	cursorUpUnix  = "\x1b[0A"

	eraseCharWindows  = "\x1b[2K\r"
	cursorLeftWindows = "\x1b[0A"
)

func (s *MySuite) TestStepStartAndStepEnd_ColoredLogger(c *C) {
	Initialize(true, "Debug")
	cl := newColoredConsoleWriter()
	b := &bytes.Buffer{}
	cl.writer.Out = b

	input := "* Say hello to all"
	cl.StepStart(input)

	expectedStepStartOutput := spaces(cl.indentation) + "* Say hello to all\n"
	c.Assert(b.String(), Equals, expectedStepStartOutput)
	b.Reset()

	cl.StepEnd(true)

	if runtime.GOOS == "windows" {
		expectedStepEndOutput := strings.Repeat(cursorLeftWindows+eraseCharWindows, len(expectedStepStartOutput)) + spaces(cl.indentation) + "* Say hello to all\t ...[FAIL]\n"
		c.Assert(b.String(), Equals, expectedStepEndOutput)
	} else {
		expectedStepEndOutput := cursorUpUnix + eraseLineUnix + spaces(stepIndentation) + "* Say hello to all\t ...[FAIL]\n"
		c.Assert(b.String(), Equals, expectedStepEndOutput)
	}
}

func (s *MySuite) TestScenarioStartAndScenarioEndInColoredDebugMode(c *C) {
	Initialize(true, "Debug")
	cl := newColoredConsoleWriter()
	b := &bytes.Buffer{}
	cl.writer.Out = b

	cl.ScenarioStart("First Scenario")
	input := "* Say hello to all"
	cl.StepStart(input)

	twoLevelIndentation := spaces(scenarioIndentation) + spaces(stepIndentation)
	expectedStepStartOutput := twoLevelIndentation + input + newline
	c.Assert(b.String(), Equals, expectedStepStartOutput)
	b.Reset()

	cl.StepEnd(false)

	if runtime.GOOS == "windows" {
		c.Assert(b.String(), Equals, strings.Repeat(cursorLeftWindows+eraseCharWindows, len(expectedStepStartOutput))+twoLevelIndentation+"* Say hello to all\t ...[PASS]\n")
	} else {
		c.Assert(b.String(), Equals, cursorUpUnix+eraseLineUnix+twoLevelIndentation+"* Say hello to all\t ...[PASS]\n")
	}
	cl.ScenarioEnd(false)
	c.Assert(cl.headingText.String(), Equals, "")
	c.Assert(cl.buffer.String(), Equals, "")

}

func (s *MySuite) TestStacktraceConsoleFormat(c *C) {
	Initialize(true, "Debug")
	b := &bytes.Buffer{}
	cl := newColoredConsoleWriter()
	cl.writer.Out = b
	stacktrace := "Stacktrace: [StepImplementation.fail(StepImplementation.java:21)\n" +
		"sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)\n" +
		"com.thoughtworks.gauge.execution.HookExecutionStage.execute(HookExecutionStage.java:42)\n" +
		"com.thoughtworks.gauge.execution.ExecutionPipeline.start(ExecutionPipeline.java:31)\n" +
		"com.thoughtworks.gauge.processor.ExecuteStepProcessor.process(ExecuteStepProcessor.java:37)\n" +
		"com.thoughtworks.gauge.connection.MessageDispatcher.dispatchMessages(MessageDispatcher.java:72)\n" +
		"com.thoughtworks.gauge.GaugeRuntime.main(GaugeRuntime.java:37)\n" +
		"]          "

	fmt.Fprint(cl, stacktrace)

	formattedStacktrace := spaces(sysoutIndentation) + "Stacktrace: [StepImplementation.fail(StepImplementation.java:21)\n" +
		spaces(sysoutIndentation) + "sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.execution.HookExecutionStage.execute(HookExecutionStage.java:42)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.execution.ExecutionPipeline.start(ExecutionPipeline.java:31)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.processor.ExecuteStepProcessor.process(ExecuteStepProcessor.java:37)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.connection.MessageDispatcher.dispatchMessages(MessageDispatcher.java:72)\n" +
		spaces(sysoutIndentation) + "com.thoughtworks.gauge.GaugeRuntime.main(GaugeRuntime.java:37)\n" +
		spaces(sysoutIndentation) + "]\n"
	c.Assert(b.String(), Equals, formattedStacktrace)
	c.Assert(cl.buffer.String(), Equals, formattedStacktrace)
}

func (s *MySuite) TestConceptStartAndEnd(c *C) {
	Initialize(true, "Debug")
	b := &bytes.Buffer{}
	cl := newColoredConsoleWriter()
	cl.writer.Out = b
	cl.indentation = noIndentation

	cl.ConceptStart("my concept")
	cl.indentation = stepIndentation

	cl.ConceptStart("my concept1")
	cl.indentation = stepIndentation + stepIndentation

	cl.ConceptEnd(true)
	cl.indentation = stepIndentation

	cl.ConceptEnd(true)
	cl.indentation = noIndentation

}
