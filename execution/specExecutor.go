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
	"strings"

	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
	"github.com/golang/protobuf/proto"
)

type specExecutor struct {
	specification        *gauge.Specification
	dataTableIndex       indexRange
	runner               *runner.TestRunner
	pluginHandler        *plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	specResult           *result.SpecResult
	currentTableRow      int
	consoleReporter      reporter.Reporter
	errMap               *validationErrMaps
}

type indexRange struct {
	start int
	end   int
}

func newSpecExecutor(specToExecute *gauge.Specification, runner *runner.TestRunner, pluginHandler *plugin.Handler, tableRows indexRange, reporter reporter.Reporter, errMaps *validationErrMaps) *specExecutor {
	specExecutor := new(specExecutor)
	specExecutor.initialize(specToExecute, runner, pluginHandler, tableRows, reporter, errMaps)
	return specExecutor
}

func (e *specExecutor) initialize(specificationToExecute *gauge.Specification, runner *runner.TestRunner, pluginHandler *plugin.Handler, tableRows indexRange, consoleReporter reporter.Reporter, errMap *validationErrMaps) {
	e.specification = specificationToExecute
	e.runner = runner
	e.pluginHandler = pluginHandler
	e.dataTableIndex = tableRows
	e.consoleReporter = consoleReporter
	e.errMap = errMap
}

func (e *specExecutor) executeBeforeSpecHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionStarting.Enum(),
		SpecExecutionStartingRequest: &gauge_messages.SpecExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message, e.specResult)
}

func (e *specExecutor) initSpecDataStore() error {
	initSpecDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecDataStoreInit.Enum(),
		SpecDataStoreInitRequest: &gauge_messages.SpecDataStoreInitRequest{}}
	initResult := executeAndGetStatus(e.runner, initSpecDataStoreMessage)
	if initResult.GetFailed() {
		return fmt.Errorf("Spec data store didn't get initialized : %s\n", initResult.GetErrorMessage())
	}
	return nil
}

func (e *specExecutor) executeAfterSpecHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionEnding.Enum(),
		SpecExecutionEndingRequest: &gauge_messages.SpecExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message, e.specResult)
}

func (e *specExecutor) executeHook(message *gauge_messages.Message, execTimeTracker result.ExecTimeTracker) *gauge_messages.ProtoExecutionResult {
	e.pluginHandler.NotifyPlugins(message)
	executionResult := executeAndGetStatus(e.runner, message)
	execTimeTracker.AddExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func (e *specExecutor) getSkippedSpecResult() *result.SpecResult {
	var scenarioResults []*result.ScenarioResult
	for _, scenario := range e.specification.Scenarios {
		scenarioResults = append(scenarioResults, e.getSkippedScenarioResult(scenario))
	}
	e.specResult.AddScenarioResults(scenarioResults)
	e.specResult.Skipped = true
	return e.specResult
}

func (e *specExecutor) getSkippedScenarioResult(scenario *gauge.Scenario) *result.ScenarioResult {
	scenarioResult := &result.ScenarioResult{ProtoScenario: gauge.NewProtoScenario(scenario)}
	e.addAllItemsForScenarioExecution(scenario, scenarioResult)
	e.setSkipInfoInResult(scenarioResult, scenario)
	return scenarioResult
}

func (e *specExecutor) execute() *result.SpecResult {
	specInfo := &gauge_messages.SpecInfo{Name: proto.String(e.specification.Heading.Value),
		FileName: proto.String(e.specification.FileName),
		IsFailed: proto.Bool(false), Tags: getTagValue(e.specification.Tags)}
	e.currentExecutionInfo = &gauge_messages.ExecutionInfo{CurrentSpec: specInfo}
	e.specResult = gauge.NewSpecResult(e.specification)
	resolvedSpecItems := e.resolveItems(e.specification.GetSpecItems())
	e.specResult.AddSpecItems(resolvedSpecItems)
	if _, ok := e.errMap.specErrs[e.specification]; ok {
		return e.getSkippedSpecResult()
	}
	err := e.initSpecDataStore()
	if err != nil {
		return e.handleSpecDataStoreFailure(err)
	}
	e.consoleReporter.SpecStart(specInfo.GetName())
	beforeSpecHookStatus := e.executeBeforeSpecHook()
	if beforeSpecHookStatus.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, beforeSpecHookStatus, result.AddPreHook, e.consoleReporter)
	} else {
		dataTableRowCount := e.specification.DataTable.Table.GetRowCount()
		if dataTableRowCount == 0 {
			scenarioResult := e.executeScenarios()
			e.specResult.AddScenarioResults(scenarioResult)
		} else {
			e.executeTableDrivenSpec()
		}
	}

	afterSpecHookStatus := e.executeAfterSpecHook()
	if afterSpecHookStatus.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, afterSpecHookStatus, result.AddPostHook, e.consoleReporter)
	}
	e.specResult.Skipped = e.specResult.ScenarioSkippedCount > 0
	e.consoleReporter.SpecEnd()
	return e.specResult
}

