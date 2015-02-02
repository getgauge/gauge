// This file is part of twist
package main

import (
	"github.com/getgauge/gauge/gauge_messages"
	"time"
)

type simpleExecution struct {
	manifest             *manifest
	runner               *testRunner
	specifications       []*specification
	pluginHandler        *pluginHandler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	suiteResult          *suiteResult
}

type execution interface {
	start() *suiteResult
	finish()
}

type executionInfo struct {
	currentSpec specification
}

func newExecution(manifest *manifest, specifications []*specification, runner *testRunner, pluginHandler *pluginHandler, inParallel bool) execution {
	if inParallel {
		return &parallelSpecExecution{manifest: manifest, specifications: specifications, runner: runner, pluginHandler: pluginHandler}
	}
	return &simpleExecution{manifest: manifest, specifications: specifications, runner: runner, pluginHandler: pluginHandler}
}

func (e *simpleExecution) startExecution() *(gauge_messages.ProtoExecutionResult) {
	initSuiteDataStoreMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteDataStoreInit.Enum(),
		SuiteDataStoreInitRequest: &gauge_messages.SuiteDataStoreInitRequest{}}
	initResult := executeAndGetStatus(e.runner, initSuiteDataStoreMessage)
	if initResult.GetFailed() {
		log.Warning("Suite data store didn't get initialized")
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
	executionResult := executeAndGetStatus(e.runner, message)
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
	beforeSuiteHookExecResult := exe.startExecution()
	if beforeSuiteHookExecResult.GetFailed() {
		addPreHook(exe.suiteResult, beforeSuiteHookExecResult)
		exe.suiteResult.setFailure()
	} else {
		for _, specificationToExecute := range exe.specifications {
			executor := newSpecExecutor(specificationToExecute, exe.runner, exe.pluginHandler)
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
	if err := e.runner.kill(); err != nil {
		log.Error("Failed to kill Runner. %s\n", err.Error())
	}
}

func newSpecExecutor(specToExecute *specification, runner *testRunner, pluginHandler *pluginHandler) *specExecutor {
	specExecutor := new(specExecutor)
	specExecutor.initialize(specToExecute, runner, pluginHandler)
	return specExecutor
}
