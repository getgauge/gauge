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

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
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

type Scenario struct {
	Heading  *Heading
	Steps    []*Step
	Comments []*Comment
	Tags     *Tags
	Items    []Item
}

type ArgType string

const (
	Static               ArgType = "static"
	Dynamic              ArgType = "dynamic"
	TableArg             ArgType = "table"
	SpecialString        ArgType = "special_string"
	SpecialTable         ArgType = "special_table"
	ParameterPlaceholder         = "{}"
)

type ArgLookup struct {
	//helps to access the index of an arg at O(1)
	ParamIndexMap map[string]int
	paramValue    []paramNameValue
}
type paramNameValue struct {
	name    string
	stepArg *StepArg
}

func (paramNameValue paramNameValue) String() string {
	return fmt.Sprintf("ParamName: %s, stepArg: %s", paramNameValue.name, paramNameValue.stepArg)
}

func (argLookup ArgLookup) String() string {
	return fmt.Sprintln(argLookup.paramValue)
}

type StepArg struct {
	Name    string
	Value   string
	ArgType ArgType
	Table   Table
}

func (stepArg *StepArg) String() string {
	return fmt.Sprintf("{Name: %s,value %s,argType %s,table %v}", stepArg.Name, stepArg.Value, string(stepArg.ArgType), stepArg.Table)
}

type Step struct {
	LineNo         int
	Value          string
	LineText       string
	Args           []*StepArg
	IsConcept      bool
	Lookup         ArgLookup
	ConceptSteps   []*Step
	Fragments      []*gauge_messages.Fragment
	Parent         *Step
	HasInlineTable bool
	Items          []Item
	PreComments    []*Comment
}

type TearDown struct {
	LineNo int
	Value  string
}

type StepValue struct {
	Args                   []string
	StepValue              string
	ParameterizedStepValue string
}

func (step *Step) GetArg(name string) *StepArg {
	arg := step.Lookup.GetArg(name)
	// Return static values
	if arg != nil && arg.ArgType != Dynamic {
		return arg
	}
	if step.Parent == nil {
		return step.Lookup.GetArg(name)
	}
	return step.Parent.GetArg(step.Lookup.GetArg(name).Value)
}

func (step *Step) getLineText() string {
	if step.HasInlineTable {
		return fmt.Sprintf("%s <%s>", step.LineText, TableArg)
	}
	return step.LineText
}

func (step *Step) Rename(oldStep Step, newStep Step, isRefactored bool, orderMap map[int]int, isConcept *bool) bool {
	if strings.TrimSpace(step.Value) != strings.TrimSpace(oldStep.Value) {
		return isRefactored
	}
	if step.IsConcept {
		*isConcept = true
	}
	step.Value = newStep.Value

	step.Args = step.getArgsInOrder(newStep, orderMap)
	return true
}

func (step *Step) getArgsInOrder(newStep Step, orderMap map[int]int) []*StepArg {
	args := make([]*StepArg, len(newStep.Args))
	for key, value := range orderMap {
		arg := &StepArg{Value: newStep.Args[key].Value, ArgType: Static}
		if step.IsConcept {
			arg = &StepArg{Value: newStep.Args[key].Value, ArgType: Dynamic}
		}
		if value != -1 {
			arg = step.Args[value]
		}
		args[key] = arg
	}
	return args
}

func (step *Step) deepCopyStepArgs() []*StepArg {
	copiedStepArgs := make([]*StepArg, 0)
	for _, conceptStepArg := range step.Args {
		temp := new(StepArg)
		*temp = *conceptStepArg
		copiedStepArgs = append(copiedStepArgs, temp)
	}
	return copiedStepArgs
}

func (step *Step) ReplaceArgsWithDynamic(args []*StepArg) {
	for i, arg := range step.Args {
		for _, conceptArg := range args {
			if arg.String() == conceptArg.String() {
				if conceptArg.ArgType == SpecialString || conceptArg.ArgType == SpecialTable {
					reg := regexp.MustCompile(".*:")
					step.Args[i] = &StepArg{Value: reg.ReplaceAllString(conceptArg.Name, ""), ArgType: Dynamic}
					continue
				}
				step.Args[i] = &StepArg{Value: replaceParamChar(conceptArg.Value), ArgType: Dynamic}
			}
		}
	}
}

