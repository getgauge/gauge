package main

type scenarioIndexFilterToRetain struct {
	indexToNotFilter     int
	currentScenarioIndex int
}
type ScenarioFilterBasedOnTags struct {
	tagsToNotFilter []string
	specTags        []string
}

func newScenarioIndexFilterToRetain(index int) *scenarioIndexFilterToRetain {
	return &scenarioIndexFilterToRetain{index, 0}
}

func (filter *scenarioIndexFilterToRetain) filter(item item) bool {
	if item.kind() == scenarioKind {
		if filter.currentScenarioIndex != filter.indexToNotFilter {
			filter.currentScenarioIndex++
			return true
		} else {
			filter.currentScenarioIndex++
			return false
		}
	}
	return false
}

func newScenarioFilterBasedOnTags(tagList []string, specTags []string) *ScenarioFilterBasedOnTags {
	return &ScenarioFilterBasedOnTags{tagList, specTags}
}

func (filter *ScenarioFilterBasedOnTags) filter(item item) bool {
	if item.kind() == scenarioKind {
		if isPresent(filter.specTags, filter.tagsToNotFilter) {
			return false
		}
		if filter.areValidTags(item.(*scenario).tags) {
			return false
		}
		return true
	}
	return false
}

func (filter *ScenarioFilterBasedOnTags) areValidTags(tagsOfCurrentScenario *tags) bool {
	if tagsOfCurrentScenario == nil {
		return false
	}
	tagsOfScenario := tagsOfCurrentScenario.values
	for _, tagToExecute := range filter.tagsToNotFilter {
		if arrayContains(tagsOfScenario, tagToExecute) == false {
			return false
		}
	}
	return true
}

func isPresent(specTags []string, tagsToNotFilter []string) bool {
	for _, tagInFilter := range tagsToNotFilter {
		if arrayContains(specTags, tagInFilter) == false {
			return false
		}
	}
	return true
}
