/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"net"
	"testing"
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/runner"
)

func TestSuiteHooksAreExecutedOncePerRun(t *testing.T) {
	specs := createSpecsList(6)
	var receivedMesseges []*gauge_messages.Message
	runner1 := &fakeGrpcRunner{messageCount: make(map[gauge_messages.Message_MessageType]int)}
	runner2 := &fakeGrpcRunner{messageCount: make(map[gauge_messages.Message_MessageType]int)}
	e := parallelExecution{
		numberOfExecutionStreams: 5,
		specCollection:           gauge.NewSpecCollection(specs, false),
		runners:                  []runner.Runner{runner1, runner2},
		pluginHandler: &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {
			receivedMesseges = append(receivedMesseges, m)
		}},
	}

	t.Run("BeforeSuite", func(t *testing.T) {
		receivedMesseges = []*gauge_messages.Message{}
		e.suiteResult = result.NewSuiteResult("", time.Now())
		runner1.mockResult = &gauge_messages.ProtoExecutionResult{}
		e.notifyBeforeSuite()
		r1count := runner1.messageCount[gauge_messages.Message_ExecutionStarting]
		if r1count != 1 {
			t.Errorf("Expected runner1 to have received 1 ExecutionStarting request, got %d", r1count)
		}
		r2count := runner2.messageCount[gauge_messages.Message_ExecutionStarting]
		if r2count != 0 {
			t.Errorf("Expected runner2 to have received 0 ExecutionStarting request, got %d", r2count)
		}
		if len(receivedMesseges) != 2 {
			t.Errorf("Expected plugins to have received 2 ExecutionStarting notifications, got %d", len(receivedMesseges))
		}
	})

	t.Run("AfterSuite", func(t *testing.T) {
		receivedMesseges = []*gauge_messages.Message{}
		e.notifyAfterSuite()
		e.suiteResult = result.NewSuiteResult("", time.Now())
		runner1.mockResult = &gauge_messages.ProtoExecutionResult{}
		r1count := runner1.messageCount[gauge_messages.Message_ExecutionEnding]
		if r1count != 1 {
			t.Errorf("Expected runner1 to have received 1 ExecutionEnding request, got %d", r1count)
		}
		r2count := runner2.messageCount[gauge_messages.Message_ExecutionEnding]
		if r2count != 0 {
			t.Errorf("Expected runner2 to have received 0 ExecutionEnding request, got %d", r2count)
		}
		if len(receivedMesseges) != 2 {
			t.Errorf("Expected plugins to have received 2 ExecutionStarting notifications, got %d", len(receivedMesseges))
		}
	})
}

type fakeGrpcRunner struct {
	isMultiThreaded bool
	messageCount    map[gauge_messages.Message_MessageType]int
	mockResult      *gauge_messages.ProtoExecutionResult
}

func (f *fakeGrpcRunner) ExecuteMessageWithTimeout(m *gauge_messages.Message) (*gauge_messages.Message, error) {
	return nil, nil
}

func (f *fakeGrpcRunner) ExecuteAndGetStatus(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
	f.messageCount[m.MessageType]++
	return f.mockResult
}
func (f *fakeGrpcRunner) Alive() bool {
	return false
}
func (f *fakeGrpcRunner) Kill() error {
	return nil
}
func (f *fakeGrpcRunner) Connection() net.Conn {
	return nil
}
func (f *fakeGrpcRunner) IsMultithreaded() bool {
	return f.isMultiThreaded
}

func (f *fakeGrpcRunner) Info() *runner.RunnerInfo {
	return nil
}
func (f *fakeGrpcRunner) Pid() int {
	return 0
}
