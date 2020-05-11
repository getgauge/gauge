/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
