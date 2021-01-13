/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package result

import "github.com/getgauge/gauge-proto/go/gauge_messages"

// ConceptResult represents result of a concept execution
type ConceptResult struct {
	ProtoConcept *gauge_messages.ProtoConcept
}

// NewConceptResult creates a ConceptResult with given ProtoConcept
func NewConceptResult(con *gauge_messages.ProtoConcept) *ConceptResult {
	return &ConceptResult{ProtoConcept: con}
}

// SetFailure sets the conceptResult as failed
func (conceptResult *ConceptResult) SetFailure() {
	conceptResult.ProtoConcept.ConceptExecutionResult.ExecutionResult.Failed = true
}

// GetFailed returns the state of the concept result
func (conceptResult *ConceptResult) GetFailed() bool {
	return conceptResult.ProtoConcept.GetConceptExecutionResult().GetExecutionResult().GetFailed()
}

// GetRecoverable returns the state of the concept result
func (conceptResult *ConceptResult) GetRecoverable() bool {
	return conceptResult.ProtoConcept.GetConceptExecutionResult().GetExecutionResult().GetRecoverableError()
}

// ExecTime returns the time taken for concept execution
func (conceptResult *ConceptResult) ExecTime() int64 {
	return conceptResult.ProtoConcept.GetConceptExecutionResult().GetExecutionResult().GetExecutionTime()
}

// SetConceptExecResult sets the conceptExecResult as result of concept execution as well as result of ConceptStep
func (conceptResult *ConceptResult) SetConceptExecResult(conceptExecResult *gauge_messages.ProtoStepExecutionResult) {
	conceptResult.ProtoConcept.ConceptExecutionResult = conceptExecResult
	conceptResult.ProtoConcept.ConceptStep.StepExecutionResult = conceptExecResult
}

// UpdateConceptExecResult sets the result of Concept execution
func (conceptResult *ConceptResult) UpdateConceptExecResult() {
	var failed, recoverable bool
	protoConcept := conceptResult.ProtoConcept
	var conceptExecutionTime int64
	for _, step := range protoConcept.GetSteps() {
		if step.GetItemType() == gauge_messages.ProtoItem_Concept {
			stepExecResult := step.GetConcept().GetConceptExecutionResult().GetExecutionResult()
			conceptExecutionTime += stepExecResult.GetExecutionTime()
			if step.GetConcept().GetConceptExecutionResult().GetExecutionResult().GetFailed() {
				failed = true
				conceptExecutionResult := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: step.GetConcept().GetConceptExecutionResult().GetExecutionResult(), Skipped: false}
				conceptExecutionResult.ExecutionResult.ExecutionTime = conceptExecutionTime
				protoConcept.ConceptExecutionResult = conceptExecutionResult
				protoConcept.ConceptStep.StepExecutionResult = conceptExecutionResult
				recoverable = step.GetConcept().GetConceptExecutionResult().GetExecutionResult().GetRecoverableError()
				if !recoverable {
					return
				}
			}
		} else if step.GetItemType() == gauge_messages.ProtoItem_Step {
			stepExecResult := step.GetStep().GetStepExecutionResult().GetExecutionResult()
			conceptExecutionTime += stepExecResult.GetExecutionTime()
			if stepExecResult.GetFailed() {
				failed = true
				conceptExecutionResult := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: stepExecResult, Skipped: false}
				conceptExecutionResult.ExecutionResult.ExecutionTime = conceptExecutionTime
				protoConcept.ConceptExecutionResult = conceptExecutionResult
				protoConcept.ConceptStep.StepExecutionResult = conceptExecutionResult
				recoverable = step.GetStep().GetStepExecutionResult().GetExecutionResult().GetRecoverableError()
				if !recoverable {
					return
				}
			}
		}
	}

	conceptResult.SetConceptExecResult(&gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{RecoverableError: recoverable, Failed: failed, ExecutionTime: conceptExecutionTime}})
	protoConcept.ConceptStep.StepExecutionResult.Skipped = false
}

func (conceptResult *ConceptResult) GetPreHook() []*gauge_messages.ProtoHookFailure {
	return nil
}

func (conceptResult *ConceptResult) GetPostHook() []*gauge_messages.ProtoHookFailure {
	return nil
}

func (conceptResult *ConceptResult) AddPreHook(_ ...*gauge_messages.ProtoHookFailure) {
}

func (conceptResult *ConceptResult) AddPostHook(_ ...*gauge_messages.ProtoHookFailure) {
}

func (conceptResult *ConceptResult) Item() interface{} {
	return conceptResult.ProtoConcept
}
