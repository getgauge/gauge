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

type SuiteResult struct {
	SpecResults       []*SpecResult
	PreSuite          *(gauge_messages.ProtoHookFailure)
	PostSuite         *(gauge_messages.ProtoHookFailure)
	IsFailed          bool
	SpecsFailedCount  int
	ExecutionTime     int64 //in milliseconds
	UnhandledErrors   []error
	Environment       string
	Tags              string
	ProjectName       string
	Timestamp         string
	SpecsSkippedCount int
}

type SpecResult struct {
	ProtoSpec            *gauge_messages.ProtoSpec
	ScenarioFailedCount  int
	ScenarioCount        int
	IsFailed             bool
	FailedDataTableRows  []int32
	ExecutionTime        int64
	Skipped              bool
	ScenarioSkippedCount int
}

type ScenarioResult struct {
	ProtoScenario *gauge_messages.ProtoScenario
}

type Result interface {
	getPreHook() **(gauge_messages.ProtoHookFailure)
	getPostHook() **(gauge_messages.ProtoHookFailure)
	SetFailure()
}

type ExecTimeTracker interface {
	AddExecTime(int64)
}

func (suiteResult *SuiteResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &suiteResult.PreSuite
}

func (suiteResult *SuiteResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &suiteResult.PostSuite
}

func (suiteResult *SuiteResult) SetFailure() {
	suiteResult.IsFailed = true
}

func (specResult *SpecResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &specResult.ProtoSpec.PreHookFailure
}

func (specResult *SpecResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &specResult.ProtoSpec.PostHookFailure
}

func (specResult *SpecResult) SetFailure() {
	specResult.IsFailed = true
}

func (scenarioResult *ScenarioResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &scenarioResult.ProtoScenario.PreHookFailure
}

func (scenarioResult *ScenarioResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &scenarioResult.ProtoScenario.PostHookFailure
}

func (scenarioResult *ScenarioResult) SetFailure() {
	scenarioResult.ProtoScenario.Failed = proto.Bool(true)
}

func (scenarioResult *ScenarioResult) GetFailure() bool {
	return scenarioResult.ProtoScenario.GetFailed()
}

func (specResult *SpecResult) AddSpecItems(resolvedItems []*gauge_messages.ProtoItem) {
	specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, resolvedItems...)
}

func NewSuiteResult() *SuiteResult {
	result := new(SuiteResult)
	result.SpecResults = make([]*SpecResult, 0)
	return result
}

func AddPreHook(result Result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.getPreHook()) = GetProtoHookFailure(executionResult)
		result.SetFailure()
	}
}

func AddPostHook(result Result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.getPostHook()) = GetProtoHookFailure(executionResult)
		result.SetFailure()
	}
}

func (suiteResult *SuiteResult) AddSpecResult(specResult *SpecResult) {
	if specResult.IsFailed {
		suiteResult.IsFailed = true
		suiteResult.SpecsFailedCount++
	}
	suiteResult.ExecutionTime += specResult.ExecutionTime
	suiteResult.SpecResults = append(suiteResult.SpecResults, specResult)

}

func GetProtoHookFailure(executionResult *gauge_messages.ProtoExecutionResult) *(gauge_messages.ProtoHookFailure) {
	return &gauge_messages.ProtoHookFailure{StackTrace: executionResult.StackTrace, ErrorMessage: executionResult.ErrorMessage, ScreenShot: executionResult.ScreenShot}
}

func (specResult *SpecResult) setFileName(fileName string) {
	specResult.ProtoSpec.FileName = proto.String(fileName)
}

func (specResult *SpecResult) AddScenarioResults(scenarioResults []*ScenarioResult) {
	for _, scenarioResult := range scenarioResults {
		if scenarioResult.ProtoScenario.GetFailed() {
			specResult.IsFailed = true
			specResult.ScenarioFailedCount++
		}
		specResult.AddExecTime(scenarioResult.ProtoScenario.GetExecutionTime())
		specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Scenario.Enum(), Scenario: scenarioResult.ProtoScenario})
	}
	specResult.ScenarioCount += len(scenarioResults)
}

func (specResult *SpecResult) AddTableDrivenScenarioResult(scenarioResults [][](*ScenarioResult)) {
	numberOfScenarios := len(scenarioResults[0])

	for scenarioIndex := 0; scenarioIndex < numberOfScenarios; scenarioIndex++ {
		protoTableDrivenScenario := &gauge_messages.ProtoTableDrivenScenario{Scenarios: make([]*gauge_messages.ProtoScenario, 0)}
		scenarioFailed := false
		for rowIndex, eachRow := range scenarioResults {
			protoScenario := eachRow[scenarioIndex].ProtoScenario
			protoTableDrivenScenario.Scenarios = append(protoTableDrivenScenario.GetScenarios(), protoScenario)
			specResult.AddExecTime(protoScenario.GetExecutionTime())
			if protoScenario.GetFailed() {
				scenarioFailed = true
				specResult.FailedDataTableRows = append(specResult.FailedDataTableRows, int32(rowIndex))
			}
		}
		if scenarioFailed {
			specResult.ScenarioFailedCount++
			specResult.IsFailed = true
		}
		protoItem := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_TableDrivenScenario.Enum(), TableDrivenScenario: protoTableDrivenScenario}
		specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, protoItem)
	}
	specResult.ProtoSpec.IsTableDriven = proto.Bool(true)
	specResult.ScenarioCount += numberOfScenarios
}

func (specResult *SpecResult) AddExecTime(execTime int64) {
	specResult.ExecutionTime += execTime
}

func (scenarioResult *ScenarioResult) AddItems(protoItems []*gauge_messages.ProtoItem) {
	scenarioResult.ProtoScenario.ScenarioItems = append(scenarioResult.ProtoScenario.ScenarioItems, protoItems...)
}

func (scenarioResult *ScenarioResult) AddContexts(contextProtoItems []*gauge_messages.ProtoItem) {
	scenarioResult.ProtoScenario.Contexts = append(scenarioResult.ProtoScenario.Contexts, contextProtoItems...)
}

func (scenarioResult *ScenarioResult) UpdateExecutionTime() {
	scenarioResult.updateExecutionTimeFromItems(scenarioResult.ProtoScenario.GetContexts())
	scenarioResult.updateExecutionTimeFromItems(scenarioResult.ProtoScenario.GetScenarioItems())
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

func (scenarioResult *ScenarioResult) AddExecTime(execTime int64) {
	currentScenarioExecTime := scenarioResult.ProtoScenario.GetExecutionTime()
	scenarioResult.ProtoScenario.ExecutionTime = proto.Int64(currentScenarioExecTime + execTime)
}
