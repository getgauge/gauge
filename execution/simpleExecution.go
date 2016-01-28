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
	"time"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/reporter"

	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
)

var ExecuteTags = ""
var TableRows = ""

type simpleExecution struct {
	manifest             *manifest.Manifest
	runner               *runner.TestRunner
	specStore            *specStore
	pluginHandler        *plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	suiteResult          *result.SuiteResult
	consoleReporter      reporter.Reporter
	errMaps              *validationErrMaps
	startTime            time.Time
}

func newSimpleExecution(executionInfo *executionInfo) *simpleExecution {
	return &simpleExecution{manifest: executionInfo.manifest, specStore: executionInfo.specStore,
		runner: executionInfo.runner, pluginHandler: executionInfo.pluginHandler, consoleReporter: executionInfo.consoleReporter, errMaps: executionInfo.errMaps}
}

func (e *simpleExecution) startExecution() *(gauge_messages.ProtoExecutionResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionStarting.Enum(),
		ExecutionStartingRequest: &gauge_messages.ExecutionStartingRequest{}}
	return e.executeHook(message)
}

func (e *simpleExecution) initializeSuiteDataStore() *(gauge_messages.ProtoExecutionResult) {
	initSuiteDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteDataStoreInit.Enum(),
		SuiteDataStoreInitRequest: &gauge_messages.SuiteDataStoreInitRequest{}}
	initResult := executeAndGetStatus(e.runner, initSuiteDataStoreMessage)
	return initResult
}

func (e *simpleExecution) endExecution() *(gauge_messages.ProtoExecutionResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionEnding.Enum(),
		ExecutionEndingRequest: &gauge_messages.ExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message)
}

func (e *simpleExecution) executeHook(message *gauge_messages.Message) *(gauge_messages.ProtoExecutionResult) {
	e.pluginHandler.NotifyPlugins(message)
	executionResult := executeAndGetStatus(e.runner, message)
	e.addExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func (e *simpleExecution) addExecTime(execTime int64) {
	e.suiteResult.ExecutionTime += execTime
}

func (e *simpleExecution) notifyExecutionResult() {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{SuiteResult: gauge.ConvertToProtoSuiteResult(e.suiteResult)}}
	e.pluginHandler.NotifyPlugins(message)
}

func (e *simpleExecution) notifyExecutionStop() {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_KillProcessRequest.Enum(),
		KillProcessRequest: &gauge_messages.KillProcessRequest{}}
	e.pluginHandler.NotifyPlugins(message)
	e.pluginHandler.GracefullyKillPlugins()
}

func (e *simpleExecution) killPlugins() {
	e.pluginHandler.GracefullyKillPlugins()
}

func (e *simpleExecution) start() {
	e.pluginHandler = plugin.StartPlugins(e.manifest)
	e.startTime = time.Now()
}

func (e *simpleExecution) run() *result.SuiteResult {
	e.suiteResult = result.NewSuiteResult(ExecuteTags, e.startTime)
	initSuiteDataStoreResult := e.initializeSuiteDataStore()
	if initSuiteDataStoreResult.GetFailed() {
		e.consoleReporter.Error("Failed to initialize suite datastore. Error: %s", initSuiteDataStoreResult.GetErrorMessage())
	} else {
		beforeSuiteHookExecResult := e.startExecution()
		if beforeSuiteHookExecResult.GetFailed() {
			handleHookFailure(e.suiteResult, beforeSuiteHookExecResult, result.AddPreHook, e.consoleReporter)
		} else {
			for e.specStore.hasNext() {
				e.executeSpec(e.specStore.next())
			}
		}
		afterSuiteHookExecResult := e.endExecution()
		if afterSuiteHookExecResult.GetFailed() {
			handleHookFailure(e.suiteResult, afterSuiteHookExecResult, result.AddPostHook, e.consoleReporter)
		}
	}
	e.suiteResult.ExecutionTime = int64(time.Since(e.startTime) / 1e6)
	e.suiteResult.SpecsSkippedCount = len(e.errMaps.specErrs)
	return e.suiteResult
}

func handleHookFailure(result result.Result, execResult *gauge_messages.ProtoExecutionResult, predicate func(result.Result, *gauge_messages.ProtoExecutionResult), reporter reporter.Reporter) {
	predicate(result, execResult)
	result.SetFailure()
	printStatus(execResult, reporter)
}

func getDataTableRows(rowCount int) indexRange {
	if TableRows == "" {
		return indexRange{start: 0, end: rowCount - 1}
	}
	indexes, err := getDataTableRowsRange(TableRows, rowCount)
	if err != nil {
		logger.Error("Table rows validation failed. %s\n", err.Error())
	}
	return indexes
}

func (e *simpleExecution) finish() {
	e.notifyExecutionResult()
	e.stopAllPlugins()
}

func (e *simpleExecution) stopAllPlugins() {
	e.notifyExecutionStop()
	if err := e.runner.Kill(); err != nil {
		e.consoleReporter.Error("Failed to kill Runner: %s", err.Error())
	}
}

func (e *simpleExecution) executeSpec(specificationToExecute *gauge.Specification) {
	executor := newSpecExecutor(specificationToExecute, e.runner, e.pluginHandler, getDataTableRows(specificationToExecute.DataTable.Table.GetRowCount()), e.consoleReporter, e.errMaps)
	protoSpecResult := executor.execute()
	e.suiteResult.AddSpecResult(protoSpecResult)
}
