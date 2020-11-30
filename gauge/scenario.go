/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import (
	"strings"
)

type Scenario struct {
	Heading                   *Heading
	Steps                     []*Step
	Comments                  []*Comment
	Tags                      *Tags
	Items                     []Item
	DataTable                 DataTable
	SpecDataTableRow          Table
	SpecDataTableRowIndex     int
	ScenarioDataTableRow      Table
	ScenarioDataTableRowIndex int
	Span                      *Span
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

func (scenario *Scenario) AddExternalDataTable(externalTable *DataTable) {
	scenario.DataTable = *externalTable
	scenario.AddItem(externalTable)
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

func (scenario *Scenario) AddDataTable(table *Table) {
	scenario.DataTable.Table = table
	scenario.AddItem(&scenario.DataTable)
}

func (scenario *Scenario) InSpan(lineNumber int) bool {
	return scenario.Span.isInRange(lineNumber)
}

func (scenario *Scenario) renameSteps(oldStep *Step, newStep *Step, orderMap map[int]int) ([]*StepDiff, bool) {
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

// skipcq CRT-P0003
func (scenario Scenario) Kind() TokenKind {
	return ScenarioKind
}

func (scn *Scenario) HasAnyHeading(headings []string) bool {
	for _, heading := range headings {
		if strings.Compare(scn.Heading.Value, heading) == 0 {
			return true
		}
	}
	return false
}
