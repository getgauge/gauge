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
	runner               runner.Runner
	pluginHandler        plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	specResult           *result.SpecResult
	errMap               *gauge.BuildErrors
	stream               int
	scenarioExecutor     executor
}

func newSpecExecutor(s *gauge.Specification, r runner.Runner, ph plugin.Handler, e *gauge.BuildErrors, stream int) *specExecutor {
	ei := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     s.Heading.Value,
			FileName: s.FileName,
			IsFailed: false,
			Tags:     getTagValue(s.Tags)},
	}

	return &specExecutor{
		specification:        s,
		runner:               r,
		pluginHandler:        ph,
		errMap:               e,
		stream:               stream,
		currentExecutionInfo: ei,
		scenarioExecutor:     newScenarioExecutor(r, ph, ei, e, s.Contexts, s.TearDownSteps, stream),
	}
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

func (e *specExecutor) execute(executeBefore, execute, executeAfter bool) *result.SpecResult {
	e.specResult = gauge.NewSpecResult(e.specification)
	if errs, ok := e.errMap.SpecErrs[e.specification]; ok {
		if hasParseError(errs) {
			e.failSpec()
			return e.specResult
		}
	}
	resolvedSpecItems, err := e.resolveItems(e.specification.GetSpecItems())
	if err != nil {
		logger.Fatalf(true, "Failed to resolve Specifications : %s", err.Error())
	}
	e.specResult.AddSpecItems(resolvedSpecItems)
	if executeBefore {
		event.Notify(event.NewExecutionEvent(event.SpecStart, e.specification, e.specResult, e.stream, *e.currentExecutionInfo))
		if _, ok := e.errMap.SpecErrs[e.specification]; !ok {
			if res := e.initSpecDataStore(); res.GetFailed() {
				e.skipSpecForError(fmt.Errorf("Failed to initialize spec datastore. Error: %s", res.GetErrorMessage()))
			} else {
				e.notifyBeforeSpecHook()
			}
		} else {
			e.specResult.SetSkipped(true)
			e.specResult.Errors = e.convertErrors(e.errMap.SpecErrs[e.specification])
		}
	}
	if execute && !e.specResult.GetFailed() {
		if e.specification.DataTable.Table.GetRowCount() == 0 {
			scenarioResults, err := e.executeScenarios(e.specification.Scenarios)
			if err != nil {
				logger.Fatalf(true, "Failed to resolve Specifications : %s", err.Error())
			}
			e.specResult.AddScenarioResults(scenarioResults)
		} else {
			e.executeSpec()
		}
	}
	e.specResult.SetSkipped(e.specResult.Skipped || e.specResult.ScenarioSkippedCount == len(e.specification.Scenarios))
	if executeAfter {
		if _, ok := e.errMap.SpecErrs[e.specification]; !ok {
			e.notifyAfterSpecHook()
		}
		event.Notify(event.NewExecutionEvent(event.SpecEnd, e.specification, e.specResult, e.stream, *e.currentExecutionInfo))
	}
	return e.specResult
}

func (e *specExecutor) executeTableRelatedScenarios(scenarios []*gauge.Scenario) error {
	if len(scenarios) > 0 {
		index := e.specification.Scenarios[0].SpecDataTableRowIndex
		sceRes, err := e.executeScenarios(scenarios)
		if err != nil {
			return err
		}
		result := [][]result.Result{sceRes}
		e.specResult.AddTableRelatedScenarioResult(result, index)
	}
	return nil
}

func (e *specExecutor) executeSpec() error {
	parser.GetResolvedDataTablerows(e.specification.DataTable.Table)
	nonTableRelatedScenarios, tableRelatedScenarios := parser.FilterTableRelatedScenarios(e.specification.Scenarios, func(s *gauge.Scenario) bool {
		return s.SpecDataTableRow.IsInitialized()
	})
	res, err := e.executeScenarios(nonTableRelatedScenarios)
	if err != nil {
		return err
	}
	e.specResult.AddScenarioResults(res)
	e.executeTableRelatedScenarios(tableRelatedScenarios)
	return nil
}

func (e *specExecutor) resolveItems(items []gauge.Item) ([]*gauge_messages.ProtoItem, error) {
	var protoItems []*gauge_messages.ProtoItem
	for _, item := range items {
		if item.Kind() != gauge.TearDownKind {
			protoItem, err := e.resolveToProtoItem(item)
			if err != nil {
				return nil, err
			}
			protoItems = append(protoItems, protoItem)
		}
	}
	return protoItems, nil
}

