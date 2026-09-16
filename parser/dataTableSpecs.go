/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
)

// GetSpecsForDataTableRows creates a spec for each data table row
func GetSpecsForDataTableRows(s []*gauge.Specification, errMap *gauge.BuildErrors) (specs []*gauge.Specification) {
	for _, spec := range s {
		if spec.DataTable.IsInitialized() {
			if spec.UsesArgsInContextTeardown(spec.DataTable.Table.Headers...) {
				specs = append(specs, createSpecsForSpecTableRows(spec, spec.Scenarios, errMap)...)
			} else {
				nonSpecTableRelatedScenarios, specTableRelatedScenarios := FilterTableRelatedScenarios(spec.Scenarios, func(scenario *gauge.Scenario) bool {
					return scenario.UsesArgsInSteps(spec.DataTable.Table.Headers...)
				})
				if len(specTableRelatedScenarios) > 0 {
					s := createSpecsForSpecTableRows(spec, specTableRelatedScenarios, errMap)
					copiedNonSpecTableScenarios := copyScenarios(nonSpecTableRelatedScenarios, gauge.Table{}, -1, errMap)
					s[0].Scenarios = append(s[0].Scenarios, copiedNonSpecTableScenarios...)
					for _, scn := range copiedNonSpecTableScenarios { // nolint
						s[0].Items = append(s[0].Items, scn)
					}
					specs = append(specs, s...)
				} else {
					specs = append(specs, createSpec(copyScenarios(nonSpecTableRelatedScenarios, gauge.Table{}, -1, errMap), &gauge.Table{}, spec, errMap))
				}
			}
		} else {
			specs = append(specs, createSpec(copyScenarios(spec.Scenarios, gauge.Table{}, -1, errMap), &gauge.Table{}, spec, errMap))
		}
	}
	return
}

func createSpecsForSpecTableRows(spec *gauge.Specification, scns []*gauge.Scenario, errMap *gauge.BuildErrors) (specs []*gauge.Specification) {
	for i := range spec.DataTable.Table.Rows() {
		t := getTableWithOneRow(spec.DataTable.Table, i)
		newSpec := createSpec(copyScenarios(scns, *t, i, errMap), t, spec, errMap)
		specs = append(specs, newSpec)
	}
	return
}

func createSpec(scns []*gauge.Scenario, table *gauge.Table, spec *gauge.Specification, errMap *gauge.BuildErrors) *gauge.Specification {
	dt := &gauge.DataTable{Table: table, Value: spec.DataTable.Value, LineNo: spec.DataTable.LineNo, IsExternal: spec.DataTable.IsExternal}
	s := &gauge.Specification{DataTable: *dt, FileName: spec.FileName, Heading: spec.Heading, Scenarios: scns, Contexts: spec.Contexts, TearDownSteps: spec.TearDownSteps, Tags: spec.Tags}
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
	if len(errMap.SpecErrs[spec]) > 0 {
		errMap.SpecErrs[s] = errMap.SpecErrs[spec]
	}
	return s
}

func copyScenarios(scenarios []*gauge.Scenario, table gauge.Table, specTableRowIndex int, errMap *gauge.BuildErrors) (scns []*gauge.Scenario) {
	var create = func(scn *gauge.Scenario, scnTableRow gauge.Table, scnTableRowIndex int, assignSpecTable bool) *gauge.Scenario {
		newScn := &gauge.Scenario{
			Steps:    scn.Steps,
			Items:    scn.Items,
			Heading:  scn.Heading,
			Tags:     scn.Tags,
			Comments: scn.Comments,
			Span:     scn.Span,
		}
		if assignSpecTable {
			newScn.SpecDataTableRow = table
			newScn.SpecDataTableRowIndex = specTableRowIndex
		} else {
			newScn.SpecDataTableRowIndex = -1
		}
		if scnTableRow.IsInitialized() {
			newScn.ScenarioDataTableRow = scnTableRow
			newScn.ScenarioDataTableRowIndex = scnTableRowIndex
			newScn.DataTable = scn.DataTable
		} else {
			newScn.ScenarioDataTableRowIndex = -1
		}
		if len(errMap.ScenarioErrs[scn]) > 0 {
			errMap.ScenarioErrs[newScn] = errMap.ScenarioErrs[scn]
		}
		return newScn
	}
	for _, scn := range scenarios {
		if scn.DataTable.IsInitialized() {
			usesSpecParams := table.IsInitialized() && scn.UsesArgsInSteps(table.Headers...)
			scenarioTableRows := scn.DataTable.Table.GetRowCount()

			// Warn about nested tables with large datasets
			if usesSpecParams && table.GetRowCount() > 0 {
				specTableRows := table.GetRowCount()
				totalIterations := specTableRows * scenarioTableRows
				if totalIterations > 100 {
					logger.Warningf(true, "Scenario '%s' has nested data tables (spec: %d rows x scenario: %d rows = %d total iterations). This may impact performance and memory usage.",
						scn.Heading.Value, specTableRows, scenarioTableRows, totalIterations)
				}
			}

			for i := range scn.DataTable.Table.Rows() {
				t := getTableWithOneRow(scn.DataTable.Table, i)
				scns = append(scns, create(scn, *t, i, usesSpecParams))
			}
		} else {
			scns = append(scns, create(scn, gauge.Table{}, -1, table.IsInitialized()))
		}
	}
	return
}

func getTableWithOneRow(t *gauge.Table, i int) *gauge.Table {
	var row [][]gauge.TableCell
	for _, c := range t.Columns {
		row = append(row, []gauge.TableCell{c[i]})
	}
	return gauge.NewTable(t.Headers, row, t.LineNo)
}

// FilterTableRelatedScenarios filters Scenarios that are using dynamic params from data table.
func FilterTableRelatedScenarios(scenarios []*gauge.Scenario, fun func(*gauge.Scenario) bool) (otherScenarios, tableRelatedScenarios []*gauge.Scenario) {
	for _, scenario := range scenarios {
		if fun(scenario) {
			tableRelatedScenarios = append(tableRelatedScenarios, scenario)
		} else {
			otherScenarios = append(otherScenarios, scenario)
		}
	}
	return
}
