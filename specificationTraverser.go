package main

type specTraverser interface {
	specHeading(*heading)
	specTags(*tags)
	dataTable(*table)
	contextStep(*step)
	scenario(*scenario)
	scenarioHeading(*heading)
	scenarioTags(*tags)
	step(*step)
	comment(*comment)
}

type scenarioTraverser interface {
	scenarioHeading(*heading)
	scenarioTags(*tags)
	step(*step)
	comment(*comment)
}

func (spec *specification) traverse(traverser specTraverser) {
	traverser.specHeading(spec.heading)
	traverser.specTags(spec.tags)
	for _, item := range spec.items {
		switch item.kind() {
		case scenarioKind:
			traverser.scenario(item.(*scenario))
			item.(*scenario).traverse(traverser)
		case stepKind:
			traverser.contextStep(item.(*step))
		case commentKind:
			traverser.comment(item.(*comment))
		case tableKind:
			traverser.dataTable(item.(*table))
		}
	}
}

func (scenario *scenario) traverse(traverser scenarioTraverser) {
	traverser.scenarioHeading(scenario.heading)
	traverser.scenarioTags(scenario.tags)
	for _, item := range scenario.items {
		switch item.kind() {
		case stepKind:
			traverser.step(item.(*step))
		case commentKind:
			traverser.comment(item.(*comment))
		}
	}
}
