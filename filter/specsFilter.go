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
	"math/rand"
	"sort"
	"time"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
)

type specsFilter interface {
	filter([]*gauge.Specification) []*gauge.Specification
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

func (tagsFilter *tagsFilter) filter(specs []*gauge.Specification) []*gauge.Specification {
	if tagsFilter.tagExp != "" {
		validateTagExpression(tagsFilter.tagExp)
		specs = filterSpecsByTags(specs, tagsFilter.tagExp)
	}
	return specs
}

func (groupFilter *specsGroupFilter) filter(specs []*gauge.Specification) []*gauge.Specification {
	if groupFilter.group == -1 {
		return specs
	}
	logger.Info("Using the -g flag will make the distribution strategy 'eager'. The --strategy setting will be overridden.")
	if groupFilter.group < 1 || groupFilter.group > groupFilter.execStreams {
		return make([]*gauge.Specification, 0)
	}
	group := DistributeSpecs(sortSpecsList(specs), groupFilter.execStreams)[groupFilter.group-1]
	if group == nil {
		return make([]*gauge.Specification, 0)
	}
	return group.Specs
}

func DistributeSpecs(specifications []*gauge.Specification, distributions int) []*SpecCollection {
	specCollections := make([]*SpecCollection, distributions)
	for i := 0; i < len(specifications); i++ {
		mod := i % distributions
		if specCollections[mod] == nil {
			specCollections[mod] = &SpecCollection{Specs: make([]*gauge.Specification, 0)}
		}
		specCollections[mod].Specs = append(specCollections[mod].Specs, specifications[i])
	}
	return specCollections
}

type SpecCollection struct {
	Specs []*gauge.Specification
}

func (s *SpecCollection) SpecNames() []string {
	specNames := make([]string, 0)
	for _, spec := range s.Specs {
		specNames = append(specNames, spec.FileName)
	}
	return specNames
}

func (randomizer *specRandomizer) filter(specs []*gauge.Specification) []*gauge.Specification {
	if !randomizer.dontRandomize {
		return shuffleSpecs(specs)
	}
	return specs
}

func shuffleSpecs(allSpecs []*gauge.Specification) []*gauge.Specification {
	dest := make([]*gauge.Specification, len(allSpecs))
	rand.Seed(int64(time.Now().Nanosecond()))
	perm := rand.Perm(len(allSpecs))
	for i, v := range perm {
		dest[v] = allSpecs[i]
	}
	return dest
}

type ByFileName []*gauge.Specification

func (s ByFileName) Len() int {
	return len(s)
}

func (s ByFileName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByFileName) Less(i, j int) bool {
	return s[i].FileName < s[j].FileName
}

func sortSpecsList(allSpecs []*gauge.Specification) []*gauge.Specification {
	sort.Sort(ByFileName(allSpecs))
	return allSpecs
}
