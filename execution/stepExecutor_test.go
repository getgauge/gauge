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

	"github.com/getgauge/gauge/gauge"

	"github.com/getgauge/gauge/gauge_messages"
)

func TestStepExecutionShouldAddStepHookMessages(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_StepExecutionStarting {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"Before Step Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		} else if m.MessageType == gauge_messages.Message_ExecuteStep {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"Step Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		} else if m.MessageType == gauge_messages.Message_StepExecutionEnding {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"After Step Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &gauge_messages.ExecutionInfo{
		CurrentStep: &gauge_messages.StepInfo{
			Step: &gauge_messages.ExecuteStepRequest{
				ActualStepText:  "a simple step",
				ParsedStepText:  "a simple step",
				ScenarioFailing: false,
			},
			IsFailed: false,
		},
	}
	se := &stepExecutor{runner: r, pluginHandler: h, currentExecutionInfo: ei, stream: 0}
	step := &gauge.Step{
		Value:     "a simple step",
		LineText:  "a simple step",
		Fragments: []*gauge_messages.Fragment{{FragmentType: gauge_messages.Fragment_Text, Text: "a simple step"}},
	}
	protoStep := gauge.ConvertToProtoItem(step).GetStep()
	protoStep.StepExecutionResult = &gauge_messages.ProtoStepExecutionResult{}

	stepResult := se.executeStep(step, protoStep)
	gotMessages := stepResult.ProtoStepExecResult().ExecutionResult.Message

	if len(gotMessages) != 3 {
		t.Errorf("Expected 3 message, got : %d", len(gotMessages))
	}
	if gotMessages[0] != "Before Step Called" {
		t.Errorf("Expected `Before Step Called` message, got : %s", gotMessages[0])
	}
	if gotMessages[1] != "Step Called" {
		t.Errorf("Expected `Step Called` message, got : %s", gotMessages[1])
	}
	if gotMessages[2] != "After Step Called" {
		t.Errorf("Expected `After Step Called` message, got : %s", gotMessages[2])
	}
}
