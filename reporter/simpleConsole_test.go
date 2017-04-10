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

func setupSimpleConsole() (*dummyWriter, *simpleConsole) {
	dw := newDummyWriter()
	sc := newSimpleConsole(dw)
	return dw, sc
}

func (s *MySuite) TestSpecStart_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.SpecStart("Specification heading")
	c.Assert(dw.output, Equals, "# Specification heading\n")
}

func (s *MySuite) TestSpecEnd_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()

	sc.SpecEnd(&DummyResult{})
	c.Assert(dw.output, Equals, "\n")
}

func (s *MySuite) TestScenarioStart_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.ScenarioStart("First Scenario")
	c.Assert(dw.output, Equals, "  ## First Scenario\n")
}

func (s *MySuite) TestScenarioEnd_SimpleConsole(c *C) {
	_, sc := setupSimpleConsole()
	sc.indentation = 2
	res := &DummyResult{IsFailed: true}

	sc.ScenarioEnd(res)

	c.Assert(sc.indentation, Equals, 0)
}

func (s *MySuite) TestStepStartInVerboseMode_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 2
	Verbose = true

	sc.StepStart("* Say hello to gauge")

	c.Assert(dw.output, Equals, "      * Say hello to gauge\n")
}

func (s *MySuite) TestStepStartInNonVerboseMode_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 2
	Verbose = false

	sc.StepStart("* Say hello to gauge")

	c.Assert(dw.output, Equals, "")
}

func (s *MySuite) TestStepEnd_SimpleConsole(c *C) {
	_, sc := setupSimpleConsole()
	sc.indentation = 6
	specInfo := gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{Name: "hello.spec"}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{}})
	sc.StepEnd(gauge.Step{LineText: ""}, stepRes, specInfo)

	c.Assert(sc.indentation, Equals, 2)
}

func (s *MySuite) TestSingleConceptStartInVerboseMode_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 2
	Verbose = true

	sc.ConceptStart("* my first concept")

	c.Assert(dw.output, Equals, fmt.Sprintf("%s* my first concept\n", spaces(6)))
}

func (s *MySuite) TestNestedConceptStartInVerboseMode_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 2
	Verbose = true

	sc.ConceptStart("* my first concept")
	dw.output = ""
	sc.ConceptStart("* my second concept")

	c.Assert(dw.output, Equals, fmt.Sprintf("%s* my second concept\n", spaces(10)))
}

func (s *MySuite) TestNestedConceptStartInVerboseMode_case2(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 2
	Verbose = true

	sc.ConceptStart("* my first concept")
	dw.output = ""
	sc.StepStart("* do foo bar")

	c.Assert(dw.output, Equals, fmt.Sprintf("%s* do foo bar\n", spaces(10)))
}

func (s *MySuite) TestNestedConceptStartInVerboseMode_case3(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 2
	Verbose = true

	sc.ConceptStart("* my first concept")
	sc.ConceptStart("* my second concept")
	dw.output = ""
	sc.StepStart("* do foo bar")

	c.Assert(dw.output, Equals, fmt.Sprintf("%s* do foo bar\n", spaces(14)))
}

func (s *MySuite) TestConceptEnd_SimpleConsole(c *C) {
	_, sc := setupSimpleConsole()
	sc.indentation = 6
	Verbose = true
	res := &DummyResult{IsFailed: false}

	sc.ConceptEnd(res)

	c.Assert(sc.indentation, Equals, 2)
}

func (s *MySuite) TestDataTable_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 2
	Verbose = true
	table := `|Product|Description                  |
|-------|-----------------------------|
|Gauge  |Test automation with ease    |`

	want := `|Product|Description                  |
|-------|-----------------------------|
|Gauge  |Test automation with ease    |`

	sc.DataTable(table)

	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestError_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 6
	Verbose = true

	sc.Errorf("Failed %s", "network error")

	c.Assert(dw.output, Equals, fmt.Sprintf("%sFailed network error\n", spaces(sc.indentation+errorIndentation)))
}

func (s *MySuite) TestWrite_VerboseSimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 6
	Verbose = true
	input := "hello, gauge"

	_, err := sc.Write([]byte(input))

	c.Assert(err, Equals, nil)
	c.Assert(dw.output, Equals, input)
}

func (s *MySuite) TestWrite_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 6
	Verbose = false
	input := "hello, gauge"

	_, err := sc.Write([]byte(input))

	c.Assert(err, Equals, nil)
	c.Assert(dw.output, Equals, input)
}

func (s *MySuite) TestSpecReporting_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	Verbose = true

	sc.SpecStart("Specification heading")
	sc.ScenarioStart("My First scenario")
	sc.StepStart("* do foo bar")
	sc.Write([]byte("doing foo bar"))
	res := &DummyResult{IsFailed: false}
	specInfo := gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{Name: "hello.spec"}}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: false}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})

	sc.StepEnd(gauge.Step{LineText: "* do foo bar"}, stepRes, specInfo)
	sc.ScenarioEnd(res)
	sc.SpecEnd(res)

	want := `# Specification heading
  ## My First scenario
      * do foo bar
doing foo bar
`

	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestStepEndWithPreHookFailure_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 6
	errMsg := "pre hook failure message"
	stackTrace := "my stacktrace"
	specInfo := gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{Name: "hello.spec"}}
	preHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: errMsg, StackTrace: stackTrace}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{PreHookFailure: preHookFailure}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})

	sc.StepEnd(gauge.Step{LineText: "* my step"}, stepRes, specInfo)

	c.Assert(sc.indentation, Equals, 2)
	c.Assert(dw.output, Equals, fmt.Sprintf("%sError Message: %s\n%sStacktrace: \n%s%s\n", spaces(8), errMsg, spaces(8), spaces(8), stackTrace))
}

