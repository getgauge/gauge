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

import (
	"fmt"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
	"regexp"
	"strings"
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

type StepArg struct {
	Name    string
	Value   string
	ArgType ArgType
	Table   Table
}

func (stepArg *StepArg) String() string {
	return fmt.Sprintf("{Name: %s,value %s,argType %s,table %v}", stepArg.Name, stepArg.Value, string(stepArg.ArgType), stepArg.Table)
}

type paramNameValue struct {
	name    string
	stepArg *StepArg
}

func (paramNameValue paramNameValue) String() string {
	return fmt.Sprintf("ParamName: %s, stepArg: %s", paramNameValue.name, paramNameValue.stepArg)
}

type ArgLookup struct {
	//helps to access the index of an arg at O(1)
	paramIndexMap map[string]int
	paramValue    []paramNameValue
}

type SpecItemFilter interface {
	Filter(Item) bool
}

func (argLookup ArgLookup) String() string {
	return fmt.Sprintln(argLookup.paramValue)
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

func (step *Step) getArg(name string) *StepArg {
	arg := step.Lookup.getArg(name)
	// Return static values
	if arg != nil && arg.ArgType != Dynamic {
		return arg
	}
	if step.Parent == nil {
		return step.Lookup.getArg(name)
	}
	return step.Parent.getArg(step.Lookup.getArg(name).Value)
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

func CreateStepFromStepRequest(stepReq *gauge_messages.ExecuteStepRequest) *Step {
	args := createStepArgsFromProtoArguments(stepReq.GetParameters())
	step := &Step{Value: stepReq.GetParsedStepText(),
		LineText: stepReq.GetActualStepText()}
	step.addArgs(args...)
	return step
}

func createStepArgsFromProtoArguments(parameters []*gauge_messages.Parameter) []*StepArg {
	stepArgs := make([]*StepArg, 0)
	for _, parameter := range parameters {
		var arg *StepArg
		switch parameter.GetParameterType() {
		case gauge_messages.Parameter_Static:
			arg = &StepArg{ArgType: Static, Value: parameter.GetValue(), Name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Dynamic:
			arg = &StepArg{ArgType: Dynamic, Value: parameter.GetValue(), Name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Special_String:
			arg = &StepArg{ArgType: SpecialString, Value: parameter.GetValue(), Name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Table:
			arg = &StepArg{ArgType: TableArg, Table: *(TableFrom(parameter.GetTable())), Name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Special_Table:
			arg = &StepArg{ArgType: SpecialTable, Table: *(TableFrom(parameter.GetTable())), Name: parameter.GetName()}
			break
		}
		stepArgs = append(stepArgs, arg)
	}
	return stepArgs
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

type HeadingType int

const (
	SpecHeading     = 0
	ScenarioHeading = 1
)

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

type ParseResult struct {
	ParseError *ParseError
	Warnings   []*Warning
	Ok         bool
	FileName   string
}

func converterFn(predicate func(token *Token, state *int) bool, apply func(token *Token, spec *Specification, state *int) ParseResult) func(*Token, *int, *Specification) ParseResult {

	return func(token *Token, state *int, spec *Specification) ParseResult {
		if !predicate(token, state) {
			return ParseResult{Ok: true}
		}
		return apply(token, spec, state)
	}

}

func (specParser *SpecParser) CreateSpecification(tokens []*Token, conceptDictionary *ConceptDictionary) (*Specification, *ParseResult) {
	specParser.conceptDictionary = conceptDictionary
	converters := specParser.initializeConverters()
	specification := &Specification{}
	finalResult := &ParseResult{}
	state := initial

	for _, token := range tokens {
		for _, converter := range converters {
			result := converter(token, &state, specification)
			if !result.Ok {
				if result.ParseError != nil {
					finalResult.Ok = false
					finalResult.ParseError = result.ParseError
					return nil, finalResult
				}
			}
			if result.Warnings != nil {
				if finalResult.Warnings == nil {
					finalResult.Warnings = make([]*Warning, 0)
				}
				finalResult.Warnings = append(finalResult.Warnings, result.Warnings...)
			}
		}
	}

	specification.processConceptStepsFrom(conceptDictionary)
	validationError := specParser.validateSpec(specification)
	if validationError != nil {
		finalResult.Ok = false
		finalResult.ParseError = validationError
		return nil, finalResult
	}
	finalResult.Ok = true
	return specification, finalResult
}

func (specParser *SpecParser) initializeConverters() []func(*Token, *int, *Specification) ParseResult {
	specConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == SpecKind
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		if spec.Heading != nil {
			return ParseResult{Ok: false, ParseError: &ParseError{token.LineNo, "Parse error: Multiple spec headings found in same file", token.LineText}}
		}

		spec.addHeading(&Heading{LineNo: token.LineNo, Value: token.Value})
		addStates(state, specScope)
		return ParseResult{Ok: true}
	})

	scenarioConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == ScenarioKind
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		if spec.Heading == nil {
			return ParseResult{Ok: false, ParseError: &ParseError{token.LineNo, "Parse error: Scenario should be defined after the spec heading", token.LineText}}
		}
		for _, scenario := range spec.Scenarios {
			if strings.ToLower(scenario.Heading.Value) == strings.ToLower(token.Value) {
				return ParseResult{Ok: false, ParseError: &ParseError{token.LineNo, "Parse error: Duplicate scenario definition '" + scenario.Heading.Value + "' found in the same specification", token.LineText}}
			}
		}
		scenario := &Scenario{}
		scenario.addHeading(&Heading{Value: token.Value, LineNo: token.LineNo})
		spec.addScenario(scenario)

		retainStates(state, specScope)
		addStates(state, scenarioScope)
		return ParseResult{Ok: true}
	})

	stepConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == StepKind && isInState(*state, scenarioScope)
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		latestScenario := spec.latestScenario()
		stepToAdd, parseDetails := spec.createStep(token)
		if parseDetails != nil && parseDetails.Error != nil {
			return ParseResult{ParseError: parseDetails.Error, Ok: false, Warnings: parseDetails.Warnings}
		}
		latestScenario.addStep(stepToAdd)
		retainStates(state, specScope, scenarioScope)
		addStates(state, stepScope)
		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	contextConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == StepKind && !isInState(*state, scenarioScope) && isInState(*state, specScope) && !isInState(*state, tearDownScope)
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		stepToAdd, parseDetails := spec.createStep(token)
		if parseDetails != nil && parseDetails.Error != nil {
			return ParseResult{ParseError: parseDetails.Error, Ok: false, Warnings: parseDetails.Warnings}
		}
		spec.addContext(stepToAdd)
		retainStates(state, specScope)
		addStates(state, contextScope)
		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	tearDownConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == TearDownKind
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		retainStates(state, specScope)
		addStates(state, tearDownScope)
		spec.addItem(&TearDown{LineNo: token.LineNo, Value: token.Value})
		return ParseResult{Ok: true}
	})

	tearDownStepConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == StepKind && isInState(*state, tearDownScope)
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		stepToAdd, parseDetails := spec.createStep(token)
		if parseDetails != nil && parseDetails.Error != nil {
			return ParseResult{ParseError: parseDetails.Error, Ok: false, Warnings: parseDetails.Warnings}
		}
		spec.TearDownSteps = append(spec.TearDownSteps, stepToAdd)
		spec.addItem(stepToAdd)
		retainStates(state, specScope, tearDownScope)

		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	commentConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == CommentKind
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		comment := &Comment{token.Value, token.LineNo}
		if isInState(*state, scenarioScope) {
			spec.latestScenario().addComment(comment)
		} else {
			spec.addComment(comment)
		}
		retainStates(state, specScope, scenarioScope, tearDownScope)
		addStates(state, commentScope)
		return ParseResult{Ok: true}
	})

	keywordConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == DataTableKind
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		resolvedArg, _ := newSpecialTypeResolver().resolve(token.Value)
		if isInState(*state, specScope) && !spec.DataTable.IsInitialized() {
			externalTable := &DataTable{}
			externalTable.Table = resolvedArg.Table
			externalTable.LineNo = token.LineNo
			externalTable.Value = token.Value
			externalTable.IsExternal = true
			spec.addExternalDataTable(externalTable)
		} else if isInState(*state, specScope) && spec.DataTable.IsInitialized() {
			value := "Multiple data table present, ignoring table"
			spec.addComment(&Comment{token.LineText, token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{value, token.LineNo}}}
		} else {
			value := "Data table not associated with spec"
			spec.addComment(&Comment{token.LineText, token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{value, token.LineNo}}}
		}
		retainStates(state, specScope)
		addStates(state, keywordScope)
		return ParseResult{Ok: true}
	})

	tableHeaderConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == TableHeader && isInState(*state, specScope)
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		if isInState(*state, stepScope) {
			latestScenario := spec.latestScenario()
			latestStep := latestScenario.latestStep()
			addInlineTableHeader(latestStep, token)
		} else if isInState(*state, contextScope) {
			latestContext := spec.latestContext()
			addInlineTableHeader(latestContext, token)
		} else if isInState(*state, tearDownScope) {
			latestTeardown := spec.latestTeardown()
			addInlineTableHeader(latestTeardown, token)
		} else if !isInState(*state, scenarioScope) {
			if !spec.DataTable.Table.IsInitialized() {
				dataTable := &Table{}
				dataTable.LineNo = token.LineNo
				dataTable.AddHeaders(token.Args)
				spec.addDataTable(dataTable)
			} else {
				value := "Multiple data table present, ignoring table"
				spec.addComment(&Comment{token.LineText, token.LineNo})
				return ParseResult{Ok: false, Warnings: []*Warning{&Warning{value, token.LineNo}}}
			}
		} else {
			value := "Table not associated with a step, ignoring table"
			spec.latestScenario().addComment(&Comment{token.LineText, token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{value, token.LineNo}}}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tearDownScope)
		addStates(state, tableScope)
		return ParseResult{Ok: true}
	})

	tableRowConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == TableRow
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		var result ParseResult
		//When table is to be treated as a comment
		if !isInState(*state, tableScope) {
			if isInState(*state, scenarioScope) {
				spec.latestScenario().addComment(&Comment{token.LineText, token.LineNo})
			} else {
				spec.addComment(&Comment{token.LineText, token.LineNo})
			}
		} else if areUnderlined(token.Args) {
			// skip table separator
			result = ParseResult{Ok: true}
		} else if isInState(*state, stepScope) {
			latestScenario := spec.latestScenario()
			latestStep := latestScenario.latestStep()
			result = addInlineTableRow(latestStep, token, new(ArgLookup).fromDataTable(&spec.DataTable.Table))
		} else if isInState(*state, contextScope) {
			latestContext := spec.latestContext()
			result = addInlineTableRow(latestContext, token, new(ArgLookup).fromDataTable(&spec.DataTable.Table))
		} else if isInState(*state, tearDownScope) {
			latestTeardown := spec.latestTeardown()
			result = addInlineTableRow(latestTeardown, token, new(ArgLookup).fromDataTable(&spec.DataTable.Table))
		} else {
			//todo validate datatable rows also
			spec.DataTable.Table.AddRowValues(token.Args)
			result = ParseResult{Ok: true}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tearDownScope, tableScope)
		return result
	})

	tagConverter := converterFn(func(token *Token, state *int) bool {
		return (token.Kind == TagKind)
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		tags := &Tags{Values: token.Args}
		if isInState(*state, scenarioScope) {
			spec.latestScenario().addTags(tags)
		} else {
			spec.addTags(tags)
		}
		return ParseResult{Ok: true}
	})

	converter := []func(*Token, *int, *Specification) ParseResult{
		specConverter, scenarioConverter, stepConverter, contextConverter, commentConverter, tableHeaderConverter, tableRowConverter, tagConverter, keywordConverter, tearDownConverter, tearDownStepConverter,
	}

	return converter
}

