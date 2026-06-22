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

func TestStepExecutionShouldRetryOnFailureWhenConditionIsNotSet(t *testing.T) {
	MaxStepRetriesCount = 3
	RetryStepOn = ""
	defer resetStepRetryOptions()

	r := &mockRunner{}
	attempts := 0
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ExecuteStep {
			attempts++
			if attempts < MaxStepRetriesCount {
				return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "transient failure"}
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}

	se, step, protoStep := newStepExecutorForRetryTests(r)
	stepResult := se.executeStep(step, protoStep)

	if attempts != MaxStepRetriesCount {
		t.Errorf("Expected step to execute %d times, got %d", MaxStepRetriesCount, attempts)
	}
	if stepResult.GetFailed() {
		t.Errorf("Expected successful step execution result")
	}
}

func TestStepExecutionShouldRetryOnlyWhenConditionMatches(t *testing.T) {
	MaxStepRetriesCount = 3
	RetryStepOn = "TimeoutException|429"
	defer resetStepRetryOptions()

	r := &mockRunner{}
	attempts := 0
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ExecuteStep {
			attempts++
			if attempts == 1 {
				return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "HTTP 429 TimeoutException"}
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}

	se, step, protoStep := newStepExecutorForRetryTests(r)
	stepResult := se.executeStep(step, protoStep)

	if attempts != 2 {
		t.Errorf("Expected step to execute twice, got %d", attempts)
	}
	if stepResult.GetFailed() {
		t.Errorf("Expected successful step execution result")
	}
}

func TestStepExecutionShouldNotRetryWhenConditionDoesNotMatch(t *testing.T) {
	MaxStepRetriesCount = 3
	RetryStepOn = "TimeoutException"
	defer resetStepRetryOptions()

	r := &mockRunner{}
	attempts := 0
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ExecuteStep {
			attempts++
			return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "AssertionError"}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}

	se, step, protoStep := newStepExecutorForRetryTests(r)
	stepResult := se.executeStep(step, protoStep)

	if attempts != 1 {
		t.Errorf("Expected step to execute once, got %d", attempts)
	}
	if !stepResult.GetFailed() {
		t.Errorf("Expected failed step execution result")
	}
}

func newStepExecutorForRetryTests(r *mockRunner) (*stepExecutor, *gauge.Step, *gauge_messages.ProtoStep) {
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	ei := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Example Spec",
			FileName: "example.spec",
			IsFailed: false,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Example Scenario",
			IsFailed: false,
		},
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
	return se, step, protoStep
}

func resetStepRetryOptions() {
	MaxStepRetriesCount = 1
	RetryStepOn = ""
}
