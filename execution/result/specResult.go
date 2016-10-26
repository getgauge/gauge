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

func (specResult *SpecResult) SetFailure() {
	specResult.IsFailed = true
}

func (specResult *SpecResult) SetSkipped(skipped bool) {
	specResult.Skipped = skipped
}

func (specResult *SpecResult) AddSpecItems(resolvedItems []*gauge_messages.ProtoItem) {
	specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, resolvedItems...)
}

func (specResult *SpecResult) AddScenarioResults(scenarioResults []Result) {
	for _, scenarioResult := range scenarioResults {
		if scenarioResult.GetFailed() {
			specResult.IsFailed = true
			specResult.ScenarioFailedCount++
		}
		specResult.AddExecTime(scenarioResult.ExecTime())
		specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Scenario.Enum(), Scenario: scenarioResult.Item().(*gauge_messages.ProtoScenario)})
	}
	specResult.ScenarioCount += len(scenarioResults)
}

func (specResult *SpecResult) AddTableDrivenScenarioResult(scenarioResults [][]Result, rowIndexStart int) {
	numberOfScenarios := len(scenarioResults[0])

	for scenarioIndex := 0; scenarioIndex < numberOfScenarios; scenarioIndex++ {
		scenarioFailed := false
		for _, eachRow := range scenarioResults {
			protoScenario := eachRow[scenarioIndex].Item().(*gauge_messages.ProtoScenario)
			specResult.AddExecTime(protoScenario.GetExecutionTime())
			if protoScenario.GetExecutionStatus() == gauge_messages.ExecutionStatus_FAILED {
				scenarioFailed = true
				specResult.FailedDataTableRows = append(specResult.FailedDataTableRows, int32(rowIndexStart))
			}
			protoTableDrivenScenario := &gauge_messages.ProtoTableDrivenScenario{Scenario: protoScenario, TableRowIndex: proto.Int32(int32(rowIndexStart))}
			rowIndexStart++
			protoItem := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_TableDrivenScenario.Enum(), TableDrivenScenario: protoTableDrivenScenario}
			specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, protoItem)
		}
		if scenarioFailed {
			specResult.ScenarioFailedCount++
			specResult.IsFailed = true
		}
	}
	specResult.ProtoSpec.IsTableDriven = proto.Bool(true)
	specResult.ScenarioCount += numberOfScenarios
}

func (specResult *SpecResult) AddExecTime(execTime int64) {
	specResult.ExecutionTime += execTime
}

func (specResult *SpecResult) GetPreHook() **(gauge_messages.ProtoHookFailure) {
	return &specResult.ProtoSpec.PreHookFailure
}

func (specResult *SpecResult) GetPostHook() **(gauge_messages.ProtoHookFailure) {
	return &specResult.ProtoSpec.PostHookFailure
}

func (specResult *SpecResult) setFileName(fileName string) {
	specResult.ProtoSpec.FileName = proto.String(fileName)
}

func (specResult *SpecResult) ExecTime() int64 {
	return specResult.ExecutionTime
}

func (specResult *SpecResult) GetFailed() bool {
	return specResult.IsFailed
}

func (specResult *SpecResult) Item() interface{} {
	return specResult.ProtoSpec
}
