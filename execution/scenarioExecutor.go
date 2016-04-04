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
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/validation"
	"github.com/golang/protobuf/proto"
)

func (e *specExecutor) executeScenarios() []*result.ScenarioResult {
	var scenarioResults []*result.ScenarioResult
	for _, scenario := range e.specification.Scenarios {
		scenarioResults = append(scenarioResults, e.executeScenario(scenario))
	}
	return scenarioResults
}

func (e *specExecutor) executeScenario(scenario *gauge.Scenario) *result.ScenarioResult {
	e.currentExecutionInfo.CurrentScenario = &gauge_messages.ScenarioInfo{
		Name:     proto.String(scenario.Heading.Value),
		Tags:     getTagValue(scenario.Tags),
		IsFailed: proto.Bool(false),
	}

	scenarioResult := &result.ScenarioResult{ProtoScenario: gauge.NewProtoScenario(scenario)}
	e.addAllItemsForScenarioExecution(scenario, scenarioResult)
	scenarioResult.ProtoScenario.Skipped = proto.Bool(false)

	if _, ok := e.errMap.ScenarioErrs[scenario]; ok {
		e.setSkipInfoInResult(scenarioResult, scenario)
		return scenarioResult
	}

	res := e.initScenarioDataStore()
	if res.GetFailed() {
		e.consoleReporter.Errorf("Failed to initialize scenario datastore. Error: %s", res.GetErrorMessage())
		e.handleScenarioDataStoreFailure(scenarioResult, scenario, fmt.Errorf(res.GetErrorMessage()))
		return scenarioResult
	}
	e.consoleReporter.ScenarioStart(scenario.Heading.Value)

	e.notifyBeforeScenarioHook(scenarioResult)
	if !e.specResult.IsFailed {
		stepExec := newStepExecutor(e.runner, e.pluginHandler, e.currentExecutionInfo, e.consoleReporter)
		stepExec.execute(scenarioResult)
	}

	e.notifyAfterScenarioHook(scenarioResult)
	scenarioResult.UpdateExecutionTime()

	e.consoleReporter.ScenarioEnd(scenarioResult.GetFailure())
	return scenarioResult
}

func (e *specExecutor) initScenarioDataStore() *gauge_messages.ProtoExecutionResult {
	initScenarioDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioDataStoreInit.Enum(),
		ScenarioDataStoreInitRequest: &gauge_messages.ScenarioDataStoreInitRequest{}}
	return executeAndGetStatus(e.runner, initScenarioDataStoreMessage)
}

func (e *specExecutor) notifyBeforeScenarioHook(scenarioResult *result.ScenarioResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionStarting.Enum(),
		ScenarioExecutionStartingRequest: &gauge_messages.ScenarioExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := e.executeHook(message, scenarioResult)
	if res.GetFailed() {
		setScenarioFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPreHook, e.consoleReporter)
	}
}

func (e *specExecutor) notifyAfterScenarioHook(scenarioResult *result.ScenarioResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionEnding.Enum(),
		ScenarioExecutionEndingRequest: &gauge_messages.ScenarioExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := e.executeHook(message, scenarioResult)
	if res.GetFailed() {
		setScenarioFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPostHook, e.consoleReporter)
	}
}

func (e *specExecutor) addAllItemsForScenarioExecution(scenario *gauge.Scenario, scenarioResult *result.ScenarioResult) {
	scenarioResult.AddContexts(e.getContextItemsForScenarioExecution(e.specification.Contexts))
	scenarioResult.AddTearDownSteps(e.getContextItemsForScenarioExecution(e.specification.TearDownSteps))
	scenarioResult.AddItems(e.resolveItems(scenario.Items))
}

func (e *specExecutor) getSkippedScenarioResult(scenario *gauge.Scenario) *result.ScenarioResult {
	scenarioResult := &result.ScenarioResult{ProtoScenario: gauge.NewProtoScenario(scenario)}
	e.addAllItemsForScenarioExecution(scenario, scenarioResult)
	e.setSkipInfoInResult(scenarioResult, scenario)
	return scenarioResult
}

func (e *specExecutor) handleScenarioDataStoreFailure(scenarioResult *result.ScenarioResult, scenario *gauge.Scenario, err error) {
	e.consoleReporter.Errorf(err.Error())
	validationError := validation.NewValidationError(&gauge.Step{LineNo: scenario.Heading.LineNo, LineText: scenario.Heading.Value},
		err.Error(), e.specification.FileName, nil)
	e.errMap.ScenarioErrs[scenario] = []*validation.StepValidationError{validationError}
	e.setSkipInfoInResult(scenarioResult, scenario)
}

func (e *specExecutor) setSkipInfoInResult(result *result.ScenarioResult, scenario *gauge.Scenario) {
	e.specResult.ScenarioSkippedCount++
	result.ProtoScenario.Skipped = proto.Bool(true)
	var errors []string
	for _, err := range e.errMap.ScenarioErrs[scenario] {
		errors = append(errors, err.Error())
	}
	result.ProtoScenario.SkipErrors = errors
}
