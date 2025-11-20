/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import (
	"testing"
)

func TestNewLazyScenarioCollection(t *testing.T) {
	template := &Scenario{
		Heading: &Heading{Value: "Test Scenario"},
		Steps:   []*Step{{LineText: "Step 1"}},
	}

	// NewTable expects columns, not rows
	scenarioTable := NewTable([]string{"header1", "header2"}, [][]TableCell{
		{{Value: "row1col1", CellType: Static}, {Value: "row2col1", CellType: Static}}, // column 1
		{{Value: "row1col2", CellType: Static}, {Value: "row2col2", CellType: Static}}, // column 2
	}, 0)

	lsc := NewLazyScenarioCollection(template, nil, scenarioTable, false, 0)

	if lsc.Template != template {
		t.Errorf("Expected template to match")
	}
	if lsc.TotalIterations != 2 {
		t.Errorf("Expected 2 iterations, got %d", lsc.TotalIterations)
	}
}

func TestLazyScenarioIterator(t *testing.T) {
	template := &Scenario{
		Heading: &Heading{Value: "Test Scenario"},
		Steps:   []*Step{{LineText: "Step 1"}},
		Span:    &Span{Start: 1, End: 2},
	}

	// NewTable expects columns, not rows - so for 3 rows and 2 columns, we need 2 columns with 3 cells each
	scenarioTable := NewTable([]string{"header1", "header2"}, [][]TableCell{
		{{Value: "row1col1", CellType: Static}, {Value: "row2col1", CellType: Static}, {Value: "row3col1", CellType: Static}}, // column 1
		{{Value: "row1col2", CellType: Static}, {Value: "row2col2", CellType: Static}, {Value: "row3col2", CellType: Static}}, // column 2
	}, 0)

	lsc := NewLazyScenarioCollection(template, nil, scenarioTable, false, 0)
	iterator := lsc.Iterator()

	count := 0
	for scenario, hasNext := iterator.Next(); hasNext; scenario, hasNext = iterator.Next() {
		if scenario.Heading.Value != "Test Scenario" {
			t.Errorf("Expected scenario heading to be 'Test Scenario', got '%s'", scenario.Heading.Value)
		}
		if scenario.ScenarioDataTableRowIndex != count {
			t.Errorf("Expected row index %d, got %d", count, scenario.ScenarioDataTableRowIndex)
		}
		if !scenario.ScenarioDataTableRow.IsInitialized() {
			t.Error("Expected scenario data table row to be initialized")
		}
		count++
	}

	if count != 3 {
		t.Errorf("Expected 3 iterations, got %d", count)
	}
}

func TestLazyScenarioWithSpecTable(t *testing.T) {
	template := &Scenario{
		Heading: &Heading{Value: "Test Scenario"},
		Steps:   []*Step{{LineText: "Step with <header1> and <specHeader>"}},
	}

	specTable := NewTable([]string{"specHeader"}, [][]TableCell{
		{{Value: "specValue", CellType: Static}},
	}, 0)

	scenarioTable := NewTable([]string{"header1"}, [][]TableCell{
		{{Value: "value1", CellType: Static}},
		{{Value: "value2", CellType: Static}},
	}, 0)

	lsc := NewLazyScenarioCollection(template, specTable, scenarioTable, true, 0)
	iterator := lsc.Iterator()

	scenario, hasNext := iterator.Next()
	if !hasNext {
		t.Fatal("Expected at least one scenario")
	}

	if !scenario.SpecDataTableRow.IsInitialized() {
		t.Error("Expected spec data table row to be initialized")
	}
	if scenario.SpecDataTableRowIndex != 0 {
		t.Error("Expected spec data table row index to be 0")
	}
	if !scenario.ScenarioDataTableRow.IsInitialized() {
		t.Error("Expected scenario data table row to be initialized")
	}
}

func TestLazyScenarioIteratorTermination(t *testing.T) {
	template := &Scenario{
		Heading: &Heading{Value: "Test Scenario"},
	}

	scenarioTable := NewTable([]string{"header1"}, [][]TableCell{
		{{Value: "value1", CellType: Static}},
	}, 0)

	lsc := NewLazyScenarioCollection(template, nil, scenarioTable, false, 0)
	iterator := lsc.Iterator()

	// First iteration should succeed
	_, hasNext := iterator.Next()
	if !hasNext {
		t.Error("Expected first iteration to succeed")
	}

	// Second iteration should fail
	_, hasNext = iterator.Next()
	if hasNext {
		t.Error("Expected iterator to terminate after all rows")
	}
}

func TestLazyScenarioWithNoTable(t *testing.T) {
	template := &Scenario{
		Heading: &Heading{Value: "Test Scenario"},
	}

	lsc := NewLazyScenarioCollection(template, nil, nil, false, 0)

	if lsc.TotalIterations != 0 {
		t.Errorf("Expected 0 iterations for nil table, got %d", lsc.TotalIterations)
	}

	iterator := lsc.Iterator()
	_, hasNext := iterator.Next()
	if hasNext {
		t.Error("Expected no iterations when scenario table is nil")
	}
}
