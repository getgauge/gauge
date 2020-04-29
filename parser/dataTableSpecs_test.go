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
				DataTable: gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
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
				DataTable: gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
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
				DataTable: gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
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
				DataTable: gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
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

func TestGetTableWithOneRow(t *testing.T) {
	table := *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
		{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
	}, 0)

	want := *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{{{Value: "row1", CellType: gauge.Static}}}, 0)

	got := *getTableWithOneRow(table, 0)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Failed: Table with 1 row. Wanted: %v, Got: %v", want, got)
	}
}

func TestCreateSpecsForTableRows(t *testing.T) {
	spec := &gauge.Specification{
		Heading:   &gauge.Heading{},
		Scenarios: []*gauge.Scenario{{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}}},
		DataTable: gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
			{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
		}, 0)},
		Contexts: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
		Items: []gauge.Item{
			&gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
				{{Value: "row1", CellType: gauge.Static}, {Value: "row2", CellType: gauge.Static}},
			}, 0)},
			&gauge.Scenario{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}},
		},
		TearDownSteps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "abc", ArgType: gauge.Static}}}},
	}

	want := []*gauge.Specification{
		{
			Heading: &gauge.Heading{},
			Scenarios: []*gauge.Scenario{{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}, SpecDataTableRow: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
				{{Value: "row1", CellType: gauge.Static}},
			}, 0), SpecDataTableRowIndex: 0}},
			DataTable: gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
				{{Value: "row1", CellType: gauge.Static}},
			}, 0)},
			Contexts: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
			Items: []gauge.Item{
				&gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
					{{Value: "row1", CellType: gauge.Static}},
				}, 0)},
				&gauge.Scenario{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}, SpecDataTableRow: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
					{{Value: "row1", CellType: gauge.Static}},
				}, 0), SpecDataTableRowIndex: 0},
			},
			TearDownSteps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "abc", ArgType: gauge.Static}}}},
		},
		{
			Heading: &gauge.Heading{},
			Scenarios: []*gauge.Scenario{{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}, SpecDataTableRow: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
				{{Value: "row2", CellType: gauge.Static}},
			}, 0), SpecDataTableRowIndex: 1}},
			DataTable: gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
				{{Value: "row2", CellType: gauge.Static}},
			}, 0)},
			Contexts: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}},
			Items: []gauge.Item{
				&gauge.DataTable{Table: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
					{{Value: "row2", CellType: gauge.Static}},
				}, 0)},
				&gauge.Scenario{Steps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "header", ArgType: gauge.Dynamic, Name: "header"}}}}, SpecDataTableRow: *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{
					{{Value: "row2", CellType: gauge.Static}},
				}, 0), SpecDataTableRowIndex: 1},
			},
			TearDownSteps: []*gauge.Step{{Args: []*gauge.StepArg{{Value: "abc", ArgType: gauge.Static}}}},
		},
	}

	got := createSpecsForTableRows(spec, spec.Scenarios, gauge.NewBuildErrors())

	if !reflect.DeepEqual(want, got) {
		gotJSON, _ := json.Marshal(got)
		wantJSON, _ := json.Marshal(want)
		t.Errorf("Failed: Create specs for table row.\n\tWanted: %v\n\tGot: %v", string(wantJSON), string(gotJSON))
	}
}
