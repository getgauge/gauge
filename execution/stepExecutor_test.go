/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"testing"

	"github.com/getgauge/gauge/gauge"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
)

func TestStepExecutionShouldAddBeforeStepHookMessages(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_StepExecutionStarting {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"Before Step Called"},
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
	beforeStepMsg := stepResult.ProtoStep.PreHookMessages

	if len(beforeStepMsg) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(beforeStepMsg))
	}
	if beforeStepMsg[0] != "Before Step Called" {
		t.Errorf("Expected `Before Step Called` message, got : %s", beforeStepMsg[0])
	}
}

func TestStepExecutionShouldAddAfterStepHookMessages(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_StepExecutionEnding {
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
	afterStepMsg := stepResult.ProtoStep.PostHookMessages

	if len(afterStepMsg) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(afterStepMsg))
	}
	if afterStepMsg[0] != "After Step Called" {
		t.Errorf("Expected `After Step Called` message, got : %s", afterStepMsg[0])
	}
}

func TestStepExecutionShouldGetScreenshotsBeforeStep(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_StepExecutionStarting {
			return &gauge_messages.ProtoExecutionResult{
				ScreenshotFiles: []string{"screenshot1.png", "screenshot2.png"},
				Failed:          false,
				ExecutionTime:   10,
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
	beforeStepScreenShots := stepResult.ProtoStep.PreHookScreenshotFiles

	expected := []string{"screenshot1.png", "screenshot2.png"}

	if len(beforeStepScreenShots) != len(expected) {
		t.Errorf("Expected 2 screenshots, got : %d", len(beforeStepScreenShots))
	}

	for i, e := range expected {
		if string(beforeStepScreenShots[i]) != e {
			t.Errorf("Expected `%s` screenshot, got : %s", e, beforeStepScreenShots[i])
		}
	}
}

func TestStepExecutionShouldGetScreenshotsAfterStep(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_StepExecutionEnding {
			return &gauge_messages.ProtoExecutionResult{
				ScreenshotFiles: []string{"screenshot1.png", "screenshot2.png"},
				Failed:          false,
				ExecutionTime:   10,
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
	afterStepScreenShots := stepResult.ProtoStep.PostHookScreenshotFiles

	expected := []string{"screenshot1.png", "screenshot2.png"}

	if len(afterStepScreenShots) != len(expected) {
		t.Errorf("Expected 2 screenshots, got : %d", len(afterStepScreenShots))
	}

	for i, e := range expected {
		if string(afterStepScreenShots[i]) != e {
			t.Errorf("Expected `%s` screenshot, got : %s", e, afterStepScreenShots[i])
		}
	}
}
