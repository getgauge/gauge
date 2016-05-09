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
	"github.com/getgauge/gauge/gauge"
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

// TODO: stepExecutor should not consume both gauge.Step and gauge_messages.ProtoStep. The usage of ProtoStep should be eliminated.
func (e *stepExecutor) executeStep(step *gauge.Step, protoStep *gauge_messages.ProtoStep) *result.StepResult {
	stepRequest := e.createStepRequest(protoStep)
	e.currentExecutionInfo.CurrentStep = &gauge_messages.StepInfo{Step: stepRequest, IsFailed: proto.Bool(false)}

	stepText := formatter.FormatStep(parser.CreateStepFromStepRequest(stepRequest))
	e.consoleReporter.StepStart(stepText)

	res := result.NewStepResult(protoStep)
	e.notifyBeforeStepHook(res)

	if !res.GetFailed() {
		executeStepMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecuteStep.Enum(), ExecuteStepRequest: stepRequest}
		stepExecutionStatus := e.runner.ExecuteAndGetStatus(executeStepMessage)
		if stepExecutionStatus.GetFailed() {
			setStepFailure(e.currentExecutionInfo)
		}
		res.SetProtoExecResult(stepExecutionStatus)
	}

	e.notifyAfterStepHook(res)

	stepFailed := res.GetFailed()
	if stepFailed {
		r := protoStep.GetStepExecutionResult().GetExecutionResult()
		e.consoleReporter.Errorf("\nFailed Step: %s", e.currentExecutionInfo.CurrentStep.Step.GetActualStepText())
		e.consoleReporter.Errorf("Error Message: %s", strings.TrimSpace(r.GetErrorMessage()))
		e.consoleReporter.Errorf("Stacktrace: \n%s", r.GetStackTrace())
	}
	e.consoleReporter.StepEnd(stepFailed)
	return res
}

func (e *stepExecutor) createStepRequest(protoStep *gauge_messages.ProtoStep) *gauge_messages.ExecuteStepRequest {
	stepRequest := &gauge_messages.ExecuteStepRequest{ParsedStepText: proto.String(protoStep.GetParsedText()), ActualStepText: proto.String(protoStep.GetActualText())}
	stepRequest.Parameters = getParameters(protoStep.GetFragments())
	return stepRequest
}

func (e *stepExecutor) notifyBeforeStepHook(stepResult *result.StepResult) {
	m := &gauge_messages.Message{
		MessageType:                  gauge_messages.Message_StepExecutionStarting.Enum(),
		StepExecutionStartingRequest: &gauge_messages.StepExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo},
	}

	res := executeHook(m, stepResult, e.runner, e.pluginHandler)
	if res.GetFailed() {
		setStepFailure(e.currentExecutionInfo)
		handleHookFailure(stepResult, res, result.AddPreHook, e.consoleReporter)
	}
}

func (e *stepExecutor) notifyAfterStepHook(stepResult *result.StepResult) {
	m := &gauge_messages.Message{
		MessageType:                gauge_messages.Message_StepExecutionEnding.Enum(),
		StepExecutionEndingRequest: &gauge_messages.StepExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo},
	}

	res := executeHook(m, stepResult, e.runner, e.pluginHandler)
	if res.GetFailed() {
		setStepFailure(e.currentExecutionInfo)
		handleHookFailure(stepResult, res, result.AddPostHook, e.consoleReporter)
	}
}
