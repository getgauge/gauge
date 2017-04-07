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
func getSpecsForDataTableRows(s []*gauge.Specification) (specs []*gauge.Specification) {
	for _, spec := range s {
		if spec.DataTable.IsInitialized() {
			if spec.UsesArgsInContextTeardown(spec.DataTable.Table.Headers...) {
				specs = append(specs, createSpecsForTableRows(spec, spec.Scenarios)...)
			} else {
				nonTableRelatedScenarios, tableRelatedScenarios := filterTableRelatedScenarios(spec.Scenarios, spec.DataTable.Table.Headers)
				if len(tableRelatedScenarios) > 0 {
					s := createSpecsForTableRows(spec, tableRelatedScenarios)
					s[0].Scenarios = append(s[0].Scenarios, nonTableRelatedScenarios...)
					specs = append(specs, s...)
				} else {
					specs = append(specs, createSpec(copyScenarios(nonTableRelatedScenarios, gauge.Table{}), spec.Heading, spec.FileName, &gauge.Table{}))
				}
			}
		} else {
			specs = append(specs, spec)
		}
	}
	return
}

func createSpecsForTableRows(spec *gauge.Specification, scns []*gauge.Scenario) (specs []*gauge.Specification) {
	for i := range spec.DataTable.Table.Rows() {
		t := getTableWith1Row(spec.DataTable.Table, i)
		specs = append(specs, createSpec(copyScenarios(scns, *t), spec.Heading, spec.FileName, t))
	}
	return
}

func createSpec(scns []*gauge.Scenario, heading *gauge.Heading, fileName string, table *gauge.Table) *gauge.Specification {
	s := &gauge.Specification{DataTable: gauge.DataTable{}, FileName: fileName}
	s.AddHeading(heading)
	s.AddDataTable(table)
	for _, scn := range scns {
		s.AddScenario(scn)
	}
	return s
}

func copyScenarios(scenarios []*gauge.Scenario, table gauge.Table) (scns []*gauge.Scenario) {
	for _, scn := range scenarios {
		scns = append(scns, &gauge.Scenario{
			Steps:             scn.Steps,
			Items:             scn.Items,
			Heading:           scn.Heading,
			DataTableRow:      table,
			DataTableRowIndex: 0,
			Tags:              scn.Tags,
			Comments:          scn.Comments,
			Span:              scn.Span,
		})
	}
	return
}

func getTableWith1Row(t gauge.Table, i int) *gauge.Table {
	var row [][]gauge.TableCell
	for _, c := range t.Columns {
		row = append(row, []gauge.TableCell{c[i]})
	}
	return gauge.NewTable(t.Headers, row, t.LineNo)
}

func filterTableRelatedScenarios(scenarios []*gauge.Scenario, headers []string) (otherScenarios, tableRelatedScenarios []*gauge.Scenario) {
	for _, scenario := range scenarios {
		if scenario.UsesArgsInSteps(headers...) {
			tableRelatedScenarios = append(tableRelatedScenarios, scenario)
		} else {
			otherScenarios = append(otherScenarios, scenario)
		}
	}
	return
}
