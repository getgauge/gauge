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

package main

import (
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type suiteResult struct {
	specResults      []*specResult
	preSuite         *(gauge_messages.ProtoHookFailure)
	postSuite        *(gauge_messages.ProtoHookFailure)
	isFailed         bool
	specsFailedCount int
	executionTime    int64 //in milliseconds
	unhandledErrors  []error
}

type specResult struct {
	protoSpec           *gauge_messages.ProtoSpec
	scenarioFailedCount int
	scenarioCount       int
	isFailed            bool
	failedDataTableRows []int32
	executionTime       int64
}

type scenarioResult struct {
	protoScenario *gauge_messages.ProtoScenario
}

type result interface {
	getPreHook() **(gauge_messages.ProtoHookFailure)
	getPostHook() **(gauge_messages.ProtoHookFailure)
	setFailure()
}

type execTimeTracker interface {
	addExecTime(int64)
}

func (suiteResult *suiteResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &suiteResult.preSuite
}

func (suiteResult *suiteResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &suiteResult.postSuite
}

func (suiteResult *suiteResult) setFailure() {
	suiteResult.isFailed = true
}

func (specResult *specResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &specResult.protoSpec.PreHookFailure
}

func (specResult *specResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &specResult.protoSpec.PostHookFailure
}

func (specResult *specResult) setFailure() {
	specResult.isFailed = true
}

func (scenarioResult *scenarioResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &scenarioResult.protoScenario.PreHookFailure
}

func (scenarioResult *scenarioResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &scenarioResult.protoScenario.PostHookFailure
}

func (scenarioResult *scenarioResult) setFailure() {
	scenarioResult.protoScenario.Failed = proto.Bool(true)
}

func (scenarioResult *scenarioResult) getFailure() bool {
	return scenarioResult.protoScenario.GetFailed()
}

func (specResult *specResult) addSpecItems(resolvedItems []*gauge_messages.ProtoItem) {
	specResult.protoSpec.Items = append(specResult.protoSpec.Items, resolvedItems...)
}

func newSuiteResult() *suiteResult {
	result := new(suiteResult)
	result.specResults = make([]*specResult, 0)
	return result
}

func addPreHook(result result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.getPreHook()) = getProtoHookFailure(executionResult)
		result.setFailure()
	}
}

func addPostHook(result result, executionResult *gauge_messages.ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.getPostHook()) = getProtoHookFailure(executionResult)
		result.setFailure()
	}
}

func (suiteResult *suiteResult) addSpecResult(specResult *specResult) {
	if specResult.isFailed {
		suiteResult.isFailed = true
		suiteResult.specsFailedCount++
	}
	suiteResult.executionTime += specResult.executionTime
	suiteResult.specResults = append(suiteResult.specResults, specResult)

}

func getProtoHookFailure(executionResult *gauge_messages.ProtoExecutionResult) *(gauge_messages.ProtoHookFailure) {
	return &gauge_messages.ProtoHookFailure{StackTrace: executionResult.StackTrace, ErrorMessage: executionResult.ErrorMessage, ScreenShot: executionResult.ScreenShot}
}

func (specResult *specResult) setFileName(fileName string) {
	specResult.protoSpec.FileName = proto.String(fileName)
}

func (specResult *specResult) addScenarioResults(scenarioResults []*scenarioResult) {
	for _, scenarioResult := range scenarioResults {
		if scenarioResult.protoScenario.GetFailed() {
			specResult.isFailed = true
			specResult.scenarioFailedCount++
		}
		specResult.addExecTime(scenarioResult.protoScenario.GetExecutionTime())
		specResult.protoSpec.Items = append(specResult.protoSpec.Items, &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Scenario.Enum(), Scenario: scenarioResult.protoScenario})
	}
	specResult.scenarioCount += len(scenarioResults)
}

func (specResult *specResult) addTableDrivenScenarioResult(scenarioResults [][](*scenarioResult)) {
	numberOfScenarios := len(scenarioResults[0])

	for scenarioIndex := 0; scenarioIndex < numberOfScenarios; scenarioIndex++ {
		protoTableDrivenScenario := &gauge_messages.ProtoTableDrivenScenario{Scenarios: make([]*gauge_messages.ProtoScenario, 0)}
		scenarioFailed := false
		for rowIndex, eachRow := range scenarioResults {
			protoScenario := eachRow[scenarioIndex].protoScenario
			protoTableDrivenScenario.Scenarios = append(protoTableDrivenScenario.GetScenarios(), protoScenario)
			specResult.addExecTime(protoScenario.GetExecutionTime())
			if protoScenario.GetFailed() {
				scenarioFailed = true
				specResult.failedDataTableRows = append(specResult.failedDataTableRows, int32(rowIndex))
			}
		}
		if scenarioFailed {
			specResult.scenarioFailedCount++
			specResult.isFailed = true
		}
		protoItem := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_TableDrivenScenario.Enum(), TableDrivenScenario: protoTableDrivenScenario}
		specResult.protoSpec.Items = append(specResult.protoSpec.Items, protoItem)
	}
	specResult.protoSpec.IsTableDriven = proto.Bool(true)
	specResult.scenarioCount += numberOfScenarios
}

func (specResult *specResult) addExecTime(execTime int64) {
	specResult.executionTime += execTime
}

func (scenarioResult *scenarioResult) addItems(protoItems []*gauge_messages.ProtoItem) {
	scenarioResult.protoScenario.ScenarioItems = append(scenarioResult.protoScenario.ScenarioItems, protoItems...)
}

func (scenarioResult *scenarioResult) addContexts(contextProtoItems []*gauge_messages.ProtoItem) {
	scenarioResult.protoScenario.Contexts = append(scenarioResult.protoScenario.Contexts, contextProtoItems...)
}

func (scenarioResult *scenarioResult) updateExecutionTime() {
	scenarioResult.updateExecutionTimeFromItems(scenarioResult.protoScenario.GetContexts())
	scenarioResult.updateExecutionTimeFromItems(scenarioResult.protoScenario.GetScenarioItems())
}

func (scenarioResult *scenarioResult) updateExecutionTimeFromItems(protoItems []*gauge_messages.ProtoItem) {
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

func (scenarioResult *scenarioResult) addExecTime(execTime int64) {
	currentScenarioExecTime := scenarioResult.protoScenario.GetExecutionTime()
	scenarioResult.protoScenario.ExecutionTime = proto.Int64(currentScenarioExecTime + execTime)
}
