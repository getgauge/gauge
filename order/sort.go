/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package order

import (
	"math/rand"
	"sort"
	"time"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
)

var SortOrder string
var RandomSeed int64

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
	switch SortOrder {
	case "alpha":
		sort.Sort(byFileName(specs))
	case "random":
		// RandomSeed should already be set by the execute() function
		// This ensures the seed is saved for --failed and --repeat
		if RandomSeed == 0 {
			// Fallback: generate seed if not already set (shouldn't happen in normal flow)
			RandomSeed = time.Now().UnixNano()
		}
		logger.Infof(true, "Randomizing execution with seed: %d", RandomSeed)

		// Create a new random source with the seed
		r := rand.New(rand.NewSource(RandomSeed))
		r.Shuffle(len(specs), func(i, j int) {
			specs[i], specs[j] = specs[j], specs[i]
		})
	}
	return specs
}
