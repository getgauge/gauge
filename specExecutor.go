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

package main

import (
	"fmt"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type specExecutor struct {
	specification        *specification
	dataTableIndex       int
	runner               *testRunner
	conceptDictionary    *conceptDictionary
	pluginHandler        *pluginHandler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	specResult           *specResult
	writer               executionLogger
}

func (specExecutor *specExecutor) initialize(specificationToExecute *specification, runner *testRunner, pluginHandler *pluginHandler, writer executionLogger) {
	specExecutor.specification = specificationToExecute
	specExecutor.runner = runner
	specExecutor.pluginHandler = pluginHandler
	specExecutor.writer = writer
}

func (e *specExecutor) executeBeforeSpecHook() *gauge_messages.ProtoExecutionResult {
	initSpecDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecDataStoreInit.Enum(),
		SpecDataStoreInitRequest: &gauge_messages.SpecDataStoreInitRequest{}}
	initResult := executeAndGetStatus(e.runner, initSpecDataStoreMessage, e.writer)
	if initResult.GetFailed() {
		e.writer.Warning("Spec data store didn't get initialized")
	}

	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionStarting.Enum(),
		SpecExecutionStartingRequest: &gauge_messages.SpecExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message, e.specResult)
}

func (e *specExecutor) executeAfterSpecHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionEnding.Enum(),
		SpecExecutionEndingRequest: &gauge_messages.SpecExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message, e.specResult)
}

func (e *specExecutor) executeHook(message *gauge_messages.Message, execTimeTracker execTimeTracker) *gauge_messages.ProtoExecutionResult {
	e.pluginHandler.notifyPlugins(message)
	executionResult := executeAndGetStatus(e.runner, message, e.writer)
	execTimeTracker.addExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func (specExecutor *specExecutor) execute() *specResult {
	specInfo := &gauge_messages.SpecInfo{Name: proto.String(specExecutor.specification.heading.value),
		FileName: proto.String(specExecutor.specification.fileName),
		IsFailed: proto.Bool(false), Tags: getTagValue(specExecutor.specification.tags)}
	specExecutor.currentExecutionInfo = &gauge_messages.ExecutionInfo{CurrentSpec: specInfo}

	specExecutor.writer.SpecHeading(specInfo.GetName())

	specExecutor.specResult = newSpecResult(specExecutor.specification)
	resolvedSpecItems := specExecutor.resolveItems(specExecutor.specification.getSpecItems())
	specExecutor.specResult.addSpecItems(resolvedSpecItems)

	beforeSpecHookStatus := specExecutor.executeBeforeSpecHook()
	if beforeSpecHookStatus.GetFailed() {
		addPreHook(specExecutor.specResult, beforeSpecHookStatus)
		setSpecFailure(specExecutor.currentExecutionInfo)
	} else {
		for _, step := range specExecutor.specification.contexts {
			specExecutor.writer.Step(step)
		}
		dataTableRowCount := specExecutor.specification.dataTable.getRowCount()
		if dataTableRowCount == 0 {
			scenarioResult := specExecutor.executeScenarios()
			specExecutor.specResult.addScenarioResults(scenarioResult)
		} else {
			specExecutor.executeTableDrivenScenarios()
		}
	}

	afterSpecHookStatus := specExecutor.executeAfterSpecHook()
	if afterSpecHookStatus.GetFailed() {
		addPostHook(specExecutor.specResult, afterSpecHookStatus)
		setSpecFailure(specExecutor.currentExecutionInfo)
	}
	return specExecutor.specResult
}

func (specExecutor *specExecutor) executeTableDrivenScenarios() {
	var dataTableScenarioExecutionResult [][]*scenarioResult
	dataTableRowCount := specExecutor.specification.dataTable.getRowCount()
	for specExecutor.dataTableIndex = 0; specExecutor.dataTableIndex < dataTableRowCount; specExecutor.dataTableIndex++ {
		dataTableScenarioExecutionResult = append(dataTableScenarioExecutionResult, specExecutor.executeScenarios())
	}
	specExecutor.specResult.addTableDrivenScenarioResult(dataTableScenarioExecutionResult)
}

func getTagValue(tags *tags) []string {
	tagValues := make([]string, 0)
	if tags != nil {
		tagValues = append(tagValues, tags.values...)
	}
	return tagValues
}

func (executor *specExecutor) executeBeforeScenarioHook(scenarioResult *scenarioResult) *gauge_messages.ProtoExecutionResult {
	initScenarioDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioDataStoreInit.Enum(),
		ScenarioDataStoreInitRequest: &gauge_messages.ScenarioDataStoreInitRequest{}}
	initResult := executeAndGetStatus(executor.runner, initScenarioDataStoreMessage, executor.writer)
	if initResult.GetFailed() {
		executor.writer.Warning("Scenario data store didn't get initialized")
	}

	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionStarting.Enum(),
		ScenarioExecutionStartingRequest: &gauge_messages.ScenarioExecutionStartingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	return executor.executeHook(message, scenarioResult)
}

