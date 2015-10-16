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
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/logger/execLogger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
	"path/filepath"
	"time"
)

var ExecuteTags = ""
var TableRows = ""

type simpleExecution struct {
	manifest             *manifest.Manifest
	runner               *runner.TestRunner
	specifications       []*parser.Specification
	pluginHandler        *plugin.PluginHandler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	suiteResult          *result.SuiteResult
	logger               *logger.GaugeLogger
	errMaps              *validationErrMaps
}

type execution interface {
	start() *result.SuiteResult
	finish()
}

type executionInfo struct {
	manifest        *manifest.Manifest
	specifications  []*parser.Specification
	runner          *runner.TestRunner
	pluginHandler   *plugin.PluginHandler
	parallelRunInfo *parallelInfo
	execLogger      *logger.GaugeLogger
	errMaps         *validationErrMaps
}

func newExecution(executionInfo *executionInfo) execution {
	if executionInfo.parallelRunInfo.inParallel {
		return &parallelSpecExecution{manifest: executionInfo.manifest, specifications: executionInfo.specifications,
			runner: executionInfo.runner, pluginHandler: executionInfo.pluginHandler,
			numberOfExecutionStreams: executionInfo.parallelRunInfo.numberOfStreams,
			logger: executionInfo.execLogger, errMaps: executionInfo.errMaps}
	}
	return &simpleExecution{manifest: executionInfo.manifest, specifications: executionInfo.specifications,
		runner: executionInfo.runner, pluginHandler: executionInfo.pluginHandler, logger: executionInfo.execLogger, errMaps: executionInfo.errMaps}
}

func newSimpleExecution(executionInfo *executionInfo) *simpleExecution {
	return &simpleExecution{manifest: executionInfo.manifest, specifications: executionInfo.specifications,
		runner: executionInfo.runner, pluginHandler: executionInfo.pluginHandler, logger: executionInfo.execLogger, errMaps: executionInfo.errMaps}
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
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{SuiteResult: parser.ConvertToProtoSuiteResult(e.suiteResult)}}
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

func (exe *simpleExecution) start() *result.SuiteResult {
	startTime := time.Now()
	exe.suiteResult = result.NewSuiteResult()
	exe.suiteResult.Timestamp = startTime.Format(config.LayoutForTimeStamp)
	exe.suiteResult.ProjectName = filepath.Base(config.ProjectRoot)
	exe.suiteResult.Environment = env.CurrentEnv
	exe.suiteResult.Tags = ExecuteTags
	initSuiteDataStoreResult := exe.initializeSuiteDataStore()
	if initSuiteDataStoreResult.GetFailed() {
		exe.logger.Error("Failed to initialize suite datastore. Error: %s", initSuiteDataStoreResult.GetErrorMessage())
	} else {
		beforeSuiteHookExecResult := exe.startExecution()
		if beforeSuiteHookExecResult.GetFailed() {
			result.AddPreHook(exe.suiteResult, beforeSuiteHookExecResult)
			exe.suiteResult.SetFailure()
			printStatus(beforeSuiteHookExecResult, exe.logger)
		} else {
			for _, specificationToExecute := range exe.specifications {
				exe.executeSpec(specificationToExecute)
			}
		}
		afterSuiteHookExecResult := exe.endExecution()
		if afterSuiteHookExecResult.GetFailed() {
			result.AddPostHook(exe.suiteResult, afterSuiteHookExecResult)
			exe.suiteResult.SetFailure()
			printStatus(afterSuiteHookExecResult, exe.logger)
		}
	}
	exe.suiteResult.ExecutionTime = int64(time.Since(startTime) / 1e6)
	return exe.suiteResult
}

func getDataTableRows(rowCount int) indexRange {
	if TableRows == "" {
		return indexRange{start: 0, end: rowCount - 1}
	}
	indexes, err := getDataTableRowsRange(TableRows, rowCount)
	if err != nil {
		execLogger.CriticalError(errors.New(fmt.Sprintf("Table rows validation failed. %s\n", err.Error())))
	}
	return indexes
}

func (exe *simpleExecution) finish() {
	exe.notifyExecutionResult()
	exe.stopAllPlugins()
}

func (e *simpleExecution) stopAllPlugins() {
	e.notifyExecutionStop()
	if err := e.runner.Kill(); err != nil {
		e.logger.Error("Failed to kill Runner: %s", err.Error())
	}
}

func newSpecExecutor(specToExecute *parser.Specification, runner *runner.TestRunner, pluginHandler *plugin.PluginHandler, tableRows indexRange, logger *logger.GaugeLogger, errMaps *validationErrMaps) *specExecutor {
	specExecutor := new(specExecutor)
	specExecutor.initialize(specToExecute, runner, pluginHandler, tableRows, logger, errMaps)
	return specExecutor
}

func (exe *simpleExecution) executeStream(specs *specList) *result.SuiteResult {
	startTime := time.Now()
	exe.suiteResult = result.NewSuiteResult()
	exe.suiteResult.Timestamp = startTime.Format(config.LayoutForTimeStamp)
	exe.suiteResult.ProjectName = filepath.Base(config.ProjectRoot)
	exe.suiteResult.Environment = env.CurrentEnv
	exe.suiteResult.Tags = ExecuteTags
	initSuiteDataStoreResult := exe.initializeSuiteDataStore()
	if initSuiteDataStoreResult.GetFailed() {
		exe.logger.Error("Failed to initialize suite datastore. Error: %s", initSuiteDataStoreResult.GetErrorMessage())
	} else {
		beforeSuiteHookExecResult := exe.startExecution()
		if beforeSuiteHookExecResult.GetFailed() {
			result.AddPreHook(exe.suiteResult, beforeSuiteHookExecResult)
			exe.suiteResult.SetFailure()
		} else {
			for !specs.isEmpty() {
				exe.executeSpec(specs.getSpec())
			}
		}
		afterSuiteHookExecResult := exe.endExecution()
		if afterSuiteHookExecResult.GetFailed() {
			result.AddPostHook(exe.suiteResult, afterSuiteHookExecResult)
			exe.suiteResult.SetFailure()
		}
	}
	exe.suiteResult.ExecutionTime = int64(time.Since(startTime) / 1e6)
	return exe.suiteResult
}

func (exe *simpleExecution) executeSpec(specificationToExecute *parser.Specification) {
	executor := newSpecExecutor(specificationToExecute, exe.runner, exe.pluginHandler, getDataTableRows(specificationToExecute.DataTable.Table.GetRowCount()), exe.logger, exe.errMaps)
	protoSpecResult := executor.execute()
	exe.suiteResult.AddSpecResult(protoSpecResult)
}
