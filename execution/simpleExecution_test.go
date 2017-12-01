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
	"testing"

	"github.com/getgauge/gauge/execution/result"

	"github.com/getgauge/gauge/gauge_messages"
)

func TestNotifyBeforeSuiteShouldAddsBeforeSuiteHookMessages(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ExecutionStarting {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"Before Suite Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h}
	simpleExecution := newSimpleExecution(ei, false)
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	simpleExecution.notifyBeforeSuite()

	gotMessages := simpleExecution.suiteResult.Message

	if len(gotMessages) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(gotMessages))
	}
	if gotMessages[0] != "Before Suite Called" {
		t.Errorf("Expected `Before Suite Called` message, got : %s", gotMessages[0])
	}
}

func TestNotifyAfterSuiteShouldAddsAfterSuiteHookMessages(t *testing.T) {
	r := &mockRunner{}
	h := &mockPluginHandler{NotifyPluginsfunc: func(m *gauge_messages.Message) {}, GracefullyKillPluginsfunc: func() {}}
	r.ExecuteAndGetStatusFunc = func(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
		if m.MessageType == gauge_messages.Message_ExecutionEnding {
			return &gauge_messages.ProtoExecutionResult{
				Message:       []string{"After Suite Called"},
				Failed:        false,
				ExecutionTime: 10,
			}
		}
		return &gauge_messages.ProtoExecutionResult{}
	}
	ei := &executionInfo{runner: r, pluginHandler: h}
	simpleExecution := newSimpleExecution(ei, false)
	simpleExecution.suiteResult = result.NewSuiteResult(ExecuteTags, simpleExecution.startTime)
	simpleExecution.notifyAfterSuite()

	gotMessages := simpleExecution.suiteResult.Message

	if len(gotMessages) != 1 {
		t.Errorf("Expected 1 message, got : %d", len(gotMessages))
	}
	if gotMessages[0] != "After Suite Called" {
		t.Errorf("Expected `After Suite Called` message, got : %s", gotMessages[0])
	}
}
