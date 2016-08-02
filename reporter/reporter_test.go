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

	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestSubscribeSpecEnd(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.SpecEnd, nil, &DummyResult{}, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(dw.output, Equals, "\n")
}

func (s *MySuite) TestSubscribeSpecStart(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	spec := &gauge.Specification{Heading: &gauge.Heading{Value: "My Spec Heading"}}

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.SpecStart, spec, nil, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(dw.output, Equals, "# My Spec Heading\n")
}

func (s *MySuite) TestSubscribeScenarioStart(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	sce := &gauge.Scenario{Heading: &gauge.Heading{Value: "My Scenario Heading"}}
	sceHeading := "My scenario heading"
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: &sceHeading})

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.ScenarioStart, sce, sceRes, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(dw.output, Equals, "  ## My Scenario Heading\n")
}

func (s *MySuite) TestSubscribeScenarioStartWithDataTable(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	dataTable := gauge.Table{}
	dataTable.AddHeaders([]string{"foo", "bar"})
	dataTable.AddRowValues([]string{"one", "two"})
	sce := &gauge.Scenario{Heading: &gauge.Heading{Value: "My Scenario Heading"}, DataTableRow: dataTable}
	sceHeading := "My scenario heading"
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: &sceHeading})

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.ScenarioStart, sce, sceRes, 0, gauge_messages.ExecutionInfo{}))
	table := `
     |foo|bar|
     |---|---|
     |one|two|`
	c.Assert(dw.output, Equals, table+newline+"  ## My Scenario Heading\n")
}

func (s *MySuite) TestSubscribeScenarioEnd(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = scenarioIndentation
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	sceHeading := "My scenario heading"
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: &sceHeading})

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.ScenarioEnd, nil, sceRes, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(dw.output, Equals, "")
	c.Assert(sc.indentation, Equals, 0)
}

func (s *MySuite) TestSubscribeStepStart(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	Verbose = true
	event.InitRegistry()
	stepText := "My first step"
	step := &gauge.Step{Value: stepText}

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.StepStart, step, nil, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(dw.output, Equals, spaces(stepIndentation)+"* "+stepText+newline)
}

func (s *MySuite) TestSubscribeStepEnd(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = stepIndentation
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	failed := false
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: &failed}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.StepEnd, gauge.Step{}, stepRes, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(dw.output, Equals, "")
	c.Assert(sc.indentation, Equals, 0)
}

func (s *MySuite) TestSubscribeFailedStepEnd(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 0
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	failed := true
	stepText := "* say hello"
	errMsg := "failure message"
	specName := "hello.spec"
	specInfo := gauge_messages.ExecutionInfo{CurrentSpec: &gauge_messages.SpecInfo{FileName: &specName}}
	stacktrace := `StepImplementation.implementation4(StepImplementation.java:77)
sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)`
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: &failed, ErrorMessage: &errMsg, StackTrace: &stacktrace}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})
	stepRes.SetStepFailure()

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.StepEnd, gauge.Step{LineText: stepText}, stepRes, 0, specInfo))
	want := spaces(errorIndentation) + newline +
		`  Failed Step: * say hello
  Specification: hello.spec:0
  Error Message: failure message
  Stacktrace:` + spaces(1) +
		`
  StepImplementation.implementation4(StepImplementation.java:77)
  sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
`
	c.Assert(dw.output, Equals, want)
}

func (s *MySuite) TestSubscribeConceptStart(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	Verbose = true
	cptText := "My last concept"
	concept := &gauge.Step{Value: cptText, IsConcept: true}

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.ConceptStart, concept, nil, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(dw.output, Equals, spaces(stepIndentation)+"* "+cptText+newline)
}

func (s *MySuite) TestSubscribeConceptEnd(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = stepIndentation
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	failed := true
	cptExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: &failed}}
	cptRes := result.NewConceptResult(&gauge_messages.ProtoConcept{ConceptExecutionResult: cptExeRes})

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.ConceptEnd, nil, cptRes, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(dw.output, Equals, "")
	c.Assert(sc.indentation, Equals, 0)
}

func (s *MySuite) TestSubscribeSuiteEnd(c *C) {
	dw, sc := setupSimpleConsole()
	sc.indentation = 0
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	suiteRes := &result.SuiteResult{UnhandledErrors: []error{fmt.Errorf("failure 1"), fmt.Errorf("failure 2")}}

	ListenExecutionEvents()
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, suiteRes, 0, gauge_messages.ExecutionInfo{}))

	c.Assert(dw.output, Equals, spaces(errorIndentation)+"failure 1\n"+spaces(errorIndentation)+"failure 2\n")
}
