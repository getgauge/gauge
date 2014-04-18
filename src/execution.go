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

func (e *execution) stopExecution() error {
	message := &Message{MessageType: Message_ExecutionEnding.Enum(),
		ExecutionEndingRequest: &ExecutionEndingRequest{}}

	_, err := getResponse(e.connection, message)
	if err != nil {
		return err
	}

	return nil
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

//TODO: Before execution and after execution hooks
func (exe *execution) start() *testExecutionStatus {
	testExecutionStatus := &testExecutionStatus{specifications: exe.specifications}
	for _, specificationToExecute := range exe.specifications {
		executor := &specExecutor{specification: specificationToExecute, connection: exe.connection}
		specExecutionStatus := executor.execute()
		testExecutionStatus.specifications = append(testExecutionStatus.specifications, specificationToExecute)
		testExecutionStatus.specExecutionStatuses = append(testExecutionStatus.specExecutionStatuses, specExecutionStatus)
	}
	//TODO: error check when hooks are in place
	exe.stopExecution()

	return testExecutionStatus
}