func (executor *specExecutor) executeAfterScenarioHook(scenarioResult *scenarioResult) *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionEnding.Enum(),
		ScenarioExecutionEndingRequest: &gauge_messages.ScenarioExecutionEndingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	return executor.executeHook(message, scenarioResult)
}

func (specExecutor *specExecutor) executeScenarios() []*scenarioResult {
	scenarioResults := make([]*scenarioResult, 0)
	for _, scenario := range specExecutor.specification.scenarios {
		scenarioResults = append(scenarioResults, specExecutor.executeScenario(scenario))
	}
	return scenarioResults
}

func (executor *specExecutor) executeScenario(scenario *scenario) *scenarioResult {
	executor.currentExecutionInfo.CurrentScenario = &gauge_messages.ScenarioInfo{Name: proto.String(scenario.heading.value), Tags: getTagValue(scenario.tags), IsFailed: proto.Bool(false)}
	executor.writer.ScenarioHeading(scenario.heading.value)

	scenarioResult := &scenarioResult{newProtoScenario(scenario)}
	executor.addAllItemsForScenarioExecution(scenario, scenarioResult)
	beforeHookExecutionStatus := executor.executeBeforeScenarioHook(scenarioResult)
	if beforeHookExecutionStatus.GetFailed() {
		addPreHook(scenarioResult, beforeHookExecutionStatus)
		setScenarioFailure(executor.currentExecutionInfo)
	} else {
		executor.executeContextItems(scenarioResult)
		if !scenarioResult.getFailure() {
			executor.executeScenarioItems(scenarioResult)
		}
	}

	afterHookExecutionStatus := executor.executeAfterScenarioHook(scenarioResult)
	addPostHook(scenarioResult, afterHookExecutionStatus)
	scenarioResult.updateExecutionTime()
	return scenarioResult
}

func (executor *specExecutor) addAllItemsForScenarioExecution(scenario *scenario, scenarioResult *scenarioResult) {
	scenarioResult.addContexts(executor.getContextItemsForScenarioExecution(executor.specification))
	scenarioResult.addItems(executor.resolveItems(scenario.items))
}

func (executor *specExecutor) getContextItemsForScenarioExecution(specification *specification) []*gauge_messages.ProtoItem {
	contextSteps := specification.contexts
	items := make([]item, len(contextSteps))
	for i, context := range contextSteps {
		items[i] = context
	}
	contextProtoItems := executor.resolveItems(items)
	return contextProtoItems
}

func (executor *specExecutor) executeContextItems(scenarioResult *scenarioResult) {
	failure := executor.executeItems(scenarioResult.protoScenario.GetContexts())
	if failure {
		scenarioResult.setFailure()
	}
}

func (executor *specExecutor) executeScenarioItems(scenarioResult *scenarioResult) {
	failure := executor.executeItems(scenarioResult.protoScenario.GetScenarioItems())
	if failure {
		scenarioResult.setFailure()
	}
}

func (executor *specExecutor) resolveItems(items []item) []*gauge_messages.ProtoItem {
	protoItems := make([]*gauge_messages.ProtoItem, 0)
	for _, item := range items {
		protoItems = append(protoItems, executor.resolveToProtoItem(item))
	}
	return protoItems
}

func (executor *specExecutor) executeItems(executingItems []*gauge_messages.ProtoItem) bool {
	for _, protoItem := range executingItems {
		failure := executor.executeItem(protoItem)
		if failure == true {
			return true
		}
	}
	return false
}

