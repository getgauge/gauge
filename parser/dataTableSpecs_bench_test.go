/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"os"
	"testing"

	"github.com/getgauge/gauge/gauge"
)

// createTestScenarioWithTable creates a scenario with a data table of specified size
func createTestScenarioWithTable(rows, cols int) *gauge.Scenario {
	headers := make([]string, cols)
	columns := make([][]gauge.TableCell, cols)

	for i := 0; i < cols; i++ {
		headers[i] = "header" + string(rune('A'+i))
		columns[i] = make([]gauge.TableCell, rows)
		for j := 0; j < rows; j++ {
			columns[i][j] = gauge.TableCell{
				Value:    "value",
				CellType: gauge.Static,
			}
		}
	}

	table := gauge.NewTable(headers, columns, 0)

	// Create step args that reference the table headers
	stepArgs := make([]*gauge.StepArg, cols)
	for i := 0; i < cols; i++ {
		stepArgs[i] = &gauge.StepArg{
			Name:    headers[i],
			Value:   headers[i],
			ArgType: gauge.Dynamic,
		}
	}

	return &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Benchmark Scenario"},
		DataTable: gauge.DataTable{
			Table: table,
		},
		Steps: []*gauge.Step{
			{
				LineText: "Step with parameters",
				Args:     stepArgs,
			},
		},
		Span: &gauge.Span{Start: 1, End: 10},
	}
}

// BenchmarkEagerScenarioCreation benchmarks eager scenario creation
func BenchmarkEagerScenarioCreation(b *testing.B) {
	benchmarks := []struct {
		name string
		rows int
		cols int
	}{
		{"10rows_2cols", 10, 2},
		{"50rows_2cols", 50, 2},
		{"100rows_2cols", 100, 2},
		{"500rows_2cols", 500, 2},
		{"1000rows_2cols", 1000, 2},
		{"100rows_5cols", 100, 5},
		{"500rows_5cols", 500, 5},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Force eager mode
			os.Setenv("scenario_init_strategy", "eager")
			defer os.Unsetenv("scenario_init_strategy")

			scenario := createTestScenarioWithTable(bm.rows, bm.cols)
			scenarios := []*gauge.Scenario{scenario}
			errMap := gauge.NewBuildErrors()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = copyScenarios(scenarios, gauge.Table{}, 0, errMap)
			}
		})
	}
}

// BenchmarkLazyScenarioCreation benchmarks lazy scenario creation
func BenchmarkLazyScenarioCreation(b *testing.B) {
	benchmarks := []struct {
		name string
		rows int
		cols int
	}{
		{"10rows_2cols", 10, 2},
		{"50rows_2cols", 50, 2},
		{"100rows_2cols", 100, 2},
		{"500rows_2cols", 500, 2},
		{"1000rows_2cols", 1000, 2},
		{"100rows_5cols", 100, 5},
		{"500rows_5cols", 500, 5},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Force lazy mode
			os.Setenv("scenario_init_strategy", "lazy")
			defer os.Unsetenv("scenario_init_strategy")

			scenario := createTestScenarioWithTable(bm.rows, bm.cols)
			scenarios := []*gauge.Scenario{scenario}
			errMap := gauge.NewBuildErrors()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = copyScenarios(scenarios, gauge.Table{}, 0, errMap)
			}
		})
	}
}

// BenchmarkNestedTableEagerCreation benchmarks eager creation with nested tables
func BenchmarkNestedTableEagerCreation(b *testing.B) {
	benchmarks := []struct {
		name      string
		specRows  int
		scenRows  int
	}{
		{"spec10_scen10", 10, 10},
		{"spec20_scen20", 20, 20},
		{"spec10_scen100", 10, 100},
		{"spec50_scen50", 50, 50},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Force eager mode
			os.Setenv("scenario_init_strategy", "eager")
			defer os.Unsetenv("scenario_init_strategy")

			// Create spec table
			specHeaders := []string{"specParam"}
			specCols := make([][]gauge.TableCell, 1)
			specCols[0] = make([]gauge.TableCell, bm.specRows)
			for i := 0; i < bm.specRows; i++ {
				specCols[0][i] = gauge.TableCell{Value: "specValue", CellType: gauge.Static}
			}
			specTable := gauge.NewTable(specHeaders, specCols, 0)

			// Create scenario with table that uses spec params
			scenario := createTestScenarioWithTable(bm.scenRows, 2)
			// Add step that uses spec param
			scenario.Steps = append(scenario.Steps, &gauge.Step{
				LineText: "Step with <specParam>",
				Args: []*gauge.StepArg{
					{Name: "specParam", Value: "specParam", ArgType: gauge.Dynamic},
				},
			})

			scenarios := []*gauge.Scenario{scenario}
			errMap := gauge.NewBuildErrors()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = copyScenarios(scenarios, *specTable, 0, errMap)
			}
		})
	}
}

