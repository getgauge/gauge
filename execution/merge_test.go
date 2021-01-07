/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"testing"

	"reflect"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
)

type stat struct {
	failed  int
	skipped int
	total   int
}

var statsTests = []struct {
	status  gm.ExecutionStatus
	want    stat
	message string
}{
	{gm.ExecutionStatus_FAILED, stat{failed: 1, total: 1}, "Scenario Failure"},
	{gm.ExecutionStatus_SKIPPED, stat{skipped: 1, total: 1}, "Scenario Skipped"},
	{gm.ExecutionStatus_PASSED, stat{total: 1}, "Scenario Passed"},
}

func TestModifySpecStats(t *testing.T) {
	for _, test := range statsTests {
		res := &result.SpecResult{}

		modifySpecStats(&gm.ProtoScenario{ExecutionStatus: test.status}, res)
		got := stat{failed: res.ScenarioFailedCount, skipped: res.ScenarioSkippedCount, total: res.ScenarioCount}

		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Modify spec stats failed for %s. Want: %v , Got: %v", test.message, test.want, got)
		}
	}
}

func TestAggregateDataTableScnStats(t *testing.T) {
	res := &result.SpecResult{}
	scns := map[string][]*gm.ProtoTableDrivenScenario{
		"heading1": {
			{Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED}},
			{Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_FAILED}},
			{Scenario: &gm.ProtoScenario{
				ExecutionStatus: gm.ExecutionStatus_SKIPPED,
				SkipErrors:      []string{"--table-rows"},
			}},
		},
		"heading2": {{Scenario: &gm.ProtoScenario{
			ExecutionStatus: gm.ExecutionStatus_SKIPPED,
			SkipErrors:      []string{""},
		}}},
		"heading3": {{Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED}}},
		"heading4": {{Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_FAILED}}},
	}

	aggregateDataTableScnStats(scns, res)

	got := stat{failed: res.ScenarioFailedCount, skipped: res.ScenarioSkippedCount, total: res.ScenarioCount}
	want := stat{failed: 2, skipped: 1, total: 5}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Aggregate data table scenario stats failed. Want: %v , Got: %v", want, got)
	}
}

func TestMergeResults(t *testing.T) {
	got := mergeResults([]*result.SpecResult{
		{
			ProtoSpec: &gm.ProtoSpec{
				SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"d"}}}}},
					{ItemType: gm.ProtoItem_Scenario, Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading1"}},
					{
						ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
							Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading2"},
							TableRowIndex: 2,
						},
					},
				},
			}, ExecutionTime: int64(1),
		},
		{
			ProtoSpec: &gm.ProtoSpec{
				SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"b"}}}}},
					{
						ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
							Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading2"},
							TableRowIndex: 0,
						},
					},
				},
			}, ExecutionTime: int64(2),
		},
		{
			ProtoSpec: &gm.ProtoSpec{
				SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"c"}}}}},
					{
						ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
							Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading2"},
							TableRowIndex: 1,
						},
					},
				},
			}, ExecutionTime: int64(2),
		},
	})
	want := &result.SpecResult{
		ProtoSpec: &gm.ProtoSpec{
			SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
			Items: []*gm.ProtoItem{
				{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"d"}}, {Cells: []string{"b"}}, {Cells: []string{"c"}}}}},
				{ItemType: gm.ProtoItem_Scenario, Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading1"}},
				{
					ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading2"},
						TableRowIndex: 0,
					},
				},
				{
					ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading2"},
						TableRowIndex: 1,
					},
				},
				{
					ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading2"},
						TableRowIndex: 2,
					},
				},
			}, IsTableDriven: false,
		},
		ScenarioCount: 4, ScenarioSkippedCount: 0, ScenarioFailedCount: 0, IsFailed: false, Skipped: false, ExecutionTime: int64(5),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge data table spec results failed.\n\tWant: %v\n\tGot: %v", want, got)
	}
}