func (executor *specExecutor) resolveToProtoItem(item item) *gauge_messages.ProtoItem {
	var protoItem *gauge_messages.ProtoItem
	switch item.kind() {
	case stepKind:
		if (item.(*step)).isConcept {
			concept := item.(*step)
			protoItem = executor.resolveToProtoConceptItem(*concept)
		} else {
			protoItem = executor.resolveToProtoStepItem(item.(*step))
		}
		break

	default:
		protoItem = convertToProtoItem(item)
	}
	return protoItem
}

func (executor *specExecutor) resolveToProtoStepItem(step *step) *gauge_messages.ProtoItem {
	protoStepItem := convertToProtoItem(step)
	paramResolver := new(paramResolver)
	parameters := paramResolver.getResolvedParams(step, nil, executor.dataTableLookup())
	updateProtoStepParameters(protoStepItem.Step, parameters)
	return protoStepItem
}

// Not passing poiter as we cannot modify the original concept step's lookup. This has to be populated for each iteration over data table.
func (executor *specExecutor) resolveToProtoConceptItem(concept step) *gauge_messages.ProtoItem {
	paramResolver := new(paramResolver)

	populateConceptDynamicParams(&concept, executor.dataTableLookup())
	protoConceptItem := convertToProtoItem(&concept)

	for stepIndex, step := range concept.conceptSteps {
		// Need to reset parent as the step.parent is pointing to a concept whose lookup is not populated yet
		if step.isConcept {
			step.parent = &concept
			protoConceptItem.GetConcept().GetSteps()[stepIndex] = executor.resolveToProtoConceptItem(*step)
		} else {
			stepParameters := paramResolver.getResolvedParams(step, &concept, executor.dataTableLookup())
			updateProtoStepParameters(protoConceptItem.Concept.Steps[stepIndex].Step, stepParameters)
		}
	}
	return protoConceptItem
}

func updateProtoStepParameters(protoStep *gauge_messages.ProtoStep, parameters []*gauge_messages.Parameter) {
	paramIndex := 0
	for fragmentIndex, fragment := range protoStep.Fragments {
		if fragment.GetFragmentType() == gauge_messages.Fragment_Parameter {
			protoStep.Fragments[fragmentIndex].Parameter = parameters[paramIndex]
			paramIndex++
		}
	}
}

func (executor *specExecutor) dataTableLookup() *argLookup {
	return new(argLookup).fromDataTableRow(&executor.specification.dataTable, executor.dataTableIndex)
}

func (executor *specExecutor) executeItem(protoItem *gauge_messages.ProtoItem) bool {
	if protoItem.GetItemType() == gauge_messages.ProtoItem_Concept {
		return executor.executeConcept(protoItem.GetConcept())
	} else if protoItem.GetItemType() == gauge_messages.ProtoItem_Step {
		return executor.executeStep(protoItem.GetStep())
	}
	return false
}

func (executor *specExecutor) executeSteps(protoSteps []*gauge_messages.ProtoStep) bool {
	for _, protoStep := range protoSteps {
		failure := executor.executeStep(protoStep)
		if failure {
			return true
		}
	}
	return false
}

func (executor *specExecutor) executeConcept(protoConcept *gauge_messages.ProtoConcept) bool {
	executor.writer.ConceptStarting(protoConcept)
	for _, step := range protoConcept.Steps {
		failure := executor.executeItem(step)
		executor.setExecutionResultForConcept(protoConcept)
		if failure {
			return true
		}
	}
	executor.writer.ConceptFinished(protoConcept)
	return protoConcept.GetConceptExecutionResult().GetExecutionResult().GetFailed()
}

