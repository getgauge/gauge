/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package refactor

import (
	"net"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/runner"
)

type mockRunner struct {
	response *gauge_messages.Message
}

func (r *mockRunner) ExecuteAndGetStatus(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
	return nil
}

func (r *mockRunner) ExecuteMessageWithTimeout(m *gauge_messages.Message) (*gauge_messages.Message, error) {
	return r.response, nil
}

func (r *mockRunner) Alive() bool {
	return true
}

func (r *mockRunner) Kill() error {
	return nil
}

func (r *mockRunner) Connection() net.Conn {
	return nil
}

func (r *mockRunner) IsMultithreaded() bool {
	return false
}

func (r *mockRunner) Info() *runner.RunnerInfo {
	return &runner.RunnerInfo{Killed: false}
}

func (r *mockRunner) Pid() int {
	return 0
}