func TestMergeResultsWithPreHookFailure(t *testing.T) {
	got := mergeResults([]*result.SpecResult{
		{
			ProtoSpec: &gm.ProtoSpec{
				PreHookFailures: []*gm.ProtoHookFailure{{StackTrace: "stacktrace"}},
				SpecHeading:     "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"b"}}}}},
				},
			}, ExecutionTime: int64(1),
		},
		{
			ProtoSpec: &gm.ProtoSpec{
				PreHookFailures: []*gm.ProtoHookFailure{{StackTrace: "stacktrace1"}},
				SpecHeading:     "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"d"}}}}},
				},
			}, ExecutionTime: int64(2),
		},
		{
			ProtoSpec: &gm.ProtoSpec{
				PreHookFailures: []*gm.ProtoHookFailure{{StackTrace: "stacktrace2"}},
				SpecHeading:     "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"c"}}}}},
				},
			}, ExecutionTime: int64(2),
		},
	})
	want := &result.SpecResult{
		ProtoSpec: &gm.ProtoSpec{
			PreHookFailures: []*gm.ProtoHookFailure{{StackTrace: "stacktrace"}, {StackTrace: "stacktrace1", TableRowIndex: 1}, {StackTrace: "stacktrace2", TableRowIndex: 2}},
			SpecHeading:     "heading", FileName: "filename", Tags: []string{"tags"},
			Items: []*gm.ProtoItem{
				{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"b"}}, {Cells: []string{"d"}}, {Cells: []string{"c"}}}}},
			}, IsTableDriven: false,
		},
		ScenarioCount: 0, ScenarioSkippedCount: 0, ScenarioFailedCount: 0, IsFailed: false, Skipped: false, ExecutionTime: int64(5),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge data table spec results failed.\n\tWant: %v\n\tGot: %v", want, got)
	}
}

func TestMergeSkippedResults(t *testing.T) {
	got := mergeResults([]*result.SpecResult{
		{
			ProtoSpec: &gm.ProtoSpec{
				SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"b"}}}}},
					{ItemType: gm.ProtoItem_Scenario, Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_SKIPPED, ScenarioHeading: "scenario Heading1", SkipErrors: []string{"error"}}},
					{
						ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
							Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_SKIPPED, ScenarioHeading: "scenario Heading2", SkipErrors: []string{"error"}},
							TableRowIndex: 0,
						},
					},
				},
			}, ExecutionTime: int64(1),
			Skipped: true,
		},
		{
			ProtoSpec: &gm.ProtoSpec{
				SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"c"}}}}},
					{
						ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
							Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_SKIPPED, ScenarioHeading: "scenario Heading2", SkipErrors: []string{"error"}},
							TableRowIndex: 1,
						},
					},
				},
			}, ExecutionTime: int64(2),
			Skipped: true,
		},
	})
	want := &result.SpecResult{
		ProtoSpec: &gm.ProtoSpec{
			SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
			Items: []*gm.ProtoItem{
				{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"b"}}, {Cells: []string{"c"}}}}},
				{ItemType: gm.ProtoItem_Scenario, Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_SKIPPED, SkipErrors: []string{"error"}, ScenarioHeading: "scenario Heading1"}},
				{
					ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_SKIPPED, SkipErrors: []string{"error"}, ScenarioHeading: "scenario Heading2"},
						TableRowIndex: 0,
					},
				},
				{
					ItemType: gm.ProtoItem_TableDrivenScenario, TableDrivenScenario: &gm.ProtoTableDrivenScenario{
						Scenario:      &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_SKIPPED, SkipErrors: []string{"error"}, ScenarioHeading: "scenario Heading2"},
						TableRowIndex: 1,
					},
				},
			}, IsTableDriven: false,
		},
		ScenarioCount: 3, ScenarioSkippedCount: 3, ScenarioFailedCount: 0, IsFailed: false, Skipped: true, ExecutionTime: int64(3),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge data table spec results failed.\n\tWant: %v\n\tGot: %v", want, got)
	}
}

func TestMergeResultsExecutionTimeInParallel(t *testing.T) {
	InParallel = true

	got := mergeResults([]*result.SpecResult{
		{
			ProtoSpec: &gm.ProtoSpec{
				SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
			}, ExecutionTime: int64(1),
		},
		{
			ProtoSpec: &gm.ProtoSpec{
				SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
			}, ExecutionTime: int64(2),
		},
	})

	want := int64(2)
	InParallel = false

	if !reflect.DeepEqual(got.ExecutionTime, want) {
		t.Errorf("Execution time in parallel data table spec results.\n\tWant: %v\n\tGot: %v", want, got.ExecutionTime)
	}
}

