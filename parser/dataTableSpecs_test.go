/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"testing"

	"encoding/json"
	"reflect"

	"github.com/getgauge/gauge/gauge"
)

type DataTableSpecTest struct {
	specs   []*gauge.Specification
	want    int
	message string
}

var tests = []DataTableSpecTest{
	{
		specs: []*gauge.Specification{
			{
				Heading:   &gauge.Heading{},
				Scenarios: []*gauge.Scenario{{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}}},
				DataTable: gauge.DataTable{Table: gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
					{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
				}, 0)},
			},
		},
		want:    2,
		message: "Create specs for each data table row",
	},
	{
		specs: []*gauge.Specification{
			{
				Heading:   &gauge.Heading{},
				Scenarios: []*gauge.Scenario{{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}}},
			},
		},
		want:    1,
		message: "Create non data table driven specs",
	},
	{
		specs: []*gauge.Specification{
			{
				Heading:   &gauge.Heading{},
				Scenarios: []*gauge.Scenario{{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}}},
				DataTable: gauge.DataTable{Table: gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
					{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
				}, 0)},
			},
			{
				Heading:   &gauge.Heading{},
				Scenarios: []*gauge.Scenario{{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}}},
			},
		},
		want:    3,
		message: "Create data table driven and non data table driven specs",
	},
	{
		specs: []*gauge.Specification{
			{
				Heading:   &gauge.Heading{},
				Scenarios: []*gauge.Scenario{{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "abc", ArgType: gauge.Static}}}}}},
				DataTable: gauge.DataTable{Table: gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
					{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
				}, 0)},
				Contexts: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
			},
		},
		want:    2,
		message: "Create specs with context steps using table param",
	},
	{
		specs: []*gauge.Specification{
			{
				Heading:   &gauge.Heading{},
				Scenarios: []*gauge.Scenario{{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "abc", ArgType: gauge.Static}}}}}},
				DataTable: gauge.DataTable{Table: gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
					{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
				}, 0)},
				TearDownSteps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
			},
		},
		want:    2,
		message: "Create specs with Teardown steps using table param",
	},
}

func TestGetSpecsForDataTableRows(t *testing.T) {
	for _, test := range tests {
		got := GetSpecsForDataTableRows(test.specs, gauge.NewBuildErrors())

		if len(got) != test.want {
			t.Errorf("Failed: %s. Wanted: %d specs, Got: %d specs", test.message, test.want, len(got))
		}
	}
}

func TestGetSpecsForDataTableRowsShouldHaveEqualNumberOfScenearioInSpecsScenariosAndItemCollection(t *testing.T) {
	specs := []*gauge.Specification{
		{
			Heading: &gauge.Heading{},
			Scenarios: []*gauge.Scenario{
				{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}},
				{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "param1", ArgType: gauge.Static, Name: "param1"}}}}},
			},
			DataTable: gauge.DataTable{Table: gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
				{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
			}, 0)},
		},
	}
	actualSpecs := GetSpecsForDataTableRows(specs, gauge.NewBuildErrors())
	if !containsScenario(actualSpecs[0].Scenarios, actualSpecs[0].Items) {
		itemsJSON, _ := json.Marshal(actualSpecs[0].Items)
		scnJSON, _ := json.Marshal(actualSpecs[0].Scenarios)
		t.Errorf("Failed: Wanted items:\n\n%s\n\nto contain all scenarios: \n\n%s", itemsJSON, scnJSON)
	}
}

func containsScenario(scenarios []*gauge.Scenario, items []gauge.Item) bool {
	for _, scenario := range scenarios {
		contains := false
		for _, item := range items {
			if item.Kind() == gauge.ScenarioKind && reflect.DeepEqual(scenario, item.(*gauge.Scenario)) {
				contains = true
			}
		}
		if !contains {
			return false
		}
	}
	return true
}

