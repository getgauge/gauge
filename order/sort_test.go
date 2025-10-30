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
