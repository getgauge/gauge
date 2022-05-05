/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"fmt"

	"errors"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/validation"
)

type scenarioExecutor struct {
	runner               runner.Runner
	pluginHandler        plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	errMap               *gauge.BuildErrors
	stream               int
	contexts             []*gauge.Step
	teardowns            []*gauge.Step
}

func newScenarioExecutor(r runner.Runner, ph plugin.Handler, ei *gauge_messages.ExecutionInfo, errMap *gauge.BuildErrors, contexts []*gauge.Step, teardowns []*gauge.Step, stream int) *scenarioExecutor {
	return &scenarioExecutor{
		runner:               r,
		pluginHandler:        ph,
		currentExecutionInfo: ei,
		errMap:               errMap,
		stream:               stream,
		contexts:             contexts,
		teardowns:            teardowns,
	}
}

func (e *scenarioExecutor) execute(i gauge.Item, r result.Result) {
	scenario := i.(*gauge.Scenario)
	scenarioResult := r.(*result.ScenarioResult)
	scenarioResult.ProtoScenario.ExecutionStatus = gauge_messages.ExecutionStatus_PASSED
	if e.runner.Info().Killed {
		e.errMap.ScenarioErrs[scenario] = append([]error{errors.New("skipped Reason: Runner is not alive")}, e.errMap.ScenarioErrs[scenario]...)
		setSkipInfoInResult(scenarioResult, scenario, e.errMap)
		return
	}
	if scenario.SpecDataTableRow.IsInitialized() && !shouldExecuteForRow(scenario.SpecDataTableRowIndex) {
		e.errMap.ScenarioErrs[scenario] = append([]error{errors.New("skipped Reason: Doesn't satisfy --table-rows flag condition")}, e.errMap.ScenarioErrs[scenario]...)
		setSkipInfoInResult(scenarioResult, scenario, e.errMap)
		return
	}
	if _, ok := e.errMap.ScenarioErrs[scenario]; ok {
		setSkipInfoInResult(scenarioResult, scenario, e.errMap)
		event.Notify(event.NewExecutionEvent(event.ScenarioStart, scenario, scenarioResult, e.stream, e.currentExecutionInfo))
		event.Notify(event.NewExecutionEvent(event.ScenarioEnd, scenario, scenarioResult, e.stream, e.currentExecutionInfo))
		return
	}
	event.Notify(event.NewExecutionEvent(event.ScenarioStart, scenario, scenarioResult, e.stream, e.currentExecutionInfo))
	defer event.Notify(event.NewExecutionEvent(event.ScenarioEnd, scenario, scenarioResult, e.stream, e.currentExecutionInfo))

	res := e.initScenarioDataStore()
	if res.GetFailed() {
		e.handleScenarioDataStoreFailure(scenarioResult, scenario, fmt.Errorf("Failed to initialize scenario datastore. Error: %s", res.GetErrorMessage()))
		return
	}
	e.notifyBeforeScenarioHook(scenarioResult)

	if !scenarioResult.GetFailed() {
		protoContexts := scenarioResult.ProtoScenario.GetContexts()
		protoScenItems := scenarioResult.ProtoScenario.GetScenarioItems()
		// context and steps are not appended together since sometime it cause the issue and the steps in step list and proto step list differs.
		// This is done to fix https://github.com/getgauge/gauge/issues/1629
		if e.executeSteps(e.contexts, protoContexts, scenarioResult) {
			e.executeSteps(scenario.Steps, protoScenItems, scenarioResult)
		}
		// teardowns are not appended to previous call to executeSteps to ensure they are run irrespective of context/step failure
		e.executeSteps(e.teardowns, scenarioResult.ProtoScenario.GetTearDownSteps(), scenarioResult)
	}

	e.notifyAfterScenarioHook(scenarioResult)
	scenarioResult.UpdateExecutionTime()
}

func (e *scenarioExecutor) initScenarioDataStore() *gauge_messages.ProtoExecutionResult {
	msg := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioDataStoreInit,
		ScenarioDataStoreInitRequest: &gauge_messages.ScenarioDataStoreInitRequest{Stream: int32(e.stream)}}
	return e.runner.ExecuteAndGetStatus(msg)
}

func (e *scenarioExecutor) handleScenarioDataStoreFailure(scenarioResult *result.ScenarioResult, scenario *gauge.Scenario, err error) {
	logger.Errorf(true, err.Error())
	validationError := validation.NewStepValidationError(&gauge.Step{LineNo: scenario.Heading.LineNo, LineText: scenario.Heading.Value},
		err.Error(), e.currentExecutionInfo.CurrentSpec.GetFileName(), nil, "")
	e.errMap.ScenarioErrs[scenario] = []error{validationError}
	setSkipInfoInResult(scenarioResult, scenario, e.errMap)
}

func setSkipInfoInResult(scenarioResult *result.ScenarioResult, scenario *gauge.Scenario, errMap *gauge.BuildErrors) {
	scenarioResult.ProtoScenario.ExecutionStatus = gauge_messages.ExecutionStatus_SKIPPED
	var errs []string
	for _, err := range errMap.ScenarioErrs[scenario] {
		errs = append(errs, err.Error())
	}
	scenarioResult.ProtoScenario.SkipErrors = errs
}

