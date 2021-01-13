/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
)

type stepExecutor struct {
	runner               runner.Runner
	pluginHandler        plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	stream               int
}

// TODO: stepExecutor should not consume both gauge.Step and gauge_messages.ProtoStep. The usage of ProtoStep should be eliminated.
func (e *stepExecutor) executeStep(step *gauge.Step, protoStep *gauge_messages.ProtoStep) *result.StepResult {
	stepRequest := e.createStepRequest(protoStep)
	e.currentExecutionInfo.CurrentStep = &gauge_messages.StepInfo{Step: stepRequest, IsFailed: false}
	stepResult := result.NewStepResult(protoStep)
	for i := range step.GetFragments() {
		stepFragmet := step.GetFragments()[i]
		protoStepFragmet := protoStep.GetFragments()[i]
		if stepFragmet.FragmentType == gauge_messages.Fragment_Parameter && stepFragmet.Parameter.ParameterType == gauge_messages.Parameter_Dynamic {
			stepFragmet.GetParameter().Value = protoStepFragmet.GetParameter().Value
		}
	}
	event.Notify(event.NewExecutionEvent(event.StepStart, step, nil, e.stream, e.currentExecutionInfo))

	e.notifyBeforeStepHook(stepResult)
	if !stepResult.GetFailed() {
		executeStepMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecuteStep, ExecuteStepRequest: stepRequest}
		stepExecutionStatus := e.runner.ExecuteAndGetStatus(executeStepMessage)
		stepExecutionStatus.Message = append(stepResult.ProtoStepExecResult().GetExecutionResult().Message, stepExecutionStatus.Message...)
		stepExecutionStatus.Screenshots = append(stepResult.ProtoStepExecResult().GetExecutionResult().Screenshots, stepExecutionStatus.Screenshots...)
		if stepExecutionStatus.GetFailed() {
			e.currentExecutionInfo.CurrentStep.ErrorMessage = stepExecutionStatus.GetErrorMessage()
			e.currentExecutionInfo.CurrentStep.StackTrace = stepExecutionStatus.GetStackTrace()
			setStepFailure(e.currentExecutionInfo)
			stepResult.SetStepFailure()
		}
		stepResult.SetProtoExecResult(stepExecutionStatus)
	}
	e.notifyAfterStepHook(stepResult)

	event.Notify(event.NewExecutionEvent(event.StepEnd, *step, stepResult, e.stream, e.currentExecutionInfo))
	defer e.currentExecutionInfo.CurrentStep.Reset()
	return stepResult
}

func (e *stepExecutor) createStepRequest(protoStep *gauge_messages.ProtoStep) *gauge_messages.ExecuteStepRequest {
	stepRequest := &gauge_messages.ExecuteStepRequest{ParsedStepText: protoStep.GetParsedText(), ActualStepText: protoStep.GetActualText(), Stream: int32(e.stream)}
	stepRequest.Parameters = getParameters(protoStep.GetFragments())
	return stepRequest
}

func (e *stepExecutor) notifyBeforeStepHook(stepResult *result.StepResult) {
	m := &gauge_messages.Message{
		MessageType:                  gauge_messages.Message_StepExecutionStarting,
		StepExecutionStartingRequest: &gauge_messages.StepExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo, Stream: int32(e.stream)},
	}
	e.pluginHandler.NotifyPlugins(m)
	res := executeHook(m, stepResult, e.runner)
	stepResult.ProtoStep.PreHookMessages = res.Message
	stepResult.ProtoStep.PreHookScreenshotFiles = res.ScreenshotFiles
	stepResult.ProtoStep.PreHookScreenshots = res.Screenshots
	if res.GetFailed() {
		setStepFailure(e.currentExecutionInfo)
		handleHookFailure(stepResult, res, result.AddPreHook)
	}
	m.StepExecutionStartingRequest.StepResult = gauge.ConvertToProtoStepResult(stepResult)
	e.pluginHandler.NotifyPlugins(m)
}

func (e *stepExecutor) notifyAfterStepHook(stepResult *result.StepResult) {
	m := &gauge_messages.Message{
		MessageType:                gauge_messages.Message_StepExecutionEnding,
		StepExecutionEndingRequest: &gauge_messages.StepExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo, Stream: int32(e.stream)},
	}

	res := executeHook(m, stepResult, e.runner)
	stepResult.ProtoStep.PostHookMessages = res.Message
	stepResult.ProtoStep.PostHookScreenshotFiles = res.ScreenshotFiles
	stepResult.ProtoStep.PostHookScreenshots = res.Screenshots
	if res.GetFailed() {
		setStepFailure(e.currentExecutionInfo)
		handleHookFailure(stepResult, res, result.AddPostHook)
	}
	m.StepExecutionEndingRequest.StepResult = gauge.ConvertToProtoStepResult(stepResult)
	e.pluginHandler.NotifyPlugins(m)
}