func replaceParamChar(text string) string {
	return strings.Replace(strings.Replace(text, "<", "{", -1), ">", "}", -1)
}

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
}

type Item interface {
	Kind() TokenKind
}

type Heading struct {
	Value       string
	LineNo      int
	HeadingType HeadingType
}

type Comment struct {
	Value  string
	LineNo int
}

type Tags struct {
	Values []string
}

type Warning struct {
	Message string
	LineNo  int
}

func (specification *Specification) ProcessConceptStepsFrom(conceptDictionary *ConceptDictionary) {
	for _, step := range specification.Contexts {
		specification.processConceptStep(step, conceptDictionary)
	}
	for _, scenario := range specification.Scenarios {
		for _, step := range scenario.Steps {
			specification.processConceptStep(step, conceptDictionary)
		}
	}
	for _, step := range specification.TearDownSteps {
		specification.processConceptStep(step, conceptDictionary)
	}
}

func (specification *Specification) processConceptStep(step *Step, conceptDictionary *ConceptDictionary) {
	if conceptFromDictionary := conceptDictionary.Search(step.Value); conceptFromDictionary != nil {
		specification.createConceptStep(conceptFromDictionary.ConceptStep, step)
	}
}

func (specification *Specification) createConceptStep(concept *Step, originalStep *Step) {
	stepCopy := concept.GetCopy()
	originalArgs := originalStep.Args
	originalStep.CopyFrom(stepCopy)
	originalStep.Args = originalArgs

	// set parent of all concept steps to be the current concept (referred as originalStep here)
	// this is used to fetch from parent's lookup when nested
	for _, conceptStep := range originalStep.ConceptSteps {
		conceptStep.Parent = originalStep
	}

	specification.PopulateConceptLookup(&originalStep.Lookup, concept.Args, originalStep.Args)
}

func (specification *Specification) AddItem(itemToAdd Item) {
	if specification.Items == nil {
		specification.Items = make([]Item, 0)
	}

	specification.Items = append(specification.Items, itemToAdd)
}

func (specification *Specification) AddHeading(heading *Heading) {
	heading.HeadingType = SpecHeading
	specification.Heading = heading
}

func (specification *Specification) AddScenario(scenario *Scenario) {
	specification.Scenarios = append(specification.Scenarios, scenario)
	specification.AddItem(scenario)
}

func (specification *Specification) AddContext(contextStep *Step) {
	specification.Contexts = append(specification.Contexts, contextStep)
	specification.AddItem(contextStep)
}

func (specification *Specification) AddComment(comment *Comment) {
	specification.Comments = append(specification.Comments, comment)
	specification.AddItem(comment)
}

func (specification *Specification) AddDataTable(table *Table) {
	specification.DataTable.Table = *table
	specification.AddItem(&specification.DataTable)
}

func (specification *Specification) AddExternalDataTable(externalTable *DataTable) {
	specification.DataTable = *externalTable
	specification.AddItem(externalTable)
}

func (specification *Specification) AddTags(tags *Tags) {
	specification.Tags = tags
	specification.AddItem(tags)
}

func (specification *Specification) LatestScenario() *Scenario {
	return specification.Scenarios[len(specification.Scenarios)-1]
}

func (specification *Specification) LatestContext() *Step {
	return specification.Contexts[len(specification.Contexts)-1]
}

func (specification *Specification) LatestTeardown() *Step {
	return specification.TearDownSteps[len(specification.TearDownSteps)-1]
}

func (step *Step) AddArgs(args ...*StepArg) {
	step.Args = append(step.Args, args...)
	step.PopulateFragments()
}

func (step *Step) AddInlineTableHeaders(headers []string) {
	tableArg := &StepArg{ArgType: TableArg}
	tableArg.Table.AddHeaders(headers)
	step.AddArgs(tableArg)
}

func (step *Step) AddInlineTableRow(row []TableCell) {
	lastArg := step.Args[len(step.Args)-1]
	lastArg.Table.addRows(row)
	step.PopulateFragments()
}

