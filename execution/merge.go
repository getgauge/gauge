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

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge_messages"
)

func mergeDataTableSpecResults(sResult *result.SuiteResult) *result.SuiteResult {
	suiteRes := result.NewSuiteResult(sResult.Tags, time.Now())
	suiteRes.IsFailed = sResult.IsFailed
	suiteRes.ExecutionTime = sResult.ExecutionTime
	suiteRes.PostSuite = sResult.PostSuite
	suiteRes.PreSuite = sResult.PreSuite
	suiteRes.UnhandledErrors = sResult.UnhandledErrors
	suiteRes.Timestamp = sResult.Timestamp

	combinedResults := make(map[string][]*result.SpecResult)
	for _, res := range sResult.SpecResults {
		fileName := res.ProtoSpec.GetFileName()
		if _, ok := combinedResults[fileName]; !ok {
			combinedResults[fileName] = make([]*result.SpecResult, 0)
		}
		combinedResults[fileName] = append(combinedResults[fileName], res)
	}
	for _, res := range combinedResults {
		mergedRes := res[0]
		if len(res) > 1 {
			mergedRes = mergeResults(res)
		}
		modifySuiteStats(mergedRes, suiteRes)
		suiteRes.SpecResults = append(suiteRes.SpecResults, mergedRes)
	}
	return suiteRes
}
func modifySuiteStats(res *result.SpecResult, suiteRes *result.SuiteResult) {
	if res.GetFailed() {
		suiteRes.SpecsFailedCount++
	} else if res.Skipped {
		suiteRes.SpecsSkippedCount++
	}
}
func modifySpecStats(scn *gauge_messages.ProtoScenario, specRes *result.SpecResult) {
	if scn.Skipped {
		specRes.ScenarioSkippedCount++
		return
	}
	if scn.GetFailed() {
		specRes.ScenarioFailedCount++
	}
	specRes.ScenarioCount++
}

func mergeResults(results []*result.SpecResult) *result.SpecResult {
	specResult := &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{IsTableDriven: true}}
	var scnResults []*gauge_messages.ProtoItem
	table := &gauge_messages.ProtoTable{}

	dataTableScnResults := make(map[string][]*gauge_messages.ProtoTableDrivenScenario)

	sortResults := make([]*result.SpecResult, len(results))
	for i, res := range results {
		if len(res.GetPreHook()) > 0 {
			sortResults[i] = res
			continue
		}
		for _, item := range res.ProtoSpec.Items {
			if item.ItemType == gauge_messages.ProtoItem_TableDrivenScenario {
				sortResults[item.TableDrivenScenario.TableRowIndex] = res
			}
		}
	}

	specResult.ProtoSpec.FileName = results[0].ProtoSpec.FileName
	specResult.ProtoSpec.SpecHeading = results[0].ProtoSpec.SpecHeading

	for _, res := range sortResults {
		if res.GetFailed() {
			specResult.IsFailed = true
		}
		if len(res.GetPreHook()) > 0 {
			specResult.AddPreHook(res.GetPreHook()[0])
		}
		for _, item := range res.ProtoSpec.Items {
			switch item.ItemType {
			case gauge_messages.ProtoItem_Scenario:
				scnResults = append(scnResults, item)
				modifySpecStats(item.Scenario, specResult)
			case gauge_messages.ProtoItem_TableDrivenScenario:
				scnResults = append(scnResults, item)
				heading := item.TableDrivenScenario.Scenario.ScenarioHeading
				dataTableScnResults[heading] = append(dataTableScnResults[heading], item.TableDrivenScenario)
			case gauge_messages.ProtoItem_Table:
				table.Headers = item.Table.Headers
				table.Rows = append(table.Rows, item.Table.Rows...)
			}
		}
		if len(res.GetPostHook()) > 0 {
			specResult.AddPostHook(res.GetPostHook()[0])
		}
	}
	for _, dResult := range dataTableScnResults {
		isFailed := false
		isSkipped := false
		for _, res := range dResult {
			if res.Scenario.Failed {
				isFailed = true
			} else if res.Scenario.Skipped {
				isSkipped = true
			}
		}
		if isSkipped {
			specResult.ScenarioSkippedCount++
			continue
		} else if isFailed {
			specResult.ScenarioFailedCount++
		}
		specResult.ScenarioCount++
	}
	specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Table, Table: table})
	specResult.ProtoSpec.Items = append(specResult.ProtoSpec.Items, scnResults...)
	return specResult
}
