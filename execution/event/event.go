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
)

// ExecutionEvent represents an event raised during various phases of the
// execution lifecycle. This is only for execution and excludes parsing, validation, etc.
type ExecutionEvent struct {
	Topic  Topic
	Result result.Result
	item   gauge.Item
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

var eventRegistry map[Topic][]chan ExecutionEvent

// Register registers the given channel to the given list of topics. Any updates for the given topics
// will be sent on this channel
func Register(ch chan ExecutionEvent, topics ...Topic) {
	if eventRegistry == nil {
		eventRegistry = make(map[Topic][]chan ExecutionEvent, 0)

	}
	for _, t := range topics {
		if _, ok := eventRegistry[t]; !ok {
			eventRegistry[t] = make([]chan ExecutionEvent, 0)
		}
		eventRegistry[t] = append(eventRegistry[t], ch)
	}
}