func TestGetSpecsForDataTableRowsShouldHaveEqualNumberOfScenearioInSpecsScenariosAndItemCollectionForScenarioDataTable(t *testing.T) {
	specs := []*gauge.Specification{
		{
			Heading: &gauge.Heading{},
			Scenarios: []*gauge.Scenario{
				{
					Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
					DataTable: gauge.DataTable{Table: gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
						{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}, {Value: "row3", CellType: gauge.Static}},
					}, 0)}},
			},
		},
	}
	actualSpecs := GetSpecsForDataTableRows(specs, gauge.NewBuildErrors())

	if !containsScenario(actualSpecs[0].Scenarios, actualSpecs[0].Items) {
		itemsJSON, _ := json.Marshal(actualSpecs[0].Items)
		scnJSON, _ := json.Marshal(actualSpecs[0].Scenarios)
		t.Errorf("Failed: Wanted items:\n\n%s\n\nto contain all scenarios: \n\n%s", itemsJSON, scnJSON)
	}
}
func TestGetSpecsForDataTableRowsWithMixedScenarios(t *testing.T) {
	specs := []*gauge.Specification{
		{
			Heading: &gauge.Heading{},
			Scenarios: []*gauge.Scenario{
				{
					Heading: &gauge.Heading{Value: "Scenario with spec params"},
					Steps:   []*gauge.Step{{Args: []*gauge.StepArg{{Value: "specParam", ArgType: gauge.Dynamic, Name: "specParam"}}}},
					Span:    &gauge.Span{Start: 5, End: 6},
				},
				{
					Heading: &gauge.Heading{Value: "Scenario with own table"},
					Steps:   []*gauge.Step{{Args: []*gauge.StepArg{{Value: "scenarioParam", ArgType: gauge.Dynamic, Name: "scenarioParam"}}}},
					DataTable: gauge.DataTable{Table: gauge.NewTable([]string{"scenarioParam"}, [][]gauge.TableCell{
						{{Value: "scnRow1", CellType: gauge.Static}, {Value: "scnRow2", CellType: gauge.Static}},
					}, 0)},
					Span: &gauge.Span{Start: 8, End: 12},
				},
			},
			DataTable: gauge.DataTable{Table: gauge.NewTable([]string{"specParam"}, [][]gauge.TableCell{
				{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
			}, 0)},
		},
	}

	actualSpecs := GetSpecsForDataTableRows(specs, gauge.NewBuildErrors())

	if len(actualSpecs) != 2 {
		t.Errorf("Expected 2 specs (one per spec table row), got %d", len(actualSpecs))
	}

	firstSpec := actualSpecs[0]
	if len(firstSpec.Scenarios) != 3 {
		t.Errorf("First spec should have 3 scenarios (1 using spec param + 2 scenario table iterations), got %d", len(firstSpec.Scenarios))
	}

	secondSpec := actualSpecs[1]
	if len(secondSpec.Scenarios) != 1 {
		t.Errorf("Second spec should have 1 scenario (only the one using spec param), got %d", len(secondSpec.Scenarios))
	}

	scenarioWithOwnTableCount := 0
	for _, scn := range firstSpec.Scenarios {
		if scn.Heading.Value == "Scenario with own table" {
			scenarioWithOwnTableCount++
			if scn.SpecDataTableRow.IsInitialized() {
				t.Errorf("Scenario with own table should not have spec table row when it doesn't use spec params")
			}
		}
	}

	if scenarioWithOwnTableCount != 2 {
		t.Errorf("Expected 2 iterations of scenario with own table in first spec, got %d", scenarioWithOwnTableCount)
	}
}

func TestGetTableWithOneRow(t *testing.T) {
	table := gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
		{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
	}, 0)

	want := *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{{{Value: "row1", CellType: gauge.Static}}}, 0)

	got := *getTableWithOneRow(table, 0)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Failed: Table with 1 row. Wanted: %v, Got: %v", want, got)
	}
}