func (e *specExecutor) handleSpecDataStoreFailure(err error) *result.SpecResult {
	e.consoleReporter.Error(err.Error())
	validationErrors := []*stepValidationError{
		&stepValidationError{
			step:     &gauge.Step{LineNo: e.specification.Heading.LineNo, LineText: e.specification.Heading.Value},
			message:  err.Error(),
			fileName: e.specification.FileName,
		},
	}
	for _, scenario := range e.specification.Scenarios {
		e.errMap.scenarioErrs[scenario] = validationErrors
	}
	e.errMap.specErrs[e.specification] = validationErrors
	return e.getSkippedSpecResult()
}

func (e *specExecutor) executeTableDrivenSpec() {
	var dataTableScenarioExecutionResult [][]*result.ScenarioResult
	for e.currentTableRow = e.dataTableIndex.start; e.currentTableRow <= e.dataTableIndex.end; e.currentTableRow++ {
		var dataTable gauge.Table
		dataTable.AddHeaders(e.specification.DataTable.Table.Headers)
		dataTable.AddRowValues(e.specification.DataTable.Table.Rows()[e.currentTableRow])
		e.consoleReporter.DataTable(formatter.FormatTable(&dataTable))
		dataTableScenarioExecutionResult = append(dataTableScenarioExecutionResult, e.executeScenarios())
	}
	e.specResult.AddTableDrivenScenarioResult(dataTableScenarioExecutionResult)
}

func getTagValue(tags *gauge.Tags) []string {
	var tagValues []string
	if tags != nil {
		tagValues = append(tagValues, tags.Values...)
	}
	return tagValues
}

func (e *specExecutor) executeBeforeScenarioHook(scenarioResult *result.ScenarioResult) *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionStarting.Enum(),
		ScenarioExecutionStartingRequest: &gauge_messages.ScenarioExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message, scenarioResult)
}

func (e *specExecutor) initScenarioDataStore() error {
	initScenarioDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioDataStoreInit.Enum(),
		ScenarioDataStoreInitRequest: &gauge_messages.ScenarioDataStoreInitRequest{}}
	initResult := executeAndGetStatus(e.runner, initScenarioDataStoreMessage)
	if initResult.GetFailed() {
		return fmt.Errorf("Scenario data store didn't get initialized : %s\n", initResult.GetErrorMessage())
	}
	return nil
}

func (e *specExecutor) executeAfterScenarioHook(scenarioResult *result.ScenarioResult) *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ScenarioExecutionEnding.Enum(),
		ScenarioExecutionEndingRequest: &gauge_messages.ScenarioExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message, scenarioResult)
}

func (e *specExecutor) executeScenarios() []*result.ScenarioResult {
	var scenarioResults []*result.ScenarioResult
	for _, scenario := range e.specification.Scenarios {
		scenarioResults = append(scenarioResults, e.executeScenario(scenario))
	}
	return scenarioResults
}

func (e *specExecutor) executeScenario(scenario *gauge.Scenario) *result.ScenarioResult {
	e.currentExecutionInfo.CurrentScenario = &gauge_messages.ScenarioInfo{Name: proto.String(scenario.Heading.Value), Tags: getTagValue(scenario.Tags), IsFailed: proto.Bool(false)}
	scenarioResult := &result.ScenarioResult{ProtoScenario: gauge.NewProtoScenario(scenario)}
	e.addAllItemsForScenarioExecution(scenario, scenarioResult)
	scenarioResult.ProtoScenario.Skipped = proto.Bool(false)
	if _, ok := e.errMap.scenarioErrs[scenario]; ok {
		e.setSkipInfoInResult(scenarioResult, scenario)
		return scenarioResult
	}
	err := e.initScenarioDataStore()
	if err != nil {
		e.handleScenarioDataStoreFailure(scenarioResult, scenario, err)
		return scenarioResult
	}
	e.consoleReporter.ScenarioStart(scenario.Heading.Value)
	beforeHookExecutionStatus := e.executeBeforeScenarioHook(scenarioResult)
	if beforeHookExecutionStatus.GetFailed() {
		handleHookFailure(scenarioResult, beforeHookExecutionStatus, result.AddPreHook, e.consoleReporter)
		setScenarioFailure(e.currentExecutionInfo)
	} else {
		e.executeContextItems(scenarioResult)
		if !scenarioResult.GetFailure() {
			e.executeScenarioItems(scenarioResult)
		}
		e.executeTearDownItems(scenarioResult)
	}
	afterHookExecutionStatus := e.executeAfterScenarioHook(scenarioResult)
	scenarioResult.UpdateExecutionTime()
	if afterHookExecutionStatus.GetFailed() {
		handleHookFailure(scenarioResult, afterHookExecutionStatus, result.AddPostHook, e.consoleReporter)
		setScenarioFailure(e.currentExecutionInfo)
	}
	e.consoleReporter.ScenarioEnd(scenarioResult.GetFailure())

	return scenarioResult
}

