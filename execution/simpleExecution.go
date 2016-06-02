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
	"time"

	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/validation"

	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
)

var ExecuteTags = ""
var TableRows = ""

type simpleExecution struct {
	manifest             *manifest.Manifest
	runner               runner.Runner
	specCollection       *gauge.SpecCollection
	pluginHandler        *plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	suiteResult          *result.SuiteResult
	errMaps              *validation.ValidationErrMaps
	startTime            time.Time
	stream               int
}

func newSimpleExecution(executionInfo *executionInfo) *simpleExecution {
	return &simpleExecution{
		manifest:       executionInfo.manifest,
		specCollection: executionInfo.specs,
		runner:         executionInfo.runner,
		pluginHandler:  executionInfo.pluginHandler,
		errMaps:        executionInfo.errMaps,
		stream:         executionInfo.stream,
	}
}

func (e *simpleExecution) run() *result.SuiteResult {
	e.start()
	e.execute()
	e.finish()
	return e.suiteResult
}

func (e *simpleExecution) execute() {
	e.suiteResult = result.NewSuiteResult(ExecuteTags, e.startTime)
	setResultMeta := func() {
		e.suiteResult.UpdateExecTime(e.startTime)
		e.suiteResult.SetSpecsSkippedCount()
	}

	initSuiteDataStoreResult := e.initSuiteDataStore()
	if initSuiteDataStoreResult.GetFailed() {
		e.suiteResult.AddUnhandledError(fmt.Errorf("Failed to initialize suite datastore. Error: %s", initSuiteDataStoreResult.GetErrorMessage()))
		setResultMeta()
		return
	}

	e.notifyBeforeSuite()
	if !e.suiteResult.GetFailed() {
		results := e.executeSpecs(e.specCollection)
		e.suiteResult.AddSpecResults(results)
	}
	e.notifyAfterSuite()

	setResultMeta()
}

func (e *simpleExecution) start() {
	e.startTime = time.Now()
	event.Notify(event.NewExecutionEvent(event.SuiteStart, nil, nil, 0, gauge_messages.ExecutionInfo{}))
	e.pluginHandler = plugin.StartPlugins(e.manifest)
}

func (e *simpleExecution) finish() {
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, e.suiteResult, 0, gauge_messages.ExecutionInfo{}))
	e.notifyExecutionResult()
	e.stopAllPlugins()
}

func (e *simpleExecution) stopAllPlugins() {
	e.notifyExecutionStop()
	if err := e.runner.Kill(); err != nil {
		logger.Errorf("Failed to kill Runner: %s", err.Error())
	}
}

func (e *simpleExecution) executeSpecs(specs *gauge.SpecCollection) []*result.SpecResult {
	var results []*result.SpecResult
	for specs.HasNext() {
		s := specs.Next()
		ex := newSpecExecutor(s, e.runner, e.pluginHandler, getDataTableRows(s.DataTable.Table.GetRowCount()), e.errMaps, e.stream)
		results = append(results, ex.execute())
	}
	return results
}

func (e *simpleExecution) notifyBeforeSuite() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionStarting.Enum(),
		ExecutionStartingRequest: &gauge_messages.ExecutionStartingRequest{}}
	res := e.executeHook(m)
	if res.GetFailed() {
		handleHookFailure(e.suiteResult, res, result.AddPreHook)
	}
}

func (e *simpleExecution) notifyAfterSuite() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionEnding.Enum(),
		ExecutionEndingRequest: &gauge_messages.ExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	res := e.executeHook(m)
	if res.GetFailed() {
		handleHookFailure(e.suiteResult, res, result.AddPostHook)
	}
}

func (e *simpleExecution) initSuiteDataStore() *(gauge_messages.ProtoExecutionResult) {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteDataStoreInit.Enum(),
		SuiteDataStoreInitRequest: &gauge_messages.SuiteDataStoreInitRequest{}}
	return e.runner.ExecuteAndGetStatus(m)
}

func (e *simpleExecution) executeHook(m *gauge_messages.Message) *(gauge_messages.ProtoExecutionResult) {
	e.pluginHandler.NotifyPlugins(m)
	return e.runner.ExecuteAndGetStatus(m)
}

func (e *simpleExecution) notifyExecutionResult() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{SuiteResult: gauge.ConvertToProtoSuiteResult(e.suiteResult)}}
	e.pluginHandler.NotifyPlugins(m)
}

func (e *simpleExecution) notifyExecutionStop() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_KillProcessRequest.Enum(),
		KillProcessRequest: &gauge_messages.KillProcessRequest{}}
	e.pluginHandler.NotifyPlugins(m)
	e.pluginHandler.GracefullyKillPlugins()
}

func handleHookFailure(result result.Result, execResult *gauge_messages.ProtoExecutionResult, f func(result.Result, *gauge_messages.ProtoExecutionResult)) {
	f(result, execResult)
}

func getDataTableRows(rowCount int) indexRange {
	if TableRows == "" {
		return indexRange{start: 0, end: rowCount - 1}
	}
	indexes, err := getDataTableRowsRange(TableRows, rowCount)
	if err != nil {
		logger.Errorf("Table rows validation failed. %s\n", err.Error())
	}
	return indexes
}
