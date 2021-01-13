/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package reporter

import (
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func setupColoredConsole() (*dummyWriter, *coloredConsole) {
	dw := newDummyWriter()
	cc := newColoredConsole(dw)
	return dw, cc
}

func (s *MySuite) TestScenarioStartInNonVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 2
	scnRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_PASSED})

	cc.ScenarioStart(&gauge.Scenario{Heading: &gauge.Heading{Value: "my first scenario"}}, &gauge_messages.ExecutionInfo{}, scnRes)

	c.Assert(dw.output, Equals, "    ## my first scenario\t")
}

func (s *MySuite) TestScenarioEndInNonVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 2
	scnRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_FAILED, Failed: true})
	cc.ScenarioStart(&gauge.Scenario{Heading: &gauge.Heading{Value: "failing step"}}, &gauge_messages.ExecutionInfo{}, scnRes)
	dw.output = ""

	_, err := cc.Write([]byte("fail reason: blah"))
	c.Assert(err, IsNil)
	cc.ScenarioEnd(nil, scnRes, &gauge_messages.ExecutionInfo{})

	c.Assert(dw.output, Equals, "fail reason: blah\n")
}

func (s *MySuite) TestFailingStepEnd_NonVerbose(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 2
	stepText := "* say hello"
	errMsg := "pre hook failure message"
	stacktrace := "my stacktrace"
	specName := "hello.spec"
	specInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: specName}}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{ErrorMessage: errMsg, StackTrace: stacktrace}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()
	cc.StepStart(stepText)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: "* say hello"}, stepRes, specInfo)

	c.Assert(dw.output, Equals, getFailureSymbol())
}

func (s *MySuite) TestPassingStepEndInNonVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 2
	cc.StepStart("* say hello")
	dw.output = ""

	specName := "hello.spec"
	specInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: specName}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{}})

	cc.StepEnd(gauge.Step{LineText: "* say hello"}, stepRes, specInfo)

	c.Assert(dw.output, Equals, getSuccessSymbol())
}
