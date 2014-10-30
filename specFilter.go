package main

type scenarioIndexFilter struct {
	indexToFilter    int
	currentScenarioIndex int
}

func newScenarioIndexFilter(index int) *scenarioIndexFilter {
	return &scenarioIndexFilter{index, 0}
}

func (filter *scenarioIndexFilter) filter(item item) bool {
	if item.kind() == scenarioKind {
		if filter.currentScenarioIndex == filter.indexToFilter {
			return true
		}
		filter.currentScenarioIndex ++
	}
	return false
}
