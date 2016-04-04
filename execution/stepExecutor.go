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
	runner               *runner.TestRunner
	pluginHandler        *plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	consoleReporter      reporter.Reporter
}

func newStepExecutor(r *runner.TestRunner, ph *plugin.Handler, ei *gauge_messages.ExecutionInfo, rep reporter.Reporter) *stepExecutor {
	return &stepExecutor{
		runner:               r,
		pluginHandler:        ph,
		currentExecutionInfo: ei,
		consoleReporter:      rep,
	}
}

func (e *stepExecutor) execute(scenarioResult *result.ScenarioResult) {
	e.executeContextSteps(scenarioResult)
	if !scenarioResult.GetFailure() {
		e.executeScenarioSteps(scenarioResult)
	}
	e.executeTearDownSteps(scenarioResult)
}

func (e *stepExecutor) executeContextSteps(scenarioResult *result.ScenarioResult) {
	failure := e.executeItems(scenarioResult.ProtoScenario.GetContexts())
	if failure {
		scenarioResult.SetFailure()
	}
}

func (e *stepExecutor) executeScenarioSteps(scenarioResult *result.ScenarioResult) {
	failure := e.executeItems(scenarioResult.ProtoScenario.GetScenarioItems())
	if failure {
		scenarioResult.SetFailure()
	}
}

func (e *stepExecutor) executeTearDownSteps(scenarioResult *result.ScenarioResult) {
	failure := e.executeItems(scenarioResult.ProtoScenario.TearDownSteps)
	if failure {
		scenarioResult.SetFailure()
	}
}

func (e *stepExecutor) executeItems(executingItems []*gauge_messages.ProtoItem) bool {
	for _, protoItem := range executingItems {
		failure := e.executeItem(protoItem)
		if failure == true {
			return true
		}
	}
	return false
}

func (e *stepExecutor) executeItem(protoItem *gauge_messages.ProtoItem) bool {
	if protoItem.GetItemType() == gauge_messages.ProtoItem_Concept {
		return e.executeConcept(protoItem.GetConcept())
	} else if protoItem.GetItemType() == gauge_messages.ProtoItem_Step {
		return e.executeStep(protoItem.GetStep())
	}
	return false
}

func (e *stepExecutor) executeStep(protoStep *gauge_messages.ProtoStep) bool {
	stepRequest := e.createStepRequest(protoStep)
	stepText := formatter.FormatStep(parser.CreateStepFromStepRequest(stepRequest))
	e.consoleReporter.StepStart(stepText)

	protoStepExecResult := &gauge_messages.ProtoStepExecutionResult{}
	e.currentExecutionInfo.CurrentStep = &gauge_messages.StepInfo{Step: stepRequest, IsFailed: proto.Bool(false)}

	beforeHookStatus := e.notifyBeforeStepHook()
	if beforeHookStatus.GetFailed() {
		protoStepExecResult.PreHookFailure = result.GetProtoHookFailure(beforeHookStatus)
		protoStepExecResult.ExecutionResult = &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(true)}
		setStepFailure(e.currentExecutionInfo)
		printStatus(beforeHookStatus, e.consoleReporter)
	} else {
		executeStepMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecuteStep.Enum(), ExecuteStepRequest: stepRequest}
		stepExecutionStatus := executeAndGetStatus(e.runner, executeStepMessage)
		if stepExecutionStatus.GetFailed() {
			setStepFailure(e.currentExecutionInfo)
		}
		protoStepExecResult.ExecutionResult = stepExecutionStatus
	}
	afterStepHookStatus := e.notifyAfterStepHook()
	addExecutionTimes(protoStepExecResult, beforeHookStatus, afterStepHookStatus)
	if afterStepHookStatus.GetFailed() {
		setStepFailure(e.currentExecutionInfo)
		printStatus(afterStepHookStatus, e.consoleReporter)
		protoStepExecResult.PostHookFailure = result.GetProtoHookFailure(afterStepHookStatus)
		protoStepExecResult.ExecutionResult.Failed = proto.Bool(true)
	}
	protoStepExecResult.ExecutionResult.Message = afterStepHookStatus.Message
	protoStepExecResult.Skipped = protoStep.StepExecutionResult.Skipped
	protoStepExecResult.SkippedReason = protoStep.StepExecutionResult.SkippedReason
	protoStep.StepExecutionResult = protoStepExecResult

	stepFailed := protoStep.GetStepExecutionResult().GetExecutionResult().GetFailed()
	if stepFailed {
		result := protoStep.GetStepExecutionResult().GetExecutionResult()
		e.consoleReporter.Errorf("\nFailed Step: %s", e.currentExecutionInfo.CurrentStep.Step.GetActualStepText())
		e.consoleReporter.Errorf("Error Message: %s", strings.TrimSpace(result.GetErrorMessage()))
		e.consoleReporter.Errorf("Stacktrace: \n%s", result.GetStackTrace())
	}
	e.consoleReporter.StepEnd(stepFailed)
	return stepFailed
}

