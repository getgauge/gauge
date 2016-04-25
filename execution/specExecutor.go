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

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/validation"
	"github.com/golang/protobuf/proto"
)

type specExecutor struct {
	specification        *gauge.Specification
	dataTableIndex       indexRange
	runner               runner.Runner
	pluginHandler        *plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	specResult           *result.SpecResult
	currentTableRow      int
	consoleReporter      reporter.Reporter
	errMap               *validation.ValidationErrMaps
}

type indexRange struct {
	start int
	end   int
}

func newSpecExecutor(s *gauge.Specification, r runner.Runner, ph *plugin.Handler, tr indexRange, rep reporter.Reporter, e *validation.ValidationErrMaps) *specExecutor {
	return &specExecutor{specification: s, runner: r, pluginHandler: ph, dataTableIndex: tr, consoleReporter: rep, errMap: e}
}

func (e *specExecutor) execute() *result.SpecResult {
	specInfo := &gauge_messages.SpecInfo{Name: proto.String(e.specification.Heading.Value),
		FileName: proto.String(e.specification.FileName),
		IsFailed: proto.Bool(false), Tags: getTagValue(e.specification.Tags)}
	e.currentExecutionInfo = &gauge_messages.ExecutionInfo{CurrentSpec: specInfo}
	e.specResult = gauge.NewSpecResult(e.specification)
	resolvedSpecItems := e.resolveItems(e.specification.GetSpecItems())
	e.specResult.AddSpecItems(resolvedSpecItems)
	if _, ok := e.errMap.SpecErrs[e.specification]; ok {
		e.skipSpec()
		return e.specResult
	}

	res := e.initSpecDataStore()
	if res.GetFailed() {
		e.consoleReporter.Errorf("Failed to initialize spec datastore. Error: %s", res.GetErrorMessage())
		e.skipSpecForError(fmt.Errorf(res.GetErrorMessage()))
		return e.specResult
	}
	if len(e.specification.Scenarios) == 0 {
		e.skipSpecForError(fmt.Errorf("No scenarios found in spec: %s\n", e.specification.FileName))
		return e.specResult
	}

	e.consoleReporter.SpecStart(specInfo.GetName())
	e.notifyBeforeSpecHook()
	if !e.specResult.IsFailed {
		dataTableRowCount := e.specification.DataTable.Table.GetRowCount()
		if dataTableRowCount == 0 {
			scenarioResults := e.executeScenarios()
			e.specResult.AddScenarioResults(scenarioResults)
		} else {
			e.executeTableDrivenSpec()
		}
	}

	e.notifyAfterSpecHook()
	e.specResult.Skipped = e.specResult.ScenarioSkippedCount > 0
	e.consoleReporter.SpecEnd()
	return e.specResult
}

func (e *specExecutor) executeTableDrivenSpec() {
	var dataTableScenarioExecutionResult [][]result.Result
	for e.currentTableRow = e.dataTableIndex.start; e.currentTableRow <= e.dataTableIndex.end; e.currentTableRow++ {
		var dataTable gauge.Table
		dataTable.AddHeaders(e.specification.DataTable.Table.Headers)
		dataTable.AddRowValues(e.specification.DataTable.Table.Rows()[e.currentTableRow])
		e.consoleReporter.DataTable(formatter.FormatTable(&dataTable))
		dataTableScenarioExecutionResult = append(dataTableScenarioExecutionResult, e.executeScenarios())
	}
	e.specResult.AddTableDrivenScenarioResult(dataTableScenarioExecutionResult)
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
	protoConceptItem.Concept.ConceptStep.StepExecutionResult.Skipped = proto.Bool(false)
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
	initSpecDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecDataStoreInit.Enum(),
		SpecDataStoreInitRequest: &gauge_messages.SpecDataStoreInitRequest{}}
	return e.runner.ExecuteAndGetStatus(initSpecDataStoreMessage)
}

func (e *specExecutor) notifyBeforeSpecHook() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionStarting.Enum(),
		SpecExecutionStartingRequest: &gauge_messages.SpecExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := executeHook(m, e.specResult, e.runner, e.pluginHandler)
	if res.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPreHook, e.consoleReporter)
	}
}

func (e *specExecutor) notifyAfterSpecHook() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionEnding.Enum(),
		SpecExecutionEndingRequest: &gauge_messages.SpecExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := executeHook(m, e.specResult, e.runner, e.pluginHandler)
	if res.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPostHook, e.consoleReporter)
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
	validationError := validation.NewValidationError(&gauge.Step{LineNo: e.specification.Heading.LineNo, LineText: e.specification.Heading.Value},
		err.Error(), e.specification.FileName, nil)
	for _, scenario := range e.specification.Scenarios {
		e.errMap.ScenarioErrs[scenario] = []*validation.StepValidationError{validationError}
	}
	e.errMap.SpecErrs[e.specification] = []*validation.StepValidationError{validationError}
	e.skipSpec()
}

func (e *specExecutor) skipSpec() {
	var scenarioResults []result.Result
	for _, scenario := range e.specification.Scenarios {
		scenarioResults = append(scenarioResults, e.getSkippedScenarioResult(scenario))
	}
	e.specResult.AddScenarioResults(scenarioResults)
	e.specResult.Skipped = true
}

func (e *specExecutor) setSkipInfo(protoStep *gauge_messages.ProtoStep, step *gauge.Step) {
	protoStep.StepExecutionResult = &gauge_messages.ProtoStepExecutionResult{}
	protoStep.StepExecutionResult.Skipped = proto.Bool(false)
	if _, ok := e.errMap.StepErrs[step]; ok {
		protoStep.StepExecutionResult.Skipped = proto.Bool(true)
		protoStep.StepExecutionResult.SkippedReason = proto.String("Step implemenatation not found")
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

func (e *specExecutor) executeScenarios() []result.Result {
	var scenarioResults []result.Result
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

	scenarioExec := newScenarioExecutor(e.runner, e.pluginHandler, e.currentExecutionInfo, e.consoleReporter, e.errMap)
	scenarioExec.execute(scenarioResult, scenario)
	if scenarioResult.ProtoScenario.GetSkipped() {
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

func printStatus(executionResult *gauge_messages.ProtoExecutionResult, reporter reporter.Reporter) {
	reporter.Errorf("Error Message: %s", executionResult.GetErrorMessage())
	reporter.Errorf("Stacktrace: \n%s", executionResult.GetStackTrace())
}

func getTagValue(tags *gauge.Tags) []string {
	var tagValues []string
	if tags != nil {
		tagValues = append(tagValues, tags.Values...)
	}
	return tagValues
}

func setSpecFailure(executionInfo *gauge_messages.ExecutionInfo) {
	executionInfo.CurrentSpec.IsFailed = proto.Bool(true)
}

func getDataTableRowsRange(tableRows string, rowCount int) (indexRange, error) {
	var startIndex, endIndex int
	var err error
	indexRanges := strings.Split(tableRows, "-")
	if len(indexRanges) == 2 {
		startIndex, endIndex, err = validation.ValidateTableRowsRange(indexRanges[0], indexRanges[1], rowCount)
	} else if len(indexRanges) == 1 {
		startIndex, endIndex, err = validation.ValidateTableRowsRange(tableRows, tableRows, rowCount)
	} else {
		return indexRange{start: 0, end: 0}, errors.New("Table rows range validation failed.")
	}
	if err != nil {
		return indexRange{start: 0, end: 0}, err
	}
	return indexRange{start: startIndex, end: endIndex}, nil
}
