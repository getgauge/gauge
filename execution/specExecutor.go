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
	"errors"
	"fmt"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger/execLogger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
	"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
)

type specExecutor struct {
	specification        *parser.Specification
	dataTableIndex       indexRange
	runner               *runner.TestRunner
	conceptDictionary    *parser.ConceptDictionary
	pluginHandler        *plugin.PluginHandler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	specResult           *result.SpecResult
	writer               execLogger.ExecutionLogger
	currentTableRow      int
}

type indexRange struct {
	start int
	end   int
}

func (specExecutor *specExecutor) initialize(specificationToExecute *parser.Specification, runner *runner.TestRunner, pluginHandler *plugin.PluginHandler, writer execLogger.ExecutionLogger, tableRows indexRange) {
	specExecutor.specification = specificationToExecute
	specExecutor.runner = runner
	specExecutor.pluginHandler = pluginHandler
	specExecutor.writer = writer
	specExecutor.dataTableIndex = tableRows
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

func (e *specExecutor) executeHook(message *gauge_messages.Message, execTimeTracker result.ExecTimeTracker) *gauge_messages.ProtoExecutionResult {
	e.pluginHandler.NotifyPlugins(message)
	executionResult := executeAndGetStatus(e.runner, message, e.writer)
	execTimeTracker.AddExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func (specExecutor *specExecutor) execute() *result.SpecResult {
	specInfo := &gauge_messages.SpecInfo{Name: proto.String(specExecutor.specification.Heading.Value),
		FileName: proto.String(specExecutor.specification.FileName),
		IsFailed: proto.Bool(false), Tags: getTagValue(specExecutor.specification.Tags)}
	specExecutor.currentExecutionInfo = &gauge_messages.ExecutionInfo{CurrentSpec: specInfo}

	specExecutor.writer.SpecHeading(specInfo.GetName())

	specExecutor.specResult = parser.NewSpecResult(specExecutor.specification)
	resolvedSpecItems := specExecutor.resolveItems(specExecutor.specification.Items)
	specExecutor.specResult.AddSpecItems(resolvedSpecItems)

	beforeSpecHookStatus := specExecutor.executeBeforeSpecHook()
	if beforeSpecHookStatus.GetFailed() {
		result.AddPreHook(specExecutor.specResult, beforeSpecHookStatus)
		setSpecFailure(specExecutor.currentExecutionInfo)
	} else {
		for _, step := range specExecutor.specification.Contexts {
			specExecutor.writer.Step(step)
		}
		dataTableRowCount := specExecutor.specification.DataTable.Table.GetRowCount()
		if dataTableRowCount == 0 {
			scenarioResult := specExecutor.executeScenarios()
			specExecutor.specResult.AddScenarioResults(scenarioResult)
		} else {
			specExecutor.executeTableDrivenScenarios()
		}
	}

	afterSpecHookStatus := specExecutor.executeAfterSpecHook()
	if afterSpecHookStatus.GetFailed() {
		result.AddPostHook(specExecutor.specResult, afterSpecHookStatus)
		setSpecFailure(specExecutor.currentExecutionInfo)
	}
	return specExecutor.specResult
}

func (specExecutor *specExecutor) executeTableDrivenScenarios() {
	var dataTableScenarioExecutionResult [][]*result.ScenarioResult
	//	dataTableRowCount := specExecutor.specification.DataTable.table.GetRowCount()
	for specExecutor.currentTableRow = specExecutor.dataTableIndex.start; specExecutor.currentTableRow <= specExecutor.dataTableIndex.end; specExecutor.currentTableRow++ {
		dataTableScenarioExecutionResult = append(dataTableScenarioExecutionResult, specExecutor.executeScenarios())
	}
	specExecutor.specResult.AddTableDrivenScenarioResult(dataTableScenarioExecutionResult)
}

func getTagValue(tags *parser.Tags) []string {
	tagValues := make([]string, 0)
	if tags != nil {
		tagValues = append(tagValues, tags.Values...)
	}
	return tagValues
}

func (executor *specExecutor) executeBeforeScenarioHook(scenarioResult *result.ScenarioResult) *gauge_messages.ProtoExecutionResult {
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

func (executor *specExecutor) executeAfterScenarioHook(scenarioResult *result.ScenarioResult) *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionEnding.Enum(),
		ScenarioExecutionEndingRequest: &gauge_messages.ScenarioExecutionEndingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	return executor.executeHook(message, scenarioResult)
}

func (specExecutor *specExecutor) executeScenarios() []*result.ScenarioResult {
	scenarioResults := make([]*result.ScenarioResult, 0)
	for _, scenario := range specExecutor.specification.Scenarios {
		scenarioResults = append(scenarioResults, specExecutor.executeScenario(scenario))
	}
	return scenarioResults
}

func (executor *specExecutor) executeScenario(scenario *parser.Scenario) *result.ScenarioResult {
	executor.currentExecutionInfo.CurrentScenario = &gauge_messages.ScenarioInfo{Name: proto.String(scenario.Heading.Value), Tags: getTagValue(scenario.Tags), IsFailed: proto.Bool(false)}
	executor.writer.ScenarioHeading(scenario.Heading.Value)

	scenarioResult := &result.ScenarioResult{parser.NewProtoScenario(scenario)}
	executor.addAllItemsForScenarioExecution(scenario, scenarioResult)
	beforeHookExecutionStatus := executor.executeBeforeScenarioHook(scenarioResult)
	if beforeHookExecutionStatus.GetFailed() {
		result.AddPreHook(scenarioResult, beforeHookExecutionStatus)
		setScenarioFailure(executor.currentExecutionInfo)
	} else {
		executor.executeContextItems(scenarioResult)
		if !scenarioResult.GetFailure() {
			executor.executeScenarioItems(scenarioResult)
		}
	}

	afterHookExecutionStatus := executor.executeAfterScenarioHook(scenarioResult)
	result.AddPostHook(scenarioResult, afterHookExecutionStatus)
	scenarioResult.UpdateExecutionTime()
	return scenarioResult
}

func (executor *specExecutor) addAllItemsForScenarioExecution(scenario *parser.Scenario, scenarioResult *result.ScenarioResult) {
	scenarioResult.AddContexts(executor.getContextItemsForScenarioExecution(executor.specification))
	scenarioResult.AddItems(executor.resolveItems(scenario.Items))
}

func (executor *specExecutor) getContextItemsForScenarioExecution(specification *parser.Specification) []*gauge_messages.ProtoItem {
	contextSteps := specification.Contexts
	items := make([]parser.Item, len(contextSteps))
	for i, context := range contextSteps {
		items[i] = context
	}
	contextProtoItems := executor.resolveItems(items)
	return contextProtoItems
}

func (executor *specExecutor) executeContextItems(scenarioResult *result.ScenarioResult) {
	failure := executor.executeItems(scenarioResult.ProtoScenario.GetContexts())
	if failure {
		scenarioResult.SetFailure()
	}
}

func (executor *specExecutor) executeScenarioItems(scenarioResult *result.ScenarioResult) {
	failure := executor.executeItems(scenarioResult.ProtoScenario.GetScenarioItems())
	if failure {
		scenarioResult.SetFailure()
	}
}

func (executor *specExecutor) resolveItems(items []parser.Item) []*gauge_messages.ProtoItem {
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

func (executor *specExecutor) resolveToProtoItem(item parser.Item) *gauge_messages.ProtoItem {
	var protoItem *gauge_messages.ProtoItem
	switch item.Kind() {
	case parser.StepKind:
		if (item.(*parser.Step)).IsConcept {
			concept := item.(*parser.Step)
			protoItem = executor.resolveToProtoConceptItem(*concept)
		} else {
			protoItem = executor.resolveToProtoStepItem(item.(*parser.Step))
		}
		break

	default:
		protoItem = parser.ConvertToProtoItem(item)
	}
	return protoItem
}

func (executor *specExecutor) resolveToProtoStepItem(step *parser.Step) *gauge_messages.ProtoItem {
	protoStepItem := parser.ConvertToProtoItem(step)
	paramResolver := new(parser.ParamResolver)
	parameters := paramResolver.GetResolvedParams(step, nil, executor.dataTableLookup())
	updateProtoStepParameters(protoStepItem.Step, parameters)
	return protoStepItem
}

// Not passing pointer as we cannot modify the original concept step's lookup. This has to be populated for each iteration over data table.
func (executor *specExecutor) resolveToProtoConceptItem(concept parser.Step) *gauge_messages.ProtoItem {
	paramResolver := new(parser.ParamResolver)

	parser.PopulateConceptDynamicParams(&concept, executor.dataTableLookup())
	protoConceptItem := parser.ConvertToProtoItem(&concept)

	for stepIndex, step := range concept.ConceptSteps {
		// Need to reset parent as the step.parent is pointing to a concept whose lookup is not populated yet
		if step.IsConcept {
			step.Parent = &concept
			protoConceptItem.GetConcept().GetSteps()[stepIndex] = executor.resolveToProtoConceptItem(*step)
		} else {
			stepParameters := paramResolver.GetResolvedParams(step, &concept, executor.dataTableLookup())
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

func (executor *specExecutor) dataTableLookup() *parser.ArgLookup {
	return new(parser.ArgLookup).FromDataTableRow(&executor.specification.DataTable.Table, executor.currentTableRow)
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

func printStatus(executionResult *gauge_messages.ProtoExecutionResult, writer execLogger.ExecutionLogger) {
	writer.PrintError(executionResult.GetErrorMessage())
	writer.PrintError(executionResult.GetStackTrace())
}

func (executor *specExecutor) executeStep(protoStep *gauge_messages.ProtoStep) bool {
	stepRequest := executor.createStepRequest(protoStep)
	stepWithResolvedArgs := parser.CreateStepFromStepRequest(stepRequest)
	executor.writer.StepStarting(stepWithResolvedArgs)

	protoStepExecResult := &gauge_messages.ProtoStepExecutionResult{}
	executor.currentExecutionInfo.CurrentStep = &gauge_messages.StepInfo{Step: stepRequest, IsFailed: proto.Bool(false)}

	beforeHookStatus := executor.executeBeforeStepHook()
	if beforeHookStatus.GetFailed() {
		protoStepExecResult.PreHookFailure = result.GetProtoHookFailure(beforeHookStatus)
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
		protoStepExecResult.PostHookFailure = result.GetProtoHookFailure(afterStepHookStatus)
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
	executor.pluginHandler.NotifyPlugins(message)
	return executeAndGetStatus(executor.runner, message, executor.writer)
}

func (executor *specExecutor) executeAfterStepHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepExecutionEnding.Enum(),
		StepExecutionEndingRequest: &gauge_messages.StepExecutionEndingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	executor.pluginHandler.NotifyPlugins(message)
	return executeAndGetStatus(executor.runner, message, executor.writer)
}

func (executor *specExecutor) createStepRequest(protoStep *gauge_messages.ProtoStep) *gauge_messages.ExecuteStepRequest {
	stepRequest := &gauge_messages.ExecuteStepRequest{ParsedStepText: proto.String(protoStep.GetParsedText()), ActualStepText: proto.String(protoStep.GetActualText())}
	stepRequest.Parameters = getParameters(protoStep.GetFragments())
	return stepRequest
}

func (executor *specExecutor) getCurrentDataTableValueFor(columnName string) string {
	return executor.specification.DataTable.Table.Get(columnName)[executor.currentTableRow].Value
}

func executeAndGetStatus(runner *runner.TestRunner, message *gauge_messages.Message, writer execLogger.ExecutionLogger) *gauge_messages.ProtoExecutionResult {
	response, err := conn.GetResponseForGaugeMessage(message, runner.Connection)
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

func getDataTableRowsRange(tableRows string, rowCount int) (indexRange, error) {
	var startIndex, endIndex int
	var err error
	indexRanges := strings.Split(tableRows, "-")
	if len(indexRanges) == 2 {
		startIndex, endIndex, err = validateTableRowsRange(indexRanges[0], indexRanges[1], rowCount)
	} else if len(indexRanges) == 1 {
		startIndex, endIndex, err = validateTableRowsRange(tableRows, tableRows, rowCount)
	} else {
		return indexRange{start: 0, end: 0}, errors.New("Table rows range validation failed.")
	}
	if err != nil {
		return indexRange{start: 0, end: 0}, err
	}
	return indexRange{start: startIndex, end: endIndex}, nil
}

func validateTableRowsRange(start string, end string, rowCount int) (int, int, error) {
	message := "Table rows range validation failed."
	startRow, err := strconv.Atoi(start)
	if err != nil {
		return 0, 0, errors.New(message)
	}
	endRow, err := strconv.Atoi(end)
	if err != nil {
		return 0, 0, errors.New(message)
	}
	if startRow > endRow || endRow > rowCount || startRow < 1 || endRow < 1 {
		return 0, 0, errors.New(message)
	}
	return startRow - 1, endRow - 1, nil
}
