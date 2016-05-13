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
	"time"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	. "gopkg.in/check.v1"
)

var (
	eraseLine = "\x1b[2K\r"
	cursorUp  = "\x1b[0A"
)

type DummyResult struct {
	PreHookFailure  **(gauge_messages.ProtoHookFailure)
	PostHookFailure **(gauge_messages.ProtoHookFailure)
	IsFailed        bool
}

func (r *DummyResult) GetPreHook() **(gauge_messages.ProtoHookFailure) {
	return r.PreHookFailure
}
func (r *DummyResult) GetPostHook() **(gauge_messages.ProtoHookFailure) {
	return r.PostHookFailure
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

	res := &DummyResult{IsFailed: false}
	cc.SpecEnd(res)
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
	res := &DummyResult{IsFailed: true}
	cc.ScenarioEnd(res)

	c.Assert(cc.pluginMessagesBuffer.String(), Equals, "fail reason: blah")
}

func (s *MySuite) TestScenarioStartAndScenarioEnd_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	sceHeading := "First Scenario"
	stepText := "* Say hello to all"
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{}})
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: &sceHeading})

	cc.ScenarioStart(sceHeading)
	c.Assert(dw.output, Equals, spaces(scenarioIndentation)+"## First Scenario\t\n")
	dw.output = ""

	cc.StepStart(stepText)

	twoLevelIndentation := spaces(scenarioIndentation + stepIndentation)
	expectedStepStartOutput := twoLevelIndentation + stepText + newline
	c.Assert(dw.output, Equals, expectedStepStartOutput)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes)
	c.Assert(dw.output, Equals, cursorUp+eraseLine+twoLevelIndentation+stepText+"\t ...[PASS]\n")

	cc.ScenarioEnd(sceRes)
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
	stepText := "* say hello"
	cc.StepStart(stepText)
	dw.output = ""
	errMsg := "pre hook failure message"
	stackTrace := "my stacktrace"
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{ErrorMessage: &errMsg, StackTrace: &stackTrace}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes)

	expectedErrMsg := `        ` + `
        Failed Step: * say hello
        Error Message: pre hook failure message
        Stacktrace:` + spaces(1) + `
        my stacktrace
`
	c.Assert(dw.output, Equals, cursorUp+eraseLine+"      "+stepText+"\t ...[FAIL]\n"+expectedErrMsg)
}

func (s *MySuite) TestFailingStepEnd_NonVerbose(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = false
	cc.indentation = 2
	stepText := "* say hello"
	errMsg := "pre hook failure message"
	stacktrace := "my stacktrace"
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{ErrorMessage: &errMsg, StackTrace: &stacktrace}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()
	cc.StepStart(stepText)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: "* say hello"}, stepRes)

	expectedErrMsg := spaces(8) + `
        Failed Step: ` + stepText + `
        Error Message: ` + errMsg + `
        Stacktrace:` + spaces(1) + `
        ` + stacktrace + `
`
	c.Assert(dw.output, Equals, getFailureSymbol()+newline+expectedErrMsg)
}

func (s *MySuite) TestPassingStepEndInNonVerbose_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = false
	cc.indentation = 2
	cc.StepStart("* say hello")
	dw.output = ""
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{}})

	cc.StepEnd(gauge.Step{LineText: "* say hello"}, stepRes)

	c.Assert(dw.output, Equals, getSuccessSymbol())
}

func (s *MySuite) TestStepStartAndStepEnd_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 2
	stepText := "* Say hello to all"
	errMsg := "pre hook failure message"
	stacktrace := "my stacktrace"
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{ErrorMessage: &errMsg, StackTrace: &stacktrace}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()

	cc.StepStart(stepText)

	expectedStepStartOutput := spaces(cc.indentation) + stepText + newline
	c.Assert(dw.output, Equals, expectedStepStartOutput)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes)

	expectedErrMsg := spaces(8) + `
        Failed Step: ` + stepText + `
        Error Message: ` + errMsg + `
        Stacktrace:` + spaces(1) + `
        ` + stacktrace + `
`
	expectedStepEndOutput := cursorUp + eraseLine + spaces(6) + stepText + "\t ...[FAIL]\n" + expectedErrMsg
	c.Assert(dw.output, Equals, expectedStepEndOutput)
}

func (s *MySuite) TestStepFailure_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
	cc.indentation = 2
	errMsg := "pre hook failure message"
	stacktrace := "my stacktrace"
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{ErrorMessage: &errMsg, StackTrace: &stacktrace}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()
	stepText := "* Say hello to all"
	cc.StepStart(stepText)

	expectedStepStartOutput := spaces(cc.indentation) + stepText + newline
	c.Assert(dw.output, Equals, expectedStepStartOutput)
	dw.output = ""

	cc.Errorf("Failed!")
	c.Assert(dw.output, Equals, spaces(cc.indentation+errorIndentation)+"Failed!\n")
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes)

	expectedErrMsg := spaces(8) + `
        Failed Step: ` + stepText + `
        Error Message: ` + errMsg + `
        Stacktrace:` + spaces(1) + `
        ` + stacktrace + `
`
	expectedStepEndOutput := cursorUp + eraseLine + cursorUp + eraseLine + spaces(6) + "* Say hello to all\t ...[FAIL]\n" + spaces(8) + "Failed!\n" + expectedErrMsg
	c.Assert(dw.output, Equals, expectedStepEndOutput)
}

