/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
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
		ProjectName:              filepath.Base(config.ProjectRoot),
		NumberOfExecutionStreams: int32(NumberOfExecutionStreams),
		RunnerId:                 int32(stream),
		ExecutionArgs:            gauge.ConvertToProtoExecutionArg(ExecutionArgs),
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

func (e *specExecutor) execute(executeBefore, execute, executeAfter bool) *result.SpecResult {
	e.specResult = gauge.NewSpecResult(e.specification)
	if e.runner.Info().Killed {
		e.specResult.SetSkipped(true)
		return e.specResult
	}
	if errs, ok := e.errMap.SpecErrs[e.specification]; ok {
		if hasParseError(errs) {
			e.failSpec()
			return e.specResult
		}
	}
	lookup, err := e.dataTableLookup()
	if err != nil {
		logger.Fatalf(true, "Failed to resolve Specifications : %s", err.Error())
	}
	resolvedSpecItems, err := resolveItems(e.specification.GetSpecItems(), lookup, e.setSkipInfo)
	if err != nil {
		logger.Fatalf(true, "Failed to resolve Specifications : %s", err.Error())
	}
	e.specResult.AddSpecItems(resolvedSpecItems)
	if executeBefore {
		event.Notify(event.NewExecutionEvent(event.SpecStart, e.specification, e.specResult, e.stream, e.currentExecutionInfo))
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
			others, tableDriven := parser.FilterTableRelatedScenarios(e.specification.Scenarios, func(s *gauge.Scenario) bool {
				return s.ScenarioDataTableRow.IsInitialized()
			})
			results, err := e.executeScenarios(others)
			if err != nil {
				logger.Fatalf(true, "Failed to resolve Specifications : %s", err.Error())
			}
			e.specResult.AddScenarioResults(results)
			scnMap := make(map[int]bool)

			// Execute eager table-driven scenarios
			for _, s := range tableDriven {
				if _, ok := scnMap[s.Span.Start]; !ok {
					scnMap[s.Span.Start] = true
				}
				r, err := e.executeScenario(s)
				if err != nil {
					logger.Fatalf(true, "Failed to resolve Specifications : %s", err.Error())
				}
				e.specResult.AddTableDrivenScenarioResult(r, gauge.ConvertToProtoTable(s.DataTable.Table),
					s.ScenarioDataTableRowIndex, s.SpecDataTableRowIndex, s.SpecDataTableRow.IsInitialized())
			}

			// Execute lazy scenario collections
			for _, lazyCollection := range e.specification.LazyScenarios {
				if _, ok := scnMap[lazyCollection.Template.Span.Start]; !ok {
					scnMap[lazyCollection.Template.Span.Start] = true
				}

				iterator := lazyCollection.Iterator()
				for scenario, hasNext := iterator.Next(); hasNext; scenario, hasNext = iterator.Next() {
					r, err := e.executeScenario(scenario)
					if err != nil {
						logger.Fatalf(true, "Failed to resolve Specifications : %s", err.Error())
					}
					e.specResult.AddTableDrivenScenarioResult(r, gauge.ConvertToProtoTable(scenario.DataTable.Table),
						scenario.ScenarioDataTableRowIndex, scenario.SpecDataTableRowIndex, scenario.SpecDataTableRow.IsInitialized())
				}
			}

			e.specResult.ScenarioCount += len(scnMap)
		} else {
			err := e.executeSpec()
			if err != nil {
				logger.Fatalf(true, "Failed to execute Specification %s : %s", e.specification.Heading.Value, err.Error())
			}
		}
	}
	e.specResult.SetSkipped(e.specResult.Skipped || e.specResult.ScenarioSkippedCount == len(e.specification.Scenarios))
	if executeAfter {
		if _, ok := e.errMap.SpecErrs[e.specification]; !ok {
			e.notifyAfterSpecHook()
		}
		event.Notify(event.NewExecutionEvent(event.SpecEnd, e.specification, e.specResult, e.stream, e.currentExecutionInfo))
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
		specResult := [][]result.Result{sceRes}
		e.specResult.AddTableRelatedScenarioResult(specResult, index)
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
	err = e.executeTableRelatedScenarios(tableRelatedScenarios)
	if err != nil {
		return err
	}
	return nil
}

func (e *specExecutor) initSpecDataStore() *gauge_messages.ProtoExecutionResult {
	initSpecDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecDataStoreInit,
		SpecDataStoreInitRequest: &gauge_messages.SpecDataStoreInitRequest{Stream: int32(e.stream)}}
	return e.runner.ExecuteAndGetStatus(initSpecDataStoreMessage)
}

func (e *specExecutor) notifyBeforeSpecHook() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionStarting,
		SpecExecutionStartingRequest: &gauge_messages.SpecExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo, Stream: int32(e.stream)}}
	e.pluginHandler.NotifyPlugins(m)
	res := executeHook(m, e.specResult, e.runner)
	e.specResult.ProtoSpec.PreHookMessages = res.Message
	e.specResult.ProtoSpec.PreHookScreenshotFiles = res.ScreenshotFiles
	if res.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPreHook)
	}
	m.SpecExecutionStartingRequest.SpecResult = gauge.ConvertToProtoSpecResult(e.specResult)
	e.pluginHandler.NotifyPlugins(m)
}

