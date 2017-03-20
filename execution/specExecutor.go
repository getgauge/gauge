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

	"strconv"
	"strings"

	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/validation"
)

type specExecutor struct {
	specification        *gauge.Specification
	dataTableIndexes     []int
	runner               runner.Runner
	pluginHandler        *plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	specResult           *result.SpecResult
	currentTableRow      int
	errMap               *gauge.BuildErrors
	stream               int
}

func newSpecExecutor(s *gauge.Specification, r runner.Runner, ph *plugin.Handler, e *gauge.BuildErrors, stream int) *specExecutor {
	return &specExecutor{specification: s, runner: r, pluginHandler: ph, errMap: e, stream: stream}
}

func hasParseError(errs []error) bool {
	for _, e := range errs {
		switch e.(type) {
		case parser.ParseError:
			return true
		}
	}
	return false
}

func (e *specExecutor) execute() *result.SpecResult {
	specInfo := &gauge_messages.SpecInfo{Name: e.specification.Heading.Value,
		FileName: e.specification.FileName,
		IsFailed: false, Tags: getTagValue(e.specification.Tags)}
	e.currentExecutionInfo = &gauge_messages.ExecutionInfo{CurrentSpec: specInfo}
	e.specResult = gauge.NewSpecResult(e.specification)
	if errs, ok := e.errMap.SpecErrs[e.specification]; ok {
		if hasParseError(errs) {
			e.failSpec()
			return e.specResult
		}
	}

	resolvedSpecItems := e.resolveItems(e.specification.GetSpecItems())
	e.specResult.AddSpecItems(resolvedSpecItems)
	if _, ok := e.errMap.SpecErrs[e.specification]; ok {
		e.skipSpec()
		return e.specResult
	}
	e.dataTableIndexes = getDataTableRows(e.specification.DataTable.Table.GetRowCount())

	if len(e.specification.Scenarios) == 0 {
		e.skipSpecForError(fmt.Errorf("%s: No scenarios found in spec\n", e.specification.FileName))
		return e.specResult
	}

	event.Notify(event.NewExecutionEvent(event.SpecStart, e.specification, nil, e.stream, *e.currentExecutionInfo))
	defer event.Notify(event.NewExecutionEvent(event.SpecEnd, nil, e.specResult, e.stream, *e.currentExecutionInfo))

	res := e.initSpecDataStore()
	if res.GetFailed() {
		e.skipSpecForError(fmt.Errorf("Failed to initialize spec datastore. Error: %s", res.GetErrorMessage()))
		return e.specResult
	}

	e.notifyBeforeSpecHook()
	if !e.specResult.GetFailed() {
		if e.specification.DataTable.Table.GetRowCount() == 0 {
			scenarioResults := e.executeScenarios(e.specification.Scenarios)
			e.specResult.AddScenarioResults(scenarioResults)
		} else {
			e.executeTableRelatedSpec()
		}
	}
	e.notifyAfterSpecHook()

	e.specResult.SetSkipped(e.specResult.ScenarioSkippedCount == len(e.specification.Scenarios))
	return e.specResult
}

func (e *specExecutor) executeTableRelatedScenarios(scenarios []*gauge.Scenario) (result [][]result.Result, executedRowIndexes []int) {
	for _, tableRowIndex := range e.dataTableIndexes {
		e.currentTableRow = tableRowIndex
		result = append(result, e.executeScenarios(scenarios))
		executedRowIndexes = append(executedRowIndexes, e.currentTableRow)
	}
	return
}

func (e *specExecutor) executeNonTableRelatedScenarios(scenarios []*gauge.Scenario) (scenarioResults []result.Result){
	for _, scenario := range scenarios {
		scenarioResults = append(scenarioResults, e.executeScenario(scenario))
	}
	return
}

func filterTableRelatedScenarios(scenarios []*gauge.Scenario, headers []string) (otherScenarios, tablRelatedScenarios []*gauge.Scenario) {
	for _, scenario := range scenarios {
		if scenario.UsesArgsInSteps(headers...) {
			tablRelatedScenarios = append(tablRelatedScenarios, scenario)
		} else {
			otherScenarios = append(otherScenarios, scenario)
		}
	}
	return
}

