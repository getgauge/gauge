// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.
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
