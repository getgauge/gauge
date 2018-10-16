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
)

// SpecResult represents the result of spec execution
type SpecResult struct {
	ProtoSpec            *gauge_messages.ProtoSpec
	ScenarioFailedCount  int
	ScenarioCount        int
	IsFailed             bool
	FailedDataTableRows  []int32
	ExecutionTime        int64
	Skipped              bool
	ScenarioSkippedCount int
	Errors               []*gauge_messages.Error
}

// SetFailure sets the result to failed
func (specResult *SpecResult) SetFailure() {
	specResult.IsFailed = true
}

func (specResult *SpecResult) SetSkipped(skipped bool) {
	specResult.Skipped = skipped
}

func (specResult *SpecResult) AddSpecItems(resolvedItems []*gauge_messages.ProtoItem) {
	specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, resolvedItems...)
}

// AddScenarioResults adds the scenario result to the spec result.
func (specResult *SpecResult) AddScenarioResults(scenarioResults []Result) {
	for _, scenarioResult := range scenarioResults {
		if scenarioResult.GetFailed() {
			specResult.IsFailed = true
			specResult.ScenarioFailedCount++
		}
		specResult.AddExecTime(scenarioResult.ExecTime())
		specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Scenario, Scenario: scenarioResult.Item().(*gauge_messages.ProtoScenario)})
	}
	specResult.ScenarioCount += len(scenarioResults)
}

// AddTableRelatedScenarioResult aggregates the data table driven spec results.
func (specResult *SpecResult) AddTableRelatedScenarioResult(scenarioResults [][]Result, index int, specDataTable bool) {
	numberOfScenarios := len(scenarioResults[0])

	for scenarioIndex := 0; scenarioIndex < numberOfScenarios; scenarioIndex++ {
		scenarioFailed := false
		for _, eachRow := range scenarioResults {
			protoScenario := eachRow[scenarioIndex].Item().(*gauge_messages.ProtoScenario)
			specResult.AddExecTime(protoScenario.GetExecutionTime())
			if protoScenario.GetExecutionStatus() == gauge_messages.ExecutionStatus_FAILED {
				scenarioFailed = true
				specResult.FailedDataTableRows = append(specResult.FailedDataTableRows, int32(index))
			}
			protoTableDrivenScenario := &gauge_messages.ProtoTableDrivenScenario{Scenario: protoScenario, TableRowIndex: int32(index)}
			protoItem := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_TableDrivenScenario, TableDrivenScenario: protoTableDrivenScenario}
			specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, protoItem)
		}
		if scenarioFailed {
			specResult.ScenarioFailedCount++
			specResult.IsFailed = true
		}
	}
	specResult.ProtoSpec.IsTableDriven = specDataTable
	specResult.ScenarioCount += numberOfScenarios
}

func (specResult *SpecResult) AddExecTime(execTime int64) {
	specResult.ExecutionTime += execTime
}

func (specResult *SpecResult) GetPreHook() []*gauge_messages.ProtoHookFailure {
	return specResult.ProtoSpec.PreHookFailures
}

func (specResult *SpecResult) GetPostHook() []*gauge_messages.ProtoHookFailure {
	return specResult.ProtoSpec.PostHookFailures
}

func (specResult *SpecResult) AddPreHook(f ...*gauge_messages.ProtoHookFailure) {
	specResult.ProtoSpec.PreHookFailures = append(specResult.ProtoSpec.PreHookFailures, f...)
}

func (specResult *SpecResult) AddPostHook(f ...*gauge_messages.ProtoHookFailure) {
	specResult.ProtoSpec.PostHookFailures = append(specResult.ProtoSpec.PostHookFailures, f...)
}

func (specResult *SpecResult) ExecTime() int64 {
	return specResult.ExecutionTime
}

// GetFailed returns the state of the result
func (specResult *SpecResult) GetFailed() bool {
	return specResult.IsFailed
}

func (specResult *SpecResult) Item() interface{} {
	return specResult.ProtoSpec
}
