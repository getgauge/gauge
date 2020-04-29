/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package filter

import (
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
)

type specsFilter interface {
	filter([]*gauge.Specification) []*gauge.Specification
}

type tagsFilter struct {
	tagExp string
}

type tagFilterForParallelRun struct {
	tagExp string
}

type specsGroupFilter struct {
	group       int
	execStreams int
}

type scenariosFilter struct {
	scenarios []string
}

func (tf *tagFilterForParallelRun) filter(specs []*gauge.Specification) ([]*gauge.Specification, []*gauge.Specification) {
	return filterByTags(tf.tagExp, specs)
}

func (tagsFilter *tagsFilter) filter(specs []*gauge.Specification) []*gauge.Specification {
	specs, _ = filterByTags(tagsFilter.tagExp, specs)
	return specs
}

func filterByTags(tagExpression string, specs []*gauge.Specification) ([]*gauge.Specification, []*gauge.Specification) {
	if tagExpression != "" {
		logger.Debugf(true, "Applying tags filter: %s", tagExpression)
		validateTagExpression(tagExpression)
		return filterSpecsByTags(specs, tagExpression)
	}
	return specs, specs
}

func (groupFilter *specsGroupFilter) filter(specs []*gauge.Specification) []*gauge.Specification {
	if groupFilter.group == -1 {
		return specs
	}
	logger.Infof(true, "Using the -g flag will make the distribution strategy 'eager'. The --strategy setting will be overridden.")
	if groupFilter.group < 1 || groupFilter.group > groupFilter.execStreams {
		return make([]*gauge.Specification, 0)
	}
	logger.Debugf(true, "Applying group filter: %d", groupFilter.group)
	group := DistributeSpecs(specs, groupFilter.execStreams)[groupFilter.group-1]
	if group == nil {
		return make([]*gauge.Specification, 0)
	}
	return group.Specs()
}

func (scenarioFilter *scenariosFilter) filter(specs []*gauge.Specification) []*gauge.Specification {
	if len(scenarioFilter.scenarios) != 0 {
		logger.Debugf(true, "Applying scenarios filter: %s", strings.Join(scenarioFilter.scenarios, ", "))
		specs = filterSpecsByScenarioName(specs, scenarioFilter.scenarios)
	}
	return specs
}

func DistributeSpecs(specifications []*gauge.Specification, distributions int) []*gauge.SpecCollection {
	s := make([]*gauge.SpecCollection, distributions)
	for i := 0; i < len(specifications); i++ {
		mod := i % distributions
		if s[mod] == nil {
			s[mod] = gauge.NewSpecCollection(make([]*gauge.Specification, 0), false)
		}
		s[mod].Add(specifications[i])
	}
	return s
}