func (spec *Specification) createStep(stepToken *Token) (*Step, *ParseDetailResult) {
	dataTableLookup := new(ArgLookup).fromDataTable(&spec.DataTable.Table)
	stepToAdd, parseDetails := spec.CreateStepUsingLookup(stepToken, dataTableLookup)

	if parseDetails != nil && parseDetails.Error != nil {
		return nil, parseDetails
	}
	return stepToAdd, parseDetails
}

func (spec *Specification) CreateStepUsingLookup(stepToken *Token, lookup *ArgLookup) (*Step, *ParseDetailResult) {
	stepValue, argsType := extractStepValueAndParameterTypes(stepToken.Value)
	if argsType != nil && len(argsType) != len(stepToken.Args) {
		return nil, &ParseDetailResult{Error: &ParseError{stepToken.LineNo, "Step text should not have '{static}' or '{dynamic}' or '{special}'", stepToken.LineText}, Warnings: nil}
	}
	step := &Step{LineNo: stepToken.LineNo, Value: stepValue, LineText: strings.TrimSpace(stepToken.LineText)}
	arguments := make([]*StepArg, 0)
	var warnings []*Warning
	for i, argType := range argsType {
		argument, parseDetails := spec.createStepArg(stepToken.Args[i], argType, stepToken, lookup)
		if parseDetails != nil && parseDetails.Error != nil {
			return nil, parseDetails
		}
		arguments = append(arguments, argument)
		if parseDetails != nil && parseDetails.Warnings != nil {
			for _, warn := range parseDetails.Warnings {
				warnings = append(warnings, warn)
			}
		}
	}
	step.addArgs(arguments...)
	return step, &ParseDetailResult{Warnings: warnings}
}

