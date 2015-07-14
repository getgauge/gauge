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
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/parser"
	"math/rand"
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
	return execution.DistributeSpecs(specs, groupFilter.execStreams)[groupFilter.group-1].Specs
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
