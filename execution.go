// This file is part of twist
package main

import "fmt"

type simpleExecution struct {
	manifest             *manifest
	runner               *testRunner
	specifications       []*specification
	pluginHandler        *pluginHandler
	currentExecutionInfo *ExecutionInfo
	suiteResult          *suiteResult
}

type execution interface {
	start() *suiteResult
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

func (e *simpleExecution) startExecution() *ProtoExecutionResult {
	initSuiteDataStoreMessage := &Message{MessageType: Message_SuiteDataStoreInit.Enum(),
		SuiteDataStoreInitRequest: &SuiteDataStoreInitRequest{}}
	initResult := executeAndGetStatus(e.runner, initSuiteDataStoreMessage)
	if initResult.GetFailed() {
		fmt.Println("[Warning] Suite data store didn't get initialized")
	}
	message := &Message{MessageType: Message_ExecutionStarting.Enum(),
		ExecutionStartingRequest: &ExecutionStartingRequest{}}
	return e.executeHook(message)
}

func (e *simpleExecution) endExecution() *ProtoExecutionResult {
	message := &Message{MessageType: Message_ExecutionEnding.Enum(),
		ExecutionEndingRequest: &ExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message)
}

func (e *simpleExecution) executeHook(message *Message) *ProtoExecutionResult {
	e.pluginHandler.notifyPlugins(message)
	executionResult := executeAndGetStatus(e.runner, message)
	e.addExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func (e *simpleExecution) addExecTime(execTime int64) {
	e.suiteResult.executionTime += execTime
}

func (e *simpleExecution) notifyExecutionResult() {
	message := &Message{MessageType: Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &SuiteExecutionResult{SuiteResult: convertToProtoSuiteResult(e.suiteResult)}}
	e.pluginHandler.notifyPlugins(message)
}

func (e *simpleExecution) notifyExecutionStop() {
	message := &Message{MessageType: Message_KillProcessRequest.Enum(),
		KillProcessRequest: &KillProcessRequest{}}
	e.pluginHandler.notifyPlugins(message)
	e.pluginHandler.gracefullyKillPlugins()
}

func (e *simpleExecution) killPlugins() {
	e.pluginHandler.gracefullyKillPlugins()
}

func (exe *simpleExecution) start() *suiteResult {
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
	exe.notifyExecutionResult()
	exe.stopAllPlugins()
	return exe.suiteResult
}

func (e *simpleExecution) stopAllPlugins() {
	e.notifyExecutionStop()
	if err := e.runner.kill(); err != nil {
		fmt.Printf("[Error] Failed to kill Runner. %s\n", err.Error())
	}
}

func newSpecExecutor(specToExecute *specification, runner *testRunner, pluginHandler *pluginHandler) *specExecutor {
	specExecutor := new(specExecutor)
	specExecutor.initialize(specToExecute, runner, pluginHandler)
	return specExecutor
}
