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

package gauge

type Scenario struct {
	Heading           *Heading
	Steps             []*Step
	Comments          []*Comment
	Tags              *Tags
	Items             []Item
	DataTableRow      Table
	DataTableRowIndex int
	Span              *Span
}

// Span represents scope of Scenario based on line number
type Span struct {
	Start int
	End   int
}

func (s *Span) isInRange(lineNumber int) bool {
	return s.Start <= lineNumber && s.End >= lineNumber
}

func (scenario *Scenario) AddHeading(heading *Heading) {
	heading.HeadingType = ScenarioHeading
	scenario.Heading = heading
}

func (scenario *Scenario) AddStep(step *Step) {
	scenario.Steps = append(scenario.Steps, step)
	scenario.AddItem(step)
}

func (scenario *Scenario) AddTags(tags *Tags) {
	scenario.Tags = tags
	scenario.AddItem(tags)
}

func (scenario *Scenario) NTags() int {
	if scenario.Tags == nil {
		return 0
	}
	return len(scenario.Tags.Values())
}

func (scenario *Scenario) AddComment(comment *Comment) {
	scenario.Comments = append(scenario.Comments, comment)
	scenario.AddItem(comment)
}

func (scenario *Scenario) InSpan(lineNumber int) bool {
	return scenario.Span.isInRange(lineNumber)
}

func (scenario *Scenario) renameSteps(oldStep Step, newStep Step, orderMap map[int]int) ([]*StepDiff, bool) {
	isRefactored := false
	diffs := []*StepDiff{}
	isConcept := false
	for _, step := range scenario.Steps {
		diff, refactor := step.Rename(oldStep, newStep, isRefactored, orderMap, &isConcept)
		if diff != nil {
			diffs = append(diffs, diff)
		}
		if refactor {
			isRefactored = refactor
		}
	}
	return diffs, isRefactored
}

func (scenario *Scenario) AddItem(itemToAdd Item) {
	if scenario.Items == nil {
		scenario.Items = make([]Item, 0)
	}
	scenario.Items = append(scenario.Items, itemToAdd)
}

func (scenario *Scenario) LatestStep() *Step {
	return scenario.Steps[len(scenario.Steps)-1]
}

func (scenario *Scenario) UsesArgsInSteps(args ...string) bool {
	return UsesArgs(scenario.Steps, args...)
}

func (scenario Scenario) Kind() TokenKind {
	return ScenarioKind
}
