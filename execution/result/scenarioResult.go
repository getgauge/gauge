/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package result

import (
	"github.com/getgauge/gauge-proto/go/gauge_messages"
)

type ScenarioResult struct {
	ProtoScenario             *gauge_messages.ProtoScenario
	ScenarioDataTableRow      *gauge_messages.ProtoTable
	ScenarioDataTableRowIndex int
	ScenarioDataTable         *gauge_messages.ProtoTable
}

func NewScenarioResult(sce *gauge_messages.ProtoScenario) *ScenarioResult {
	return &ScenarioResult{ProtoScenario: sce}
}

// SetFailure sets the scenarioResult as failed
func (s ScenarioResult) SetFailure() {
	s.ProtoScenario.ExecutionStatus = gauge_messages.ExecutionStatus_FAILED
	s.ProtoScenario.Failed = true
}

// GetFailed returns the state of the scenario result
func (s ScenarioResult) GetFailed() bool {
	return s.ProtoScenario.GetExecutionStatus() == gauge_messages.ExecutionStatus_FAILED
}

func (s ScenarioResult) AddItems(protoItems []*gauge_messages.ProtoItem) {
	s.ProtoScenario.ScenarioItems = append(s.ProtoScenario.ScenarioItems, protoItems...)
}

func (s ScenarioResult) AddContexts(contextProtoItems []*gauge_messages.ProtoItem) {
	s.ProtoScenario.Contexts = append(s.ProtoScenario.Contexts, contextProtoItems...)
}

func (s ScenarioResult) AddTearDownSteps(tearDownProtoItems []*gauge_messages.ProtoItem) {
	s.ProtoScenario.TearDownSteps = append(s.ProtoScenario.TearDownSteps, tearDownProtoItems...)
}

func (s ScenarioResult) UpdateExecutionTime() {
	s.updateExecutionTimeFromItems(s.ProtoScenario.GetContexts())
	s.updateExecutionTimeFromItems(s.ProtoScenario.GetScenarioItems())
}

func (s ScenarioResult) AddExecTime(execTime int64) {
	currentScenarioExecTime := s.ProtoScenario.GetExecutionTime()
	s.ProtoScenario.ExecutionTime = currentScenarioExecTime + execTime
}

// ExecTime returns the time taken for scenario execution
func (s ScenarioResult) ExecTime() int64 {
	return s.ProtoScenario.ExecutionTime
}

func (s ScenarioResult) updateExecutionTimeFromItems(protoItems []*gauge_messages.ProtoItem) {
	for _, item := range protoItems {
		if item.GetItemType() == gauge_messages.ProtoItem_Step {
			stepExecTime := item.GetStep().GetStepExecutionResult().GetExecutionResult().GetExecutionTime()
			s.AddExecTime(stepExecTime)
		} else if item.GetItemType() == gauge_messages.ProtoItem_Concept {
			conceptExecTime := item.GetConcept().GetConceptExecutionResult().GetExecutionResult().GetExecutionTime()
			s.AddExecTime(conceptExecTime)
		}
	}
}

func (s ScenarioResult) GetPreHook() []*gauge_messages.ProtoHookFailure {
	if s.ProtoScenario.PreHookFailure == nil {
		return []*gauge_messages.ProtoHookFailure{}
	}
	return []*gauge_messages.ProtoHookFailure{s.ProtoScenario.PreHookFailure}
}

func (s ScenarioResult) GetPostHook() []*gauge_messages.ProtoHookFailure {
	if s.ProtoScenario.PostHookFailure == nil {
		return []*gauge_messages.ProtoHookFailure{}
	}
	return []*gauge_messages.ProtoHookFailure{s.ProtoScenario.PostHookFailure}
}

func (s ScenarioResult) AddPreHook(f ...*gauge_messages.ProtoHookFailure) {
	s.ProtoScenario.PreHookFailure = f[0]
}

func (s ScenarioResult) AddPostHook(f ...*gauge_messages.ProtoHookFailure) {
	s.ProtoScenario.PostHookFailure = f[0]
}

func (s ScenarioResult) Item() interface{} {
	return s.ProtoScenario
}