func (s *MySuite) TestConceptStartAndEnd_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	Verbose = true
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
	dw, cc := setupColoredConsole()
	cc.indentation = 2
	Verbose = true
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
	dw, cc := setupColoredConsole()
	initialIndentation := 6
	cc.indentation = initialIndentation
	Verbose = true

	cc.Errorf("Failed %s", "network error")

	c.Assert(dw.output, Equals, fmt.Sprintf("%sFailed network error\n", spaces(initialIndentation+errorIndentation)))
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

func (s *MySuite) TestStepEndWithPreHookFailure_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 2
	Verbose = true
	errMsg := "pre hook failure message"
	stackTrace := "my stacktrace"
	stepText := "* my step"
	preHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: &errMsg, StackTrace: &stackTrace}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{PreHookFailure: preHookFailure}})
	cc.StepStart(stepText)
	dw.output = ""

	cc.StepEnd(gauge.Step{LineText: stepText}, stepRes)

	c.Assert(cc.indentation, Equals, 2)
	expectedErrMsg := spaces(8) + `Error Message: ` + errMsg + `
        Stacktrace:` + spaces(1) + `
        ` + stackTrace + `
`
	c.Assert(dw.output, Equals, cursorUp+eraseLine+"      "+stepText+"\t ...[PASS]\n"+expectedErrMsg)
}

func (s *MySuite) TestStepEndWithPostHookFailure_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 6
	errMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: &errMsg, StackTrace: &stackTrace}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{PostHookFailure: postHookFailure}})

	cc.StepEnd(gauge.Step{LineText: "* my step"}, stepRes)

	c.Assert(cc.indentation, Equals, 2)
	expectedErrMsg := spaces(8) + `Error Message: ` + errMsg + `
        Stacktrace:` + spaces(1) + `
        ` + stackTrace + `
`
	c.Assert(dw.output, Equals, "\t ...[PASS]\n"+expectedErrMsg)
}

func (s *MySuite) TestStepEndWithPreAndPostHookFailure_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 6
	preHookErrMsg := "pre hook failure message"
	postHookErrMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	preHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: &preHookErrMsg, StackTrace: &stackTrace}
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: &postHookErrMsg, StackTrace: &stackTrace}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{PostHookFailure: postHookFailure, PreHookFailure: preHookFailure}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})

	cc.StepEnd(gauge.Step{LineText: "* my step"}, stepRes)

	c.Assert(cc.indentation, Equals, 2)
	err1 := fmt.Sprintf("%sError Message: %s\n%sStacktrace: \n%s%s\n", spaces(8), preHookErrMsg, spaces(8), spaces(8), stackTrace)
	err2 := fmt.Sprintf("%sError Message: %s\n%sStacktrace: \n%s%s\n", spaces(8), postHookErrMsg, spaces(8), spaces(8), stackTrace)
	c.Assert(dw.output, Equals, "\t ...[PASS]\n"+err1+err2)
}

func (s *MySuite) TestSubscribeScenarioEndPreHookFailure_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = scenarioIndentation
	currentReporter = cc
	preHookErrMsg := "pre hook failure message"
	stackTrace := "my stacktrace"
	preHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: &preHookErrMsg, StackTrace: &stackTrace}
	res := &DummyResult{PreHookFailure: &preHookFailure}

	cc.ScenarioEnd(res)

	ind := spaces(scenarioIndentation + errorIndentation)
	want := ind + "Error Message: " + preHookErrMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline
	c.Assert(dw.output, Equals, want)
	c.Assert(cc.indentation, Equals, 0)
}

func (s *MySuite) TestSpecEndWithPostHookFailure_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 0
	errMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: &errMsg, StackTrace: &stackTrace}
	res := &DummyResult{PostHookFailure: &postHookFailure}

	cc.SpecEnd(res)

	c.Assert(cc.indentation, Equals, 0)
	ind := spaces(errorIndentation)
	want := newline + ind + "Error Message: " + errMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline
	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestSuiteEndWithPostHookFailure_ColoredConsole(c *C) {
	dw, cc := setupColoredConsole()
	cc.indentation = 0
	errMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	res := result.NewSuiteResult("", time.Now())
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: &errMsg, StackTrace: &stackTrace}
	res.PostSuite = postHookFailure

	cc.SuiteEnd(res)

	c.Assert(cc.indentation, Equals, 0)
	ind := spaces(errorIndentation)
	want := ind + "Error Message: " + errMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline
	c.Assert(dw.output, Equals, want)
}
