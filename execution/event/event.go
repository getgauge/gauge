/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package event

import (
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
)

// ExecutionEvent represents an event raised during various phases of the
// execution lifecycle. This is only for execution and excludes parsing, validation, etc.
type ExecutionEvent struct {
	Topic         Topic
	Item          gauge.Item
	Result        result.Result
	Stream        int
	ExecutionInfo *gauge_messages.ExecutionInfo
}

// NewExecutionEvent creates a new execution event.
func NewExecutionEvent(t Topic, i gauge.Item, r result.Result, stream int, executionInfo *gauge_messages.ExecutionInfo) ExecutionEvent {
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
