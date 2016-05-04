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

package event

import (
	"testing"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestInitRegistry(c *C) {
	InitRegistry()

	c.Assert(len(subscriberRegistry), Equals, 8)
}

func (s *MySuite) TestRegisterForOneTopic(c *C) {
	InitRegistry()
	ch := make(chan ExecutionEvent)

	Register(ch, StepEnd)

	c.Assert(subscriberRegistry[StepEnd][0], Equals, ch)
}

func (s *MySuite) TestRegisterForMultipleTopics(c *C) {
	InitRegistry()
	ch := make(chan ExecutionEvent)

	Register(ch, StepEnd, StepStart, SpecEnd, SpecStart)

	c.Assert(subscriberRegistry[StepEnd][0], Equals, ch)
	c.Assert(subscriberRegistry[StepStart][0], Equals, ch)
	c.Assert(subscriberRegistry[SpecEnd][0], Equals, ch)
	c.Assert(subscriberRegistry[SpecEnd][0], Equals, ch)
}

func (s *MySuite) TestMultipleSubscribersRegisteringForMultipleEvent(c *C) {
	InitRegistry()

	ch1 := make(chan ExecutionEvent)
	Register(ch1, StepStart, StepEnd)

	ch2 := make(chan ExecutionEvent)
	Register(ch2, StepStart, StepEnd)

	ch3 := make(chan ExecutionEvent)
	Register(ch3, SpecStart, SpecEnd, StepStart, StepEnd)

	c.Assert(len(subscriberRegistry[SpecStart]), Equals, 1)
	c.Assert(contains(subscriberRegistry[SpecStart], ch3), Equals, true)

	c.Assert(len(subscriberRegistry[SpecEnd]), Equals, 1)
	c.Assert(contains(subscriberRegistry[SpecEnd], ch3), Equals, true)

	c.Assert(len(subscriberRegistry[StepStart]), Equals, 3)
	c.Assert(contains(subscriberRegistry[StepStart], ch1), Equals, true)
	c.Assert(contains(subscriberRegistry[StepStart], ch2), Equals, true)
	c.Assert(contains(subscriberRegistry[StepStart], ch3), Equals, true)

	c.Assert(len(subscriberRegistry[StepEnd]), Equals, 3)
	c.Assert(contains(subscriberRegistry[StepEnd], ch1), Equals, true)
	c.Assert(contains(subscriberRegistry[StepEnd], ch2), Equals, true)
	c.Assert(contains(subscriberRegistry[StepEnd], ch3), Equals, true)
}

func (s *MySuite) TestNotify(c *C) {
	InitRegistry()

	ch1 := make(chan ExecutionEvent, 2)
	Register(ch1, StepStart, StepEnd)

	ch2 := make(chan ExecutionEvent, 2)
	Register(ch2, StepStart, StepEnd)

	stepText := "Hello World"
	protoStep := &gauge_messages.ProtoStep{ActualText: &stepText}
	stepRes := result.NewStepResult(protoStep)

	step := &gauge.Step{Value: stepText}
	stepStartEvent := NewExecutionEvent(StepStart, nil, step)
	stepEndEvent := NewExecutionEvent(StepEnd, stepRes, nil)

	Notify(stepStartEvent)
	c.Assert(<-ch1, DeepEquals, stepStartEvent)
	c.Assert(<-ch2, DeepEquals, stepStartEvent)

	Notify(stepEndEvent)
	c.Assert(<-ch1, DeepEquals, stepEndEvent)
	c.Assert(<-ch2, DeepEquals, stepEndEvent)
}

func contains(arr []chan ExecutionEvent, key chan ExecutionEvent) bool {
	for _, k := range arr {
		if k == key {
			return true
		}
	}
	return false
}