func (e *specExecutor) handleScenarioDataStoreFailure(scenarioResult *result.ScenarioResult, scenario *gauge.Scenario, err error) {
	e.consoleReporter.Error(err.Error())
	e.errMap.scenarioErrs[scenario] = []*stepValidationError{
		&stepValidationError{
			step:     &gauge.Step{LineNo: scenario.Heading.LineNo, LineText: scenario.Heading.Value},
			message:  err.Error(),
			fileName: e.specification.FileName,
		},
	}
	e.setSkipInfoInResult(scenarioResult, scenario)
}

func (e *specExecutor) setSkipInfoInResult(result *result.ScenarioResult, scenario *gauge.Scenario) {
	e.specResult.ScenarioSkippedCount++
	result.ProtoScenario.Skipped = proto.Bool(true)
	var errors []string
	for _, err := range e.errMap.scenarioErrs[scenario] {
		errors = append(errors, fmt.Sprintf("%s:%d: %s. %s", err.fileName, err.step.LineNo, err.Error(), err.step.LineText))
	}
	result.ProtoScenario.SkipErrors = errors
}

func (e *specExecutor) addAllItemsForScenarioExecution(scenario *gauge.Scenario, scenarioResult *result.ScenarioResult) {
	scenarioResult.AddContexts(e.getContextItemsForScenarioExecution(e.specification.Contexts))
	scenarioResult.AddTearDownSteps(e.getContextItemsForScenarioExecution(e.specification.TearDownSteps))
	scenarioResult.AddItems(e.resolveItems(scenario.Items))
}

func (e *specExecutor) getContextItemsForScenarioExecution(steps []*gauge.Step) []*gauge_messages.ProtoItem {
	items := make([]gauge.Item, len(steps))
	for i, context := range steps {
		items[i] = context
	}
	return e.resolveItems(items)
}

func (e *specExecutor) executeContextItems(scenarioResult *result.ScenarioResult) {
	failure := e.executeItems(scenarioResult.ProtoScenario.GetContexts())
	if failure {
		scenarioResult.SetFailure()
	}
}

func (e *specExecutor) executeTearDownItems(scenarioResult *result.ScenarioResult) {
	failure := e.executeItems(scenarioResult.ProtoScenario.TearDownSteps)
	if failure {
		scenarioResult.SetFailure()
	}
}

func (e *specExecutor) executeScenarioItems(scenarioResult *result.ScenarioResult) {
	failure := e.executeItems(scenarioResult.ProtoScenario.GetScenarioItems())
	if failure {
		scenarioResult.SetFailure()
	}
}

func (e *specExecutor) resolveItems(items []gauge.Item) []*gauge_messages.ProtoItem {
	var protoItems []*gauge_messages.ProtoItem
	for _, item := range items {
		if item.Kind() != gauge.TearDownKind {
			protoItems = append(protoItems, e.resolveToProtoItem(item))
		}
	}
	return protoItems
}

func (e *specExecutor) executeItems(executingItems []*gauge_messages.ProtoItem) bool {
	for _, protoItem := range executingItems {
		failure := e.executeItem(protoItem)
		if failure == true {
			return true
		}
	}
	return false
}

func (e *specExecutor) resolveToProtoItem(item gauge.Item) *gauge_messages.ProtoItem {
	var protoItem *gauge_messages.ProtoItem
	switch item.Kind() {
	case gauge.StepKind:
		if (item.(*gauge.Step)).IsConcept {
			concept := item.(*gauge.Step)
			protoItem = e.resolveToProtoConceptItem(*concept)
		} else {
			protoItem = e.resolveToProtoStepItem(item.(*gauge.Step))
		}
		break

	default:
		protoItem = gauge.ConvertToProtoItem(item)
	}
	return protoItem
}