func (e *specExecutor) executeTableRelatedSpec() {
	if e.specification.UsesArgsInContextTeardown(e.specification.DataTable.Table.Headers...){
		res, executedRowIndexes := e.executeTableRelatedScenarios(e.specification.Scenarios)
		e.specResult.AddTableRelatedScenarioResult(res, executedRowIndexes)

	}else {
		nonTableRelatedScenarios, tableRelatedScenarios := filterTableRelatedScenarios(e.specification.Scenarios, e.specification.DataTable.Table.Headers)
		res := e.executeNonTableRelatedScenarios(nonTableRelatedScenarios)
		e.specResult.AddNonTableRelatedScenarioResult(res)
		tableDrivenRes, executedRowIndexes := e.executeTableRelatedScenarios(tableRelatedScenarios)
		e.specResult.AddTableRelatedScenarioResult(tableDrivenRes, executedRowIndexes)
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
	protoConceptItem.Concept.ConceptStep.StepExecutionResult.Skipped = false
	return protoConceptItem
}

func (e *specExecutor) resolveToProtoStepItem(step *gauge.Step) *gauge_messages.ProtoItem {
	protoStepItem := gauge.ConvertToProtoItem(step)
	paramResolver := new(parser.ParamResolver)
	parameters := paramResolver.GetResolvedParams(step, nil, e.dataTableLookup())
	updateProtoStepParameters(protoStepItem.Step, parameters)
	e.setSkipInfo(protoStepItem.Step, step)
	return protoStepItem
}

func (e *specExecutor) initSpecDataStore() *gauge_messages.ProtoExecutionResult {
	initSpecDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecDataStoreInit,
		SpecDataStoreInitRequest: &gauge_messages.SpecDataStoreInitRequest{}}
	return e.runner.ExecuteAndGetStatus(initSpecDataStoreMessage)
}

func (e *specExecutor) notifyBeforeSpecHook() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionStarting,
		SpecExecutionStartingRequest: &gauge_messages.SpecExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := executeHook(m, e.specResult, e.runner, e.pluginHandler)
	if res.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPreHook)
	}
}

func (e *specExecutor) notifyAfterSpecHook() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionEnding,
		SpecExecutionEndingRequest: &gauge_messages.SpecExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := executeHook(m, e.specResult, e.runner, e.pluginHandler)
	if res.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPostHook)
	}
}

