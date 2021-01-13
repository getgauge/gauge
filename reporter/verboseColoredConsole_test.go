/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package reporter

import (
	"fmt"
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

var (
	eraseLine = "\x1b[2K\r"
	cursorUp  = "\x1b[0A"
)

type DummyResult struct {
	PreHookFailure  []*gauge_messages.ProtoHookFailure
	PostHookFailure []*gauge_messages.ProtoHookFailure
	IsFailed        bool
}

func (r *DummyResult) GetPreHook() []*gauge_messages.ProtoHookFailure {
	return r.PreHookFailure
}
func (r *DummyResult) GetPostHook() []*gauge_messages.ProtoHookFailure {
	return r.PostHookFailure
}
func (r *DummyResult) AddPreHook(f ...*gauge_messages.ProtoHookFailure) {
}
func (r *DummyResult) AddPostHook(f ...*gauge_messages.ProtoHookFailure) {
}
func (r *DummyResult) SetFailure() {
	r.IsFailed = true
}
func (r *DummyResult) GetFailed() bool {
	return r.IsFailed
}
func (r *DummyResult) Item() interface{} {
	return nil
}
func (r *DummyResult) ExecTime() int64 {
	return 0
}

func setupVerboseColoredConsole() (*dummyWriter, *verboseColoredConsole) {
	dw := newDummyWriter()
	cc := newVerboseColoredConsole(dw)
	return dw, cc
}

func (s *MySuite) TestSpecStart_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()

	cc.SpecStart(&gauge.Specification{Heading: &gauge.Heading{Value: "Spec heading"}}, &result.SpecResult{Skipped: false})

	c.Assert(dw.output, Equals, "# Spec heading\n")
}

func (s *MySuite) TestSpecEnd_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()

	res := &result.SpecResult{Skipped: false, ProtoSpec: &gauge_messages.ProtoSpec{}, IsFailed: false}
	cc.SpecEnd(&gauge.Specification{}, res)
	c.Assert(dw.output, Equals, "\n")
}

func (s *MySuite) TestScenarioStartInVerbose_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = 2
	scnRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_PASSED})

	cc.ScenarioStart(&gauge.Scenario{Heading: &gauge.Heading{Value: "my first scenario"}}, &gauge_messages.ExecutionInfo{}, scnRes)

	c.Assert(dw.output, Equals, "    ## my first scenario\t\n")
}

func (s *MySuite) TestScenarioStartAndScenarioEnd_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	sceHeading := "First Scenario"
	stepText := "* Say hello to all"
	specInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: "hello.spec"}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{}})
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: sceHeading})
	scnRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ExecutionStatus: gauge_messages.ExecutionStatus_PASSED})

	cc.ScenarioStart(&gauge.Scenario{Heading: &gauge.Heading{Value: sceHeading}}, &gauge_messages.ExecutionInfo{}, scnRes)
	c.Assert(dw.output, Equals, spaces(scenarioIndentation)+"## First Scenario\t\n")
	cc.StepStart(stepText)

	twoLevelIndentation := spaces(scenarioIndentation + stepIndentation)
	expectedStepStartOutput := twoLevelIndentation + stepText
	c.Assert(cc.headingBuffer.String(), Equals, expectedStepStartOutput)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes, specInfo)
	c.Assert(dw.output, Equals, twoLevelIndentation+stepText+"\t ...[PASS]\n")

	cc.ScenarioEnd(nil, sceRes, &gauge_messages.ExecutionInfo{})
	c.Assert(cc.headingBuffer.String(), Equals, "")
	c.Assert(cc.pluginMessagesBuffer.String(), Equals, "")
}

func (s *MySuite) TestStepStart_Verbose(c *C) {
	_, cc := setupVerboseColoredConsole()
	cc.indentation = 2

	cc.StepStart("* say hello")
	c.Assert(cc.headingBuffer.String(), Equals, "      * say hello")
}

func (s *MySuite) TestFailingStepEndInVerbose_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = 2
	stepText := "* say hello"
	cc.StepStart(stepText)
	dw.output = ""
	errMsg := "pre hook failure message"
	stackTrace := "my stacktrace"
	specInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: "hello.spec"}}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{ErrorMessage: errMsg, StackTrace: stackTrace}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes, specInfo)

	expectedErrMsg := `        ` + `
        Failed Step: * say hello
        Specification: hello.spec:0
        Error Message: pre hook failure message
        Stacktrace:` + spaces(1) + `
        my stacktrace
`
	c.Assert(dw.output, Equals, "      "+stepText+"\t ...[FAIL]\n"+expectedErrMsg)
}

