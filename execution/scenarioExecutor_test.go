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

package execution

import (
	"testing"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"

	"github.com/getgauge/gauge/gauge_messages"
)

func TestNotifyBeforeScenarioShouldAddBeforeScenarioHookMessages(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ScenarioExecutionStarting {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"Before Scenario Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &gauge_messages.ExecutionInfo{}
	sce := newScenarioExecutor(r, h, ei, nil, nil, nil, 0)
	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "A scenario"},
		Span:    &gauge.Span{Start: 2, End: 10},
	}
	scenarioResult := result.NewScenarioResult(gauge.NewProtoScenario(scenario))
	sce.notifyBeforeScenarioHook(scenarioResult)
	gotMessages := scenarioResult.ProtoScenario.Message

	if len(gotMessages) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(gotMessages))
	}
	if gotMessages[0] != "Before Scenario Called" {
		t.Errorf("Expected `Before Scenario Called` message, got : %s", gotMessages[0])
	}
}

func TestNotifyAfterScenarioShouldAddAfterScenarioHookMessages(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ScenarioExecutionEnding {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"After Scenario Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &gauge_messages.ExecutionInfo{}
	sce := newScenarioExecutor(r, h, ei, nil, nil, nil, 0)
	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "A scenario"},
		Span:    &gauge.Span{Start: 2, End: 10},
	}
	scenarioResult := result.NewScenarioResult(gauge.NewProtoScenario(scenario))
	sce.notifyAfterScenarioHook(scenarioResult)
	gotMessages := scenarioResult.ProtoScenario.Message

	if len(gotMessages) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(gotMessages))
	}
	if gotMessages[0] != "After Scenario Called" {
		t.Errorf("Expected `After Scenario Called` message, got : %s", gotMessages[0])
	}
}
