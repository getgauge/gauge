// This file is part of twist
package main

import "net"

type execution struct {
	manifest       *manifest
	connection     net.Conn
	specifications []*specification
}

func newExecution(manifest *manifest, specifications []*specification, conn net.Conn) *execution {
	e := execution{manifest: manifest, specifications: specifications, connection: conn}
	return &e
}

func (e *execution) startExecution() *ExecutionStatus {
	message := &Message{MessageType: Message_ExecutionStarting.Enum(),
		ExecutionStartingRequest: &ExecutionStartingRequest{}}

	return executeAndGetStatus(e.connection, message)
}

func (e *execution) stopExecution() *ExecutionStatus {
	message := &Message{MessageType: Message_ExecutionEnding.Enum(),
		ExecutionEndingRequest: &ExecutionEndingRequest{}}

	return executeAndGetStatus(e.connection, message)
}

func (e *execution) killProcess() error {
	message := &Message{MessageType: Message_KillProcessRequest.Enum(),
		KillProcessRequest: &KillProcessRequest{}}

	_, err := getResponse(e.connection, message)
	return err
}

type testExecutionStatus struct {
	specifications         []*specification
	specExecutionStatuses  []*specExecutionStatus
	hooksExecutionStatuses []*ExecutionStatus
}

func (t *testExecutionStatus) isFailed() bool {
	if t.hooksExecutionStatuses != nil {
		for _, s := range t.hooksExecutionStatuses {
			if !s.GetPassed() {
				return true
			}
		}
	}

	for _, specStatus := range t.specExecutionStatuses {
		if specStatus.isFailed() {
			return true
		}
	}

	return false
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

func (exe *execution) start() *testExecutionStatus {
	testExecutionStatus := &testExecutionStatus{specifications: exe.specifications}
	beforeSuiteHookExecStatus := exe.startExecution()
	if beforeSuiteHookExecStatus.GetPassed() {
		for _, specificationToExecute := range exe.specifications {
			executor := newSpecExecutor(specificationToExecute, exe.connection)
			specExecutionStatus := executor.execute()
			testExecutionStatus.specifications = append(testExecutionStatus.specifications, specificationToExecute)
			testExecutionStatus.specExecutionStatuses = append(testExecutionStatus.specExecutionStatuses, specExecutionStatus)
		}
	}

	afterSuiteHookExecStatus := exe.stopExecution()
	testExecutionStatus.hooksExecutionStatuses = append(testExecutionStatus.hooksExecutionStatuses, beforeSuiteHookExecStatus, afterSuiteHookExecStatus)

	return testExecutionStatus
}

func newSpecExecutor(specToExecute *specification, connection net.Conn) *specExecutor {
	specExecutor := new(specExecutor)
	specExecutor.initialize(specToExecute, connection)
	return specExecutor
}
