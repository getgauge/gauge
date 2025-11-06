/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package order

import (
	"testing"

	"github.com/getgauge/gauge/gauge"
)

func TestToSortSpecs(t *testing.T) {
	spec1 := &gauge.Specification{FileName: "ab"}
	spec2 := &gauge.Specification{FileName: "b"}
	spec3 := &gauge.Specification{FileName: "c"}
	var specs []*gauge.Specification
	specs = append(specs, spec3)
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	SortOrder = "alpha"
	got := Sort(specs)
	expected := []*gauge.Specification{spec1, spec2, spec3}

	for i, s := range got {
		if expected[i].FileName != s.FileName {
			t.Errorf("Expected '%s' at position %d, got %s", expected[i].FileName, i, s.FileName)
		}
	}
}

func TestNoSortWhenSortOrderIsEmpty(t *testing.T) {
	spec1 := &gauge.Specification{FileName: "c"}
	spec2 := &gauge.Specification{FileName: "a"}
	spec3 := &gauge.Specification{FileName: "b"}
	specs := []*gauge.Specification{spec1, spec2, spec3}

	SortOrder = ""
	got := Sort(specs)

	// Should maintain original order
	if got[0].FileName != "c" || got[1].FileName != "a" || got[2].FileName != "b" {
		t.Errorf("Expected original order to be preserved, got %v", got)
	}
}

func TestRandomSortWithSameSeedProducesSameOrder(t *testing.T) {
	// Create specs
	specs1 := createTestSpecs(10)
	specs2 := createTestSpecs(10)

	// Sort both with same seed
	SortOrder = "random"
	RandomSeed = 12345
	result1 := Sort(specs1)

	RandomSeed = 12345
	result2 := Sort(specs2)

	// Verify same order
	for i := range result1 {
		if result1[i].FileName != result2[i].FileName {
			t.Errorf("Same seed should produce same order. Position %d: expected '%s', got '%s'",
				i, result1[i].FileName, result2[i].FileName)
		}
	}
}

func TestRandomSortWithDifferentSeedsProducesDifferentOrders(t *testing.T) {
	// Create specs
	specs1 := createTestSpecs(10)
	specs2 := createTestSpecs(10)

	// Sort with different seeds
	SortOrder = "random"
	RandomSeed = 111
	result1 := Sort(specs1)

	RandomSeed = 222
	result2 := Sort(specs2)

	// Verify different order (at least one difference)
	foundDifference := false
	for i := range result1 {
		if result1[i].FileName != result2[i].FileName {
			foundDifference = true
			break
		}
	}

	if !foundDifference {
		t.Errorf("Different seeds should produce different orders")
	}
}

func TestRandomSortActuallyRandomizes(t *testing.T) {
	specs := createTestSpecs(10)
	original := make([]*gauge.Specification, len(specs))
	copy(original, specs)

	// Randomize
	SortOrder = "random"
	RandomSeed = 999
	result := Sort(specs)

	// Check that order changed (statistically very unlikely to be same with 10 items)
	same := true
	for i := range result {
		if result[i].FileName != original[i].FileName {
			same = false
			break
		}
	}

	if same {
		t.Errorf("Random sort should change the order (may fail rarely due to randomness)")
	}
}

func TestRandomSortHandlesEmptySlice(t *testing.T) {
	specs := []*gauge.Specification{}

	SortOrder = "random"
	RandomSeed = 123
	result := Sort(specs)

	if len(result) != 0 {
		t.Errorf("Empty slice should remain empty after random sort")
	}
}

func TestRandomSortHandlesSingleSpec(t *testing.T) {
	spec := &gauge.Specification{FileName: "single.spec"}
	specs := []*gauge.Specification{spec}

	SortOrder = "random"
	RandomSeed = 456
	result := Sort(specs)

	if len(result) != 1 || result[0].FileName != "single.spec" {
		t.Errorf("Single spec should remain unchanged after random sort")
	}
}

func TestAutoGenerateSeedWhenZero(t *testing.T) {
	specs := createTestSpecs(5)

	SortOrder = "random"
	RandomSeed = 0 // Not set
	Sort(specs)

	// After sorting, RandomSeed should be set to non-zero
	if RandomSeed == 0 {
		t.Errorf("RandomSeed should be auto-generated when set to 0")
	}
}

// Helper function to create test specs
func createTestSpecs(count int) []*gauge.Specification {
	specs := make([]*gauge.Specification, count)
	for i := 0; i < count; i++ {
		specs[i] = &gauge.Specification{FileName: string(rune('a' + i)) + ".spec"}
	}
	return specs
}

