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

package main

type specsFilter interface {
	filter([]*specification) []*specification
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

func (tagsFilter *tagsFilter) filter(specs []*specification) []*specification {
	if tagsFilter.tagExp != "" {
		validateTagExpression(tagsFilter.tagExp)
		specs = filterSpecsByTags(specs, tagsFilter.tagExp)
	}
	return specs
}

func (groupFilter *specsGroupFilter) filter(specs []*specification) []*specification {
	if groupFilter.group == -1 {
		return specs
	}

	if groupFilter.group < 1 || groupFilter.group > groupFilter.execStreams {
		return make([]*specification, 0)
	}
	execution := &parallelSpecExecution{specifications: specs}
	return execution.distributeSpecs(groupFilter.execStreams)[groupFilter.group-1].specs
}

func (randomizer *specRandomizer) filter(specs []*specification) []*specification {
	if !randomizer.dontRandomize {
		return shuffleSpecs(specs)
	}
	return specs
}