func TestCreateSpecsForSpecTableRows(t *testing.T) {
	spec := &gauge.Specification{
		Heading:   &gauge.Heading{},
		Scenarios: []*gauge.Scenario{
			{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}},
		},
		DataTable: gauge.DataTable{
			Table: gauge.NewTable(
				[]string{"header"},
				[][]gauge.TableCell{{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}}},
				0,
			),
		},
		Contexts: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
		Items: []gauge.Item{
			&gauge.DataTable{
				Table: gauge.NewTable(
					[]string{"header"},
					[][]gauge.TableCell{{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}}},
					0,
				),
			},
			&gauge.Scenario{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}},
		},
		TearDownSteps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "abc", ArgType: gauge.Static}}}},
	}

	want := []*gauge.Specification{
		{
			Heading: &gauge.Heading{},
			Scenarios: []*gauge.Scenario{
				{
					Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
					SpecDataTableRow: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{{{Value: "row1", CellType: gauge.Static}},}, 0),
					SpecDataTableRowIndex: 0,
					ScenarioDataTableRowIndex: -1,
				},
			},
			DataTable: gauge.DataTable{
				Table: gauge.NewTable(
					[]string{"header"},
					[][]gauge.TableCell{{{Value: "row1", CellType: gauge.Static}}},
					0,
				),
			},
			Contexts: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
			Items: []gauge.Item{
				&gauge.DataTable{
					Table: gauge.NewTable(
						[]string{"header"},
						[][]gauge.TableCell{{{Value: "row1", CellType: gauge.Static}}},
						0,
					),
				},
				&gauge.Scenario{
					Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
					SpecDataTableRow: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{{{Value: "row1", CellType: gauge.Static}}}, 0),
					SpecDataTableRowIndex: 0,
					ScenarioDataTableRowIndex: -1,
				},
			},
			TearDownSteps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "abc", ArgType: gauge.Static}}}},
		},
		{
			Heading: &gauge.Heading{},
			Scenarios: []*gauge.Scenario{
				{
					Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
					SpecDataTableRow: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{{{Value: "row2", CellType: gauge.Static}},}, 0),
					SpecDataTableRowIndex: 1,
					ScenarioDataTableRowIndex: -1,
				},
			},
			DataTable: gauge.DataTable{
				Table: gauge.NewTable(
					[]string{"header"},
					[][]gauge.TableCell{{{Value: "row2", CellType: gauge.Static}}},
					0,
				),
			},
			Contexts: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
			Items: []gauge.Item{
				&gauge.DataTable{
					Table: gauge.NewTable(
						[]string{"header"},
						[][]gauge.TableCell{{{Value: "row2", CellType: gauge.Static}}},
						0,
					),
				},
				&gauge.Scenario{
					Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
					SpecDataTableRow: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{{{Value: "row2", CellType: gauge.Static}}}, 0),
					SpecDataTableRowIndex: 1,
					ScenarioDataTableRowIndex: -1,
				},
			},
			TearDownSteps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "abc", ArgType: gauge.Static}}}},
		},
	}

	got := createSpecsForSpecTableRows(spec, spec.Scenarios, gauge.NewBuildErrors())

	if !reflect.DeepEqual(want, got) {
		gotJSON, _ := json.Marshal(got)
		wantJSON, _ := json.Marshal(want)
		t.Errorf("Failed: Create specs for table row.\n\tWanted: %v\n\tGot: %v", string(wantJSON), string(gotJSON))
	}
}

