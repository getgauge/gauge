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

func (exe *execution) start() error {
	for _, specificationToExecute := range exe.specifications {
		executor := &specExecutor{specification: specificationToExecute, connection: exe.connection}
		executor.execute()
	}
	return exe.stopExecution()
}
