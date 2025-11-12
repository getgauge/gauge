/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package result

import (
	"github.com/getgauge/gauge-proto/go/gauge_messages"
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

func (specResult *SpecResult) AddTableDrivenScenarioResult(r *ScenarioResult, t *gauge_messages.ProtoTable, scenarioRowIndex int, specRowIndex int, specTableDriven bool) {
	if r.GetFailed() {
		specResult.IsFailed = true
		specResult.ScenarioFailedCount++
	}
	specResult.AddExecTime(r.ExecTime())
	pItem := &gauge_messages.ProtoItem{ // nolint
		ItemType: gauge_messages.ProtoItem_TableDrivenScenario,
		TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{
			Scenario:              r.Item().(*gauge_messages.ProtoScenario),
			IsScenarioTableDriven: true,
			ScenarioTableRowIndex: int32(scenarioRowIndex),
			IsSpecTableDriven:     specTableDriven,
			ScenarioDataTable:     t,
			TableRowIndex:         int32(specRowIndex),
			ScenarioTableRow:      r.ScenarioDataTableRow,
		},
	}
	specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, pItem)
}

// AddTableRelatedScenarioResult aggregates the data table driven spec results.
func (specResult *SpecResult) AddTableRelatedScenarioResult(scenarioResults [][]Result, index int) {
	numberOfScenarios := len(scenarioResults[0])

	for scenarioIndex := 0; scenarioIndex < numberOfScenarios; scenarioIndex++ {
		scenarioFailed := false
		for _, eachRow := range scenarioResults {
			protoScenario := eachRow[scenarioIndex].Item().(*gauge_messages.ProtoScenario)
			result := eachRow[scenarioIndex].(*ScenarioResult)
			specResult.AddExecTime(protoScenario.GetExecutionTime())
			if protoScenario.GetExecutionStatus() == gauge_messages.ExecutionStatus_FAILED {
				scenarioFailed = true
				specResult.FailedDataTableRows = append(specResult.FailedDataTableRows, int32(index))
			}
			protoTableDrivenScenario := &gauge_messages.ProtoTableDrivenScenario{
				Scenario:          protoScenario,
				TableRowIndex:     int32(index),
				ScenarioTableRow:  eachRow[scenarioIndex].(*ScenarioResult).ScenarioDataTableRow,
				IsSpecTableDriven: true,
			}

			if result.GetTableDrivenScenario() { // If nested table driven scenario within table driven spec
				protoTableDrivenScenario.IsScenarioTableDriven = true
				protoTableDrivenScenario.ScenarioTableRowIndex = int32(result.ScenarioDataTableRowIndex)
				protoTableDrivenScenario.ScenarioDataTable = result.ScenarioDataTable
			}
			protoItem := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_TableDrivenScenario, TableDrivenScenario: protoTableDrivenScenario} // nolint
			specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, protoItem)
		}
		if scenarioFailed {
			specResult.ScenarioFailedCount++
			specResult.IsFailed = true
		}
	}
	specResult.ProtoSpec.IsTableDriven = true
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
