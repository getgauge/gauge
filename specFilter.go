package main

type scenarioIndexFilterToRetain struct {
	indexToNotFilter     int
	currentScenarioIndex int
}
type ScenarioFilterBasedOnTags struct {
	tagsToNotFilter []string
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

func newScenarioFilterBasedOnTags(tagList []string) *ScenarioFilterBasedOnTags {
	return &ScenarioFilterBasedOnTags{tagList}
}

func (filter *ScenarioFilterBasedOnTags) filter(item item) bool {
	if item.kind() == scenarioKind {
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
