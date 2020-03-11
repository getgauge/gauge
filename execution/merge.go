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
	"time"

	"strings"

	"github.com/getgauge/gauge/execution/result"
	m "github.com/getgauge/gauge/gauge_messages"
)

func mergeDataTableSpecResults(sResult *result.SuiteResult) *result.SuiteResult {
	suiteRes := result.NewSuiteResult(sResult.Tags, time.Now())
	suiteRes.IsFailed = sResult.IsFailed
	suiteRes.ExecutionTime = sResult.ExecutionTime
	suiteRes.PostSuite = sResult.PostSuite
	suiteRes.PreSuite = sResult.PreSuite
	suiteRes.UnhandledErrors = sResult.UnhandledErrors
	suiteRes.Timestamp = sResult.Timestamp
	suiteRes.ProjectName = sResult.ProjectName
	suiteRes.Environment = sResult.Environment
	suiteRes.Tags = sResult.Tags
	suiteRes.PreHookMessages = append(suiteRes.PreHookMessages, sResult.PreHookMessages...)
	suiteRes.PostHookMessages = append(suiteRes.PostHookMessages, sResult.PostHookMessages...)
	suiteRes.PreHookScreenshotFiles = append(suiteRes.PreHookScreenshotFiles, sResult.PreHookScreenshotFiles...)
	suiteRes.PostHookScreenshotFiles = append(suiteRes.PostHookScreenshotFiles, sResult.PostHookScreenshotFiles...)
	suiteRes.PreHookScreenshots = append(suiteRes.PreHookScreenshots, sResult.PreHookScreenshots...)
	suiteRes.PostHookScreenshots = append(suiteRes.PostHookScreenshots, sResult.PostHookScreenshots...)
	combinedResults := make(map[string][]*result.SpecResult)
	for _, res := range sResult.SpecResults {
		fileName := res.ProtoSpec.GetFileName()
		combinedResults[fileName] = append(combinedResults[fileName], res)
	}
	for _, res := range combinedResults {
		mergedRes := res[0]
		if len(res) > 1 {
			mergedRes = mergeResults(res)
		}
		if mergedRes.GetFailed() {
			suiteRes.SpecsFailedCount++
		} else if mergedRes.Skipped {
			suiteRes.SpecsSkippedCount++
		}
		suiteRes.SpecResults = append(suiteRes.SpecResults, mergedRes)
	}
	return suiteRes
}

func hasTableDrivenSpec(results []*result.SpecResult) bool {
	for _, r := range results {
		if r.ProtoSpec.GetIsTableDriven() {
			return true
		}
	}
	return false
}

func mergeResults(results []*result.SpecResult) *result.SpecResult {
	specResult := &result.SpecResult{ProtoSpec: &m.ProtoSpec{IsTableDriven: hasTableDrivenSpec(results)}}
	var scnResults []*m.ProtoItem
	table := &m.ProtoTable{}
	dataTableScnResults := make(map[string][]*m.ProtoTableDrivenScenario)
	includedTableRowIndexMap := make(map[int32]bool)
	max := results[0].ExecutionTime
	for _, res := range results {
		specResult.ExecutionTime += res.ExecutionTime
		specResult.Errors = res.Errors
		if res.ExecutionTime > max {
			max = res.ExecutionTime
		}
		if res.GetFailed() {
			specResult.IsFailed = true
		}
		var tableRowIndex int32
		var tableRows []*m.ProtoTableRow
		for _, item := range res.ProtoSpec.Items {
			switch item.ItemType {
			case m.ProtoItem_Scenario:
				scnResults = append(scnResults, item)
				modifySpecStats(item.Scenario, specResult)
			case m.ProtoItem_TableDrivenScenario:
				tableRowIndex = item.TableDrivenScenario.TableRowIndex
				scnResults = append(scnResults, item)
				heading := item.TableDrivenScenario.Scenario.ScenarioHeading
				dataTableScnResults[heading] = append(dataTableScnResults[heading], item.TableDrivenScenario)
			case m.ProtoItem_Table:
				table.Headers = item.Table.Headers
				tableRows = item.Table.GetRows()
				if len(res.GetPreHook()) > 0 {
					tableRowIndex = res.GetPreHook()[0].GetTableRowIndex()
				}
			}
		}
		if _, ok := includedTableRowIndexMap[tableRowIndex]; !ok {
			table.Rows = append(table.Rows, tableRows...)
			includedTableRowIndexMap[tableRowIndex] = true
		}
		addHookFailure(tableRowIndex, res.GetPreHook(), specResult.AddPreHook)
		addHookFailure(tableRowIndex, res.GetPostHook(), specResult.AddPostHook)
	}
	if InParallel {
		specResult.ExecutionTime = max
	}
	aggregateDataTableScnStats(dataTableScnResults, specResult)
	specResult.ProtoSpec.FileName = results[0].ProtoSpec.FileName
	specResult.ProtoSpec.Tags = results[0].ProtoSpec.Tags
	specResult.ProtoSpec.SpecHeading = results[0].ProtoSpec.SpecHeading
	specResult.ProtoSpec.Items = getItems(table, scnResults, results)
	return specResult
}

func addHookFailure(tableRowIndex int32, f []*m.ProtoHookFailure, add func(...*m.ProtoHookFailure)) {
	for _, h := range f {
		h.TableRowIndex = tableRowIndex
	}
	add(f...)
}

func getItems(table *m.ProtoTable, scnResults []*m.ProtoItem, results []*result.SpecResult) (items []*m.ProtoItem) {
	index := 0
	for _, item := range results[0].ProtoSpec.Items {
		switch item.ItemType {
		case m.ProtoItem_Scenario, m.ProtoItem_TableDrivenScenario:
			items = append(items, scnResults[index])
			index++
		case m.ProtoItem_Table:
			items = append(items, &m.ProtoItem{ItemType: m.ProtoItem_Table, Table: table})
		default:
			items = append(items, item)
		}
	}
	items = append(items, scnResults[index:]...)
	return
}

func aggregateDataTableScnStats(results map[string][]*m.ProtoTableDrivenScenario, specResult *result.SpecResult) {
	for _, dResult := range results {
		for _, res := range dResult {
			isTableIndicesExcluded := false
			if res.Scenario.ExecutionStatus == m.ExecutionStatus_FAILED {
				specResult.ScenarioFailedCount++
			} else if res.Scenario.ExecutionStatus == m.ExecutionStatus_SKIPPED &&
				!strings.Contains(res.Scenario.SkipErrors[0], "--table-rows") {
				specResult.ScenarioSkippedCount++
				specResult.Skipped = true
			} else if res.Scenario.ExecutionStatus == m.ExecutionStatus_SKIPPED &&
				strings.Contains(res.Scenario.SkipErrors[0], "--table-rows") {
				isTableIndicesExcluded = true
			}
			if !isTableIndicesExcluded {
				specResult.ScenarioCount++
			}
		}
	}
}

func modifySpecStats(scn *m.ProtoScenario, specRes *result.SpecResult) {
	switch scn.ExecutionStatus {
	case m.ExecutionStatus_SKIPPED:
		specRes.ScenarioSkippedCount++
	case m.ExecutionStatus_FAILED:
		specRes.ScenarioFailedCount++
	}
	specRes.ScenarioCount++
}
