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

package result

import "github.com/getgauge/gauge/gauge_messages"

// Result represents execution result
type Result interface {
	GetPreHook() **(gauge_messages.ProtoHookFailure)
	GetPostHook() **(gauge_messages.ProtoHookFailure)
	SetFailure()
	GetFailed() bool
	item() interface{}
	ExecTime() int64
}

// ExecTimeTracker is an interface for tracking execution time
type ExecTimeTracker interface {
	AddExecTime(int64)
}

// GetProtoHookFailure returns the failure result of hook execution
func GetProtoHookFailure(executionResult *gauge_messages.ProtoExecutionResult) *(gauge_messages.ProtoHookFailure) {
	return &gauge_messages.ProtoHookFailure{StackTrace: executionResult.StackTrace, ErrorMessage: executionResult.ErrorMessage, ScreenShot: executionResult.ScreenShot}
}

// AddPreHook adds the before hook execution result to the actual result object
func AddPreHook(result Result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.GetPreHook()) = GetProtoHookFailure(executionResult)
		result.SetFailure()
	}
}

// AddPostHook adds the after hook execution result to the actual result object
func AddPostHook(result Result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.GetPostHook()) = GetProtoHookFailure(executionResult)
		result.SetFailure()
	}
}