func (e *specExecutor) resolveToProtoItem(item gauge.Item) (*gauge_messages.ProtoItem, error) {
	var protoItem *gauge_messages.ProtoItem
	var err error
	switch item.Kind() {
	case gauge.StepKind:
		if (item.(*gauge.Step)).IsConcept {
			concept := item.(*gauge.Step)
			lookup, err := e.dataTableLookup()
			if err != nil {
				return nil, err
			}
			protoItem, err = e.resolveToProtoConceptItem(*concept, lookup)
		} else {
			protoItem, err = e.resolveToProtoStepItem(item.(*gauge.Step))
		}
		break

	default:
		protoItem = gauge.ConvertToProtoItem(item)
	}
	return protoItem, err
}

// Not passing pointer as we cannot modify the original concept step's lookup. This has to be populated for each iteration over data table.
func (e *specExecutor) resolveToProtoConceptItem(concept gauge.Step, lookup *gauge.ArgLookup) (*gauge_messages.ProtoItem, error) {
	if err := parser.PopulateConceptDynamicParams(&concept, lookup); err != nil {
		return nil, err
	}
	protoConceptItem := gauge.ConvertToProtoItem(&concept)
	protoConceptItem.Concept.ConceptStep.StepExecutionResult = &gauge_messages.ProtoStepExecutionResult{}
	for stepIndex, step := range concept.ConceptSteps {
		// Need to reset parent as the step.parent is pointing to a concept whose lookup is not populated yet
		if step.IsConcept {
			step.Parent = &concept
			protoItem, err := e.resolveToProtoConceptItem(*step, &concept.Lookup)
			if err != nil {
				return nil, err
			}
			protoConceptItem.GetConcept().GetSteps()[stepIndex] = protoItem
		} else {
			stepParameters, err := parser.GetResolvedParams(step, &concept, &concept.Lookup)
			if err != nil {
				return nil, err
			}
			updateProtoStepParameters(protoConceptItem.Concept.Steps[stepIndex].Step, stepParameters)
			e.setSkipInfo(protoConceptItem.Concept.Steps[stepIndex].Step, step)
		}
	}
	protoConceptItem.Concept.ConceptStep.StepExecutionResult.Skipped = false
	return protoConceptItem, nil
}

func (e *specExecutor) resolveToProtoStepItem(step *gauge.Step) (*gauge_messages.ProtoItem, error) {
	protoStepItem := gauge.ConvertToProtoItem(step)
	lookup, err := e.dataTableLookup()
	if err != nil {
		return nil, err
	}
	parameters, err := parser.GetResolvedParams(step, nil, lookup)
	if err != nil {
		return nil, err
	}
	updateProtoStepParameters(protoStepItem.Step, parameters)
	e.setSkipInfo(protoStepItem.Step, step)
	return protoStepItem, err
}

func (e *specExecutor) initSpecDataStore() *gauge_messages.ProtoExecutionResult {
	initSpecDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecDataStoreInit,
		SpecDataStoreInitRequest: &gauge_messages.SpecDataStoreInitRequest{}}
	return e.runner.ExecuteAndGetStatus(initSpecDataStoreMessage)
}

func (e *specExecutor) notifyBeforeSpecHook() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionStarting,
		SpecExecutionStartingRequest: &gauge_messages.SpecExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	e.pluginHandler.NotifyPlugins(m)
	res := executeHook(m, e.specResult, e.runner)
	e.specResult.ProtoSpec.PreHookMessages = res.Message
	e.specResult.ProtoSpec.PreHookScreenshots = res.Screenshots
	if res.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPreHook)
	}
}

func (e *specExecutor) notifyAfterSpecHook() {
	e.currentExecutionInfo.CurrentScenario = nil
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionEnding,
		SpecExecutionEndingRequest: &gauge_messages.SpecExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := executeHook(m, e.specResult, e.runner)
	e.specResult.ProtoSpec.PostHookMessages = res.Message
	e.specResult.ProtoSpec.PostHookScreenshots = res.Screenshots
	if res.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPostHook)
	}
	e.pluginHandler.NotifyPlugins(m)
}