func (executor *specExecutor) setExecutionResultForConcept(protoConcept *gauge_messages.ProtoConcept) {
	var conceptExecutionTime int64
	for _, step := range protoConcept.GetSteps() {
		if step.GetItemType() == gauge_messages.ProtoItem_Concept {
			stepExecResult := step.GetConcept().GetConceptExecutionResult().GetExecutionResult()
			conceptExecutionTime += stepExecResult.GetExecutionTime()
			if step.GetConcept().GetConceptExecutionResult().GetExecutionResult().GetFailed() {
				conceptExecutionResult := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: step.GetConcept().GetConceptExecutionResult().GetExecutionResult()}
				conceptExecutionResult.ExecutionResult.ExecutionTime = proto.Int64(conceptExecutionTime)
				protoConcept.ConceptExecutionResult = conceptExecutionResult
				protoConcept.ConceptStep.StepExecutionResult = conceptExecutionResult
				return
			}
		} else if step.GetItemType() == gauge_messages.ProtoItem_Step {
			stepExecResult := step.GetStep().GetStepExecutionResult().GetExecutionResult()
			conceptExecutionTime += stepExecResult.GetExecutionTime()
			if stepExecResult.GetFailed() {
				conceptExecutionResult := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: stepExecResult}
				conceptExecutionResult.ExecutionResult.ExecutionTime = proto.Int64(conceptExecutionTime)
				protoConcept.ConceptExecutionResult = conceptExecutionResult
				protoConcept.ConceptStep.StepExecutionResult = conceptExecutionResult
				return
			}
		}
	}
	protoConcept.ConceptExecutionResult = &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(false), ExecutionTime: proto.Int64(conceptExecutionTime)}}
	protoConcept.ConceptStep.StepExecutionResult = protoConcept.ConceptExecutionResult
}

func printStatus(executionResult *gauge_messages.ProtoExecutionResult, writer executionLogger) {
	writer.PrintError(executionResult.GetErrorMessage())
	writer.PrintError(executionResult.GetStackTrace())
}

func (executor *specExecutor) executeStep(protoStep *gauge_messages.ProtoStep) bool {
	stepRequest := executor.createStepRequest(protoStep)
	stepWithResolvedArgs := createStepFromStepRequest(stepRequest)
	executor.writer.StepStarting(stepWithResolvedArgs)

	protoStepExecResult := &gauge_messages.ProtoStepExecutionResult{}
	executor.currentExecutionInfo.CurrentStep = &gauge_messages.StepInfo{Step: stepRequest, IsFailed: proto.Bool(false)}

	beforeHookStatus := executor.executeBeforeStepHook()
	if beforeHookStatus.GetFailed() {
		protoStepExecResult.PreHookFailure = getProtoHookFailure(beforeHookStatus)
		protoStepExecResult.ExecutionResult = &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(true)}
		setStepFailure(executor.currentExecutionInfo)
		printStatus(beforeHookStatus, executor.writer)
	} else {
		executeStepMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecuteStep.Enum(), ExecuteStepRequest: stepRequest}
		stepExecutionStatus := executeAndGetStatus(executor.runner, executeStepMessage, executor.writer)
		if stepExecutionStatus.GetFailed() {
			setStepFailure(executor.currentExecutionInfo)
			printStatus(stepExecutionStatus, executor.writer)
		}
		protoStepExecResult.ExecutionResult = stepExecutionStatus
	}
	afterStepHookStatus := executor.executeAfterStepHook()
	addExecutionTimes(protoStepExecResult, beforeHookStatus, afterStepHookStatus)
	if afterStepHookStatus.GetFailed() {
		setStepFailure(executor.currentExecutionInfo)
		printStatus(afterStepHookStatus, executor.writer)
		protoStepExecResult.PostHookFailure = getProtoHookFailure(afterStepHookStatus)
		protoStepExecResult.ExecutionResult.Failed = proto.Bool(true)
	}

	executor.writer.StepFinished(stepWithResolvedArgs, protoStepExecResult.GetExecutionResult().GetFailed())
	protoStep.StepExecutionResult = protoStepExecResult
	return protoStep.GetStepExecutionResult().GetExecutionResult().GetFailed()
}

func addExecutionTimes(stepExecResult *gauge_messages.ProtoStepExecutionResult, execResults ...*gauge_messages.ProtoExecutionResult) {
	for _, execResult := range execResults {
		currentScenarioExecTime := stepExecResult.ExecutionResult.ExecutionTime
		if currentScenarioExecTime == nil {
			stepExecResult.ExecutionResult.ExecutionTime = proto.Int64(execResult.GetExecutionTime())
		} else {
			stepExecResult.ExecutionResult.ExecutionTime = proto.Int64(*currentScenarioExecTime + execResult.GetExecutionTime())
		}
	}
}

