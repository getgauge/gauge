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

	event.Notify(event.NewExecutionEvent(event.SpecEnd, nil, nil))
	c.Assert(dw.output, Equals, "\n")
}

func (s *MySuite) TestSubscribeSpecStart(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	spec := &gauge.Specification{Heading: &gauge.Heading{Value: "My Spec Heading"}}

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.SpecStart, spec, nil))
	c.Assert(dw.output, Equals, "# My Spec Heading\n")
}

func (s *MySuite) TestSubscribeScenarioStart(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	sce := &gauge.Scenario{Heading: &gauge.Heading{Value: "My Scenario Heading"}}

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.ScenarioStart, sce, nil))
	c.Assert(dw.output, Equals, "  ## My Scenario Heading\n")
}

func (s *MySuite) TestSubscribeScenarioEnd(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	sceHeading := "My scenario heading"
	sceRes := result.NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: &sceHeading})

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.ScenarioEnd, nil, sceRes))
	c.Assert(dw.output, Equals, "")
}

func (s *MySuite) TestSubscribeStepStart(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	stepText := "My first step"
	step := &gauge.Step{Value: stepText}

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.StepStart, step, nil))
	c.Assert(dw.output, Equals, spaces(stepIndentation)+"* "+stepText+newline)
}

func (s *MySuite) TestSubscribeStepEnd(c *C) {
	dw, sc := setupSimpleConsole()
	currentReporter = sc
	SimpleConsoleOutput = true
	event.InitRegistry()
	failed := true
	stepExeRes := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: &failed}}
	stepRes := result.NewStepResult(&gauge_messages.ProtoStep{StepExecutionResult: stepExeRes})

	ListenExecutionEvents()

	event.Notify(event.NewExecutionEvent(event.StepEnd, nil, stepRes))
	c.Assert(dw.output, Equals, "")
}
