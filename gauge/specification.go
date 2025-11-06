/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import (
	"reflect"
)

type HeadingType int

const (
	SpecHeading     = 0
	ScenarioHeading = 1
)

type TokenKind int

const (
	SpecKind TokenKind = iota
	TagKind
	ScenarioKind
	CommentKind
	StepKind
	TableHeader
	TableRow
	HeadingKind
	TableKind
	DataTableKind
	TearDownKind
)

type Specification struct {
	Heading       *Heading
	Scenarios     []*Scenario
	Comments      []*Comment
	DataTable     DataTable
	Contexts      []*Step
	FileName      string
	Tags          *Tags
	Items         []Item
	TearDownSteps []*Step
	LazyScenarios []*LazyScenarioCollection
}

type Item interface {
	Kind() TokenKind
}

func (spec *Specification) Kind() TokenKind {
	return SpecKind
}

// Steps gives all the steps present in Specification
func (spec *Specification) Steps() []*Step {
	steps := spec.Contexts
	for _, scen := range spec.Scenarios {
		steps = append(steps, scen.Steps...)
	}
	return append(steps, spec.TearDownSteps...)
}

func (spec *Specification) ProcessConceptStepsFrom(conceptDictionary *ConceptDictionary) error {
	for _, step := range spec.Contexts {
		if err := spec.processConceptStep(step, conceptDictionary); err != nil {
			return err
		}
	}
	for _, scenario := range spec.Scenarios {
		for _, step := range scenario.Steps {
			if err := spec.processConceptStep(step, conceptDictionary); err != nil {
				return err
			}
		}
	}
	for _, step := range spec.TearDownSteps {
		if err := spec.processConceptStep(step, conceptDictionary); err != nil {
			return err
		}
	}
	return nil
}

func (spec *Specification) processConceptStep(step *Step, conceptDictionary *ConceptDictionary) error {
	if conceptFromDictionary := conceptDictionary.Search(step.Value); conceptFromDictionary != nil {
		return spec.createConceptStep(conceptFromDictionary.ConceptStep, step)
	}
	return nil
}

func (spec *Specification) createConceptStep(concept *Step, originalStep *Step) error {
	stepCopy, err := concept.GetCopy()
	if err != nil {
		return err
	}
	originalArgs := originalStep.Args
	originalStep.CopyFrom(stepCopy)
	originalStep.Args = originalArgs

	// set parent of all concept steps to be the current concept (referred as originalStep here)
	// this is used to fetch from parent's lookup when nested
	for _, conceptStep := range originalStep.ConceptSteps {
		conceptStep.Parent = originalStep
	}

	return spec.PopulateConceptLookup(&originalStep.Lookup, concept.Args, originalStep.Args)
}

func (spec *Specification) AddItem(itemToAdd Item) {
	if spec.Items == nil {
		spec.Items = make([]Item, 0)
	}

	spec.Items = append(spec.Items, itemToAdd)
}

func (spec *Specification) AddHeading(heading *Heading) {
	heading.HeadingType = SpecHeading
	spec.Heading = heading
}

func (spec *Specification) AddScenario(scenario *Scenario) {
	spec.Scenarios = append(spec.Scenarios, scenario)
	spec.AddItem(scenario)
}

func (spec *Specification) AddContext(contextStep *Step) {
	spec.Contexts = append(spec.Contexts, contextStep)
	spec.AddItem(contextStep)
}

func (spec *Specification) AddComment(comment *Comment) {
	spec.Comments = append(spec.Comments, comment)
	spec.AddItem(comment)
}

func (spec *Specification) AddDataTable(table *Table) {
	spec.DataTable.Table = table
	spec.AddItem(&spec.DataTable)
}

func (spec *Specification) AddExternalDataTable(externalTable *DataTable) {
	spec.DataTable = *externalTable
	spec.AddItem(externalTable)
}

func (spec *Specification) AddTags(tags *Tags) {
	spec.Tags = tags
	spec.AddItem(spec.Tags)
}

func (spec *Specification) NTags() int {
	if spec.Tags == nil {
		return 0
	}
	return len(spec.Tags.Values())
}

func (spec *Specification) LatestScenario() *Scenario {
	return spec.Scenarios[len(spec.Scenarios)-1]
}

func (spec *Specification) LatestContext() *Step {
	return spec.Contexts[len(spec.Contexts)-1]
}

func (spec *Specification) LatestTeardown() *Step {
	return spec.TearDownSteps[len(spec.TearDownSteps)-1]
}

func (spec *Specification) removeItem(itemIndex int) {
	item := spec.Items[itemIndex]
	items := make([]Item, len(spec.Items))
	copy(items, spec.Items)
	if len(spec.Items)-1 == itemIndex {
		spec.Items = items[:itemIndex]
	} else if itemIndex == 0 {
		spec.Items = items[itemIndex+1:]
	} else {
		spec.Items = append(items[:itemIndex], items[itemIndex+1:]...)
	}
	if item.Kind() == ScenarioKind {
		spec.removeScenario(item.(*Scenario))
	}
}

