// Copyright 2014 ThoughtWorks, Inc.

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