func (e *stepExecutor) executeConcept(protoConcept *gauge_messages.ProtoConcept) bool {
	e.consoleReporter.ConceptStart(formatter.FormatConcept(protoConcept))
	for _, step := range protoConcept.Steps {
		failure := e.executeItem(step)
		e.setExecutionResultForConcept(protoConcept)
		if failure {
			return true
		}
	}
	conceptFailed := protoConcept.GetConceptExecutionResult().GetExecutionResult().GetFailed()
	e.consoleReporter.ConceptEnd(conceptFailed)
	return conceptFailed
}

func (e *stepExecutor) createStepRequest(protoStep *gauge_messages.ProtoStep) *gauge_messages.ExecuteStepRequest {
	stepRequest := &gauge_messages.ExecuteStepRequest{ParsedStepText: proto.String(protoStep.GetParsedText()), ActualStepText: proto.String(protoStep.GetActualText())}
	stepRequest.Parameters = getParameters(protoStep.GetFragments())
	return stepRequest
}

func (e *stepExecutor) notifyBeforeStepHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepExecutionStarting.Enum(),
		StepExecutionStartingRequest: &gauge_messages.StepExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	e.pluginHandler.NotifyPlugins(message)
	return executeAndGetStatus(e.runner, message)
}

func (e *stepExecutor) notifyAfterStepHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepExecutionEnding.Enum(),
		StepExecutionEndingRequest: &gauge_messages.StepExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	e.pluginHandler.NotifyPlugins(message)
	return executeAndGetStatus(e.runner, message)
}

func setStepFailure(executionInfo *gauge_messages.ExecutionInfo) {
	setScenarioFailure(executionInfo)
	executionInfo.CurrentStep.IsFailed = proto.Bool(true)
}

func (e *stepExecutor) setExecutionResultForConcept(protoConcept *gauge_messages.ProtoConcept) {
	var conceptExecutionTime int64
	for _, step := range protoConcept.GetSteps() {
		if step.GetItemType() == gauge_messages.ProtoItem_Concept {
			stepExecResult := step.GetConcept().GetConceptExecutionResult().GetExecutionResult()
			conceptExecutionTime += stepExecResult.GetExecutionTime()
			if step.GetConcept().GetConceptExecutionResult().GetExecutionResult().GetFailed() {
				conceptExecutionResult := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: step.GetConcept().GetConceptExecutionResult().GetExecutionResult(), Skipped: proto.Bool(false)}
				conceptExecutionResult.ExecutionResult.ExecutionTime = proto.Int64(conceptExecutionTime)
				protoConcept.ConceptExecutionResult = conceptExecutionResult
				protoConcept.ConceptStep.StepExecutionResult = conceptExecutionResult
				return
			}
		} else if step.GetItemType() == gauge_messages.ProtoItem_Step {
			stepExecResult := step.GetStep().GetStepExecutionResult().GetExecutionResult()
			conceptExecutionTime += stepExecResult.GetExecutionTime()
			if stepExecResult.GetFailed() {
				conceptExecutionResult := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: stepExecResult, Skipped: proto.Bool(false)}
				conceptExecutionResult.ExecutionResult.ExecutionTime = proto.Int64(conceptExecutionTime)
				protoConcept.ConceptExecutionResult = conceptExecutionResult
				protoConcept.ConceptStep.StepExecutionResult = conceptExecutionResult
				return
			}
		}
	}
	protoConcept.ConceptExecutionResult = &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(false), ExecutionTime: proto.Int64(conceptExecutionTime)}}
	protoConcept.ConceptStep.StepExecutionResult = protoConcept.ConceptExecutionResult
	protoConcept.ConceptStep.StepExecutionResult.Skipped = proto.Bool(false)
}

func getParameters(fragments []*gauge_messages.Fragment) []*gauge_messages.Parameter {
	var parameters []*gauge_messages.Parameter
	for _, fragment := range fragments {
		if fragment.GetFragmentType() == gauge_messages.Fragment_Parameter {
			parameters = append(parameters, fragment.GetParameter())
		}
	}
	return parameters
}

func setScenarioFailure(executionInfo *gauge_messages.ExecutionInfo) {
	setSpecFailure(executionInfo)
	executionInfo.CurrentScenario.IsFailed = proto.Bool(true)
}