func executeHook(message *gauge_messages.Message, execTimeTracker result.ExecTimeTracker, r runner.Runner, ph *plugin.Handler) *gauge_messages.ProtoExecutionResult {
	ph.NotifyPlugins(message)
	executionResult := r.ExecuteAndGetStatus(message)
	execTimeTracker.AddExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func (e *specExecutor) skipSpecForError(err error) {
	logger.Errorf(err.Error())
	validationError := validation.NewStepValidationError(&gauge.Step{LineNo: e.specification.Heading.LineNo, LineText: e.specification.Heading.Value},
		err.Error(), e.specification.FileName, nil)
	for _, scenario := range e.specification.Scenarios {
		e.errMap.ScenarioErrs[scenario] = []error{validationError}
	}
	e.errMap.SpecErrs[e.specification] = []error{validationError}
	e.skipSpec()
}

func (e *specExecutor) accumulateSkippedScenarioResults() []result.Result {
	var scenarioResults []result.Result
	for _, scenario := range e.specification.Scenarios {
		scenarioResults = append(scenarioResults, e.getSkippedScenarioResult(scenario))
	}
	return scenarioResults
}

func (e *specExecutor) failSpec() {
	e.specResult.Errors = e.convertErrors(e.errMap.SpecErrs[e.specification])
	e.specResult.SetFailure()
}

func (e *specExecutor) skipSpec() {
	if e.specResult.ProtoSpec.GetIsTableDriven() {
		res := make([][]result.Result, 0)
		executedRowIndexes := make([]int, 0)
		for i := 0; i < e.specification.DataTable.Table.GetRowCount(); i++ {
			e.currentTableRow = i
			res = append(res, e.accumulateSkippedScenarioResults())
			executedRowIndexes = append(executedRowIndexes, e.currentTableRow)
		}
		e.specResult.AddTableRelatedScenarioResult(res, executedRowIndexes)
	} else {
		e.specResult.AddScenarioResults(e.accumulateSkippedScenarioResults())
	}
	e.specResult.Errors = e.convertErrors(e.errMap.SpecErrs[e.specification])
	e.specResult.Skipped = true
}

func (e *specExecutor) convertErrors(specErrors []error) []*gauge_messages.Error {
	var errors []*gauge_messages.Error
	for _, e := range specErrors {
		switch e.(type) {
		case parser.ParseError:
			err := e.(parser.ParseError)
			errors = append(errors, &gauge_messages.Error{
				Message:    err.Error(),
				LineNumber: int32(err.LineNo),
				Filename:   err.FileName,
				Type:       gauge_messages.Error_PARSE_ERROR,
			})
		case validation.StepValidationError, validation.SpecValidationError:
			errors = append(errors, &gauge_messages.Error{
				Message: e.Error(),
				Type:    gauge_messages.Error_VALIDATION_ERROR,
			})
		}
	}
	return errors
}

func (e *specExecutor) setSkipInfo(protoStep *gauge_messages.ProtoStep, step *gauge.Step) {
	protoStep.StepExecutionResult = &gauge_messages.ProtoStepExecutionResult{}
	protoStep.StepExecutionResult.Skipped = false
	if _, ok := e.errMap.StepErrs[step]; ok {
		protoStep.StepExecutionResult.Skipped = true
		protoStep.StepExecutionResult.SkippedReason = "Step implementation not found"
	}
}

func (e *specExecutor) getItemsForScenarioExecution(steps []*gauge.Step) []*gauge_messages.ProtoItem {
	items := make([]gauge.Item, len(steps))
	for i, context := range steps {
		items[i] = context
	}
	return e.resolveItems(items)
}

func (e *specExecutor) dataTableLookup() *gauge.ArgLookup {
	return new(gauge.ArgLookup).FromDataTableRow(&e.specification.DataTable.Table, e.currentTableRow)
}

func (e *specExecutor) getCurrentDataTableValueFor(columnName string) string {
	return e.specification.DataTable.Table.Get(columnName)[e.currentTableRow].Value
}

func (e *specExecutor) executeScenarios(scenarios []*gauge.Scenario) []result.Result {
	var scenarioResults []result.Result
	for _, scenario := range scenarios {
		scenarioResults = append(scenarioResults, e.executeScenario(scenario))
	}
	return scenarioResults
}

func (e *specExecutor) executeScenario(scenario *gauge.Scenario) *result.ScenarioResult {
	e.currentExecutionInfo.CurrentScenario = &gauge_messages.ScenarioInfo{
		Name:     scenario.Heading.Value,
		Tags:     getTagValue(scenario.Tags),
		IsFailed: false,
	}

	scenarioResult := result.NewScenarioResult(gauge.NewProtoScenario(scenario))

	// TODO: During data driven execution, scenario holds the last row of datatable in scenario.DataTableRow.
	// This can be eliminated by creating a new scenario instance for each of the table row execution.
	if e.specification.DataTable.Table.GetRowCount() != 0 {
		var dataTable gauge.Table
		dataTable.AddHeaders(e.specification.DataTable.Table.Headers)
		dataTable.AddRowValues(e.specification.DataTable.Table.Rows()[e.currentTableRow])
		scenario.DataTableRow = dataTable
		scenario.DataTableRowIndex = e.currentTableRow
	}

	e.addAllItemsForScenarioExecution(scenario, scenarioResult)

	scenarioExec := newScenarioExecutor(e.runner, e.pluginHandler, e.currentExecutionInfo, e.errMap, e.stream)
	scenarioExec.execute(scenarioResult, scenario, e.specification.Contexts, e.specification.TearDownSteps)
	if scenarioResult.ProtoScenario.GetExecutionStatus() == gauge_messages.ExecutionStatus_SKIPPED {
		e.specResult.ScenarioSkippedCount++
	}
	return scenarioResult
}

func (e *specExecutor) addAllItemsForScenarioExecution(scenario *gauge.Scenario, scenarioResult *result.ScenarioResult) {
	scenarioResult.AddContexts(e.getItemsForScenarioExecution(e.specification.Contexts))
	scenarioResult.AddTearDownSteps(e.getItemsForScenarioExecution(e.specification.TearDownSteps))
	scenarioResult.AddItems(e.resolveItems(scenario.Items))
}

func (e *specExecutor) getSkippedScenarioResult(scenario *gauge.Scenario) *result.ScenarioResult {
	scenarioResult := &result.ScenarioResult{ProtoScenario: gauge.NewProtoScenario(scenario)}
	e.addAllItemsForScenarioExecution(scenario, scenarioResult)
	setSkipInfoInResult(scenarioResult, scenario, e.errMap)
	e.specResult.ScenarioSkippedCount++
	return scenarioResult
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

func getTagValue(tags *gauge.Tags) []string {
	var tagValues []string
	if tags != nil {
		tagValues = append(tagValues, tags.Values...)
	}
	return tagValues
}

func setSpecFailure(executionInfo *gauge_messages.ExecutionInfo) {
	executionInfo.CurrentSpec.IsFailed = true
}

func getDataTableRows(rowCount int) []int {
	var tableRowIndexes []int
	if rowCount == 0 && TableRows == "" {
		tableRowIndexes = []int{}
	} else if TableRows == "" {
		for i := 0; i < rowCount; i++ {
			tableRowIndexes = append(tableRowIndexes, i)
		}
	} else if strings.Contains(TableRows, "-") {
		indexes := strings.Split(TableRows, "-")
		startRow, _ := strconv.Atoi(strings.TrimSpace(indexes[0]))
		endRow, _ := strconv.Atoi(strings.TrimSpace(indexes[1]))
		for i := startRow - 1; i < endRow; i++ {
			tableRowIndexes = append(tableRowIndexes, i)
		}
	} else {
		indexes := strings.Split(TableRows, ",")
		for _, i := range indexes {
			rowNumber, _ := strconv.Atoi(strings.TrimSpace(i))
			tableRowIndexes = append(tableRowIndexes, rowNumber-1)
		}
	}
	return tableRowIndexes
}
