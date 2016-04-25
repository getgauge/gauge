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
	"fmt"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/validation"
	"github.com/golang/protobuf/proto"
)

type scenarioExecutor struct {
	runner               runner.Runner
	pluginHandler        *plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	consoleReporter      reporter.Reporter
	stepExecutor         *stepExecutor
	errMap               *validation.ValidationErrMaps
}

func newScenarioExecutor(r runner.Runner, ph *plugin.Handler, ei *gauge_messages.ExecutionInfo, rep reporter.Reporter, errMap *validation.ValidationErrMaps) *scenarioExecutor {
	return &scenarioExecutor{
		runner:               r,
		pluginHandler:        ph,
		currentExecutionInfo: ei,
		consoleReporter:      rep,
		errMap:               errMap,
	}
}

func (e *scenarioExecutor) execute(scenarioResult *result.ScenarioResult, scenario *gauge.Scenario) {
	scenarioResult.ProtoScenario.Skipped = proto.Bool(false)
	if _, ok := e.errMap.ScenarioErrs[scenario]; ok {
		setSkipInfoInResult(scenarioResult, scenario, e.errMap)
		return
	}
	res := e.initScenarioDataStore()
	if res.GetFailed() {
		e.consoleReporter.Errorf("Failed to initialize scenario datastore. Error: %s", res.GetErrorMessage())
		e.handleScenarioDataStoreFailure(scenarioResult, scenario, fmt.Errorf(res.GetErrorMessage()))
		return
	}
	e.consoleReporter.ScenarioStart(scenarioResult.ProtoScenario.GetScenarioHeading())
	e.notifyBeforeScenarioHook(scenarioResult)
	if !scenarioResult.ProtoScenario.GetFailed() {
		e.executeContextSteps(scenarioResult)
		if !scenarioResult.GetFailed() {
			e.executeScenarioSteps(scenarioResult)
		}
		e.executeTearDownSteps(scenarioResult)
	}
	e.notifyAfterScenarioHook(scenarioResult)
	scenarioResult.UpdateExecutionTime()
	e.consoleReporter.ScenarioEnd(scenarioResult.GetFailed())
}

func (e *scenarioExecutor) initScenarioDataStore() *gauge_messages.ProtoExecutionResult {
	initScenarioDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioDataStoreInit.Enum(),
		ScenarioDataStoreInitRequest: &gauge_messages.ScenarioDataStoreInitRequest{}}
	return e.runner.ExecuteAndGetStatus(initScenarioDataStoreMessage)
}

func (e *scenarioExecutor) handleScenarioDataStoreFailure(scenarioResult *result.ScenarioResult, scenario *gauge.Scenario, err error) {
	e.consoleReporter.Errorf(err.Error())
	validationError := validation.NewValidationError(&gauge.Step{LineNo: scenario.Heading.LineNo, LineText: scenario.Heading.Value},
		err.Error(), e.currentExecutionInfo.CurrentSpec.GetFileName(), nil)
	e.errMap.ScenarioErrs[scenario] = []*validation.StepValidationError{validationError}
	setSkipInfoInResult(scenarioResult, scenario, e.errMap)
}

func setSkipInfoInResult(result *result.ScenarioResult, scenario *gauge.Scenario, errMap *validation.ValidationErrMaps) {
	result.ProtoScenario.Skipped = proto.Bool(true)
	var errors []string
	for _, err := range errMap.ScenarioErrs[scenario] {
		errors = append(errors, err.Error())
	}
	result.ProtoScenario.SkipErrors = errors
}

func (e *scenarioExecutor) notifyBeforeScenarioHook(scenarioResult *result.ScenarioResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionStarting.Enum(),
		ScenarioExecutionStartingRequest: &gauge_messages.ScenarioExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := executeHook(message, scenarioResult, e.runner, e.pluginHandler)
	if res.GetFailed() {
		setScenarioFailure(e.currentExecutionInfo)
		handleHookFailure(scenarioResult, res, result.AddPreHook, e.consoleReporter)
	}
}

func (e *scenarioExecutor) notifyAfterScenarioHook(scenarioResult *result.ScenarioResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionEnding.Enum(),
		ScenarioExecutionEndingRequest: &gauge_messages.ScenarioExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := executeHook(message, scenarioResult, e.runner, e.pluginHandler)
	if res.GetFailed() {
		setScenarioFailure(e.currentExecutionInfo)
		handleHookFailure(scenarioResult, res, result.AddPreHook, e.consoleReporter)
	}
}

func (e *scenarioExecutor) executeContextSteps(scenarioResult *result.ScenarioResult) {
	failure := e.executeItems(scenarioResult.ProtoScenario.GetContexts())
	if failure {
		scenarioResult.SetFailure()
	}
}

func (e *scenarioExecutor) executeScenarioSteps(scenarioResult *result.ScenarioResult) {
	failure := e.executeItems(scenarioResult.ProtoScenario.GetScenarioItems())
	if failure {
		scenarioResult.SetFailure()
	}
}

func (e *scenarioExecutor) executeTearDownSteps(scenarioResult *result.ScenarioResult) {
	failure := e.executeItems(scenarioResult.ProtoScenario.TearDownSteps)
	if failure {
		scenarioResult.SetFailure()
	}
}

func (e *scenarioExecutor) executeItems(executingItems []*gauge_messages.ProtoItem) bool {
	for _, protoItem := range executingItems {
		res := e.executeItem(protoItem)
		if res != nil && res.GetExecutionResult().GetFailed() {
			return true
		}
	}
	return false
}

func (e *scenarioExecutor) executeItem(protoItem *gauge_messages.ProtoItem) *gauge_messages.ProtoStepExecutionResult {
	if protoItem.GetItemType() == gauge_messages.ProtoItem_Concept {
		return e.executeConcept(protoItem.GetConcept())
	} else if protoItem.GetItemType() == gauge_messages.ProtoItem_Step {
		se := &stepExecutor{runner: e.runner, pluginHandler: e.pluginHandler, currentExecutionInfo: e.currentExecutionInfo, consoleReporter: e.consoleReporter}
		protoItem.GetStep().StepExecutionResult = se.executeStep(protoItem.GetStep())
		return protoItem.GetStep().StepExecutionResult
	}
	return nil
}

func (e *scenarioExecutor) executeConcept(protoConcept *gauge_messages.ProtoConcept) *gauge_messages.ProtoStepExecutionResult {
	e.consoleReporter.ConceptStart(formatter.FormatConcept(protoConcept))
	for _, step := range protoConcept.Steps {
		res := e.executeItem(step)
		e.setExecutionResultForConcept(protoConcept)
		if res.GetExecutionResult().GetFailed() {
			return protoConcept.GetConceptExecutionResult()
		}
	}
	conceptFailed := protoConcept.GetConceptExecutionResult().GetExecutionResult().GetFailed()
	e.consoleReporter.ConceptEnd(conceptFailed)
	return protoConcept.GetConceptExecutionResult()
}

func setStepFailure(executionInfo *gauge_messages.ExecutionInfo) {
	setScenarioFailure(executionInfo)
	executionInfo.CurrentStep.IsFailed = proto.Bool(true)
}

func (e *scenarioExecutor) setExecutionResultForConcept(protoConcept *gauge_messages.ProtoConcept) {
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
