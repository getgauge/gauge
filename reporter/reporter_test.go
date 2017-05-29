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
	"sync"

	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	. "gopkg.in/check.v1"
)

var dataTableEvent = event.Topic(100)

func (s *MySuite) TestSubscribeSpecEnd(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	SimpleConsoleOutput = true
	event.InitRegistry()

	ListenExecutionEvents(&sync.WaitGroup{})

	event.Notify(event.NewExecutionEvent(event.SpecEnd, nil, &DummyResult{}, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(<-e, Equals, event.SpecEnd)
}

func (s *MySuite) TestSubscribeSpecStart(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	event.InitRegistry()
	spec := &gauge.Specification{Heading: &gauge.Heading{Value: "My Spec Heading"}}

	ListenExecutionEvents(&sync.WaitGroup{})

	event.Notify(event.NewExecutionEvent(event.SpecStart, spec, nil, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(<-e, Equals, event.SpecStart)
}

func (s *MySuite) TestSubscribeScenarioStart(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	event.InitRegistry()
	sceHeading := "My scenario heading"
	sce := &gauge.Scenario{Heading: &gauge.Heading{Value: sceHeading}}
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: sceHeading})

	ListenExecutionEvents(&sync.WaitGroup{})

	event.Notify(event.NewExecutionEvent(event.ScenarioStart, sce, sceRes, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(<-e, Equals, event.ScenarioStart)
}

func (s *MySuite) TestSubscribeScenarioStartWithDataTable(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	event.InitRegistry()
	dataTable := gauge.Table{}
	dataTable.AddHeaders([]string{"foo", "bar"})
	dataTable.AddRowValues([]string{"one", "two"})
	sceHeading := "My scenario heading"
	step := &gauge.Step{
		Args: []*gauge.StepArg{&gauge.StepArg{Name: "foo",
			Value:   "foo",
			ArgType: gauge.Dynamic}},
	}
	sce := &gauge.Scenario{Heading: &gauge.Heading{Value: sceHeading}, DataTableRow: dataTable, Steps: []*gauge.Step{step}}
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: sceHeading})

	ListenExecutionEvents(&sync.WaitGroup{})

	event.Notify(event.NewExecutionEvent(event.ScenarioStart, sce, sceRes, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(<-e, Equals, dataTableEvent)
	c.Assert(<-e, Equals, event.ScenarioStart)
}

func (s *MySuite) TestSubscribeScenarioEnd(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	event.InitRegistry()
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: "My scenario heading"})

	ListenExecutionEvents(&sync.WaitGroup{})

	event.Notify(event.NewExecutionEvent(event.ScenarioEnd, nil, sceRes, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(<-e, Equals, event.ScenarioEnd)
}

func (s *MySuite) TestSubscribeStepStart(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	event.InitRegistry()
	stepText := "My first step"
	step := &gauge.Step{Value: stepText}

	ListenExecutionEvents(&sync.WaitGroup{})

	event.Notify(event.NewExecutionEvent(event.StepStart, step, nil, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(<-e, Equals, event.StepStart)
}

func (s *MySuite) TestSubscribeStepEnd(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	event.InitRegistry()
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: false}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})

	ListenExecutionEvents(&sync.WaitGroup{})

	event.Notify(event.NewExecutionEvent(event.StepEnd, gauge.Step{}, stepRes, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(<-e, Equals, event.StepEnd)
}

func (s *MySuite) TestSubscribeConceptStart(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	SimpleConsoleOutput = true
	event.InitRegistry()
	Verbose = true
	cptText := "My last concept"
	concept := &gauge.Step{Value: cptText, IsConcept: true}

	ListenExecutionEvents(&sync.WaitGroup{})

	event.Notify(event.NewExecutionEvent(event.ConceptStart, concept, nil, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(<-e, Equals, event.ConceptStart)
}

func (s *MySuite) TestSubscribeConceptEnd(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	event.InitRegistry()
	cptExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: true}}
	cptRes := result.NewConceptResult(&gauge_messages.ProtoConcept{ConceptExecutionResult: cptExeRes})

	ListenExecutionEvents(&sync.WaitGroup{})

	event.Notify(event.NewExecutionEvent(event.ConceptEnd, nil, cptRes, 0, gauge_messages.ExecutionInfo{}))
	c.Assert(<-e, Equals, event.ConceptEnd)
}

func (s *MySuite) TestSubscribeSuiteEnd(c *C) {
	e := make(chan event.Topic)
	currentReporter = &dummyConsole{event: e}
	event.InitRegistry()
	suiteRes := &result.SuiteResult{UnhandledErrors: []error{}}

	ListenExecutionEvents(&sync.WaitGroup{})
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, suiteRes, 0, gauge_messages.ExecutionInfo{}))

	c.Assert(<-e, Equals, event.SuiteEnd)
}

type dummyConsole struct {
	event chan event.Topic
}

func (dc *dummyConsole) SpecStart(heading string) {
	dc.event <- event.SpecStart
}

func (dc *dummyConsole) SpecEnd(res result.Result) {
	dc.event <- event.SpecEnd
}

func (dc *dummyConsole) ScenarioStart(heading string) {
	dc.event <- event.ScenarioStart
}

func (dc *dummyConsole) ScenarioEnd(res result.Result) {
	dc.event <- event.ScenarioEnd
}

func (dc *dummyConsole) StepStart(stepText string) {
	dc.event <- event.StepStart
}

func (dc *dummyConsole) StepEnd(step gauge.Step, res result.Result, execInfo gauge_messages.ExecutionInfo) {
	dc.event <- event.StepEnd
}

func (dc *dummyConsole) ConceptStart(conceptHeading string) {
	dc.event <- event.ConceptStart
}

func (dc *dummyConsole) ConceptEnd(res result.Result) {
	dc.event <- event.ConceptEnd
}

func (dc *dummyConsole) SuiteEnd(res result.Result) {
	dc.event <- event.SuiteEnd
}

func (dc *dummyConsole) DataTable(table string) {
	dc.event <- dataTableEvent
}

func (dc *dummyConsole) Errorf(err string, args ...interface{}) {
}

func (dc *dummyConsole) Write(b []byte) (int, error) {
	return len(b), nil
}