// Helper function to create test specs with scenarios
func createTestSpecsWithScenarios(specCount, scenarioCount int) []*gauge.Specification {
	specs := make([]*gauge.Specification, specCount)
	for i := 0; i < specCount; i++ {
		scenarios := make([]*gauge.Scenario, scenarioCount)
		for j := 0; j < scenarioCount; j++ {
			scenarios[j] = &gauge.Scenario{
				Heading: &gauge.Heading{Value: string(rune('A'+i)) + string(rune('1'+j))},
			}
		}
		specs[i] = &gauge.Specification{
			FileName:  string(rune('a'+i)) + ".spec",
			Scenarios: scenarios,
		}
	}
	return specs
}

func TestRandomSortShufflesScenariosWithinSpec(t *testing.T) {
	// Create 2 specs with 5 scenarios each
	specs := createTestSpecsWithScenarios(2, 5)

	// Record original scenario order for first spec
	originalScenarioOrder := make([]string, len(specs[0].Scenarios))
	for i, scenario := range specs[0].Scenarios {
		originalScenarioOrder[i] = scenario.Heading.Value
	}

	// Randomize
	SortOrder = "random"
	RandomSeed = 888
	result := Sort(specs)

	// Check that scenarios were shuffled within at least one spec
	scenariosChanged := false
	for _, spec := range result {
		if len(spec.Scenarios) > 1 {
			// Check if order changed
			for i, scenario := range spec.Scenarios {
				// Compare with original order (for first spec)
				if spec == specs[0] && i < len(originalScenarioOrder) {
					if scenario.Heading.Value != originalScenarioOrder[i] {
						scenariosChanged = true
						break
					}
				}
			}
		}
	}

	if !scenariosChanged {
		t.Errorf("Expected scenarios to be shuffled within specs")
	}
}

func TestRandomSortScenariosSameSeedProducesSameOrder(t *testing.T) {
	// Create two identical sets of specs with scenarios
	specs1 := createTestSpecsWithScenarios(3, 4)
	specs2 := createTestSpecsWithScenarios(3, 4)

	// Randomize both with same seed
	SortOrder = "random"
	RandomSeed = 54321
	result1 := Sort(specs1)

	RandomSeed = 54321
	result2 := Sort(specs2)

	// Verify specs and scenarios have same order
	for i := range result1 {
		// Check spec order
		if result1[i].FileName != result2[i].FileName {
			t.Errorf("Spec order mismatch at position %d: %s vs %s",
				i, result1[i].FileName, result2[i].FileName)
		}

		// Check scenario order within each spec
		if len(result1[i].Scenarios) != len(result2[i].Scenarios) {
			t.Errorf("Scenario count mismatch in spec %d", i)
			continue
		}

		for j := range result1[i].Scenarios {
			if result1[i].Scenarios[j].Heading.Value != result2[i].Scenarios[j].Heading.Value {
				t.Errorf("Scenario order mismatch in spec %d, position %d: %s vs %s",
					i, j,
					result1[i].Scenarios[j].Heading.Value,
					result2[i].Scenarios[j].Heading.Value)
			}
		}
	}
}

func TestRandomSortScenariosDifferentSeedsProduceDifferentOrder(t *testing.T) {
	// Create two identical sets of specs with scenarios
	specs1 := createTestSpecsWithScenarios(3, 5)
	specs2 := createTestSpecsWithScenarios(3, 5)

	// Randomize with different seeds
	SortOrder = "random"
	RandomSeed = 11111
	result1 := Sort(specs1)

	RandomSeed = 22222
	result2 := Sort(specs2)

	// Verify at least one scenario order is different
	foundDifference := false
	for i := range result1 {
		for j := range result1[i].Scenarios {
			if j < len(result2[i].Scenarios) {
				if result1[i].Scenarios[j].Heading.Value != result2[i].Scenarios[j].Heading.Value {
					foundDifference = true
					break
				}
			}
		}
		if foundDifference {
			break
		}
	}

	if !foundDifference {
		t.Errorf("Expected different seeds to produce different scenario orders")
	}
}

func TestRandomSortHandlesSpecsWithNoScenarios(t *testing.T) {
	// Create specs with no scenarios
	specs := []*gauge.Specification{
		{FileName: "a.spec", Scenarios: []*gauge.Scenario{}},
		{FileName: "b.spec", Scenarios: []*gauge.Scenario{}},
	}

	SortOrder = "random"
	RandomSeed = 333
	result := Sort(specs)

	// Should not crash, just shuffle specs
	if len(result) != 2 {
		t.Errorf("Expected 2 specs, got %d", len(result))
	}
}

func TestRandomSortHandlesSpecsWithSingleScenario(t *testing.T) {
	// Create specs with single scenario each
	specs := createTestSpecsWithScenarios(3, 1)

	SortOrder = "random"
	RandomSeed = 444
	result := Sort(specs)

	// Should not crash, scenarios should remain unchanged
	for _, spec := range result {
		if len(spec.Scenarios) != 1 {
			t.Errorf("Expected 1 scenario per spec, got %d", len(spec.Scenarios))
		}
	}
}