func (specification *Specification) processConceptStepsFrom(conceptDictionary *ConceptDictionary) {
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
	if conceptFromDictionary := conceptDictionary.search(step.Value); conceptFromDictionary != nil {
		specification.createConceptStep(conceptFromDictionary.ConceptStep, step)
	}
}

func (specification *Specification) createConceptStep(concept *Step, originalStep *Step) {
	stepCopy := concept.getCopy()
	originalArgs := originalStep.Args
	originalStep.copyFrom(stepCopy)
	originalStep.Args = originalArgs

	// set parent of all concept steps to be the current concept (referred as originalStep here)
	// this is used to fetch from parent's lookup when nested
	for _, conceptStep := range originalStep.ConceptSteps {
		conceptStep.Parent = originalStep
	}

	specification.populateConceptLookup(&originalStep.Lookup, concept.Args, originalStep.Args)
}

func (specification *Specification) addItem(itemToAdd Item) {
	if specification.Items == nil {
		specification.Items = make([]Item, 0)
	}

	specification.Items = append(specification.Items, itemToAdd)
}

func (specification *Specification) addHeading(heading *Heading) {
	heading.HeadingType = SpecHeading
	specification.Heading = heading
}

func (specification *Specification) addScenario(scenario *Scenario) {
	specification.Scenarios = append(specification.Scenarios, scenario)
	specification.addItem(scenario)
}