func executeHook(message *gauge_messages.Message, execTimeTracker result.ExecTimeTracker, r runner.Runner) *gauge_messages.ProtoExecutionResult {
	executionResult := r.ExecuteAndGetStatus(message)
	execTimeTracker.AddExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func (e *specExecutor) skipSpecForError(err error) {
	logger.Errorf(true, err.Error())
	validationError := validation.NewStepValidationError(&gauge.Step{LineNo: e.specification.Heading.LineNo, LineText: e.specification.Heading.Value},
		err.Error(), e.specification.FileName, nil, "")
	for _, scenario := range e.specification.Scenarios {
		e.errMap.ScenarioErrs[scenario] = []error{validationError}
	}
	e.errMap.SpecErrs[e.specification] = []error{validationError}
	e.specResult.Errors = e.convertErrors(e.errMap.SpecErrs[e.specification])
	e.specResult.SetSkipped(true)
}

func (e *specExecutor) failSpec() {
	e.specResult.Errors = e.convertErrors(e.errMap.SpecErrs[e.specification])
	e.specResult.SetFailure()
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

func (e *specExecutor) getItemsForScenarioExecution(steps []*gauge.Step) ([]*gauge_messages.ProtoItem, error) {
	items := make([]gauge.Item, len(steps))
	for i, context := range steps {
		items[i] = context
	}
	return e.resolveItems(items)
}

func (e *specExecutor) dataTableLookup() (*gauge.ArgLookup, error) {
	return new(gauge.ArgLookup).FromDataTableRow(&e.specification.DataTable.Table, 0)
}

func (e *specExecutor) executeScenarios(scenarios []*gauge.Scenario) ([]result.Result, error) {
	var scenarioResults []result.Result
	for _, scenario := range scenarios {
		sceResult, err := e.executeScenario(scenario)
		if err != nil {
			return nil, err
		}
		scenarioResults = append(scenarioResults, sceResult)
	}
	return scenarioResults, nil
}

func (e *specExecutor) executeScenario(scenario *gauge.Scenario) (*result.ScenarioResult, error) {
	e.currentExecutionInfo.CurrentScenario = &gauge_messages.ScenarioInfo{
		Name:     scenario.Heading.Value,
		Tags:     getTagValue(scenario.Tags),
		IsFailed: false,
	}

	scenarioResult := result.NewScenarioResult(gauge.NewProtoScenario(scenario))
	if err := e.addAllItemsForScenarioExecution(scenario, scenarioResult); err != nil {
		return nil, err
	}

	e.scenarioExecutor.execute(scenario, scenarioResult)
	if scenarioResult.ProtoScenario.GetExecutionStatus() == gauge_messages.ExecutionStatus_SKIPPED {
		e.specResult.ScenarioSkippedCount++
	}
	return scenarioResult, nil
}

func (e *specExecutor) addAllItemsForScenarioExecution(scenario *gauge.Scenario, scenarioResult *result.ScenarioResult) error {
	contexts, err := e.getItemsForScenarioExecution(e.specification.Contexts)
	if err != nil {
		return err
	}
	scenarioResult.AddContexts(contexts)
	tearDownSteps, err := e.getItemsForScenarioExecution(e.specification.TearDownSteps)
	if err != nil {
		return err
	}
	scenarioResult.AddTearDownSteps(tearDownSteps)
	items, err := e.resolveItems(scenario.Items)
	if err != nil {
		return err
	}
	scenarioResult.AddItems(items)
	return nil
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
		tagValues = append(tagValues, tags.Values()...)
	}
	return tagValues
}

func setSpecFailure(executionInfo *gauge_messages.ExecutionInfo) {
	executionInfo.CurrentSpec.IsFailed = true
}

func shouldExecuteForRow(i int) bool {
	if len(tableRowsIndexes) < 1 {
		return true
	}
	for _, index := range tableRowsIndexes {
		if index == i {
			return true
		}
	}
	return false
}

func getDataTableRows(tableRows string) (tableRowIndexes []int) {
	if strings.TrimSpace(tableRows) == "" {
		return
	} else if strings.Contains(tableRows, "-") {
		indexes := strings.Split(tableRows, "-")
		startRow, _ := strconv.Atoi(strings.TrimSpace(indexes[0]))
		endRow, _ := strconv.Atoi(strings.TrimSpace(indexes[1]))
		for i := startRow - 1; i < endRow; i++ {
			tableRowIndexes = append(tableRowIndexes, i)
		}
	} else {
		indexes := strings.Split(tableRows, ",")
		for _, i := range indexes {
			rowNumber, _ := strconv.Atoi(strings.TrimSpace(i))
			tableRowIndexes = append(tableRowIndexes, rowNumber-1)
		}
	}
	return
}
