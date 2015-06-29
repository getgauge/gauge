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
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type SuiteResult struct {
	SpecResults      []*SpecResult
	PreSuite         *(gauge_messages.ProtoHookFailure)
	PostSuite        *(gauge_messages.ProtoHookFailure)
	IsFailed         bool
	SpecsFailedCount int
	ExecutionTime    int64 //in milliseconds
	UnhandledErrors  []error
	Environment      string
	Tags             string
	ProjectName      string
	Timestamp        string
}

type SpecResult struct {
	ProtoSpec           *gauge_messages.ProtoSpec
	ScenarioFailedCount int
	ScenarioCount       int
	IsFailed            bool
	FailedDataTableRows []int32
	ExecutionTime       int64
}

type ScenarioResult struct {
	ProtoScenario *gauge_messages.ProtoScenario
}

type Result interface {
	getPreHook() **(gauge_messages.ProtoHookFailure)
	getPostHook() **(gauge_messages.ProtoHookFailure)
	setFailure()
}

type execTimeTracker interface {
	addExecTime(int64)
}

func (suiteResult *SuiteResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &suiteResult.PreSuite
}

func (suiteResult *SuiteResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &suiteResult.PostSuite
}

func (suiteResult *SuiteResult) setFailure() {
	suiteResult.IsFailed = true
}

func (specResult *SpecResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &specResult.ProtoSpec.PreHookFailure
}

func (specResult *SpecResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &specResult.ProtoSpec.PostHookFailure
}

func (specResult *SpecResult) setFailure() {
	specResult.IsFailed = true
}

func (scenarioResult *ScenarioResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &scenarioResult.ProtoScenario.PreHookFailure
}

func (scenarioResult *ScenarioResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &scenarioResult.ProtoScenario.PostHookFailure
}

func (scenarioResult *ScenarioResult) setFailure() {
	scenarioResult.ProtoScenario.Failed = proto.Bool(true)
}

func (scenarioResult *ScenarioResult) getFailure() bool {
	return scenarioResult.ProtoScenario.GetFailed()
}

func (specResult *SpecResult) addSpecItems(resolvedItems []*gauge_messages.ProtoItem) {
	specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, resolvedItems...)
}

func newSuiteResult() *SuiteResult {
	result := new(SuiteResult)
	result.SpecResults = make([]*SpecResult, 0)
	return result
}

func addPreHook(result Result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.getPreHook()) = getProtoHookFailure(executionResult)
		result.setFailure()
	}
}

func addPostHook(result Result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.getPostHook()) = getProtoHookFailure(executionResult)
		result.setFailure()
	}
}

func (suiteResult *SuiteResult) addSpecResult(specResult *SpecResult) {
	if specResult.IsFailed {
		suiteResult.IsFailed = true
		suiteResult.SpecsFailedCount++
	}
	suiteResult.ExecutionTime += specResult.ExecutionTime
	suiteResult.SpecResults = append(suiteResult.SpecResults, specResult)

}

func getProtoHookFailure(executionResult *gauge_messages.ProtoExecutionResult) *(gauge_messages.ProtoHookFailure) {
	return &gauge_messages.ProtoHookFailure{StackTrace: executionResult.StackTrace, ErrorMessage: executionResult.ErrorMessage, ScreenShot: executionResult.ScreenShot}
}

func (specResult *SpecResult) setFileName(fileName string) {
	specResult.ProtoSpec.FileName = proto.String(fileName)
}

func (specResult *SpecResult) addScenarioResults(scenarioResults []*ScenarioResult) {
	for _, scenarioResult := range scenarioResults {
		if scenarioResult.ProtoScenario.GetFailed() {
			specResult.IsFailed = true
			specResult.ScenarioFailedCount++
		}
		specResult.addExecTime(scenarioResult.ProtoScenario.GetExecutionTime())
		specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Scenario.Enum(), Scenario: scenarioResult.ProtoScenario})
	}
	specResult.ScenarioCount += len(scenarioResults)
}

func (specResult *SpecResult) addTableDrivenScenarioResult(scenarioResults [][](*ScenarioResult)) {
	numberOfScenarios := len(scenarioResults[0])

	for scenarioIndex := 0; scenarioIndex < numberOfScenarios; scenarioIndex++ {
		protoTableDrivenScenario := &gauge_messages.ProtoTableDrivenScenario{Scenarios: make([]*gauge_messages.ProtoScenario, 0)}
		scenarioFailed := false
		for rowIndex, eachRow := range scenarioResults {
			protoScenario := eachRow[scenarioIndex].ProtoScenario
			protoTableDrivenScenario.Scenarios = append(protoTableDrivenScenario.GetScenarios(), protoScenario)
			specResult.addExecTime(protoScenario.GetExecutionTime())
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

func (specResult *SpecResult) addExecTime(execTime int64) {
	specResult.ExecutionTime += execTime
}

func (scenarioResult *ScenarioResult) addItems(protoItems []*gauge_messages.ProtoItem) {
	scenarioResult.ProtoScenario.ScenarioItems = append(scenarioResult.ProtoScenario.ScenarioItems, protoItems...)
}

func (scenarioResult *ScenarioResult) addContexts(contextProtoItems []*gauge_messages.ProtoItem) {
	scenarioResult.ProtoScenario.Contexts = append(scenarioResult.ProtoScenario.Contexts, contextProtoItems...)
}

func (scenarioResult *ScenarioResult) updateExecutionTime() {
	scenarioResult.updateExecutionTimeFromItems(scenarioResult.ProtoScenario.GetContexts())
	scenarioResult.updateExecutionTimeFromItems(scenarioResult.ProtoScenario.GetScenarioItems())
}

func (scenarioResult *ScenarioResult) updateExecutionTimeFromItems(protoItems []*gauge_messages.ProtoItem) {
	for _, item := range protoItems {
		if item.GetItemType() == gauge_messages.ProtoItem_Step {
			stepExecTime := item.GetStep().GetStepExecutionResult().GetExecutionResult().GetExecutionTime()
			scenarioResult.addExecTime(stepExecTime)
		} else if item.GetItemType() == gauge_messages.ProtoItem_Concept {
			conceptExecTime := item.GetConcept().GetConceptExecutionResult().GetExecutionResult().GetExecutionTime()
			scenarioResult.addExecTime(conceptExecTime)
		}
	}
}

func (scenarioResult *ScenarioResult) addExecTime(execTime int64) {
	currentScenarioExecTime := scenarioResult.ProtoScenario.GetExecutionTime()
	scenarioResult.ProtoScenario.ExecutionTime = proto.Int64(currentScenarioExecTime + execTime)
}