func (specification *Specification) addContext(contextStep *Step) {
	specification.Contexts = append(specification.Contexts, contextStep)
	specification.addItem(contextStep)
}

func (specification *Specification) addComment(comment *Comment) {
	specification.Comments = append(specification.Comments, comment)
	specification.addItem(comment)
}

func (specification *Specification) addDataTable(table *Table) {
	specification.DataTable.Table = *table
	specification.addItem(&specification.DataTable)
}

func (specification *Specification) addExternalDataTable(externalTable *DataTable) {
	specification.DataTable = *externalTable
	specification.addItem(externalTable)
}

func (specification *Specification) addTags(tags *Tags) {
	specification.Tags = tags
	specification.addItem(tags)
}

func (specification *Specification) latestScenario() *Scenario {
	return specification.Scenarios[len(specification.Scenarios)-1]
}

func (specification *Specification) latestContext() *Step {
	return specification.Contexts[len(specification.Contexts)-1]
}

func (specification *Specification) latestTeardown() *Step {
	return specification.TearDownSteps[len(specification.TearDownSteps)-1]
}

func (specParser *SpecParser) validateSpec(specification *Specification) *ParseError {
	if len(specification.Items) == 0 {
		return &ParseError{LineNo: 1, Message: "Spec does not have any elements"}
	}
	if specification.Heading == nil {
		return &ParseError{LineNo: 1, Message: "Spec heading not found"}
	}
	dataTable := specification.DataTable.Table
	if dataTable.IsInitialized() && dataTable.GetRowCount() == 0 {
		return &ParseError{LineNo: dataTable.LineNo, Message: "Data table should have at least 1 data row"}
	}
	return nil
}

func extractStepValueAndParameterTypes(stepTokenValue string) (string, []string) {
	argsType := make([]string, 0)
	r := regexp.MustCompile("{(dynamic|static|special)}")
	/*
		enter {dynamic} and {static}
		returns
		[
		["{dynamic}","dynamic"]
		["{static}","static"]
		]
	*/
	args := r.FindAllStringSubmatch(stepTokenValue, -1)

	if args == nil {
		return stepTokenValue, nil
	}
	for _, arg := range args {
		//arg[1] extracts the first group
		argsType = append(argsType, arg[1])
	}
	return r.ReplaceAllString(stepTokenValue, ParameterPlaceholder), argsType
}

func (step *Step) addArgs(args ...*StepArg) {
	step.Args = append(step.Args, args...)
	step.PopulateFragments()
}

func (step *Step) addInlineTableHeaders(headers []string) {
	tableArg := &StepArg{ArgType: TableArg}
	tableArg.Table.AddHeaders(headers)
	step.addArgs(tableArg)
}