func (step *Step) PopulateFragments() {
	r := regexp.MustCompile(ParameterPlaceholder)
	/*
		enter {} and {} bar
		returns
		[[6 8] [13 15]]
	*/
	argSplitIndices := r.FindAllStringSubmatchIndex(step.Value, -1)
	step.Fragments = make([]*gauge_messages.Fragment, 0)
	if len(step.Args) == 0 {
		step.Fragments = append(step.Fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String(step.Value)})
		return
	}

	textStartIndex := 0
	for argIndex, argIndices := range argSplitIndices {
		if textStartIndex < argIndices[0] {
			step.Fragments = append(step.Fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String(step.Value[textStartIndex:argIndices[0]])})
		}
		parameter := convertToProtoParameter(step.Args[argIndex])
		step.Fragments = append(step.Fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter.Enum(), Parameter: parameter})
		textStartIndex = argIndices[1]
	}
	if textStartIndex < len(step.Value) {
		step.Fragments = append(step.Fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String(step.Value[textStartIndex:len(step.Value)])})
	}

}

type SpecItemFilter interface {
	Filter(Item) bool
}

func (spec *Specification) Filter(filter SpecItemFilter) {
	for i := 0; i < len(spec.Items); i++ {
		if filter.Filter(spec.Items[i]) {
			spec.removeItem(i)
			i--
		}
	}
}

func (spec *Specification) removeItem(itemIndex int) {
	item := spec.Items[itemIndex]
	if len(spec.Items)-1 == itemIndex {
		spec.Items = spec.Items[:itemIndex]
	} else if 0 == itemIndex {
		spec.Items = spec.Items[itemIndex+1:]
	} else {
		spec.Items = append(spec.Items[:itemIndex], spec.Items[itemIndex+1:]...)
	}
	if item.Kind() == ScenarioKind {
		spec.removeScenario(item.(*Scenario))
	}
}

func (spec *Specification) removeScenario(scenario *Scenario) {
	index := getIndexFor(scenario, spec.Scenarios)
	if len(spec.Scenarios)-1 == index {
		spec.Scenarios = spec.Scenarios[:index]
	} else if index == 0 {
		spec.Scenarios = spec.Scenarios[index+1:]
	} else {
		spec.Scenarios = append(spec.Scenarios[:index], spec.Scenarios[index+1:]...)
	}
}

func (spec *Specification) PopulateConceptLookup(lookup *ArgLookup, conceptArgs []*StepArg, stepArgs []*StepArg) {
	for i, arg := range stepArgs {
		lookup.AddArgValue(conceptArgs[i].Value, &StepArg{Value: arg.Value, ArgType: arg.ArgType, Table: arg.Table, Name: arg.Name})
	}
}

func (spec *Specification) RenameSteps(oldStep Step, newStep Step, orderMap map[int]int) bool {
	isRefactored := false
	for _, step := range spec.Contexts {
		isConcept := false
		isRefactored = step.Rename(oldStep, newStep, isRefactored, orderMap, &isConcept)
	}
	for _, scenario := range spec.Scenarios {
		refactor := scenario.renameSteps(oldStep, newStep, orderMap)
		if refactor {
			isRefactored = refactor
		}
	}
	return isRefactored
}

func (warning *Warning) String() string {
	return fmt.Sprintf("line no: %d, %s", warning.LineNo, warning.Message)
}

