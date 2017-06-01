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

package parser

import "github.com/getgauge/gauge/gauge"

// Creates a spec for each data table row
func GetSpecsForDataTableRows(s []*gauge.Specification, errMap *gauge.BuildErrors) (specs []*gauge.Specification) {
	for _, spec := range s {
		if spec.DataTable.IsInitialized() {
			if spec.UsesArgsInContextTeardown(spec.DataTable.Table.Headers...) {
				specs = append(specs, createSpecsForTableRows(spec, spec.Scenarios, errMap)...)
			} else {
				nonTableRelatedScenarios, tableRelatedScenarios := FilterTableRelatedScenarios(spec.Scenarios, spec.DataTable.Table.Headers)
				if len(tableRelatedScenarios) > 0 {
					s := createSpecsForTableRows(spec, tableRelatedScenarios, errMap)
					s[0].Scenarios = append(s[0].Scenarios, nonTableRelatedScenarios...)
					specs = append(specs, s...)
				} else {
					specs = append(specs, createSpec(copyScenarios(nonTableRelatedScenarios, gauge.Table{}, 0, errMap), &gauge.Table{}, spec))
				}
			}
		} else {
			specs = append(specs, spec)
		}
	}
	return
}

func createSpecsForTableRows(spec *gauge.Specification, scns []*gauge.Scenario, errMap *gauge.BuildErrors) (specs []*gauge.Specification) {
	for i := range spec.DataTable.Table.Rows() {
		t := getTableWithOneRow(spec.DataTable.Table, i)
		newSpec := createSpec(copyScenarios(scns, *t, i, errMap), t, spec)
		if len(errMap.SpecErrs[spec]) > 0 {
			errMap.SpecErrs[newSpec] = errMap.SpecErrs[spec]
		}
		specs = append(specs, newSpec)
	}
	return
}

func createSpec(scns []*gauge.Scenario, table *gauge.Table, spec *gauge.Specification) *gauge.Specification {
	dt := &gauge.DataTable{Table: *table, Value: spec.DataTable.Value, LineNo: spec.DataTable.LineNo, IsExternal: spec.DataTable.IsExternal}
	s := &gauge.Specification{DataTable: *dt, FileName: spec.FileName, Heading: spec.Heading, Scenarios: scns, Contexts: spec.Contexts, TearDownSteps: spec.TearDownSteps}
	index := 0
	for _, item := range spec.Items {
		if item.Kind() == gauge.DataTableKind {
			item = dt
		} else if item.Kind() == gauge.ScenarioKind {
			if len(scns) <= index {
				continue
			}
			item = scns[index]
			index++
		}
		s.Items = append(s.Items, item)
	}
	for i := index; i < len(scns); i++ {
		s.Items = append(s.Items, scns[i])
	}
	return s
}

func copyScenarios(scenarios []*gauge.Scenario, table gauge.Table, i int, errMap *gauge.BuildErrors) (scns []*gauge.Scenario) {
	for _, scn := range scenarios {
		newScn := &gauge.Scenario{
			Steps:             scn.Steps,
			Items:             scn.Items,
			Heading:           scn.Heading,
			DataTableRow:      table,
			DataTableRowIndex: i,
			Tags:              scn.Tags,
			Comments:          scn.Comments,
			Span:              scn.Span,
		}
		if len(errMap.ScenarioErrs[scn]) > 0 {
			errMap.ScenarioErrs[newScn] = errMap.ScenarioErrs[scn]
		}
		scns = append(scns, newScn)
	}
	return
}

func getTableWithOneRow(t gauge.Table, i int) *gauge.Table {
	var row [][]gauge.TableCell
	for _, c := range t.Columns {
		row = append(row, []gauge.TableCell{c[i]})
	}
	return gauge.NewTable(t.Headers, row, t.LineNo)
}

func FilterTableRelatedScenarios(scenarios []*gauge.Scenario, headers []string) (otherScenarios, tableRelatedScenarios []*gauge.Scenario) {
	for _, scenario := range scenarios {
		if scenario.UsesArgsInSteps(headers...) {
			tableRelatedScenarios = append(tableRelatedScenarios, scenario)
		} else {
			otherScenarios = append(otherScenarios, scenario)
		}
	}
	return
}
