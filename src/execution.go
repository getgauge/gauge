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

func (exe *execution) start() *testExecutionStatus {
	testExecutionStatus := &testExecutionStatus{specifications: exe.specifications}
	beforeSuiteHookExecStatus := exe.startExecution()
	if beforeSuiteHookExecStatus.GetPassed() {
		for _, specificationToExecute := range exe.specifications {
			executor := &specExecutor{specification: specificationToExecute, connection: exe.connection}
			specExecutionStatus := executor.execute()
			testExecutionStatus.specifications = append(testExecutionStatus.specifications, specificationToExecute)
			testExecutionStatus.specExecutionStatuses = append(testExecutionStatus.specExecutionStatuses, specExecutionStatus)
		}
	}

	afterSuiteHookExecStatus := exe.stopExecution()
	testExecutionStatus.hooksExecutionStatuses = append(testExecutionStatus.hooksExecutionStatuses, beforeSuiteHookExecStatus, afterSuiteHookExecStatus)

	return testExecutionStatus
}
