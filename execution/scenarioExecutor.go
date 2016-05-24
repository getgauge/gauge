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

	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/validation"
	"github.com/golang/protobuf/proto"
)

type scenarioExecutor struct {
	runner               runner.Runner
	pluginHandler        *plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	stepExecutor         *stepExecutor
	errMap               *validation.ValidationErrMaps
	stream               int
}

func newScenarioExecutor(r runner.Runner, ph *plugin.Handler, ei *gauge_messages.ExecutionInfo, errMap *validation.ValidationErrMaps, stream int) *scenarioExecutor {
	return &scenarioExecutor{
		runner:               r,
		pluginHandler:        ph,
		currentExecutionInfo: ei,
		errMap:               errMap,
		stream:               stream,
	}
}

func (e *scenarioExecutor) execute(scenarioResult *result.ScenarioResult, scenario *gauge.Scenario, contexts []*gauge.Step, teardowns []*gauge.Step) {
	scenarioResult.ProtoScenario.Skipped = proto.Bool(false)
	if _, ok := e.errMap.ScenarioErrs[scenario]; ok {
		setSkipInfoInResult(scenarioResult, scenario, e.errMap)
		return
	}

	event.Notify(event.NewExecutionEvent(event.ScenarioStart, scenario, nil, e.stream, gauge_messages.ExecutionInfo{}))
	defer event.Notify(event.NewExecutionEvent(event.ScenarioEnd, nil, scenarioResult, e.stream, gauge_messages.ExecutionInfo{}))

	res := e.initScenarioDataStore()
	if res.GetFailed() {
		e.handleScenarioDataStoreFailure(scenarioResult, scenario, fmt.Errorf("Failed to initialize scenario datastore. Error: %s", res.GetErrorMessage()))
		return
	}

	e.notifyBeforeScenarioHook(scenarioResult)
	if !scenarioResult.GetFailed() {
		e.executeItems(contexts, scenarioResult.ProtoScenario.GetContexts(), scenarioResult)
		if !scenarioResult.GetFailed() {
			e.executeItems(scenario.Steps, scenarioResult.ProtoScenario.GetScenarioItems(), scenarioResult)
		}
		e.executeItems(teardowns, scenarioResult.ProtoScenario.GetTearDownSteps(), scenarioResult)
	}
	e.notifyAfterScenarioHook(scenarioResult)
	scenarioResult.UpdateExecutionTime()
}

func (e *scenarioExecutor) initScenarioDataStore() *gauge_messages.ProtoExecutionResult {
	initScenarioDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioDataStoreInit.Enum(),
		ScenarioDataStoreInitRequest: &gauge_messages.ScenarioDataStoreInitRequest{}}
	return e.runner.ExecuteAndGetStatus(initScenarioDataStoreMessage)
}

func (e *scenarioExecutor) handleScenarioDataStoreFailure(scenarioResult *result.ScenarioResult, scenario *gauge.Scenario, err error) {
	logger.Errorf(err.Error())
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
		handleHookFailure(scenarioResult, res, result.AddPreHook)
	}
}

func (e *scenarioExecutor) notifyAfterScenarioHook(scenarioResult *result.ScenarioResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionEnding.Enum(),
		ScenarioExecutionEndingRequest: &gauge_messages.ScenarioExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := executeHook(message, scenarioResult, e.runner, e.pluginHandler)
	if res.GetFailed() {
		setScenarioFailure(e.currentExecutionInfo)
		handleHookFailure(scenarioResult, res, result.AddPostHook)
	}
}

func (e *scenarioExecutor) executeItems(items []*gauge.Step, protoItems []*gauge_messages.ProtoItem, scenarioResult *result.ScenarioResult) {
	var itemsIndex int
	for _, protoItem := range protoItems {
		if protoItem.GetItemType() == gauge_messages.ProtoItem_Concept || protoItem.GetItemType() == gauge_messages.ProtoItem_Step {
			e.executeItem(items[itemsIndex], protoItem, scenarioResult)
			itemsIndex++
			if scenarioResult.GetFailed() {
				return
			}
		}
	}
}

func (e *scenarioExecutor) executeItem(item *gauge.Step, protoItem *gauge_messages.ProtoItem, scenarioResult *result.ScenarioResult) {
	var failed bool
	if protoItem.GetItemType() == gauge_messages.ProtoItem_Concept {
		protoConcept := protoItem.GetConcept()
		failed = e.executeConcept(item, protoConcept, scenarioResult).GetFailed()

	} else if protoItem.GetItemType() == gauge_messages.ProtoItem_Step {
		se := &stepExecutor{runner: e.runner, pluginHandler: e.pluginHandler, currentExecutionInfo: e.currentExecutionInfo, stream: e.stream}
		res := se.executeStep(item, protoItem.GetStep())
		protoItem.GetStep().StepExecutionResult = res.ProtoStepExecResult()
		failed = res.GetFailed()
	}
	if failed {
		scenarioResult.SetFailure()
	}
}

func (e *scenarioExecutor) executeConcept(item *gauge.Step, protoConcept *gauge_messages.ProtoConcept, scenarioResult *result.ScenarioResult) *result.ConceptResult {
	cptResult := result.NewConceptResult(protoConcept)
	event.Notify(event.NewExecutionEvent(event.ConceptStart, item, nil, e.stream, gauge_messages.ExecutionInfo{}))
	defer event.Notify(event.NewExecutionEvent(event.ConceptEnd, nil, cptResult, e.stream, gauge_messages.ExecutionInfo{}))

	var conceptStepIndex int
	for _, protoStep := range protoConcept.Steps {
		if protoStep.GetItemType() == gauge_messages.ProtoItem_Concept || protoStep.GetItemType() == gauge_messages.ProtoItem_Step {
			e.executeItem(item.ConceptSteps[conceptStepIndex], protoStep, scenarioResult)
			conceptStepIndex++
			if scenarioResult.GetFailed() {
				cptResult.UpdateConceptExecResult()
				return cptResult
			}
		}
	}
	cptResult.UpdateConceptExecResult()
	return cptResult
}

func setStepFailure(executionInfo *gauge_messages.ExecutionInfo) {
	setScenarioFailure(executionInfo)
	executionInfo.CurrentStep.IsFailed = proto.Bool(true)
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
