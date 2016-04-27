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

import (
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type ScenarioResult struct {
	ProtoScenario *gauge_messages.ProtoScenario
}

func (scenarioResult *ScenarioResult) SetFailure() {
	scenarioResult.ProtoScenario.Failed = proto.Bool(true)
}

func (scenarioResult *ScenarioResult) GetFailed() bool {
	return scenarioResult.ProtoScenario.GetFailed()
}

func (scenarioResult *ScenarioResult) AddItems(protoItems []*gauge_messages.ProtoItem) {
	scenarioResult.ProtoScenario.ScenarioItems = append(scenarioResult.ProtoScenario.ScenarioItems, protoItems...)
}

func (scenarioResult *ScenarioResult) AddContexts(contextProtoItems []*gauge_messages.ProtoItem) {
	scenarioResult.ProtoScenario.Contexts = append(scenarioResult.ProtoScenario.Contexts, contextProtoItems...)
}

func (scenarioResult *ScenarioResult) AddTearDownSteps(tearDownProtoItems []*gauge_messages.ProtoItem) {
	scenarioResult.ProtoScenario.TearDownSteps = append(scenarioResult.ProtoScenario.TearDownSteps, tearDownProtoItems...)
}

func (scenarioResult *ScenarioResult) UpdateExecutionTime() {
	scenarioResult.updateExecutionTimeFromItems(scenarioResult.ProtoScenario.GetContexts())
	scenarioResult.updateExecutionTimeFromItems(scenarioResult.ProtoScenario.GetScenarioItems())
}

func (scenarioResult *ScenarioResult) AddExecTime(execTime int64) {
	currentScenarioExecTime := scenarioResult.ProtoScenario.GetExecutionTime()
	scenarioResult.ProtoScenario.ExecutionTime = proto.Int64(currentScenarioExecTime + execTime)
}

func (s *ScenarioResult) ExecTime() int64 {
	return *s.ProtoScenario.ExecutionTime
}

func (scenarioResult *ScenarioResult) updateExecutionTimeFromItems(protoItems []*gauge_messages.ProtoItem) {
	for _, item := range protoItems {
		if item.GetItemType() == gauge_messages.ProtoItem_Step {
			stepExecTime := item.GetStep().GetStepExecutionResult().GetExecutionResult().GetExecutionTime()
			scenarioResult.AddExecTime(stepExecTime)
		} else if item.GetItemType() == gauge_messages.ProtoItem_Concept {
			conceptExecTime := item.GetConcept().GetConceptExecutionResult().GetExecutionResult().GetExecutionTime()
			scenarioResult.AddExecTime(conceptExecTime)
		}
	}
}

func (scenarioResult *ScenarioResult) GetPreHook() **(gauge_messages.ProtoHookFailure) {
	return &scenarioResult.ProtoScenario.PreHookFailure
}

func (scenarioResult *ScenarioResult) GetPostHook() **(gauge_messages.ProtoHookFailure) {
	return &scenarioResult.ProtoScenario.PostHookFailure
}

func (r *ScenarioResult) item() interface{} {
	return r.ProtoScenario
}
