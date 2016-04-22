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
	"strings"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
	"github.com/golang/protobuf/proto"
)

type stepExecutor struct {
	runner               runner.Runner
	pluginHandler        *plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	consoleReporter      reporter.Reporter
}

func (e *stepExecutor) executeStep(protoStep *gauge_messages.ProtoStep) *gauge_messages.ProtoStepExecutionResult {
	stepRequest := e.createStepRequest(protoStep)
	e.currentExecutionInfo.CurrentStep = &gauge_messages.StepInfo{Step: stepRequest, IsFailed: proto.Bool(false)}

	stepText := formatter.FormatStep(parser.CreateStepFromStepRequest(stepRequest))
	e.consoleReporter.StepStart(stepText)

	protoStepExecResult := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}
	e.notifyBeforeStepHook(protoStepExecResult)
	if !protoStepExecResult.ExecutionResult.GetFailed() {
		executeStepMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecuteStep.Enum(), ExecuteStepRequest: stepRequest}
		stepExecutionStatus := e.runner.ExecuteAndGetStatus(executeStepMessage)
		if stepExecutionStatus.GetFailed() {
			setStepFailure(e.currentExecutionInfo)
		}
		protoStepExecResult.ExecutionResult = stepExecutionStatus
	}

	e.notifyAfterStepHook(protoStepExecResult)

	protoStepExecResult.Skipped = protoStep.StepExecutionResult.Skipped
	protoStepExecResult.SkippedReason = protoStep.StepExecutionResult.SkippedReason

	stepFailed := protoStepExecResult.GetExecutionResult().GetFailed()
	if stepFailed {
		result := protoStep.GetStepExecutionResult().GetExecutionResult()
		e.consoleReporter.Errorf("\nFailed Step: %s", e.currentExecutionInfo.CurrentStep.Step.GetActualStepText())
		e.consoleReporter.Errorf("Error Message: %s", strings.TrimSpace(result.GetErrorMessage()))
		e.consoleReporter.Errorf("Stacktrace: \n%s", result.GetStackTrace())
	}
	e.consoleReporter.StepEnd(stepFailed)
	return protoStepExecResult
}

func (e *stepExecutor) createStepRequest(protoStep *gauge_messages.ProtoStep) *gauge_messages.ExecuteStepRequest {
	stepRequest := &gauge_messages.ExecuteStepRequest{ParsedStepText: proto.String(protoStep.GetParsedText()), ActualStepText: proto.String(protoStep.GetActualText())}
	stepRequest.Parameters = getParameters(protoStep.GetFragments())
	return stepRequest
}

func (e *stepExecutor) notifyBeforeStepHook(stepResult *gauge_messages.ProtoStepExecutionResult) {
	msg := &gauge_messages.Message{
		MessageType:                  gauge_messages.Message_StepExecutionStarting.Enum(),
		StepExecutionStartingRequest: &gauge_messages.StepExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo},
	}
	e.pluginHandler.NotifyPlugins(msg)
	execRes := e.runner.ExecuteAndGetStatus(msg)

	if execRes.GetFailed() {
		stepResult.PreHookFailure = result.GetProtoHookFailure(execRes)
		stepResult.ExecutionResult = &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(true)}
		setStepFailure(e.currentExecutionInfo)
		printStatus(execRes, e.consoleReporter)
	}

	execTime := stepResult.ExecutionResult.ExecutionTime
	if execTime == nil {
		stepResult.ExecutionResult.ExecutionTime = proto.Int64(execRes.GetExecutionTime())
	} else {
		stepResult.ExecutionResult.ExecutionTime = proto.Int64(*execTime + execRes.GetExecutionTime())
	}
}

func (e *stepExecutor) notifyAfterStepHook(stepResult *gauge_messages.ProtoStepExecutionResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepExecutionEnding.Enum(),
		StepExecutionEndingRequest: &gauge_messages.StepExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	e.pluginHandler.NotifyPlugins(message)
	execRes := e.runner.ExecuteAndGetStatus(message)

	stepResult.ExecutionResult.Message = execRes.Message

	if execRes.GetFailed() {
		stepResult.PostHookFailure = result.GetProtoHookFailure(execRes)
		stepResult.ExecutionResult.Failed = proto.Bool(true)
		setStepFailure(e.currentExecutionInfo)
		printStatus(execRes, e.consoleReporter)
	}

	execTime := stepResult.ExecutionResult.ExecutionTime
	if execTime == nil {
		stepResult.ExecutionResult.ExecutionTime = proto.Int64(execRes.GetExecutionTime())
	} else {
		stepResult.ExecutionResult.ExecutionTime = proto.Int64(*execTime + execRes.GetExecutionTime())
	}
}