func (executor *specExecutor) executeBeforeStepHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepExecutionStarting.Enum(),
		StepExecutionStartingRequest: &gauge_messages.StepExecutionStartingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	executor.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(executor.runner, message, executor.writer)
}

func (executor *specExecutor) executeAfterStepHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepExecutionEnding.Enum(),
		StepExecutionEndingRequest: &gauge_messages.StepExecutionEndingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	executor.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(executor.runner, message, executor.writer)
}

func (executor *specExecutor) createStepRequest(protoStep *gauge_messages.ProtoStep) *gauge_messages.ExecuteStepRequest {
	stepRequest := &gauge_messages.ExecuteStepRequest{ParsedStepText: proto.String(protoStep.GetParsedText()), ActualStepText: proto.String(protoStep.GetActualText())}
	stepRequest.Parameters = getParameters(protoStep.GetFragments())
	return stepRequest
}

func (executor *specExecutor) getCurrentDataTableValueFor(columnName string) string {
	return executor.specification.dataTable.get(columnName)[executor.dataTableIndex].value
}

func executeAndGetStatus(runner *testRunner, message *gauge_messages.Message, writer executionLogger) *gauge_messages.ProtoExecutionResult {
	response, err := getResponseForGaugeMessage(message, runner.connection)
	if err != nil {
		return &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(true), ErrorMessage: proto.String(err.Error())}
	}

	if response.GetMessageType() == gauge_messages.Message_ExecutionStatusResponse {
		executionResult := response.GetExecutionStatusResponse().GetExecutionResult()
		if executionResult == nil {
			errMsg := "ProtoExecutionResult obtained is nil"
			writer.Critical(errMsg)
			return errorResult(errMsg)
		}
		return executionResult
	} else {
		errMsg := fmt.Sprintf("Expected ExecutionStatusResponse. Obtained: %s", response.GetMessageType())
		writer.Critical(errMsg)
		return errorResult(errMsg)
	}
}

func errorResult(message string) *gauge_messages.ProtoExecutionResult {
	return &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(true), ErrorMessage: proto.String(message), RecoverableError: proto.Bool(false)}
}

// Creating a copy of the lookup and populating table values
func populateConceptDynamicParams(concept *step, dataTableLookup *argLookup) {
	//If it is a top level concept
	if concept.parent == nil {
		lookup := concept.lookup.getCopy()
		for key, _ := range lookup.paramIndexMap {
			conceptLookupArg := lookup.getArg(key)
			if conceptLookupArg.argType == dynamic {
				resolvedArg := dataTableLookup.getArg(conceptLookupArg.value)
				lookup.addArgValue(key, resolvedArg)
			}
		}
		concept.lookup = *lookup
	}

	//Updating values inside the concept step as well
	newArgs := make([]*stepArg, 0)
	for _, arg := range concept.args {
		if arg.argType == dynamic {
			if concept.parent != nil {
				newArgs = append(newArgs, concept.parent.getArg(arg.value))
			} else {
				newArgs = append(newArgs, dataTableLookup.getArg(arg.value))
			}
		} else {
			newArgs = append(newArgs, arg)
		}
	}
	concept.args = newArgs
	concept.populateFragments()
}

func setSpecFailure(executionInfo *gauge_messages.ExecutionInfo) {
	executionInfo.CurrentSpec.IsFailed = proto.Bool(true)
}

func setScenarioFailure(executionInfo *gauge_messages.ExecutionInfo) {
	setSpecFailure(executionInfo)
	executionInfo.CurrentScenario.IsFailed = proto.Bool(true)
}

func setStepFailure(executionInfo *gauge_messages.ExecutionInfo) {
	setScenarioFailure(executionInfo)
	executionInfo.CurrentStep.IsFailed = proto.Bool(true)
}

func getParameters(fragments []*gauge_messages.Fragment) []*gauge_messages.Parameter {
	parameters := make([]*gauge_messages.Parameter, 0)
	for _, fragment := range fragments {
		if fragment.GetFragmentType() == gauge_messages.Fragment_Parameter {
			parameters = append(parameters, fragment.GetParameter())
		}
	}
	return parameters
}