func (s *MySuite) TestStepStartAndStepEnd_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = 2
	stepText := "* Say hello to all"
	errMsg := "pre hook failure message"
	stacktrace := "my stacktrace"
	specName := "hello.spec"
	specInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: specName}}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{ErrorMessage: errMsg, StackTrace: stacktrace}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()

	cc.StepStart(stepText)

	expectedStepStartOutput := spaces(cc.indentation) + stepText
	c.Assert(cc.headingBuffer.String(), Equals, expectedStepStartOutput)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes, specInfo)

	expectedErrMsg := spaces(8) + `
        Failed Step: ` + stepText + `
        Specification: ` + specName + `:0
        Error Message: ` + errMsg + `
        Stacktrace:` + spaces(1) + `
        ` + stacktrace + `
`
	expectedStepEndOutput := spaces(6) + stepText + "\t ...[FAIL]\n" + expectedErrMsg
	c.Assert(dw.output, Equals, expectedStepEndOutput)
}

func (s *MySuite) TestStepFailure_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = 2
	errMsg := "pre hook failure message"
	stacktrace := "my stacktrace"
	specName := "hello.spec"
	specInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: specName}}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{ErrorMessage: errMsg, StackTrace: stacktrace}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()
	stepText := "* Say hello to all"
	cc.StepStart(stepText)

	expectedStepStartOutput := spaces(cc.indentation) + stepText
	c.Assert(cc.headingBuffer.String(), Equals, expectedStepStartOutput)

	cc.Errorf("Failed!")
	c.Assert(dw.output, Equals, spaces(cc.indentation+errorIndentation)+"Failed!\n")
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes, specInfo)

	expectedErrMsg := spaces(8) + `
        Failed Step: ` + stepText + `
        Specification: ` + specName + `:0
        Error Message: ` + errMsg + `
        Stacktrace:` + spaces(1) + `
        ` + stacktrace + `
`
	expectedStepEndOutput := cursorUp + eraseLine + spaces(6) + "* Say hello to all\t ...[FAIL]\n" + spaces(8) + "Failed!\n" + expectedErrMsg
	c.Assert(dw.output, Equals, expectedStepEndOutput)
}

func (s *MySuite) TestConceptStartAndEnd_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = 4
	cpt1 := "* my concept"
	cpt2 := "* my concept1"
	cptRes1 := &DummyResult{IsFailed: true}
	cptRes2 := &DummyResult{IsFailed: true}

	cc.ConceptStart(cpt1)
	c.Assert(dw.output, Equals, spaces(8)+cpt1+newline)
	c.Assert(cc.indentation, Equals, 8)

	dw.output = ""
	cc.ConceptStart(cpt2)
	c.Assert(dw.output, Equals, spaces(12)+cpt2+newline)
	c.Assert(cc.indentation, Equals, 12)

	cc.ConceptEnd(cptRes1)
	c.Assert(cc.indentation, Equals, 8)

	cc.ConceptEnd(cptRes2)
	c.Assert(cc.indentation, Equals, 4)
}

func (s *MySuite) TestDataTable_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = 2
	table := `|Product|Description                  |
|-------|-----------------------------|
|Gauge  |Test automation with ease    |`

	want := `|Product|Description                  |
|-------|-----------------------------|
|Gauge  |Test automation with ease    |`

	cc.DataTable(table)

	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestError_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	initialIndentation := 6
	cc.indentation = initialIndentation

	cc.Errorf("Failed %s", "network error")

	c.Assert(dw.output, Equals, fmt.Sprintf("%sFailed network error\n", spaces(initialIndentation+errorIndentation)))
}

func (s *MySuite) TestWrite_VerboseColoredConsole(c *C) {
	_, cc := setupVerboseColoredConsole()
	cc.indentation = 6
	input := "hello, gauge"

	_, err := cc.Write([]byte(input))

	c.Assert(err, Equals, nil)
	c.Assert(cc.pluginMessagesBuffer.String(), Equals, input)
}

