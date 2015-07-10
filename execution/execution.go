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
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	"path/filepath"
	"time"
)

type simpleExecution struct {
	manifest             *manifest
	runner               *testRunner
	specifications       []*parser.Specification
	pluginHandler        *pluginHandler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	suiteResult          *SuiteResult
	writer               executionLogger
}

type execution interface {
	start() *suiteResult
	finish()
}

type executionInfo struct {
	currentSpec specification
}

func newExecution(manifest *manifest, specifications []*parser.Specification, runner *testRunner, pluginHandler *pluginHandler, info *parallelInfo, writer executionLogger) execution {
	if info.inParallel {
		return &parallelSpecExecution{manifest: manifest, specifications: specifications, runner: runner, pluginHandler: pluginHandler, numberOfExecutionStreams: info.numberOfStreams, writer: writer}
	}
	return &simpleExecution{manifest: manifest, specifications: specifications, runner: runner, pluginHandler: pluginHandler, writer: writer}
}

func (e *simpleExecution) startExecution() *(gauge_messages.ProtoExecutionResult) {
	initSuiteDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteDataStoreInit.Enum(),
		SuiteDataStoreInitRequest: &gauge_messages.SuiteDataStoreInitRequest{}}
	initResult := executeAndGetStatus(e.runner, initSuiteDataStoreMessage, e.writer)
	if initResult.GetFailed() {
		e.writer.Warning("Suite data store didn't get initialized")
	}
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionStarting.Enum(),
		ExecutionStartingRequest: &gauge_messages.ExecutionStartingRequest{}}
	return e.executeHook(message)
}

func (e *simpleExecution) endExecution() *(gauge_messages.ProtoExecutionResult) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionEnding.Enum(),
		ExecutionEndingRequest: &gauge_messages.ExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message)
}

func (e *simpleExecution) executeHook(message *gauge_messages.Message) *(gauge_messages.ProtoExecutionResult) {
	e.pluginHandler.notifyPlugins(message)
	executionResult := executeAndGetStatus(e.runner, message, e.writer)
	e.addExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func (e *simpleExecution) addExecTime(execTime int64) {
	e.suiteResult.executionTime += execTime
}

func (e *simpleExecution) notifyExecutionResult() {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{SuiteResult: convertToProtoSuiteResult(e.suiteResult)}}
	e.pluginHandler.notifyPlugins(message)
}

func (e *simpleExecution) notifyExecutionStop() {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_KillProcessRequest.Enum(),
		KillProcessRequest: &gauge_messages.KillProcessRequest{}}
	e.pluginHandler.notifyPlugins(message)
	e.pluginHandler.gracefullyKillPlugins()
}

func (e *simpleExecution) killPlugins() {
	e.pluginHandler.gracefullyKillPlugins()
}

func (exe *simpleExecution) start() *suiteResult {
	startTime := time.Now()
	exe.suiteResult = newSuiteResult()
	exe.suiteResult.timestamp = startTime.Format(config.LayoutForTimeStamp)
	exe.suiteResult.projectName = filepath.Base(config.ProjectRoot)
	exe.suiteResult.environment = env.CurrentEnv
	exe.suiteResult.Tags = *executeTags
	beforeSuiteHookExecResult := exe.startExecution()
	if beforeSuiteHookExecResult.GetFailed() {
		addPreHook(exe.suiteResult, beforeSuiteHookExecResult)
		exe.suiteResult.setFailure()
	} else {
		for _, specificationToExecute := range exe.specifications {
			executor := newSpecExecutor(specificationToExecute, exe.runner, exe.pluginHandler, exe.writer, getDataTableRows(specificationToExecute.DataTable.table.getRowCount()))
			protoSpecResult := executor.execute()
			exe.suiteResult.addSpecResult(protoSpecResult)
		}
	}
	afterSuiteHookExecResult := exe.endExecution()
	if afterSuiteHookExecResult.GetFailed() {
		addPostHook(exe.suiteResult, afterSuiteHookExecResult)
		exe.suiteResult.setFailure()
	}
	exe.suiteResult.executionTime = int64(time.Since(startTime) / 1e6)
	return exe.suiteResult
}

func (exe *simpleExecution) finish() {
	exe.notifyExecutionResult()
	exe.stopAllPlugins()
}

func (e *simpleExecution) stopAllPlugins() {
	e.notifyExecutionStop()
	if err := e.runner.kill(e.writer); err != nil {
		e.writer.Error("Failed to kill Runner. %s\n", err.Error())
	}
}

func newSpecExecutor(specToExecute *parser.Specification, runner *testRunner, pluginHandler *pluginHandler, writer executionLogger, tableRows indexRange) *specExecutor {
	specExecutor := new(specExecutor)
	specExecutor.initialize(specToExecute, runner, pluginHandler, writer, tableRows)
	return specExecutor
}