func (s *MySuite) TestStepEndWithPostHookFailure_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 6
	errMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	specInfo := gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{Name: "hello.spec"}}
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: errMsg, StackTrace: stackTrace}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{PostHookFailure: postHookFailure}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})

	sc.StepEnd(gauge.Step{LineText: "* my step"}, stepRes, specInfo)

	c.Assert(sc.indentation, Equals, 2)
	c.Assert(dw.output, Equals, fmt.Sprintf("%sError Message: %s\n%sStacktrace: \n%s%s\n", spaces(8), errMsg, spaces(8), spaces(8), stackTrace))
}

func (s *MySuite) TestStepEndWithPreAndPostHookFailure_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 6
	preHookErrMsg := "pre hook failure message"
	postHookErrMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	specInfo := gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{Name: "hello.spec"}}
	preHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: preHookErrMsg, StackTrace: stackTrace}
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: postHookErrMsg, StackTrace: stackTrace}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{PostHookFailure: postHookFailure, PreHookFailure: preHookFailure}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})

	sc.StepEnd(gauge.Step{LineText: "* my step"}, stepRes, specInfo)

	c.Assert(sc.indentation, Equals, 2)
	err1 := fmt.Sprintf("%sError Message: %s\n%sStacktrace: \n%s%s\n", spaces(8), preHookErrMsg, spaces(8), spaces(8), stackTrace)
	err2 := fmt.Sprintf("%sError Message: %s\n%sStacktrace: \n%s%s\n", spaces(8), postHookErrMsg, spaces(8), spaces(8), stackTrace)
	c.Assert(dw.output, Equals, err1+err2)
}

func (s *MySuite) TestSubscribeScenarioEndPreHookFailure(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = scenarioIndentation
	currentReporter = sc
	preHookErrMsg := "pre hook failure message"
	stackTrace := "my stacktrace"
	preHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: preHookErrMsg, StackTrace: stackTrace}
	res := &DummyResult{PreHookFailure: []*gauge_messages.ProtoHookFailure{preHookFailure}}

	sc.ScenarioEnd(res)

	ind := spaces(scenarioIndentation + errorIndentation)
	want := ind + "Error Message: " + preHookErrMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline
	c.Assert(dw.output, Equals, want)
	c.Assert(sc.indentation, Equals, 0)
}

func (s *MySuite) TestSpecEndWithPostHookFailure_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 0
	errMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: errMsg, StackTrace: stackTrace}
	res := &DummyResult{PostHookFailure: []*gauge_messages.ProtoHookFailure{postHookFailure}}

	sc.SpecEnd(res)

	c.Assert(sc.indentation, Equals, 0)
	ind := spaces(errorIndentation)
	want := ind + "Error Message: " + errMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline + newline
	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestSuiteEndWithPostHookFailure_SimpleConsole(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 0
	errMsg := "post hook failure message"
	stackTrace := "my stacktrace"
	res := result.NewSuiteResult("", time.Now())
	postHookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: errMsg, StackTrace: stackTrace}
	res.PostSuite = postHookFailure

	sc.SuiteEnd(res)

	c.Assert(sc.indentation, Equals, 0)
	ind := spaces(errorIndentation)
	want := ind + "Error Message: " + errMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline
	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestExcludeLineNoForFailedStepInConcept(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 4
	errMsg := "failure message"
	stackTrace := "my stacktrace"
	failed := true
	specName := "hello.spec"
	stepText := "* my Step"
	parentStep := gauge.Step{LineText: "* parent step"}
	exeInfo := gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: specName}}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: failed, StackTrace: stackTrace, ErrorMessage: errMsg}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()

	sc.StepEnd(gauge.Step{LineText: stepText, Parent: &parentStep}, stepRes, exeInfo)

	c.Assert(sc.indentation, Equals, 0)
	ind := spaces(errorIndentation + 4)
	want := ind + newline + ind + "Failed Step: " + stepText + newline + ind + "Specification: " + specName + newline + ind + "Error Message: " + errMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline
	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestIncludeLineNoForFailedStep(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 4
	errMsg := "failure message"
	stackTrace := "my stacktrace"
	failed := true
	specName := "hello.spec"
	stepText := "* my Step"
	exeInfo := gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: specName}}
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: failed, StackTrace: stackTrace, ErrorMessage: errMsg}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()

	sc.StepEnd(gauge.Step{LineText: stepText, LineNo: 3}, stepRes, exeInfo)

	c.Assert(sc.indentation, Equals, 0)
	ind := spaces(errorIndentation + 4)
	want := ind + newline + ind + "Failed Step: " + stepText + newline + ind + "Specification: " + specName + ":3" + newline + ind + "Error Message: " + errMsg + newline + ind + "Stacktrace: \n" + ind + stackTrace + newline
	c.Assert(dw.output, Equals, want)
}