func (e *scenarioExecutor) notifyBeforeScenarioHook(scenarioResult *result.ScenarioResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionStarting,
		ScenarioExecutionStartingRequest: &gauge_messages.ScenarioExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo, Stream: int32(e.stream)}}
	e.pluginHandler.NotifyPlugins(message)
	res := executeHook(message, scenarioResult, e.runner)
	scenarioResult.ProtoScenario.PreHookMessages = res.Message
	scenarioResult.ProtoScenario.PreHookScreenshotFiles = res.ScreenshotFiles
	if res.GetFailed() {
		setScenarioFailure(e.currentExecutionInfo)
		handleHookFailure(scenarioResult, res, result.AddPreHook)
	}
	message.ScenarioExecutionStartingRequest.ScenarioResult = gauge.ConvertToProtoScenarioResult(scenarioResult)
	e.pluginHandler.NotifyPlugins(message)
}

func (e *scenarioExecutor) notifyAfterScenarioHook(scenarioResult *result.ScenarioResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionEnding,
		ScenarioExecutionEndingRequest: &gauge_messages.ScenarioExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo, Stream: int32(e.stream)}}
	res := executeHook(message, scenarioResult, e.runner)
	scenarioResult.ProtoScenario.PostHookMessages = res.Message
	scenarioResult.ProtoScenario.PostHookScreenshotFiles = res.ScreenshotFiles
	if res.GetFailed() {
		setScenarioFailure(e.currentExecutionInfo)
		handleHookFailure(scenarioResult, res, result.AddPostHook)
	}
	message.ScenarioExecutionEndingRequest.ScenarioResult = gauge.ConvertToProtoScenarioResult(scenarioResult)
	e.pluginHandler.NotifyPlugins(message)
}

func (e *scenarioExecutor) executeSteps(steps []*gauge.Step, protoItems []*gauge_messages.ProtoItem, scenarioResult *result.ScenarioResult) bool {
	var stepsIndex int
	for _, protoItem := range protoItems {
		if protoItem.GetItemType() == gauge_messages.ProtoItem_Concept || protoItem.GetItemType() == gauge_messages.ProtoItem_Step {
			failed, recoverable := e.executeStep(steps[stepsIndex], protoItem, scenarioResult)
			stepsIndex++
			if failed {
				scenarioResult.SetFailure()
				if !recoverable {
					return false
				}
			}
		}
	}
	return true
}

func (e *scenarioExecutor) executeStep(step *gauge.Step, protoItem *gauge_messages.ProtoItem, scenarioResult *result.ScenarioResult) (bool, bool) {
	var failed, recoverable bool
	if protoItem.GetItemType() == gauge_messages.ProtoItem_Concept {
		protoConcept := protoItem.GetConcept()
		res := e.executeConcept(step, protoConcept, scenarioResult)
		failed = res.GetFailed()
		recoverable = res.GetRecoverable()

	} else if protoItem.GetItemType() == gauge_messages.ProtoItem_Step {
		se := &stepExecutor{runner: e.runner, pluginHandler: e.pluginHandler, currentExecutionInfo: e.currentExecutionInfo, stream: e.stream}
		res := se.executeStep(step, protoItem.GetStep())
		protoItem.GetStep().StepExecutionResult = res.ProtoStepExecResult()
		failed = res.GetFailed()
		recoverable = res.ProtoStepExecResult().GetExecutionResult().GetRecoverableError()
	}
	return failed, recoverable
}

func (e *scenarioExecutor) executeConcept(item *gauge.Step, protoConcept *gauge_messages.ProtoConcept, scenarioResult *result.ScenarioResult) *result.ConceptResult {
	cptResult := result.NewConceptResult(protoConcept)
	event.Notify(event.NewExecutionEvent(event.ConceptStart, item, nil, e.stream, e.currentExecutionInfo))
	defer event.Notify(event.NewExecutionEvent(event.ConceptEnd, nil, cptResult, e.stream, e.currentExecutionInfo))

	var conceptStepIndex int
	for _, protoStep := range protoConcept.Steps {
		if protoStep.GetItemType() == gauge_messages.ProtoItem_Concept || protoStep.GetItemType() == gauge_messages.ProtoItem_Step {
			failed, recoverable := e.executeStep(item.ConceptSteps[conceptStepIndex], protoStep, scenarioResult)
			conceptStepIndex++
			if failed {
				scenarioResult.SetFailure()
				cptResult.UpdateConceptExecResult()
				if recoverable {
					continue
				}
				return cptResult
			}
		}
	}
	cptResult.UpdateConceptExecResult()
	return cptResult
}

func setStepFailure(executionInfo *gauge_messages.ExecutionInfo) {
	setScenarioFailure(executionInfo)
	executionInfo.CurrentStep.IsFailed = true
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
	executionInfo.CurrentScenario.IsFailed = true
}
