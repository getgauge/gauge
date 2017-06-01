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
	"sort"

	"github.com/getgauge/gauge/gauge"
)

var Sorted bool

type byFileName []*gauge.Specification

func (s byFileName) Len() int {
	return len(s)
}

func (s byFileName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byFileName) Less(i, j int) bool {
	return s[i].FileName < s[j].FileName
}

func Sort(specs []*gauge.Specification) []*gauge.Specification {
	if Sorted {
		sort.Sort(byFileName(specs))
	}
	return specs
}
