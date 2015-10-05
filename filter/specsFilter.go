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

package filter

import (
	"github.com/getgauge/gauge/parser"
	"math/rand"
	"sort"
	"time"
)

type specsFilter interface {
	filter([]*parser.Specification) []*parser.Specification
}

type tagsFilter struct {
	tagExp string
}

type specsGroupFilter struct {
	group       int
	execStreams int
}

type specRandomizer struct {
	dontRandomize bool
}

func (tagsFilter *tagsFilter) filter(specs []*parser.Specification) []*parser.Specification {
	if tagsFilter.tagExp != "" {
		validateTagExpression(tagsFilter.tagExp)
		specs = filterSpecsByTags(specs, tagsFilter.tagExp)
	}
	return specs
}

func (groupFilter *specsGroupFilter) filter(specs []*parser.Specification) []*parser.Specification {
	if groupFilter.group == -1 {
		return specs
	}

	if groupFilter.group < 1 || groupFilter.group > groupFilter.execStreams {
		return make([]*parser.Specification, 0)
	}
	return DistributeSpecs(sortSpecsList(specs), groupFilter.execStreams)[groupFilter.group-1].Specs
}

func DistributeSpecs(specifications []*parser.Specification, distributions int) []*SpecCollection {
	specCollections := make([]*SpecCollection, distributions)
	for i := 0; i < len(specifications); i++ {
		mod := i % distributions
		if specCollections[mod] == nil {
			specCollections[mod] = &SpecCollection{Specs: make([]*parser.Specification, 0)}
		}
		specCollections[mod].Specs = append(specCollections[mod].Specs, specifications[i])
	}
	return specCollections
}

type SpecCollection struct {
	Specs []*parser.Specification
}

func (s *SpecCollection) SpecNames() []string {
	specNames := make([]string, 0)
	for _, spec := range s.Specs {
		specNames = append(specNames, spec.FileName)
	}
	return specNames
}

func (randomizer *specRandomizer) filter(specs []*parser.Specification) []*parser.Specification {
	if !randomizer.dontRandomize {
		return shuffleSpecs(specs)
	}
	return specs
}

func shuffleSpecs(allSpecs []*parser.Specification) []*parser.Specification {
	dest := make([]*parser.Specification, len(allSpecs))
	rand.Seed(int64(time.Now().Nanosecond()))
	perm := rand.Perm(len(allSpecs))
	for i, v := range perm {
		dest[v] = allSpecs[i]
	}
	return dest
}

type ByFileName []*parser.Specification

func (s ByFileName) Len() int {
	return len(s)
}

func (s ByFileName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByFileName) Less(i, j int) bool {
	return s[i].FileName < s[j].FileName
}

func sortSpecsList(allSpecs []*parser.Specification) []*parser.Specification {
	sort.Sort(ByFileName(allSpecs))
	return allSpecs
}
