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

type ScenarioResult struct {
	ProtoScenario *gauge_messages.ProtoScenario
}

func NewScenarioResult(sce *gauge_messages.ProtoScenario) *ScenarioResult {
	return &ScenarioResult{ProtoScenario: sce}
}

func (s *ScenarioResult) SetFailure() {
	s.ProtoScenario.ExecutionStatus = gauge_messages.ExecutionStatus_FAILED
	s.ProtoScenario.Failed = true
}

func (s *ScenarioResult) GetFailed() bool {
	return s.ProtoScenario.GetExecutionStatus() == gauge_messages.ExecutionStatus_FAILED
}

func (s *ScenarioResult) AddItems(protoItems []*gauge_messages.ProtoItem) {
	s.ProtoScenario.ScenarioItems = append(s.ProtoScenario.ScenarioItems, protoItems...)
}

func (s *ScenarioResult) AddContexts(contextProtoItems []*gauge_messages.ProtoItem) {
	s.ProtoScenario.Contexts = append(s.ProtoScenario.Contexts, contextProtoItems...)
}

func (s *ScenarioResult) AddTearDownSteps(tearDownProtoItems []*gauge_messages.ProtoItem) {
	s.ProtoScenario.TearDownSteps = append(s.ProtoScenario.TearDownSteps, tearDownProtoItems...)
}

func (s *ScenarioResult) UpdateExecutionTime() {
	s.updateExecutionTimeFromItems(s.ProtoScenario.GetContexts())
	s.updateExecutionTimeFromItems(s.ProtoScenario.GetScenarioItems())
}

func (s *ScenarioResult) AddExecTime(execTime int64) {
	currentScenarioExecTime := s.ProtoScenario.GetExecutionTime()
	s.ProtoScenario.ExecutionTime = currentScenarioExecTime + execTime
}

func (s *ScenarioResult) ExecTime() int64 {
	return s.ProtoScenario.ExecutionTime
}

func (s *ScenarioResult) updateExecutionTimeFromItems(protoItems []*gauge_messages.ProtoItem) {
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

func (s *ScenarioResult) GetPreHook() []*gauge_messages.ProtoHookFailure {
	if s.ProtoScenario.PreHookFailure == nil {
		return []*gauge_messages.ProtoHookFailure{}
	}
	return []*gauge_messages.ProtoHookFailure{s.ProtoScenario.PreHookFailure}
}

func (s *ScenarioResult) GetPostHook() []*gauge_messages.ProtoHookFailure {
	if s.ProtoScenario.PostHookFailure == nil {
		return []*gauge_messages.ProtoHookFailure{}
	}
	return []*gauge_messages.ProtoHookFailure{s.ProtoScenario.PostHookFailure}
}

func (s *ScenarioResult) AddPreHook(f *gauge_messages.ProtoHookFailure) {
	s.ProtoScenario.PreHookFailure = f
}

func (s *ScenarioResult) AddPostHook(f *gauge_messages.ProtoHookFailure) {
	s.ProtoScenario.PostHookFailure = f
}

func (s *ScenarioResult) Item() interface{} {
	return s.ProtoScenario
}