func (scenario Scenario) Kind() TokenKind {
	return ScenarioKind
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

func (scenario *Scenario) AddComment(comment *Comment) {
	scenario.Comments = append(scenario.Comments, comment)
	scenario.AddItem(comment)
}

func (scenario *Scenario) renameSteps(oldStep Step, newStep Step, orderMap map[int]int) bool {
	isRefactored := false
	for _, step := range scenario.Steps {
		isConcept := false
		isRefactored = step.Rename(oldStep, newStep, isRefactored, orderMap, &isConcept)
	}
	return isRefactored
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

func (heading *Heading) Kind() TokenKind {
	return HeadingKind
}

func (comment *Comment) Kind() TokenKind {
	return CommentKind
}

func (t *TearDown) Kind() TokenKind {
	return TearDownKind
}

func (tags *Tags) Kind() TokenKind {
	return TagKind
}

func (step Step) Kind() TokenKind {
	return StepKind
}

func (specification *Specification) GetSpecItems() []Item {
	specItems := make([]Item, 0)
	for _, item := range specification.Items {
		if item.Kind() != ScenarioKind {
			specItems = append(specItems, item)
		}
		if item.Kind() == TearDownKind {
			return specItems
		}
	}
	return specItems
}

// Not copying parent as it enters an infinite loop in case of nested concepts. This is because the steps under the concept
// are copied and their parent copying again comes back to copy the same concept.
func (self *Step) GetCopy() *Step {
	if !self.IsConcept {
		return self
	}
	nestedStepsCopy := make([]*Step, 0)
	for _, nestedStep := range self.ConceptSteps {
		nestedStepsCopy = append(nestedStepsCopy, nestedStep.GetCopy())
	}

	copiedConceptStep := new(Step)
	*copiedConceptStep = *self
	copiedConceptStep.ConceptSteps = nestedStepsCopy
	copiedConceptStep.Lookup = *self.Lookup.GetCopy()
	return copiedConceptStep
}

func (self *Step) CopyFrom(another *Step) {
	self.IsConcept = another.IsConcept

	if another.Args == nil {
		self.Args = nil
	} else {
		self.Args = make([]*StepArg, len(another.Args))
		copy(self.Args, another.Args)
	}

	if another.ConceptSteps == nil {
		self.ConceptSteps = nil
	} else {
		self.ConceptSteps = make([]*Step, len(another.ConceptSteps))
		copy(self.ConceptSteps, another.ConceptSteps)
	}

	if another.Fragments == nil {
		self.Fragments = nil
	} else {
		self.Fragments = make([]*gauge_messages.Fragment, len(another.Fragments))
		copy(self.Fragments, another.Fragments)
	}

	self.LineNo = another.LineNo
	self.LineText = another.LineText
	self.HasInlineTable = another.HasInlineTable
	self.Value = another.Value
	self.Lookup = another.Lookup
	self.Parent = another.Parent
}
func (lookup *ArgLookup) AddArgName(argName string) {
	if lookup.ParamIndexMap == nil {
		lookup.ParamIndexMap = make(map[string]int)
		lookup.paramValue = make([]paramNameValue, 0)
	}
	lookup.ParamIndexMap[argName] = len(lookup.paramValue)
	lookup.paramValue = append(lookup.paramValue, paramNameValue{name: argName})
}

func (lookup *ArgLookup) AddArgValue(param string, stepArg *StepArg) {
	paramIndex, ok := lookup.ParamIndexMap[param]
	if !ok {
		panic(fmt.Sprintf("Accessing an invalid parameter (%s)", param))
	}
	lookup.paramValue[paramIndex].stepArg = stepArg
}

func (lookup *ArgLookup) ContainsArg(param string) bool {
	_, ok := lookup.ParamIndexMap[param]
	return ok
}

func (lookup *ArgLookup) GetArg(param string) *StepArg {
	paramIndex, ok := lookup.ParamIndexMap[param]
	if !ok {
		panic(fmt.Sprintf("Accessing an invalid parameter (%s)", param))
	}
	return lookup.paramValue[paramIndex].stepArg
}

func (lookup *ArgLookup) GetCopy() *ArgLookup {
	lookupCopy := new(ArgLookup)
	for key, _ := range lookup.ParamIndexMap {
		lookupCopy.AddArgName(key)
		arg := lookup.GetArg(key)
		if arg != nil {
			lookupCopy.AddArgValue(key, &StepArg{Value: arg.Value, ArgType: arg.ArgType, Table: arg.Table, Name: arg.Name})
		}
	}
	return lookupCopy
}

func (lookup *ArgLookup) FromDataTableRow(datatable *Table, index int) *ArgLookup {
	dataTableLookup := new(ArgLookup)
	if !datatable.IsInitialized() {
		return dataTableLookup
	}
	for _, header := range datatable.Headers {
		dataTableLookup.AddArgName(header)
		dataTableLookup.AddArgValue(header, &StepArg{Value: datatable.Get(header)[index].Value, ArgType: Static})
	}
	return dataTableLookup
}

//create an empty lookup with only args to resolve dynamic params for steps
func (lookup *ArgLookup) FromDataTable(datatable *Table) *ArgLookup {
	dataTableLookup := new(ArgLookup)
	if !datatable.IsInitialized() {
		return dataTableLookup
	}
	for _, header := range datatable.Headers {
		dataTableLookup.AddArgName(header)
	}
	return dataTableLookup
}

func getIndexFor(scenario *Scenario, scenarios []*Scenario) int {
	for index, anItem := range scenarios {
		if reflect.DeepEqual(scenario, anItem) {
			return index
		}
	}
	return -1
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