// BenchmarkNestedTableLazyCreation benchmarks lazy creation with nested tables
func BenchmarkNestedTableLazyCreation(b *testing.B) {
	benchmarks := []struct {
		name      string
		specRows  int
		scenRows  int
	}{
		{"spec10_scen10", 10, 10},
		{"spec20_scen20", 20, 20},
		{"spec10_scen100", 10, 100},
		{"spec50_scen50", 50, 50},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Force lazy mode
			os.Setenv("scenario_init_strategy", "lazy")
			defer os.Unsetenv("scenario_init_strategy")

			// Create spec table
			specHeaders := []string{"specParam"}
			specCols := make([][]gauge.TableCell, 1)
			specCols[0] = make([]gauge.TableCell, bm.specRows)
			for i := 0; i < bm.specRows; i++ {
				specCols[0][i] = gauge.TableCell{Value: "specValue", CellType: gauge.Static}
			}
			specTable := gauge.NewTable(specHeaders, specCols, 0)

			// Create scenario with table that uses spec params
			scenario := createTestScenarioWithTable(bm.scenRows, 2)
			// Add step that uses spec param
			scenario.Steps = append(scenario.Steps, &gauge.Step{
				LineText: "Step with <specParam>",
				Args: []*gauge.StepArg{
					{Name: "specParam", Value: "specParam", ArgType: gauge.Dynamic},
				},
			})

			scenarios := []*gauge.Scenario{scenario}
			errMap := gauge.NewBuildErrors()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = copyScenarios(scenarios, *specTable, 0, errMap)
			}
		})
	}
}

// BenchmarkScenarioIterationEager benchmarks iterating through eager scenarios
func BenchmarkScenarioIterationEager(b *testing.B) {
	os.Setenv("scenario_init_strategy", "eager")
	defer os.Unsetenv("scenario_init_strategy")

	scenario := createTestScenarioWithTable(100, 2)
	scenarios := []*gauge.Scenario{scenario}
	errMap := gauge.NewBuildErrors()

	scns, _ := copyScenarios(scenarios, gauge.Table{}, 0, errMap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range scns {
			_ = s.Heading.Value // Access scenario
		}
	}
}

// BenchmarkScenarioIterationLazy benchmarks iterating through lazy scenarios
func BenchmarkScenarioIterationLazy(b *testing.B) {
	os.Setenv("scenario_init_strategy", "lazy")
	defer os.Unsetenv("scenario_init_strategy")

	scenario := createTestScenarioWithTable(100, 2)
	scenarios := []*gauge.Scenario{scenario}
	errMap := gauge.NewBuildErrors()

	_, lazyColls := copyScenarios(scenarios, gauge.Table{}, 0, errMap)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, lazyColl := range lazyColls {
			iterator := lazyColl.Iterator()
			for s, hasNext := iterator.Next(); hasNext; s, hasNext = iterator.Next() {
				_ = s.Heading.Value // Access scenario
			}
		}
	}
}

// BenchmarkMemoryEager measures memory allocation for eager mode
func BenchmarkMemoryEager(b *testing.B) {
	os.Setenv("scenario_init_strategy", "eager")
	defer os.Unsetenv("scenario_init_strategy")

	scenario := createTestScenarioWithTable(1000, 2)
	scenarios := []*gauge.Scenario{scenario}
	errMap := gauge.NewBuildErrors()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = copyScenarios(scenarios, gauge.Table{}, 0, errMap)
	}
}

// BenchmarkMemoryLazy measures memory allocation for lazy mode
func BenchmarkMemoryLazy(b *testing.B) {
	os.Setenv("scenario_init_strategy", "lazy")
	defer os.Unsetenv("scenario_init_strategy")

	scenario := createTestScenarioWithTable(1000, 2)
	scenarios := []*gauge.Scenario{scenario}
	errMap := gauge.NewBuildErrors()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = copyScenarios(scenarios, gauge.Table{}, 0, errMap)
	}
}