func (e *specExecutor) notifyAfterSpecHook() {
	e.currentExecutionInfo.CurrentScenario = nil
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SpecExecutionEnding,
		SpecExecutionEndingRequest: &gauge_messages.SpecExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo, Stream: int32(e.stream)}}
	res := executeHook(m, e.specResult, e.runner)
	e.specResult.ProtoSpec.PostHookMessages = res.Message
	e.specResult.ProtoSpec.PostHookScreenshotFiles = res.ScreenshotFiles
	if res.GetFailed() {
		setSpecFailure(e.currentExecutionInfo)
		handleHookFailure(e.specResult, res, result.AddPostHook)
	}
	m.SpecExecutionEndingRequest.SpecResult = gauge.ConvertToProtoSpecResult(e.specResult)
	e.pluginHandler.NotifyPlugins(m)
}

func (e *specExecutor) skipSpecForError(err error) {
	logger.Error(true, err.Error())
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
		switch err := e.(type) {
		case parser.ParseError:
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
	lookup, err := e.dataTableLookup()
	if err != nil {
		return nil, err
	}
	return resolveItems(items, lookup, e.setSkipInfo)
}

func (e *specExecutor) dataTableLookup() (*gauge.ArgLookup, error) {
	l := new(gauge.ArgLookup)
	err := l.ReadDataTableRow(e.specification.DataTable.Table, 0)
	return l, err
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
	var scenarioResult *result.ScenarioResult

	shouldRetry := RetryOnlyTags == ""

	if !shouldRetry {
		spec := e.specification
		tagValues := make([]string, 0)
		if spec.Tags != nil {
			tagValues = spec.Tags.Values()
		}

		specFilter := filter.NewScenarioFilterBasedOnTags(tagValues, RetryOnlyTags)

		shouldRetry = !(specFilter.Filter(scenario))
	}
	retriesCount := 0
	for i := 0; i < MaxRetriesCount; i++ {
		e.currentExecutionInfo.CurrentScenario = &gauge_messages.ScenarioInfo{
			Name:     scenario.Heading.Value,
			Tags:     getTagValue(scenario.Tags),
			IsFailed: false,
			Retries: &gauge_messages.ScenarioRetriesInfo{
				MaxRetries:   int32(MaxRetriesCount) - 1,
				CurrentRetry: int32(retriesCount),
			},
		}

		scenarioResult = &result.ScenarioResult{
			ProtoScenario:             gauge.NewProtoScenario(scenario),
			ScenarioDataTableRow:      gauge.ConvertToProtoTable(&scenario.ScenarioDataTableRow),
			ScenarioDataTableRowIndex: scenario.ScenarioDataTableRowIndex,
			ScenarioDataTable:         gauge.ConvertToProtoTable(scenario.DataTable.Table),
		}
		if err := e.addAllItemsForScenarioExecution(scenario, scenarioResult); err != nil {
			return nil, err
		}
		e.scenarioExecutor.execute(scenario, scenarioResult)
		retriesCount++
		if scenarioResult.ProtoScenario.GetExecutionStatus() == gauge_messages.ExecutionStatus_SKIPPED {
			e.specResult.ScenarioSkippedCount++
		}

		if !(shouldRetry && scenarioResult.GetFailed()) {
			break
		}
	}
	scenarioResult.ProtoScenario.RetriesCount = int64(retriesCount)
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
	lookup, err := e.dataTableLookup()
	if err != nil {
		return err
	}
	if scenario.ScenarioDataTableRow.IsInitialized() {
		parser.GetResolvedDataTablerows(&scenario.ScenarioDataTableRow)
		if err = lookup.ReadDataTableRow(&scenario.ScenarioDataTableRow, 0); err != nil {
			return err
		}
	}
	items, err := resolveItems(scenario.Items, lookup, e.setSkipInfo)
	if err != nil {
		return err
	}
	scenarioResult.AddItems(items)
	return nil
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
	switch {
	case strings.TrimSpace(tableRows) == "":
		return
	case strings.Contains(tableRows, "-"):
		indexes := strings.Split(tableRows, "-")
		startRow, _ := strconv.Atoi(strings.TrimSpace(indexes[0]))
		endRow, _ := strconv.Atoi(strings.TrimSpace(indexes[1]))
		for i := startRow - 1; i < endRow; i++ {
			tableRowIndexes = append(tableRowIndexes, i)
		}
	default:
		indexes := strings.Split(tableRows, ",")
		for _, i := range indexes {
			rowNumber, _ := strconv.Atoi(strings.TrimSpace(i))
			tableRowIndexes = append(tableRowIndexes, rowNumber-1)
		}
	}
	return
}

func executeHook(message *gauge_messages.Message, execTimeTracker result.ExecTimeTracker, r runner.Runner) *gauge_messages.ProtoExecutionResult {
	executionResult := r.ExecuteAndGetStatus(message)
	execTimeTracker.AddExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func hasParseError(errs []error) bool {
	for _, e := range errs {
		if _, ok := e.(parser.ParseError); ok {
			return true
		}
	}
	return false
}
