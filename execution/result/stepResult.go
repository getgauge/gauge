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

// StepResult represents the result of step execution
type StepResult struct {
	protoStep  *gauge_messages.ProtoStep
	StepFailed bool
}

// NewStepResult is a constructor for StepResult
func NewStepResult(ps *gauge_messages.ProtoStep) *StepResult {
	return &StepResult{protoStep: ps}
}

func (s *StepResult) GetPreHook() []*gauge_messages.ProtoHookFailure {
	if s.protoStep.StepExecutionResult.PreHookFailure == nil {
		return []*gauge_messages.ProtoHookFailure{}
	}
	return []*gauge_messages.ProtoHookFailure{s.protoStep.StepExecutionResult.PreHookFailure}
}

func (s *StepResult) GetPostHook() []*gauge_messages.ProtoHookFailure {
	if s.protoStep.StepExecutionResult.PostHookFailure == nil {
		return []*gauge_messages.ProtoHookFailure{}
	}
	return []*gauge_messages.ProtoHookFailure{s.protoStep.StepExecutionResult.PostHookFailure}
}

func (s *StepResult) AddPreHook(f ...*gauge_messages.ProtoHookFailure) {
	s.protoStep.StepExecutionResult.PreHookFailure = f[0]
}

func (s *StepResult) AddPostHook(f ...*gauge_messages.ProtoHookFailure) {
	s.protoStep.StepExecutionResult.PostHookFailure = f[0]
}

// SetFailure sets the result to failed
func (s *StepResult) SetFailure() {
	s.protoStep.StepExecutionResult.ExecutionResult.Failed = true
}

// GetFailed returns the state of the result
func (s *StepResult) GetFailed() bool {
	return s.protoStep.StepExecutionResult.ExecutionResult.GetFailed()
}

// GetFailed returns true if the actual step failed, and not step hook.
func (s *StepResult) GetStepFailed() bool {
	return s.StepFailed
}

// GetStackTrace returns the stacktrace for step failure
func (s *StepResult) GetStackTrace() string {
	return s.protoStep.GetStepExecutionResult().GetExecutionResult().GetStackTrace()
}

// GetErrorMessage returns the error message for step failure
func (s *StepResult) GetErrorMessage() string {
	return s.protoStep.GetStepExecutionResult().GetExecutionResult().GetErrorMessage()
}

// GetStepActualText returns the Actual text of step from step result
func (s *StepResult) GetStepActualText() string {
	return s.protoStep.GetActualText()
}

// SetStepFailure sets the actual step as failed. StepResult.ProtoStep.GetFailed() returns true even if hook failed and not actual step.
func (s *StepResult) SetStepFailure() {
	s.StepFailed = true
}

func (s *StepResult) Item() interface{} {
	return s.protoStep
}

// ExecTime returns the time taken to execute the step
func (s *StepResult) ExecTime() int64 {
	return s.protoStep.StepExecutionResult.ExecutionResult.GetExecutionTime()
}

// AddExecTime increments the execution time by the given value
func (s *StepResult) AddExecTime(t int64) {
	currentTime := s.protoStep.StepExecutionResult.ExecutionResult.GetExecutionTime()
	if s.protoStep.StepExecutionResult.ExecutionResult == nil {
		s.protoStep.StepExecutionResult.ExecutionResult = &gauge_messages.ProtoExecutionResult{Failed: false}
	}

	s.protoStep.StepExecutionResult.ExecutionResult.ExecutionTime = currentTime + t
}

// ProtoStepExecResult returns the step execution result used at the proto layer
func (s *StepResult) ProtoStepExecResult() *gauge_messages.ProtoStepExecutionResult {
	return s.protoStep.StepExecutionResult
}

// SetProtoExecResult sets the execution result
func (s *StepResult) SetProtoExecResult(r *gauge_messages.ProtoExecutionResult) {
	s.protoStep.StepExecutionResult.ExecutionResult = r
}
