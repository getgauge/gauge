/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

// LazyScenarioCollection represents a collection of scenarios that should be generated lazily
type LazyScenarioCollection struct {
	Template         *Scenario
	SpecTable        *Table
	ScenarioTable    *Table
	UsesSpecParams   bool
	SpecRowIndex     int
	TotalIterations  int
}

// NewLazyScenarioCollection creates a new lazy scenario collection
func NewLazyScenarioCollection(template *Scenario, specTable *Table, scenarioTable *Table, usesSpecParams bool, specRowIndex int) *LazyScenarioCollection {
	totalIterations := 0
	if scenarioTable != nil && scenarioTable.IsInitialized() {
		totalIterations = scenarioTable.GetRowCount()
	}

	return &LazyScenarioCollection{
		Template:        template,
		SpecTable:       specTable,
		ScenarioTable:   scenarioTable,
		UsesSpecParams:  usesSpecParams,
		SpecRowIndex:    specRowIndex,
		TotalIterations: totalIterations,
	}
}

// Iterator returns an iterator for lazy scenario generation
func (lsc *LazyScenarioCollection) Iterator() *ScenarioIterator {
	return &ScenarioIterator{
		collection: lsc,
		current:    0,
	}
}

// ScenarioIterator generates scenarios on-demand
type ScenarioIterator struct {
	collection *LazyScenarioCollection
	current    int
}

// Next generates and returns the next scenario instance
func (si *ScenarioIterator) Next() (*Scenario, bool) {
	if si.current >= si.collection.TotalIterations {
		return nil, false
	}

	scenario := si.generateScenario(si.current)
	si.current++
	return scenario, true
}

// generateScenario creates a scenario instance for the given iteration
func (si *ScenarioIterator) generateScenario(iteration int) *Scenario {
	template := si.collection.Template
	scenarioRowIndex := iteration

	newScn := &Scenario{
		Steps:    template.Steps,
		Items:    template.Items,
		Heading:  template.Heading,
		Tags:     template.Tags,
		Comments: template.Comments,
		Span:     template.Span,
	}

	// Assign spec table row if scenario uses spec parameters
	if si.collection.UsesSpecParams && si.collection.SpecTable != nil && si.collection.SpecTable.IsInitialized() {
		newScn.SpecDataTableRow = *getTableWithOneRow(si.collection.SpecTable, si.collection.SpecRowIndex)
		newScn.SpecDataTableRowIndex = si.collection.SpecRowIndex
	}

	// Assign scenario table row
	if si.collection.ScenarioTable != nil && si.collection.ScenarioTable.IsInitialized() {
		newScn.ScenarioDataTableRow = *getTableWithOneRow(si.collection.ScenarioTable, scenarioRowIndex)
		newScn.ScenarioDataTableRowIndex = scenarioRowIndex
		newScn.DataTable = DataTable{Table: si.collection.ScenarioTable}
	}

	return newScn
}

// getTableWithOneRow extracts a single row from a table
func getTableWithOneRow(t *Table, rowIndex int) *Table {
	var row [][]TableCell
	for _, c := range t.Columns {
		row = append(row, []TableCell{c[rowIndex]})
	}
	return NewTable(t.Headers, row, t.LineNo)
}
