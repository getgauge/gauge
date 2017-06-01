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

	Sorted = true
	got := Sort(specs)
	expected := []*gauge.Specification{spec1, spec2, spec3}

	for i, s := range got {
		if expected[i].FileName != s.FileName {
			t.Errorf("Expected '%s' at position %d, got %s", expected[i].FileName, i, s.FileName)
		}
	}
}