func (s *MySuite) TestStepEndWithPreHookFailure_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = scenarioIndentation
	errMsg := "pre hook failure message"
	stackTrace := "my stacktrace"
	stepText := "* my step"
	specName := "hello.spec"
	specInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: specName}}
	preHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: errMsg, StackTrace: stackTrace}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{PreHookFailure: preHookFailure}})
	cc.StepStart(stepText)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes, specInfo)

	c.Assert(cc.indentation, Equals, scenarioIndentation)
	expectedErrMsg := spaces(8) + `Error Message: ` + errMsg + `
        Stacktrace:` + spaces(1) + `
        ` + stackTrace + `
`
	c.Assert(dw.output, Equals, spaces(scenarioIndentation+stepIndentation)+stepText+newline+expectedErrMsg)
}

func (s *MySuite) TestStepEndWithPostHookFailure_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = scenarioIndentation
	errMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	specName := "hello.spec"
	stepText := "* my step"
	specInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: specName}}
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: errMsg, StackTrace: stackTrace}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{PostHookFailure: postHookFailure}})
	cc.StepStart(stepText)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes, specInfo)

	c.Assert(cc.indentation, Equals, 2)
	expectedErrMsg := spaces(8) + `Error Message: ` + errMsg + `
        Stacktrace:` + spaces(1) + `
        ` + stackTrace + `
`
	c.Assert(dw.output, Equals, spaces(scenarioIndentation+stepIndentation)+stepText+newline+expectedErrMsg)
}

func (s *MySuite) TestStepEndWithPreAndPostHookFailure_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = scenarioIndentation
	preHookErrMsg := "pre hook failure message"
	postHookErrMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	specName := "hello.spec"
	stepText := "* my step"
	specInfo := &gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: specName}}
	preHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: preHookErrMsg, StackTrace: stackTrace}
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: postHookErrMsg, StackTrace: stackTrace}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{PostHookFailure: postHookFailure, PreHookFailure: preHookFailure}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	cc.StepStart(stepText)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes, specInfo)

	c.Assert(cc.indentation, Equals, scenarioIndentation)
	err1 := fmt.Sprintf("%sError Message: %s\n%sStacktrace: \n%s%s\n", spaces(8), preHookErrMsg, spaces(8), spaces(8), stackTrace)
	err2 := fmt.Sprintf("%sError Message: %s\n%sStacktrace: \n%s%s\n", spaces(8), postHookErrMsg, spaces(8), spaces(8), stackTrace)
	c.Assert(dw.output, Equals, spaces(scenarioIndentation+stepIndentation)+stepText+newline+err1+err2)
}

func (s *MySuite) TestSubscribeScenarioEndPreHookFailure_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = scenarioIndentation
	currentReporter = cc
	preHookErrMsg := "pre hook failure message"
	stackTrace := "my stacktrace"
	preHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: preHookErrMsg, StackTrace: stackTrace}
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{
		ExecutionStatus: gauge_messages.ExecutionStatus_PASSED,
		PreHookFailure:  preHookFailure,
	})

	cc.ScenarioEnd(nil, sceRes, &gauge_messages.ExecutionInfo{})

	ind := spaces(scenarioIndentation + errorIndentation)
	want := ind + "Error Message: " + preHookErrMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline
	c.Assert(dw.output, Equals, want)
	c.Assert(cc.indentation, Equals, 0)
}

func (s *MySuite) TestSpecEndWithPostHookFailure_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = 0
	errMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: errMsg, StackTrace: stackTrace}
	res := &result.SpecResult{Skipped: false, ProtoSpec: &gauge_messages.ProtoSpec{PostHookFailures: []*gauge_messages.ProtoHookFailure{postHookFailure}}}

	cc.SpecEnd(&gauge.Specification{}, res)

	c.Assert(cc.indentation, Equals, 0)
	ind := spaces(errorIndentation)
	want := ind + "Error Message: " + errMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline + newline
	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestSuiteEndWithPostHookFailure_ColoredConsole(c *C) {
	dw, cc := setupVerboseColoredConsole()
	cc.indentation = 0
	errMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	res := result.NewSuiteResult("", time.Now())
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: errMsg, StackTrace: stackTrace}
	res.PostSuite = postHookFailure

	cc.SuiteEnd(res)

	c.Assert(cc.indentation, Equals, 0)
	ind := spaces(errorIndentation)
	want := ind + "Error Message: " + errMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline
	c.Assert(dw.output, Equals, want)
}
