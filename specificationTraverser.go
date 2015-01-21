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
	for _, item := range spec.items {
		switch item.kind() {
		case scenarioKind:
			item.(*scenario).traverse(traverser)
			traverser.scenario(item.(*scenario))
		case stepKind:
			traverser.contextStep(item.(*step))
		case commentKind:
			traverser.comment(item.(*comment))
		case tableKind:
			traverser.dataTable(item.(*table))
		case tagKind:
			traverser.specTags(item.(*tags))
		}
	}
}

func (scenario *scenario) traverse(traverser scenarioTraverser) {
	traverser.scenarioHeading(scenario.heading)
	for _, item := range scenario.items {
		switch item.kind() {
		case stepKind:
			traverser.step(item.(*step))
		case commentKind:
			traverser.comment(item.(*comment))
		case tagKind:
			traverser.scenarioTags(item.(*tags))
		}
	}
}