func (e *specExecutor) resolveToProtoStepItem(step *gauge.Step) *gauge_messages.ProtoItem {
	protoStepItem := gauge.ConvertToProtoItem(step)
	paramResolver := new(parser.ParamResolver)
	parameters := paramResolver.GetResolvedParams(step, nil, e.dataTableLookup())
	updateProtoStepParameters(protoStepItem.Step, parameters)
	e.setSkipInfo(protoStepItem.Step, step)
	return protoStepItem
}

func (e *specExecutor) setSkipInfo(protoStep *gauge_messages.ProtoStep, step *gauge.Step) {
	protoStep.StepExecutionResult = &gauge_messages.ProtoStepExecutionResult{}
	protoStep.StepExecutionResult.Skipped = proto.Bool(false)
	if _, ok := e.errMap.stepErrs[step]; ok {
		protoStep.StepExecutionResult.Skipped = proto.Bool(true)
		protoStep.StepExecutionResult.SkippedReason = proto.String("Step implemenatation not found")
	}
}

// Not passing pointer as we cannot modify the original concept step's lookup. This has to be populated for each iteration over data table.
func (e *specExecutor) resolveToProtoConceptItem(concept gauge.Step) *gauge_messages.ProtoItem {
	paramResolver := new(parser.ParamResolver)

	parser.PopulateConceptDynamicParams(&concept, e.dataTableLookup())
	protoConceptItem := gauge.ConvertToProtoItem(&concept)
	protoConceptItem.Concept.ConceptStep.StepExecutionResult = &gauge_messages.ProtoStepExecutionResult{}
	for stepIndex, step := range concept.ConceptSteps {
		// Need to reset parent as the step.parent is pointing to a concept whose lookup is not populated yet
		if step.IsConcept {
			step.Parent = &concept
			protoConceptItem.GetConcept().GetSteps()[stepIndex] = e.resolveToProtoConceptItem(*step)
		} else {
			stepParameters := paramResolver.GetResolvedParams(step, &concept, e.dataTableLookup())
			updateProtoStepParameters(protoConceptItem.Concept.Steps[stepIndex].Step, stepParameters)
			e.setSkipInfo(protoConceptItem.Concept.Steps[stepIndex].Step, step)
		}
	}
	protoConceptItem.Concept.ConceptStep.StepExecutionResult.Skipped = proto.Bool(false)
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

func (e *specExecutor) dataTableLookup() *gauge.ArgLookup {
	return new(gauge.ArgLookup).FromDataTableRow(&e.specification.DataTable.Table, e.currentTableRow)
}

func (e *specExecutor) executeItem(protoItem *gauge_messages.ProtoItem) bool {
	if protoItem.GetItemType() == gauge_messages.ProtoItem_Concept {
		return e.executeConcept(protoItem.GetConcept())
	} else if protoItem.GetItemType() == gauge_messages.ProtoItem_Step {
		return e.executeStep(protoItem.GetStep())
	}
	return false
}

func (e *specExecutor) executeSteps(protoSteps []*gauge_messages.ProtoStep) bool {
	for _, protoStep := range protoSteps {
		failure := e.executeStep(protoStep)
		if failure {
			return true
		}
	}
	return false
}

func (e *specExecutor) executeConcept(protoConcept *gauge_messages.ProtoConcept) bool {
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

func (e *specExecutor) setExecutionResultForConcept(protoConcept *gauge_messages.ProtoConcept) {
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

func printStatus(executionResult *gauge_messages.ProtoExecutionResult, reporter reporter.Reporter) {
	reporter.Error("Error Message: %s", executionResult.GetErrorMessage())
	reporter.Error("Stacktrace: \n%s", executionResult.GetStackTrace())
}

func (e *specExecutor) executeStep(protoStep *gauge_messages.ProtoStep) bool {
	stepRequest := e.createStepRequest(protoStep)
	stepText := formatter.FormatStep(parser.CreateStepFromStepRequest(stepRequest))
	e.consoleReporter.StepStart(stepText)

	protoStepExecResult := &gauge_messages.ProtoStepExecutionResult{}
	e.currentExecutionInfo.CurrentStep = &gauge_messages.StepInfo{Step: stepRequest, IsFailed: proto.Bool(false)}

	beforeHookStatus := e.executeBeforeStepHook()
	if beforeHookStatus.GetFailed() {
		protoStepExecResult.PreHookFailure = result.GetProtoHookFailure(beforeHookStatus)
		protoStepExecResult.ExecutionResult = &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(true)}
		setStepFailure(e.currentExecutionInfo, e.consoleReporter)
		printStatus(beforeHookStatus, e.consoleReporter)
	} else {
		executeStepMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecuteStep.Enum(), ExecuteStepRequest: stepRequest}
		stepExecutionStatus := executeAndGetStatus(e.runner, executeStepMessage)
		if stepExecutionStatus.GetFailed() {
			setStepFailure(e.currentExecutionInfo, e.consoleReporter)
		}
		protoStepExecResult.ExecutionResult = stepExecutionStatus
	}
	afterStepHookStatus := e.executeAfterStepHook()
	addExecutionTimes(protoStepExecResult, beforeHookStatus, afterStepHookStatus)
	if afterStepHookStatus.GetFailed() {
		setStepFailure(e.currentExecutionInfo, e.consoleReporter)
		printStatus(afterStepHookStatus, e.consoleReporter)
		protoStepExecResult.PostHookFailure = result.GetProtoHookFailure(afterStepHookStatus)
		protoStepExecResult.ExecutionResult.Failed = proto.Bool(true)
	}
	protoStepExecResult.ExecutionResult.Message = afterStepHookStatus.Message
	protoStepExecResult.Skipped = protoStep.StepExecutionResult.Skipped
	protoStepExecResult.SkippedReason = protoStep.StepExecutionResult.SkippedReason
	protoStep.StepExecutionResult = protoStepExecResult

	stepFailed := protoStep.GetStepExecutionResult().GetExecutionResult().GetFailed()
	e.consoleReporter.StepEnd(stepFailed)
	if stepFailed {
		result := protoStep.GetStepExecutionResult().GetExecutionResult()
		e.consoleReporter.Error("Failed Step: %s", e.currentExecutionInfo.CurrentStep.Step.GetActualStepText())
		e.consoleReporter.Error("Error Message: %s", strings.TrimSpace(result.GetErrorMessage()))
		e.consoleReporter.Error("Stacktrace: \n%s", result.GetStackTrace())
	}
	return stepFailed
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

func (e *specExecutor) executeBeforeStepHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepExecutionStarting.Enum(),
		StepExecutionStartingRequest: &gauge_messages.StepExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	e.pluginHandler.NotifyPlugins(message)
	return executeAndGetStatus(e.runner, message)
}

func (e *specExecutor) executeAfterStepHook() *gauge_messages.ProtoExecutionResult {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepExecutionEnding.Enum(),
		StepExecutionEndingRequest: &gauge_messages.StepExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	e.pluginHandler.NotifyPlugins(message)
	return executeAndGetStatus(e.runner, message)
}

func (e *specExecutor) createStepRequest(protoStep *gauge_messages.ProtoStep) *gauge_messages.ExecuteStepRequest {
	stepRequest := &gauge_messages.ExecuteStepRequest{ParsedStepText: proto.String(protoStep.GetParsedText()), ActualStepText: proto.String(protoStep.GetActualText())}
	stepRequest.Parameters = getParameters(protoStep.GetFragments())
	return stepRequest
}

func (e *specExecutor) getCurrentDataTableValueFor(columnName string) string {
	return e.specification.DataTable.Table.Get(columnName)[e.currentTableRow].Value
}

func executeAndGetStatus(runner *runner.TestRunner, message *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
	response, err := conn.GetResponseForGaugeMessage(message, runner.Connection)
	if err != nil {
		return &gauge_messages.ProtoExecutionResult{Failed: proto.Bool(true), ErrorMessage: proto.String(err.Error())}
	}

	if response.GetMessageType() == gauge_messages.Message_ExecutionStatusResponse {
		executionResult := response.GetExecutionStatusResponse().GetExecutionResult()
		if executionResult == nil {
			errMsg := "ProtoExecutionResult obtained is nil"
			logger.Error(errMsg)
			return errorResult(errMsg)
		}
		return executionResult
	}
	errMsg := fmt.Sprintf("Expected ExecutionStatusResponse. Obtained: %s", response.GetMessageType())
	logger.Error(errMsg)
	return errorResult(errMsg)
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

func setStepFailure(executionInfo *gauge_messages.ExecutionInfo, reporter reporter.Reporter) {
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