func (spec *Specification) removeScenario(scenario *Scenario) {
	index := getIndexFor(scenario, spec.Scenarios)
	scenarios := make([]*Scenario, len(spec.Scenarios))
	copy(scenarios, spec.Scenarios)
	if len(spec.Scenarios)-1 == index {
		spec.Scenarios = scenarios[:index]
	} else if index == 0 {
		spec.Scenarios = scenarios[index+1:]
	} else {
		spec.Scenarios = append(scenarios[:index], scenarios[index+1:]...)
	}
}

func (spec *Specification) PopulateConceptLookup(lookup *ArgLookup, conceptArgs []*StepArg, stepArgs []*StepArg) error {
	for i, arg := range stepArgs {
		stepArg := StepArg{Value: arg.Value, ArgType: arg.ArgType, Table: arg.Table, Name: arg.Name}
		if err := lookup.AddArgValue(conceptArgs[i].Value, &stepArg); err != nil {
			return err
		}
	}
	return nil
}

func (spec *Specification) RenameSteps(oldStep *Step, newStep *Step, orderMap map[int]int) ([]*StepDiff, bool) {
	diffs, isRefactored := spec.rename(spec.Contexts, oldStep, newStep, false, orderMap)
	for _, scenario := range spec.Scenarios {
		scenStepDiffs, refactor := scenario.renameSteps(oldStep, newStep, orderMap)
		diffs = append(diffs, scenStepDiffs...)
		if refactor {
			isRefactored = refactor
		}
	}
	teardownStepdiffs, isRefactored := spec.rename(spec.TearDownSteps, oldStep, newStep, isRefactored, orderMap)
	return append(diffs, teardownStepdiffs...), isRefactored
}

func (spec *Specification) rename(steps []*Step, oldStep *Step, newStep *Step, isRefactored bool, orderMap map[int]int) ([]*StepDiff, bool) {
	diffs := []*StepDiff{}
	isConcept := false
	for _, step := range steps {
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

func (spec *Specification) GetSpecItems() []Item {
	specItems := make([]Item, 0)
	for _, item := range spec.Items {
		if item.Kind() != ScenarioKind {
			specItems = append(specItems, item)
		}
	}
	return specItems
}

func (spec *Specification) Traverse(processor ItemProcessor, queue *ItemQueue) {
	processor.Specification(spec)
	processor.Heading(spec.Heading)

	for queue.Peek() != nil {
		item := queue.Next()
		switch item.Kind() {
		case ScenarioKind:
			processor.Heading(item.(*Scenario).Heading)
			processor.Scenario(item.(*Scenario))
		case StepKind:
			processor.Step(item.(*Step))
		case CommentKind:
			processor.Comment(item.(*Comment))
		case TableKind:
			processor.Table(item.(*Table))
		case TagKind:
			processor.Tags(item.(*Tags))
		case TearDownKind:
			processor.TearDown(item.(*TearDown))
		case DataTableKind:
			processor.DataTable(item.(*DataTable))
		}
	}
}

func (spec *Specification) AllItems() (items []Item) {
	for _, item := range spec.Items {
		items = append(items, item)
		if item.Kind() == ScenarioKind {
			items = append(items, item.(*Scenario).Items...)
		}
	}
	return
}

func (spec *Specification) UsesArgsInContextTeardown(args ...string) bool {
	return UsesArgs(append(spec.Contexts, spec.TearDownSteps...), args...)
}

type SpecItemFilter interface {
	Filter(Item) bool
}

func (spec *Specification) Filter(filter SpecItemFilter) (*Specification, *Specification) {
	specWithFilteredItems := new(Specification)
	specWithOtherItems := new(Specification)
	*specWithFilteredItems, *specWithOtherItems = *spec, *spec
	for i := 0; i < len(specWithFilteredItems.Items); i++ {
		item := specWithFilteredItems.Items[i]
		if item.Kind() == ScenarioKind && filter.Filter(item) {
			specWithFilteredItems.removeItem(i)
			i--
		}
	}
	for i := 0; i < len(specWithOtherItems.Items); i++ {
		item := specWithOtherItems.Items[i]
		if item.Kind() == ScenarioKind && !filter.Filter(item) {
			specWithOtherItems.removeItem(i)
			i--
		}
	}
	return specWithFilteredItems, specWithOtherItems
}

func getIndexFor(scenario *Scenario, scenarios []*Scenario) int {
	for index, anItem := range scenarios {
		if reflect.DeepEqual(scenario, anItem) {
			return index
		}
	}
	return -1
}

type Heading struct {
	Value       string
	LineNo      int
	SpanEnd     int
	HeadingType HeadingType
}

func (heading *Heading) Kind() TokenKind {
	return HeadingKind
}

type Comment struct {
	Value  string
	LineNo int
}

func (comment *Comment) Kind() TokenKind {
	return CommentKind
}

type TearDown struct {
	LineNo int
	Value  string
}

func (t *TearDown) Kind() TokenKind {
	return TearDownKind
}

type Tags struct {
	RawValues [][]string
}

func (tags *Tags) Add(values []string) {
	tags.RawValues = append(tags.RawValues, values)
}

func (tags *Tags) Values() (val []string) {
	for i := range tags.RawValues {
		val = append(val, tags.RawValues[i]...)
	}
	return val
}
func (tags *Tags) Kind() TokenKind {
	return TagKind
}