func TestMergeDataTableSpecResults(t *testing.T) {
	res := &result.SuiteResult{
		Environment: "env",
		ProjectName: "name",
		Tags:        "tags",
		SpecResults: []*result.SpecResult{
			{
				ProtoSpec: &gm.ProtoSpec{
					SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
					Items: []*gm.ProtoItem{
						{ItemType: gm.ProtoItem_Scenario, Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading1"}},
					},
				},
			},
		},
	}
	got := mergeDataTableSpecResults(res)

	want := &result.SuiteResult{
		Environment: "env",
		ProjectName: "name",
		Tags:        "tags",
		SpecResults: []*result.SpecResult{
			{
				ProtoSpec: &gm.ProtoSpec{
					SpecHeading: "heading", FileName: "filename", Tags: []string{"tags"},
					Items: []*gm.ProtoItem{
						{ItemType: gm.ProtoItem_Scenario, Scenario: &gm.ProtoScenario{ExecutionStatus: gm.ExecutionStatus_PASSED, ScenarioHeading: "scenario Heading1"}},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge data table spec results failed.\n\tWant: %v\n\tGot: %v", want, got)
	}
}

func TestGetItems(t *testing.T) {
	table := &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}}
	res := []*result.SpecResult{{
		ProtoSpec: &gm.ProtoSpec{
			Items: []*gm.ProtoItem{
				{ItemType: gm.ProtoItem_Table},
				{ItemType: gm.ProtoItem_Scenario},
				{ItemType: gm.ProtoItem_TableDrivenScenario},
			},
		},
	}}
	scnRes := []*gm.ProtoItem{
		{ItemType: gm.ProtoItem_Scenario}, {ItemType: gm.ProtoItem_TableDrivenScenario}, {ItemType: gm.ProtoItem_Scenario},
	}
	got := getItems(table, scnRes, res)

	want := []*gm.ProtoItem{{ItemType: gm.ProtoItem_Table, Table: table}, {ItemType: gm.ProtoItem_Scenario}, {ItemType: gm.ProtoItem_TableDrivenScenario}, {ItemType: gm.ProtoItem_Scenario}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge data table spec results failed.\n\tWant: %v\n\tGot: %v", want, got)
	}
}

func TestHasTableDrivenSpec(t *testing.T) {
	type testcase struct {
		results []*result.SpecResult
		want    bool
	}

	cases := []testcase{
		{
			results: []*result.SpecResult{
				{
					ProtoSpec: &gm.ProtoSpec{
						IsTableDriven: false,
					},
				},
				{
					ProtoSpec: &gm.ProtoSpec{
						IsTableDriven: true,
					},
				},
			},
			want: true,
		},
		{
			results: []*result.SpecResult{
				{
					ProtoSpec: &gm.ProtoSpec{
						IsTableDriven: false,
					},
				},
				{
					ProtoSpec: &gm.ProtoSpec{
						IsTableDriven: false,
					},
				},
			},
			want: false,
		},
	}

	for _, c := range cases {
		got := hasTableDrivenSpec(c.results)
		if got != c.want {
			t.Errorf("Expected hasTableDrivenSpec to be %t, got %t", c.want, got)
		}
	}
}

func TestMergeResultWithMesages(t *testing.T) {
	got := mergeResults([]*result.SpecResult{
		{
			ProtoSpec: &gm.ProtoSpec{
				PreHookFailures: []*gm.ProtoHookFailure{{StackTrace: "stacktrace"}},
				SpecHeading:     "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"b"}}}}},
				},
				PreHookMessages: []string{"Hello"},
			}, ExecutionTime: int64(1),
		},
		{
			ProtoSpec: &gm.ProtoSpec{
				PreHookFailures: []*gm.ProtoHookFailure{{StackTrace: "stacktrace1"}},
				SpecHeading:     "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"d"}}}}},
				},
			}, ExecutionTime: int64(2),
		},
		{
			ProtoSpec: &gm.ProtoSpec{
				PreHookFailures: []*gm.ProtoHookFailure{{StackTrace: "stacktrace2"}},
				SpecHeading:     "heading", FileName: "filename", Tags: []string{"tags"},
				Items: []*gm.ProtoItem{
					{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"c"}}}}},
				},
				PostHookMessages: []string{"Bye"},
			}, ExecutionTime: int64(2),
		},
	})
	want := &result.SpecResult{
		ProtoSpec: &gm.ProtoSpec{
			PreHookFailures: []*gm.ProtoHookFailure{{StackTrace: "stacktrace"}, {StackTrace: "stacktrace1", TableRowIndex: 1}, {StackTrace: "stacktrace2", TableRowIndex: 2}},
			SpecHeading:     "heading", FileName: "filename", Tags: []string{"tags"},
			Items: []*gm.ProtoItem{
				{ItemType: gm.ProtoItem_Table, Table: &gm.ProtoTable{Headers: &gm.ProtoTableRow{Cells: []string{"a"}}, Rows: []*gm.ProtoTableRow{{Cells: []string{"b"}}, {Cells: []string{"d"}}, {Cells: []string{"c"}}}}},
			},
			PreHookMessages:  []string{"Hello"},
			PostHookMessages: []string{"Bye"},
			IsTableDriven: false,
		},
		ScenarioCount: 0, ScenarioSkippedCount: 0, ScenarioFailedCount: 0, IsFailed: false, Skipped: false, ExecutionTime: int64(5),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge data table spec results failed.\n\tWant: %v\n\tGot: %v", want, got)
	}
}