func (step *Step) addInlineTableRow(row []TableCell) {
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

func (spec *Specification) populateConceptLookup(lookup *ArgLookup, conceptArgs []*StepArg, stepArgs []*StepArg) {
	for i, arg := range stepArgs {
		lookup.addArgValue(conceptArgs[i].Value, &StepArg{Value: arg.Value, ArgType: arg.ArgType, Table: arg.Table, Name: arg.Name})
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

func (spec *Specification) createStepArg(argValue string, typeOfArg string, token *Token, lookup *ArgLookup) (*StepArg, *ParseDetailResult) {
	if typeOfArg == "special" {
		resolvedArgValue, err := newSpecialTypeResolver().resolve(argValue)
		if err != nil {
			switch err.(type) {
			case invalidSpecialParamError:
				return treatArgAsDynamic(argValue, token, lookup)
			default:
				return nil, &ParseDetailResult{Error: &ParseError{LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), LineText: token.LineText}}
			}
		}
		return resolvedArgValue, nil
	} else if typeOfArg == "static" {
		return &StepArg{ArgType: Static, Value: argValue}, nil
	} else {
		return validateDynamicArg(argValue, token, lookup)
	}
}

func treatArgAsDynamic(argValue string, token *Token, lookup *ArgLookup) (*StepArg, *ParseDetailResult) {
	parseDetailRes := &ParseDetailResult{Warnings: []*Warning{&Warning{LineNo: token.LineNo, Message: fmt.Sprintf("Could not resolve special param type <%s>. Treating it as dynamic param.", argValue)}}}
	stepArg, result := validateDynamicArg(argValue, token, lookup)
	if result != nil {
		if result.Error != nil {
			parseDetailRes.Error = result.Error
		}
		if result.Warnings != nil {
			for _, warn := range result.Warnings {
				parseDetailRes.Warnings = append(parseDetailRes.Warnings, warn)
			}
		}
	}
	return stepArg, parseDetailRes
}

func validateDynamicArg(argValue string, token *Token, lookup *ArgLookup) (*StepArg, *ParseDetailResult) {
	if !isConceptHeader(lookup) && !lookup.containsArg(argValue) {
		return nil, &ParseDetailResult{Error: &ParseError{LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), LineText: token.LineText}}
	}
	stepArgument := &StepArg{ArgType: Dynamic, Value: argValue, Name: argValue}
	return stepArgument, nil

}

//Step value is modified when inline table is found to account for the new parameter by appending {}
//todo validate headers for dynamic
func addInlineTableHeader(step *Step, token *Token) {
	step.Value = fmt.Sprintf("%s %s", step.Value, ParameterPlaceholder)
	step.HasInlineTable = true
	step.addInlineTableHeaders(token.Args)

}

func addInlineTableRow(step *Step, token *Token, argLookup *ArgLookup) ParseResult {
	dynamicArgMatcher := regexp.MustCompile("^<(.*)>$")
	tableValues := make([]TableCell, 0)
	warnings := make([]*Warning, 0)
	for _, tableValue := range token.Args {
		if dynamicArgMatcher.MatchString(tableValue) {
			match := dynamicArgMatcher.FindAllStringSubmatch(tableValue, -1)
			param := match[0][1]
			if !argLookup.containsArg(param) {
				tableValues = append(tableValues, TableCell{Value: tableValue, CellType: Static})
				warnings = append(warnings, &Warning{LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic param <%s> could not be resolved, Treating it as static param", param)})
			} else {
				tableValues = append(tableValues, TableCell{Value: param, CellType: Dynamic})
			}
		} else {
			tableValues = append(tableValues, TableCell{Value: tableValue, CellType: Static})
		}
	}
	step.addInlineTableRow(tableValues)
	return ParseResult{Ok: true, Warnings: warnings}
}

//concept header will have dynamic param and should not be resolved through lookup, so passing nil lookup
func isConceptHeader(lookup *ArgLookup) bool {
	return lookup == nil
}

func (lookup *ArgLookup) addArgName(argName string) {
	if lookup.paramIndexMap == nil {
		lookup.paramIndexMap = make(map[string]int)
		lookup.paramValue = make([]paramNameValue, 0)
	}
	lookup.paramIndexMap[argName] = len(lookup.paramValue)
	lookup.paramValue = append(lookup.paramValue, paramNameValue{name: argName})
}

func (lookup *ArgLookup) addArgValue(param string, stepArg *StepArg) {
	paramIndex, ok := lookup.paramIndexMap[param]
	if !ok {
		panic(fmt.Sprintf("Accessing an invalid parameter (%s)", param))
	}
	lookup.paramValue[paramIndex].stepArg = stepArg
}

func (lookup *ArgLookup) containsArg(param string) bool {
	_, ok := lookup.paramIndexMap[param]
	return ok
}

func (lookup *ArgLookup) getArg(param string) *StepArg {
	paramIndex, ok := lookup.paramIndexMap[param]
	if !ok {
		panic(fmt.Sprintf("Accessing an invalid parameter (%s)", param))
	}
	return lookup.paramValue[paramIndex].stepArg
}

func (lookup *ArgLookup) getCopy() *ArgLookup {
	lookupCopy := new(ArgLookup)
	for key, _ := range lookup.paramIndexMap {
		lookupCopy.addArgName(key)
		arg := lookup.getArg(key)
		if arg != nil {
			lookupCopy.addArgValue(key, &StepArg{Value: arg.Value, ArgType: arg.ArgType, Table: arg.Table, Name: arg.Name})
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
		dataTableLookup.addArgName(header)
		dataTableLookup.addArgValue(header, &StepArg{Value: datatable.Get(header)[index].Value, ArgType: Static})
	}
	return dataTableLookup
}

//create an empty lookup with only args to resolve dynamic params for steps
func (lookup *ArgLookup) fromDataTable(datatable *Table) *ArgLookup {
	dataTableLookup := new(ArgLookup)
	if !datatable.IsInitialized() {
		return dataTableLookup
	}
	for _, header := range datatable.Headers {
		dataTableLookup.addArgName(header)
	}
	return dataTableLookup
}

func (warning *Warning) String() string {
	return fmt.Sprintf("line no: %d, %s", warning.LineNo, warning.Message)
}

func (scenario Scenario) Kind() TokenKind {
	return ScenarioKind
}

func (scenario *Scenario) addHeading(heading *Heading) {
	heading.HeadingType = ScenarioHeading
	scenario.Heading = heading
}

func (scenario *Scenario) addStep(step *Step) {
	scenario.Steps = append(scenario.Steps, step)
	scenario.addItem(step)
}

func (scenario *Scenario) addTags(tags *Tags) {
	scenario.Tags = tags
	scenario.addItem(tags)
}

func (scenario *Scenario) addComment(comment *Comment) {
	scenario.Comments = append(scenario.Comments, comment)
	scenario.addItem(comment)
}

func (scenario *Scenario) renameSteps(oldStep Step, newStep Step, orderMap map[int]int) bool {
	isRefactored := false
	for _, step := range scenario.Steps {
		isConcept := false
		isRefactored = step.Rename(oldStep, newStep, isRefactored, orderMap, &isConcept)
	}
	return isRefactored
}

func (scenario *Scenario) addItem(itemToAdd Item) {
	if scenario.Items == nil {
		scenario.Items = make([]Item, 0)
	}
	scenario.Items = append(scenario.Items, itemToAdd)
}

func (scenario *Scenario) latestStep() *Step {
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
func (self *Step) getCopy() *Step {
	if !self.IsConcept {
		return self
	}
	nestedStepsCopy := make([]*Step, 0)
	for _, nestedStep := range self.ConceptSteps {
		nestedStepsCopy = append(nestedStepsCopy, nestedStep.getCopy())
	}

	copiedConceptStep := new(Step)
	*copiedConceptStep = *self
	copiedConceptStep.ConceptSteps = nestedStepsCopy
	copiedConceptStep.Lookup = *self.Lookup.getCopy()
	return copiedConceptStep
}

func (self *Step) copyFrom(another *Step) {
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

func ConvertToStepText(fragments []*gauge_messages.Fragment) string {
	stepText := ""
	for _, fragment := range fragments {
		value := ""
		if fragment.GetFragmentType() == gauge_messages.Fragment_Text {
			value = fragment.GetText()
		} else {
			switch fragment.GetParameter().GetParameterType() {
			case gauge_messages.Parameter_Static:
				value = fmt.Sprintf("\"%s\"", fragment.GetParameter().GetValue())
				break
			case gauge_messages.Parameter_Dynamic:
				value = fmt.Sprintf("<%s>", fragment.GetParameter().GetValue())
				break
			}
		}
		stepText += value
	}
	return stepText
}

func (result *ParseResult) Error() string {
	return fmt.Sprintf("[ParseError] %s : %s", result.FileName, result.ParseError.Error())
}
