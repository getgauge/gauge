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

package parser

type SpecTraverser interface {
	SpecHeading(*Heading)
	SpecTags(*Tags)
	DataTable(*Table)
	ExternalDataTable(*DataTable)
	ContextStep(*Step)
	Scenario(*Scenario)
	ScenarioHeading(*Heading)
	ScenarioTags(*Tags)
	Step(*Step)
	TearDown(*TearDown)
	Comment(*Comment)
}

type ScenarioTraverser interface {
	ScenarioHeading(*Heading)
	ScenarioTags(*Tags)
	Step(*Step)
	Comment(*Comment)
}

func (spec *Specification) Traverse(traverser SpecTraverser) {
	traverser.SpecHeading(spec.Heading)
	for _, item := range spec.Items {
		switch item.Kind() {
		case ScenarioKind:
			item.(*Scenario).Traverse(traverser)
			traverser.Scenario(item.(*Scenario))
		case StepKind:
			traverser.ContextStep(item.(*Step))
		case CommentKind:
			traverser.Comment(item.(*Comment))
		case TableKind:
			traverser.DataTable(item.(*Table))
		case TagKind:
			traverser.SpecTags(item.(*Tags))
		case TearDownKind:
			traverser.TearDown(item.(*TearDown))
		case DataTableKind:
			if !item.(*DataTable).IsExternal {
				traverser.DataTable(&item.(*DataTable).Table)
			} else {
				traverser.ExternalDataTable(item.(*DataTable))
			}
		}
	}
}

func (scenario *Scenario) Traverse(traverser ScenarioTraverser) {
	traverser.ScenarioHeading(scenario.Heading)
	for _, item := range scenario.Items {
		switch item.Kind() {
		case StepKind:
			traverser.Step(item.(*Step))
		case CommentKind:
			traverser.Comment(item.(*Comment))
		case TagKind:
			traverser.ScenarioTags(item.(*Tags))
		}
	}
}
