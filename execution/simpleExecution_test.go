/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"testing"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
)

func TestNotifyBeforeSuiteShouldAddsBeforeSuiteHookMessages(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ExecutionStarting {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"Before Suite Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h}
	simpleExecution := newSimpleExecution(ei, false, false)
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	simpleExecution.notifyBeforeSuite()

	gotMessages := simpleExecution.suiteResult.PreHookMessages

	if len(gotMessages) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(gotMessages))
	}
	if gotMessages[0] != "Before Suite Called" {
		t.Errorf("Expected `Before Suite Called` message, got : %s", gotMessages[0])
	}
}

func TestNotifyAfterSuiteShouldAddsAfterSuiteHookMessages(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ExecutionEnding {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"After Suite Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h}
	simpleExecution := newSimpleExecution(ei, false, false)
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	simpleExecution.notifyAfterSuite()

	gotMessages := simpleExecution.suiteResult.PostHookMessages

	if len(gotMessages) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(gotMessages))
	}
	if gotMessages[0] != "After Suite Called" {
		t.Errorf("Expected `After Suite Called` message, got : %s", gotMessages[0])
	}
}

func TestNotifyBeforeSuiteShouldAddsBeforeSuiteHookScreenshots(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ExecutionStarting {
			return &gauge_messages.ProtoExecutionResult{
				ScreenshotFiles: []string{"screenshot1.png", "screenshot2.png"},
				Failed:          false,
				ExecutionTime:   10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h}
	simpleExecution := newSimpleExecution(ei, false, false)
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	simpleExecution.notifyBeforeSuite()

	beforeSuiteScreenshots := simpleExecution.suiteResult.PreHookScreenshotFiles
	expected := []string{"screenshot1.png", "screenshot2.png"}

	if len(beforeSuiteScreenshots) != len(expected) {
		t.Errorf("Expected 2 screenshots, got : %d", len(beforeSuiteScreenshots))
	}
	for i, e := range expected {
		if string(beforeSuiteScreenshots[i]) != e {
			t.Errorf("Expected `%s` screenshot, got : %s", e, beforeSuiteScreenshots[i])
		}
	}
}

func TestNotifyAfterSuiteShouldAddsAfterSuiteHookScreenshots(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ExecutionEnding {
			return &gauge_messages.ProtoExecutionResult{
				ScreenshotFiles: []string{"screenshot1.png", "screenshot2.png"},
				Failed:          false,
				ExecutionTime:   10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h}
	simpleExecution := newSimpleExecution(ei, false, false)
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	simpleExecution.notifyAfterSuite()

	afterSuiteScreenshots := simpleExecution.suiteResult.PostHookScreenshotFiles
	expected := []string{"screenshot1.png", "screenshot2.png"}

	if len(afterSuiteScreenshots) != len(expected) {
		t.Errorf("Expected 2 screenshots, got : %d", len(afterSuiteScreenshots))
	}
	for i, e := range expected {
		if string(afterSuiteScreenshots[i]) != e {
			t.Errorf("Expected `%s` screenshot, got : %s", e, afterSuiteScreenshots[i])
		}
	}
}
func TestExecuteSpecsShouldAddsBeforeSpecHookFailureScreenshotFile(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_SpecExecutionStarting {
			return &gauge_messages.ProtoExecutionResult{
				Failed:                true,
				ExecutionTime:         10,
				FailureScreenshotFile: "before-spec-hook-failure-screenshot.png",
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h, errMaps: &gauge.BuildErrors{}}
	simpleExecution := newSimpleExecution(ei, false, false)
	specsC := createSpecCollection()
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	specResult := simpleExecution.executeSpecs(specsC)

	actualScreenshotFile := specResult[0].ProtoSpec.PreHookFailures[0].FailureScreenshotFile
	expectedScreenshotFile := "before-spec-hook-failure-screenshot.png"

	if actualScreenshotFile != expectedScreenshotFile {
		t.Errorf("Expected `%s` screenshot, got : %s", expectedScreenshotFile, actualScreenshotFile)
	}
}

func TestExecuteSpecsShouldAddsAfterSpecHookFailureScreenshotFile(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_SpecExecutionEnding {
			return &gauge_messages.ProtoExecutionResult{
				Failed:                true,
				ExecutionTime:         10,
				FailureScreenshotFile: "after-spec-hook-failure-screenshot.png",
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h, errMaps: &gauge.BuildErrors{}}
	simpleExecution := newSimpleExecution(ei, false, false)
	specsC := createSpecCollection()
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	specResult := simpleExecution.executeSpecs(specsC)

	actualScreenshotFile := specResult[0].ProtoSpec.PostHookFailures[0].FailureScreenshotFile
	expectedScreenshotFile := "after-spec-hook-failure-screenshot.png"

	if actualScreenshotFile != expectedScreenshotFile {
		t.Errorf("Expected `%s` screenshot, got : %s", expectedScreenshotFile, actualScreenshotFile)
	}
}

func TestExecuteSpecsShouldAddsBeforeSpecHookFailureScreenshotBytes(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_SpecExecutionStarting {
			return &gauge_messages.ProtoExecutionResult{
				Failed:            true,
				ExecutionTime:     10,
				FailureScreenshot: []byte("before spec hook failure screenshot byte"),
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h, errMaps: &gauge.BuildErrors{}}
	simpleExecution := newSimpleExecution(ei, false, false)
	specsC := createSpecCollection()
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	specResult := simpleExecution.executeSpecs(specsC)

	actualScreenshotBytes := specResult[0].ProtoSpec.PreHookFailures[0].FailureScreenshot
	expectedScreenshotBytes := "before spec hook failure screenshot byte"

	if string(actualScreenshotBytes) != expectedScreenshotBytes {
		t.Errorf("Expected `%s` screenshot, got : %s", expectedScreenshotBytes, actualScreenshotBytes)
	}
}

func TestExecuteSpecsShouldAddsAfterSpecHookFailureScreenshotBytes(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_SpecExecutionEnding {
			return &gauge_messages.ProtoExecutionResult{
				Failed:            true,
				ExecutionTime:     10,
				FailureScreenshot: []byte("after spec hook failure screenshot bytes"),
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h, errMaps: &gauge.BuildErrors{}}
	simpleExecution := newSimpleExecution(ei, false, false)
	specsC := createSpecCollection()
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	specResult := simpleExecution.executeSpecs(specsC)

	actualScreenshotBytes := specResult[0].ProtoSpec.PostHookFailures[0].FailureScreenshot
	expectedScreenshotBytes := "after spec hook failure screenshot bytes"

	if string(actualScreenshotBytes) != expectedScreenshotBytes {
		t.Errorf("Expected `%s` screenshot, got : %s", expectedScreenshotBytes, actualScreenshotBytes)
	}
}

func createSpecCollection() *gauge.SpecCollection {
	var specs []*gauge.Specification
	specs = append(specs, &gauge.Specification{
		FileName: "spec-1.spec",
		Heading: &gauge.Heading{
			Value:       "Spec heading",
			LineNo:      1,
			SpanEnd:     12,
			HeadingType: 0,
		},
	})
	return gauge.NewSpecCollection(specs, false)
}
