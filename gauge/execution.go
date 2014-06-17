// This file is part of twist
package main

import "net"

type execution struct {
	manifest             *manifest
	connection           net.Conn
	specifications       []*specification
	pluginHandler        *pluginHandler
	currentExecutionInfo *ExecutionInfo
	suiteResult          *suiteResult
}

type executionInfo struct {
	currentSpec specification
}

func newExecution(manifest *manifest, specifications []*specification, conn net.Conn, pluginHandler *pluginHandler) *execution {
	e := execution{manifest: manifest, specifications: specifications, connection: conn, pluginHandler: pluginHandler}
	return &e
}

func (e *execution) startExecution() *ProtoExecutionResult {
	message := &Message{MessageType: Message_ExecutionStarting.Enum(),
		ExecutionStartingRequest: &ExecutionStartingRequest{}}

	e.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(e.connection, message)
}

func (e *execution) endExecution() *ProtoExecutionResult {
	message := &Message{MessageType: Message_ExecutionEnding.Enum(),
		ExecutionEndingRequest: &ExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}

	e.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(e.connection, message)
}

func (e *execution) notifyExecutionResult() {
	message := &Message{MessageType: Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &SuiteExecutionResult{SuiteResult: convertToProtoSuiteResult(e.suiteResult)}}
	e.pluginHandler.notifyPlugins(message)
}

func (e *execution) notifyExecutionStop() {
	message := &Message{MessageType: Message_KillProcessRequest.Enum(),
		KillProcessRequest: &KillProcessRequest{}}

	e.pluginHandler.notifyPlugins(message)
	e.pluginHandler.gracefullyKillPlugins()
}

func (e *execution) killProcess() error {
	message := &Message{MessageType: Message_KillProcessRequest.Enum(),
		KillProcessRequest: &KillProcessRequest{}}

	_, err := getResponse(e.connection, message)
	return err
}

func (e *execution) killPlugins() {
	e.pluginHandler.gracefullyKillPlugins()
}

type executionValidationErrors map[*specification][]*stepValidationError

func (exe *execution) validate(conceptDictionary *conceptDictionary) executionValidationErrors {
	validationStatus := make(executionValidationErrors)
	for _, spec := range exe.specifications {
		executor := &specExecutor{specification: spec, connection: exe.connection, conceptDictionary: conceptDictionary}
		validationErrors := executor.validateSpecification()
		if len(validationErrors) != 0 {
			validationStatus[spec] = validationErrors
		}
	}
	if len(validationStatus) > 0 {
		return validationStatus
	} else {
		return nil
	}
}

func (exe *execution) start() *suiteResult {
	beforeSuiteHookExecStatus := exe.startExecution()
	exe.suiteResult = newSuiteResult()
	if beforeSuiteHookExecStatus.GetFailed() {
		addPreHook(exe.suiteResult, beforeSuiteHookExecStatus)
	} else {
		for _, specificationToExecute := range exe.specifications {
			executor := newSpecExecutor(specificationToExecute, exe.connection, exe.pluginHandler)
			protoSpecResult := executor.execute()
			exe.suiteResult.addSpecResult(protoSpecResult)
		}
	}

	addPostHook(exe.suiteResult, exe.endExecution())

	exe.notifyExecutionResult()
	exe.notifyExecutionStop()
	return exe.suiteResult
}

func newSpecExecutor(specToExecute *specification, connection net.Conn, pluginHandler *pluginHandler) *specExecutor {
	specExecutor := new(specExecutor)
	specExecutor.initialize(specToExecute, connection, pluginHandler)
	return specExecutor
}
