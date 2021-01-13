/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package result

import "github.com/getgauge/gauge-proto/go/gauge_messages"

// Result represents execution result
type Result interface {
	GetPreHook() []*gauge_messages.ProtoHookFailure
	GetPostHook() []*gauge_messages.ProtoHookFailure
	GetFailed() bool

	AddPreHook(...*gauge_messages.ProtoHookFailure)
	AddPostHook(...*gauge_messages.ProtoHookFailure)
	SetFailure()

	Item() interface{}
	ExecTime() int64
}

// ExecTimeTracker is an interface for tracking execution time
type ExecTimeTracker interface {
	AddExecTime(int64)
}

// GetProtoHookFailure returns the failure result of hook execution
func GetProtoHookFailure(executionResult *gauge_messages.ProtoExecutionResult) *(gauge_messages.ProtoHookFailure) {
	return &gauge_messages.ProtoHookFailure{
		StackTrace:            executionResult.StackTrace,
		ErrorMessage:          executionResult.ErrorMessage,
		FailureScreenshot:     executionResult.FailureScreenshot,
		FailureScreenshotFile: executionResult.FailureScreenshotFile,
		TableRowIndex:         -1,
	}
}

// AddPreHook adds the before hook execution result to the actual result object
func AddPreHook(result Result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		result.AddPreHook(GetProtoHookFailure(executionResult))
		result.SetFailure()
	}
}

// AddPostHook adds the after hook execution result to the actual result object
func AddPostHook(result Result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		result.AddPostHook(GetProtoHookFailure(executionResult))
		result.SetFailure()
	}
}
