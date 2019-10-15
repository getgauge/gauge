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
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
)

// ExecutionEvent represents an event raised during various phases of the
// execution lifecycle. This is only for execution and excludes parsing, validation, etc.
type ExecutionEvent struct {
	Topic         Topic
	Item          gauge.Item
	Result        result.Result
	Stream        int
	ExecutionInfo gauge_messages.ExecutionInfo
}

// NewExecutionEvent creates a new execution event.
func NewExecutionEvent(t Topic, i gauge.Item, r result.Result, stream int, executionInfo gauge_messages.ExecutionInfo) ExecutionEvent {
	return ExecutionEvent{
		Topic:         t,
		Item:          i,
		Result:        r,
		Stream:        stream,
		ExecutionInfo: executionInfo,
	}
}

// Topic indicates the topic of ExecutionEvent
type Topic int

const (
	SuiteStart Topic = iota
	SpecStart
	ScenarioStart
	ConceptStart
	StepStart
	StepEnd
	ConceptEnd
	ScenarioEnd
	SpecEnd
	SuiteEnd
)

var subscriberRegistry map[Topic][]chan ExecutionEvent

// InitRegistry is used for console reporting, execution API and rerun of specs
func InitRegistry() {
	subscriberRegistry = make(map[Topic][]chan ExecutionEvent)
	subscriberRegistry[SuiteStart] = make([]chan ExecutionEvent, 0)
	subscriberRegistry[ScenarioStart] = make([]chan ExecutionEvent, 0)
	subscriberRegistry[ConceptStart] = make([]chan ExecutionEvent, 0)
	subscriberRegistry[StepStart] = make([]chan ExecutionEvent, 0)
	subscriberRegistry[SuiteEnd] = make([]chan ExecutionEvent, 0)
	subscriberRegistry[ConceptEnd] = make([]chan ExecutionEvent, 0)
	subscriberRegistry[ScenarioEnd] = make([]chan ExecutionEvent, 0)
	subscriberRegistry[SpecEnd] = make([]chan ExecutionEvent, 0)
	subscriberRegistry[SuiteEnd] = make([]chan ExecutionEvent, 0)
}

// Register registers the given channel to the given list of topics. Any updates for the given topics
// will be sent on this channel
func Register(ch chan ExecutionEvent, topics ...Topic) {
	for _, t := range topics {
		subscriberRegistry[t] = append(subscriberRegistry[t], ch)
	}
}

// Notify notifies all the subscribers of the event about its occurrence
func Notify(e ExecutionEvent) {
	for _, c := range subscriberRegistry[e.Topic] {
		c <- e
	}
}