func TestDefaultTableRowIndicesAreSetForScenariosWithoutTableAccess(t *testing.T) {
	specs := []*gauge.Specification{
		{
			Heading: &gauge.Heading{},
			Scenarios: []*gauge.Scenario{
				{
					Heading: &gauge.Heading{Value: "Scenario without any data table"},
					Steps:   []*gauge.Step{{Args: []*gauge.StepArg{{Value: "static value", ArgType: gauge.Static}}}},
				},
				{
					Heading: &gauge.Heading{Value: "Scenario with spec data table params"},
					Steps:   []*gauge.Step{{Args: []*gauge.StepArg{{Value: "specParam", ArgType: gauge.Dynamic, Name: "specParam"}}}},
				},
				{
					Heading: &gauge.Heading{Value: "Scenario with scenario data table params"},
					Steps:   []*gauge.Step{{Args: []*gauge.StepArg{{Value: "scenarioParam", ArgType: gauge.Dynamic, Name: "scenarioParam"}}}},
					DataTable: gauge.DataTable{
						Table: gauge.NewTable(
							[]string{"scenarioParam"},
							[][]gauge.TableCell{{{Value: "scnRow1", CellType: gauge.Static}, {Value: "scnRow2", CellType: gauge.Static}}},
							0,
						),
					},
				},
				{
					Heading: &gauge.Heading{Value: "Scenario with spec and scenario data table params"},
					Steps:   []*gauge.Step{{Args: []*gauge.StepArg{
						{Value: "specParam", ArgType: gauge.Dynamic, Name: "specParam"},
						{Value: "scenarioParam", ArgType: gauge.Dynamic, Name: "scenarioParam"},
					}}},
					DataTable: gauge.DataTable{
						Table: gauge.NewTable(
							[]string{"scenarioParam"},
							[][]gauge.TableCell{{{Value: "scenarioTableValue", CellType: gauge.Static}}},
							0,
						),
					},
				},
			},
			DataTable: gauge.DataTable{
				Table: gauge.NewTable(
					[]string{"specParam"},
					[][]gauge.TableCell{{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}}},
					0,
				),
			},
		},
	}

	actualSpecs := GetSpecsForDataTableRows(specs, gauge.NewBuildErrors())

	if len(actualSpecs) != 2 {
		t.Errorf("Expected 2 spec, got %d", len(actualSpecs))
	}

	firstSpec := actualSpecs[0]
	if len(firstSpec.Scenarios) != 5 {
		t.Errorf("First spec should have 5 scenarios, got %d", len(firstSpec.Scenarios))
	}

	firstScenario := firstSpec.Scenarios[0]
	if firstScenario.Heading.Value != "Scenario with spec data table params" {
		t.Errorf("Wrong name for third scenario, found name: %v", firstScenario.Heading.Value)
	}
	isCorrect, errMessage := verifyTableRowIndices(firstScenario, 0, -1)
	if !isCorrect {
		t.Errorf("%v", errMessage)
	}

	secondScenario := firstSpec.Scenarios[1]
	if secondScenario.Heading.Value != "Scenario with spec and scenario data table params" {
		t.Errorf("Wrong name for fourth scenario, found name: %v", secondScenario.Heading.Value)
	}
	isCorrect, errMessage = verifyTableRowIndices(secondScenario, 0, 0)
	if !isCorrect {
		t.Errorf("%v", errMessage)
	}

	thirdScenario := firstSpec.Scenarios[2]
	if thirdScenario.Heading.Value != "Scenario without any data table" {
		t.Errorf("Wrong name for first scenario, found name: %v", thirdScenario.Heading.Value)
	}
	isCorrect, errMessage = verifyTableRowIndices(thirdScenario, -1, -1)
	if !isCorrect {
		t.Errorf("%v", errMessage)
	}

	fourthScenario := firstSpec.Scenarios[3]
	if fourthScenario.Heading.Value != "Scenario with scenario data table params" {
		t.Errorf("Wrong name for second scenario, found name: %v", fourthScenario.Heading.Value)
	}
	isCorrect, errMessage = verifyTableRowIndices(fourthScenario, -1, 0)
	if !isCorrect {
		t.Errorf("%v", errMessage)
	}

	secondSpec := actualSpecs[1]
	if len(secondSpec.Scenarios) != 2 {
		t.Errorf("Second spec should have 2 scenarios, got %d", len(secondSpec.Scenarios))
	}

	firstScenarioOfSecondSpec := secondSpec.Scenarios[0]
	if firstScenarioOfSecondSpec.Heading.Value != "Scenario with spec data table params" {
		t.Errorf("Wrong name for first scenario of second spec, found name: %v", firstScenarioOfSecondSpec.Heading.Value)
	}
	isCorrect, errMessage = verifyTableRowIndices(firstScenarioOfSecondSpec, 1, -1)
	if !isCorrect {
		t.Errorf("%v", errMessage)
	}

	secondScenarioOfSecondSpec := secondSpec.Scenarios[1]
	if secondScenarioOfSecondSpec.Heading.Value != "Scenario with spec and scenario data table params" {
		t.Errorf("Wrong name for first scenario of second spec, found name: %v", secondScenarioOfSecondSpec.Heading.Value)
	}
	isCorrect, errMessage = verifyTableRowIndices(secondScenarioOfSecondSpec, 1, 0)
	if !isCorrect {
		t.Errorf("%v", errMessage)
	}

}

func verifyTableRowIndices(scenario *gauge.Scenario, expectedSpecTableRowIndex int, expectedScenarioTableRowIndex int) (bool, string) {
	if scenario.SpecDataTableRowIndex != expectedSpecTableRowIndex {
		return false, "SpecTableRowIndex is wrong"
	}
	if scenario.ScenarioDataTableRowIndex != expectedScenarioTableRowIndex {
		return false, "ScenarioTableRowIndex is wrong"
	}
	return true, "all good"
}
