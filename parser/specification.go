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
	heading  *Heading
	steps    []*Step
	comments []*Comment
	tags     *Tags
	items    []Item
}

type ArgType string

const (
	Static                ArgType = "static"
	Dynamic               ArgType = "dynamic"
	TableArg              ArgType = "table"
	SpecialString         ArgType = "special_string"
	SpecialTable          ArgType = "special_table"
	PARAMETER_PLACEHOLDER         = "{}"
)

type StepArg struct {
	name    string
	value   string
	argType ArgType
	table   Table
}

func (stepArg *StepArg) String() string {
	return fmt.Sprintf("{Name: %s,value %s,argType %s,table %v}", stepArg.name, stepArg.value, string(stepArg.argType), stepArg.table)
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

type specItemFilter interface {
	filter(Item) bool
}

func (argLookup ArgLookup) String() string {
	return fmt.Sprintln(argLookup.paramValue)
}

type Step struct {
	lineNo         int
	value          string
	lineText       string
	args           []*StepArg
	isConcept      bool
	lookup         ArgLookup
	conceptSteps   []*Step
	fragments      []*gauge_messages.Fragment
	parent         *Step
	hasInlineTable bool
	items          []Item
	preComments    []*Comment
}

type StepValue struct {
	args                   []string
	stepValue              string
	parameterizedStepValue string
}

func (step *Step) getArg(name string) *StepArg {
	arg := step.lookup.getArg(name)
	// Return static values
	if arg != nil && arg.argType != Dynamic {
		return arg
	}
	if step.parent == nil {
		return step.lookup.getArg(name)
	}
	return step.parent.getArg(step.lookup.getArg(name).value)
}

func (step *Step) getLineText() string {
	if step.hasInlineTable {
		return fmt.Sprintf("%s <%s>", step.lineText, TableArg)
	}
	return step.lineText
}

func (step *Step) rename(oldStep Step, newStep Step, isRefactored bool, orderMap map[int]int, isConcept *bool) bool {
	if strings.TrimSpace(step.value) != strings.TrimSpace(oldStep.value) {
		return isRefactored
	}
	if step.isConcept {
		*isConcept = true
	}
	step.value = newStep.value

	step.args = step.getArgsInOrder(newStep, orderMap)
	return true
}

func (step *Step) getArgsInOrder(newStep Step, orderMap map[int]int) []*StepArg {
	args := make([]*StepArg, len(newStep.args))
	for key, value := range orderMap {
		arg := &StepArg{value: newStep.args[key].value, argType: Static}
		if step.isConcept {
			arg = &StepArg{value: newStep.args[key].value, argType: Dynamic}
		}
		if value != -1 {
			arg = step.args[value]
		}
		args[key] = arg
	}
	return args
}

func (step *Step) deepCopyStepArgs() []*StepArg {
	copiedStepArgs := make([]*StepArg, 0)
	for _, conceptStepArg := range step.args {
		temp := new(StepArg)
		*temp = *conceptStepArg
		copiedStepArgs = append(copiedStepArgs, temp)
	}
	return copiedStepArgs
}

func (step *Step) replaceArgsWithDynamic(args []*StepArg) {
	for i, arg := range step.args {
		for _, conceptArg := range args {
			if arg.String() == conceptArg.String() {
				step.args[i] = &StepArg{value: conceptArg.value, argType: Dynamic}
			}
		}
	}
}

func createStepFromStepRequest(stepReq *gauge_messages.ExecuteStepRequest) *Step {
	args := createStepArgsFromProtoArguments(stepReq.GetParameters())
	step := &Step{value: stepReq.GetParsedStepText(),
		lineText: stepReq.GetActualStepText()}
	step.addArgs(args...)
	return step
}

func createStepArgsFromProtoArguments(parameters []*gauge_messages.Parameter) []*StepArg {
	stepArgs := make([]*StepArg, 0)
	for _, parameter := range parameters {
		var arg *StepArg
		switch parameter.GetParameterType() {
		case gauge_messages.Parameter_Static:
			arg = &StepArg{argType: Static, value: parameter.GetValue(), name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Dynamic:
			arg = &StepArg{argType: Dynamic, value: parameter.GetValue(), name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Special_String:
			arg = &StepArg{argType: SpecialString, value: parameter.GetValue(), name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Table:
			arg = &StepArg{argType: TableArg, table: *(tableFrom(parameter.GetTable())), name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Special_Table:
			arg = &StepArg{argType: SpecialTable, table: *(tableFrom(parameter.GetTable())), name: parameter.GetName()}
			break
		}
		stepArgs = append(stepArgs, arg)
	}
	return stepArgs
}

type Specification struct {
	heading   *Heading
	scenarios []*Scenario
	comments  []*Comment
	dataTable DataTable
	contexts  []*Step
	fileName  string
	tags      *Tags
	items     []Item
}

type Item interface {
	kind() TokenKind
}

type HeadingType int

const (
	SpecHeading     = 0
	scenarioHeading = 1
)

type Heading struct {
	value       string
	lineNo      int
	headingType HeadingType
}

type Comment struct {
	value  string
	lineNo int
}

type Tags struct {
	values []string
}

type Warning struct {
	message string
	lineNo  int
}

type ParseResult struct {
	error    *parseError
	warnings []*Warning
	ok       bool
	fileName string
}

func converterFn(predicate func(token *Token, state *int) bool, apply func(token *Token, spec *Specification, state *int) ParseResult) func(*Token, *int, *Specification) ParseResult {

	return func(token *Token, state *int, spec *Specification) ParseResult {
		if !predicate(token, state) {
			return ParseResult{ok: true}
		}
		return apply(token, spec, state)
	}

}

func (specParser *SpecParser) createSpecification(tokens []*Token, conceptDictionary *ConceptDictionary) (*Specification, *ParseResult) {
	specParser.conceptDictionary = conceptDictionary
	converters := specParser.initializeConverters()
	specification := &Specification{}
	finalResult := &ParseResult{}
	state := initial

	for _, token := range tokens {
		for _, converter := range converters {
			result := converter(token, &state, specification)
			if !result.ok {
				if result.error != nil {
					finalResult.ok = false
					finalResult.error = result.error
					return nil, finalResult
				}
			}
			if result.warnings != nil {
				if finalResult.warnings == nil {
					finalResult.warnings = make([]*Warning, 0)
				}
				finalResult.warnings = append(finalResult.warnings, result.warnings...)
			}
		}
	}

	specification.processConceptStepsFrom(conceptDictionary)
	validationError := specParser.validateSpec(specification)
	if validationError != nil {
		finalResult.ok = false
		finalResult.error = validationError
		return nil, finalResult
	}
	finalResult.ok = true
	return specification, finalResult
}

func (specParser *SpecParser) initializeConverters() []func(*Token, *int, *Specification) ParseResult {
	specConverter := converterFn(func(token *Token, state *int) bool {
		return token.kind == specKind
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		if spec.heading != nil {
			return ParseResult{ok: false, error: &parseError{token.lineNo, "Parse error: Multiple spec headings found in same file", token.lineText}}
		}

		spec.addHeading(&Heading{lineNo: token.lineNo, value: token.value})
		addStates(state, specScope)
		return ParseResult{ok: true}
	})

	scenarioConverter := converterFn(func(token *Token, state *int) bool {
		return token.kind == scenarioKind
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		if spec.heading == nil {
			return ParseResult{ok: false, error: &parseError{token.lineNo, "Parse error: Scenario should be defined after the spec heading", token.lineText}}
		}
		for _, scenario := range spec.scenarios {
			if strings.ToLower(scenario.heading.value) == strings.ToLower(token.value) {
				return ParseResult{ok: false, error: &parseError{token.lineNo, "Parse error: Duplicate scenario definitions are not allowed in the same specification", token.lineText}}
			}
		}
		scenario := &Scenario{}
		scenario.addHeading(&Heading{value: token.value, lineNo: token.lineNo})
		spec.addScenario(scenario)

		retainStates(state, specScope)
		addStates(state, scenarioScope)
		return ParseResult{ok: true}
	})

	stepConverter := converterFn(func(token *Token, state *int) bool {
		return token.kind == stepKind && isInState(*state, scenarioScope)
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		latestScenario := spec.latestScenario()
		stepToAdd, parseDetails := spec.createStep(token)
		if parseDetails != nil && parseDetails.error != nil {
			return ParseResult{error: parseDetails.error, ok: false, warnings: parseDetails.warnings}
		}
		latestScenario.addStep(stepToAdd)
		retainStates(state, specScope, scenarioScope)
		addStates(state, stepScope)
		if parseDetails.warnings != nil {
			return ParseResult{ok: false, warnings: parseDetails.warnings}
		}
		return ParseResult{ok: true, warnings: parseDetails.warnings}
	})

	contextConverter := converterFn(func(token *Token, state *int) bool {
		return token.kind == stepKind && !isInState(*state, scenarioScope) && isInState(*state, specScope)
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		stepToAdd, parseDetails := spec.createStep(token)
		if parseDetails != nil && parseDetails.error != nil {
			return ParseResult{error: parseDetails.error, ok: false, warnings: parseDetails.warnings}
		}
		spec.addContext(stepToAdd)
		retainStates(state, specScope)
		addStates(state, contextScope)
		if parseDetails.warnings != nil {
			return ParseResult{ok: false, warnings: parseDetails.warnings}
		}
		return ParseResult{ok: true, warnings: parseDetails.warnings}
	})

	commentConverter := converterFn(func(token *Token, state *int) bool {
		return token.kind == commentKind
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		comment := &Comment{token.value, token.lineNo}
		if isInState(*state, scenarioScope) {
			spec.latestScenario().addComment(comment)
		} else {
			spec.addComment(comment)
		}
		retainStates(state, specScope, scenarioScope)
		addStates(state, commentScope)
		return ParseResult{ok: true}
	})

	keywordConverter := converterFn(func(token *Token, state *int) bool {
		return token.kind == dataTableKind
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		resolvedArg, _ := newSpecialTypeResolver().resolve(token.value)
		if isInState(*state, specScope) && !spec.dataTable.isInitialized() {
			externalTable := &DataTable{}
			externalTable.table = resolvedArg.table
			externalTable.lineNo = token.lineNo
			externalTable.value = token.value
			externalTable.isExternal = true
			spec.addExternalDataTable(externalTable)
		} else if isInState(*state, specScope) && spec.dataTable.isInitialized() {
			value := "Multiple data table present, ignoring table"
			spec.addComment(&Comment{token.lineText, token.lineNo})
			return ParseResult{ok: false, warnings: []*Warning{&Warning{value, token.lineNo}}}
		} else {
			value := "Data table not associated with spec"
			spec.addComment(&Comment{token.lineText, token.lineNo})
			return ParseResult{ok: false, warnings: []*Warning{&Warning{value, token.lineNo}}}
		}
		retainStates(state, specScope)
		addStates(state, keywordScope)
		return ParseResult{ok: true}
	})

	tableHeaderConverter := converterFn(func(token *Token, state *int) bool {
		return token.kind == tableHeader && isInState(*state, specScope)
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		if isInState(*state, stepScope) {
			latestScenario := spec.latestScenario()
			latestStep := latestScenario.latestStep()
			addInlineTableHeader(latestStep, token)
		} else if isInState(*state, contextScope) {
			latestContext := spec.latestContext()
			addInlineTableHeader(latestContext, token)
		} else if !isInState(*state, scenarioScope) {
			if !spec.dataTable.table.isInitialized() {
				dataTable := &Table{}
				dataTable.lineNo = token.lineNo
				dataTable.addHeaders(token.args)
				spec.addDataTable(dataTable)
			} else {
				value := "Multiple data table present, ignoring table"
				spec.addComment(&Comment{token.lineText, token.lineNo})
				return ParseResult{ok: false, warnings: []*Warning{&Warning{value, token.lineNo}}}
			}
		} else {
			value := "Table not associated with a step, ignoring table"
			spec.latestScenario().addComment(&Comment{token.lineText, token.lineNo})
			return ParseResult{ok: false, warnings: []*Warning{&Warning{value, token.lineNo}}}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope)
		addStates(state, tableScope)
		return ParseResult{ok: true}
	})

	tableRowConverter := converterFn(func(token *Token, state *int) bool {
		return token.kind == tableRow
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		var result ParseResult
		//When table is to be treated as a comment
		if !isInState(*state, tableScope) {
			if isInState(*state, scenarioScope) {
				spec.latestScenario().addComment(&Comment{token.lineText, token.lineNo})
			} else {
				spec.addComment(&Comment{token.lineText, token.lineNo})
			}
		} else if areUnderlined(token.args) {
			// skip table separator
			result = ParseResult{ok: true}
		} else if isInState(*state, stepScope) {
			latestScenario := spec.latestScenario()
			latestStep := latestScenario.latestStep()
			result = addInlineTableRow(latestStep, token, new(ArgLookup).fromDataTable(&spec.dataTable.table))
		} else if isInState(*state, contextScope) {
			latestContext := spec.latestContext()
			result = addInlineTableRow(latestContext, token, new(ArgLookup).fromDataTable(&spec.dataTable.table))
		} else {
			//todo validate datatable rows also
			spec.dataTable.table.addRowValues(token.args)
			result = ParseResult{ok: true}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tableScope)
		return result
	})

	tagConverter := converterFn(func(token *Token, state *int) bool {
		return (token.kind == tagKind)
	}, func(token *Token, spec *Specification, state *int) ParseResult {
		tags := &Tags{values: token.args}
		if isInState(*state, scenarioScope) {
			spec.latestScenario().addTags(tags)
		} else {
			spec.addTags(tags)
		}
		return ParseResult{ok: true}
	})

	converter := []func(*Token, *int, *Specification) ParseResult{
		specConverter, scenarioConverter, stepConverter, contextConverter, commentConverter, tableHeaderConverter, tableRowConverter, tagConverter, keywordConverter,
	}

	return converter
}

func (spec *Specification) createStep(stepToken *Token) (*Step, *parseDetailResult) {
	dataTableLookup := new(ArgLookup).fromDataTable(&spec.dataTable.table)
	stepToAdd, parseDetails := spec.createStepUsingLookup(stepToken, dataTableLookup)

	if parseDetails != nil && parseDetails.error != nil {
		return nil, parseDetails
	}
	return stepToAdd, parseDetails
}

func (spec *Specification) createStepUsingLookup(stepToken *Token, lookup *ArgLookup) (*Step, *parseDetailResult) {
	stepValue, argsType := extractStepValueAndParameterTypes(stepToken.value)
	if argsType != nil && len(argsType) != len(stepToken.args) {
		return nil, &parseDetailResult{error: &parseError{stepToken.lineNo, "Step text should not have '{static}' or '{dynamic}' or '{special}'", stepToken.lineText}, warnings: nil}
	}
	step := &Step{lineNo: stepToken.lineNo, value: stepValue, lineText: strings.TrimSpace(stepToken.lineText)}
	arguments := make([]*StepArg, 0)
	var warnings []*Warning
	for i, argType := range argsType {
		argument, parseDetails := spec.createStepArg(stepToken.args[i], argType, stepToken, lookup)
		if parseDetails != nil && parseDetails.error != nil {
			return nil, parseDetails
		}
		arguments = append(arguments, argument)
		if parseDetails != nil && parseDetails.warnings != nil {
			for _, warn := range parseDetails.warnings {
				warnings = append(warnings, warn)
			}
		}
	}
	step.addArgs(arguments...)
	return step, &parseDetailResult{warnings: warnings}
}

func (specification *Specification) processConceptStepsFrom(conceptDictionary *ConceptDictionary) {
	for _, step := range specification.contexts {
		specification.processConceptStep(step, conceptDictionary)
	}
	for _, scenario := range specification.scenarios {
		for _, step := range scenario.steps {
			specification.processConceptStep(step, conceptDictionary)
		}
	}
}

func (specification *Specification) processConceptStep(step *Step, conceptDictionary *ConceptDictionary) {
	if conceptFromDictionary := conceptDictionary.search(step.value); conceptFromDictionary != nil {
		specification.createConceptStep(conceptFromDictionary.conceptStep, step)
	}
}

func (specification *Specification) createConceptStep(concept *Step, originalStep *Step) {
	stepCopy := concept.getCopy()
	originalArgs := originalStep.args
	originalStep.copyFrom(stepCopy)
	originalStep.args = originalArgs

	// set parent of all concept steps to be the current concept (referred as originalStep here)
	// this is used to fetch from parent's lookup when nested
	for _, conceptStep := range originalStep.conceptSteps {
		conceptStep.parent = originalStep
	}

	specification.populateConceptLookup(&originalStep.lookup, concept.args, originalStep.args)
}

func (specification *Specification) addItem(itemToAdd Item) {
	if specification.items == nil {
		specification.items = make([]Item, 0)
	}

	specification.items = append(specification.items, itemToAdd)
}

func (specification *Specification) addHeading(heading *Heading) {
	heading.headingType = SpecHeading
	specification.heading = heading
}

func (specification *Specification) addScenario(scenario *Scenario) {
	specification.scenarios = append(specification.scenarios, scenario)
	specification.addItem(scenario)
}

func (specification *Specification) addContext(contextStep *Step) {
	specification.contexts = append(specification.contexts, contextStep)
	specification.addItem(contextStep)
}

func (specification *Specification) addComment(comment *Comment) {
	specification.comments = append(specification.comments, comment)
	specification.addItem(comment)
}

func (specification *Specification) addDataTable(table *Table) {
	specification.dataTable.table = *table
	specification.addItem(table)
}

func (specification *Specification) addExternalDataTable(externalTable *DataTable) {
	specification.dataTable = *externalTable
	specification.addItem(externalTable)
}

func (specification *Specification) addTags(tags *Tags) {
	specification.tags = tags
	specification.addItem(tags)
}

func (specification *Specification) latestScenario() *Scenario {
	return specification.scenarios[len(specification.scenarios)-1]
}

func (specification *Specification) latestContext() *Step {
	return specification.contexts[len(specification.contexts)-1]
}

func (specParser *SpecParser) validateSpec(specification *Specification) *parseError {
	if len(specification.items) == 0 {
		return &parseError{lineNo: 1, message: "Spec does not have any elements"}
	}
	if specification.heading == nil {
		return &parseError{lineNo: 1, message: "Spec heading not found"}
	}
	dataTable := specification.dataTable.table
	if dataTable.isInitialized() && dataTable.getRowCount() == 0 {
		return &parseError{lineNo: dataTable.lineNo, message: "Data table should have at least 1 data row"}
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
	return r.ReplaceAllString(stepTokenValue, PARAMETER_PLACEHOLDER), argsType
}

func (step *Step) addArgs(args ...*StepArg) {
	step.args = append(step.args, args...)
	step.populateFragments()
}

func (step *Step) addInlineTableHeaders(headers []string) {
	tableArg := &StepArg{argType: TableArg}
	tableArg.table.addHeaders(headers)
	step.addArgs(tableArg)
}

func (step *Step) addInlineTableRow(row []TableCell) {
	lastArg := step.args[len(step.args)-1]
	lastArg.table.addRows(row)
	step.populateFragments()
}

func (step *Step) populateFragments() {
	r := regexp.MustCompile(PARAMETER_PLACEHOLDER)
	/*
		enter {} and {} bar
		returns
		[[6 8] [13 15]]
	*/
	argSplitIndices := r.FindAllStringSubmatchIndex(step.value, -1)
	step.fragments = make([]*gauge_messages.Fragment, 0)
	if len(step.args) == 0 {
		step.fragments = append(step.fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String(step.value)})
		return
	}

	textStartIndex := 0
	for argIndex, argIndices := range argSplitIndices {
		if textStartIndex < argIndices[0] {
			step.fragments = append(step.fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String(step.value[textStartIndex:argIndices[0]])})
		}
		parameter := convertToProtoParameter(step.args[argIndex])
		step.fragments = append(step.fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter.Enum(), Parameter: parameter})
		textStartIndex = argIndices[1]
	}
	if textStartIndex < len(step.value) {
		step.fragments = append(step.fragments, &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String(step.value[textStartIndex:len(step.value)])})
	}

}

func (spec *Specification) filter(filter specItemFilter) {
	for i := 0; i < len(spec.items); i++ {
		if filter.filter(spec.items[i]) {
			spec.removeItem(i)
			i--
		}
	}
}

func (spec *Specification) removeItem(itemIndex int) {
	item := spec.items[itemIndex]
	if len(spec.items)-1 == itemIndex {
		spec.items = spec.items[:itemIndex]
	} else if 0 == itemIndex {
		spec.items = spec.items[itemIndex+1:]
	} else {
		spec.items = append(spec.items[:itemIndex], spec.items[itemIndex+1:]...)
	}
	if item.kind() == scenarioKind {
		spec.removeScenario(item.(*Scenario))
	}
}

func (spec *Specification) removeScenario(scenario *Scenario) {
	index := getIndexFor(scenario, spec.scenarios)
	if len(spec.scenarios)-1 == index {
		spec.scenarios = spec.scenarios[:index]
	} else if index == 0 {
		spec.scenarios = spec.scenarios[index+1:]
	} else {
		spec.scenarios = append(spec.scenarios[:index], spec.scenarios[index+1:]...)
	}
}

func (spec *Specification) populateConceptLookup(lookup *ArgLookup, conceptArgs []*StepArg, stepArgs []*StepArg) {
	for i, arg := range stepArgs {
		lookup.addArgValue(conceptArgs[i].value, &StepArg{value: arg.value, argType: arg.argType, table: arg.table, name: arg.name})
	}
}

func (spec *Specification) renameSteps(oldStep Step, newStep Step, orderMap map[int]int) bool {
	isRefactored := false
	for _, step := range spec.contexts {
		isConcept := false
		isRefactored = step.rename(oldStep, newStep, isRefactored, orderMap, &isConcept)
	}
	for _, scenario := range spec.scenarios {
		refactor := scenario.renameSteps(oldStep, newStep, orderMap)
		if refactor {
			isRefactored = refactor
		}
	}
	return isRefactored
}

func (spec *Specification) createStepArg(argValue string, typeOfArg string, token *Token, lookup *ArgLookup) (*StepArg, *parseDetailResult) {
	if typeOfArg == "special" {
		resolvedArgValue, err := newSpecialTypeResolver().resolve(argValue)
		if err != nil {
			switch err.(type) {
			case invalidSpecialParamError:
				return treatArgAsDynamic(argValue, token, lookup)
			default:
				return nil, &parseDetailResult{error: &parseError{lineNo: token.lineNo, message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), lineText: token.lineText}}
			}
		}
		return resolvedArgValue, nil
	} else if typeOfArg == "static" {
		return &StepArg{argType: Static, value: argValue}, nil
	} else {
		return validateDynamicArg(argValue, token, lookup)
	}
}

func treatArgAsDynamic(argValue string, token *Token, lookup *ArgLookup) (*StepArg, *parseDetailResult) {
	parseDetailRes := &parseDetailResult{warnings: []*Warning{&Warning{lineNo: token.lineNo, message: fmt.Sprintf("Could not resolve special param type <%s>. Treating it as dynamic param.", argValue)}}}
	stepArg, result := validateDynamicArg(argValue, token, lookup)
	if result != nil {
		if result.error != nil {
			parseDetailRes.error = result.error
		}
		if result.warnings != nil {
			for _, warn := range result.warnings {
				parseDetailRes.warnings = append(parseDetailRes.warnings, warn)
			}
		}
	}
	return stepArg, parseDetailRes
}

func validateDynamicArg(argValue string, token *Token, lookup *ArgLookup) (*StepArg, *parseDetailResult) {
	if !isConceptHeader(lookup) && !lookup.containsArg(argValue) {
		return nil, &parseDetailResult{error: &parseError{lineNo: token.lineNo, message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), lineText: token.lineText}}
	}
	stepArgument := &StepArg{argType: Dynamic, value: argValue, name: argValue}
	return stepArgument, nil

}

//Step value is modified when inline table is found to account for the new parameter by appending {}
//todo validate headers for dynamic
func addInlineTableHeader(step *Step, token *Token) {
	step.value = fmt.Sprintf("%s %s", step.value, PARAMETER_PLACEHOLDER)
	step.hasInlineTable = true
	step.addInlineTableHeaders(token.args)

}

func addInlineTableRow(step *Step, token *Token, argLookup *ArgLookup) ParseResult {
	dynamicArgMatcher := regexp.MustCompile("^<(.*)>$")
	tableValues := make([]TableCell, 0)
	warnings := make([]*Warning, 0)
	for _, tableValue := range token.args {
		if dynamicArgMatcher.MatchString(tableValue) {
			match := dynamicArgMatcher.FindAllStringSubmatch(tableValue, -1)
			param := match[0][1]
			if !argLookup.containsArg(param) {
				tableValues = append(tableValues, TableCell{value: tableValue, cellType: Static})
				warnings = append(warnings, &Warning{lineNo: token.lineNo, message: fmt.Sprintf("Dynamic param <%s> could not be resolved, Treating it as static param", param)})
			} else {
				tableValues = append(tableValues, TableCell{value: param, cellType: Dynamic})
			}
		} else {
			tableValues = append(tableValues, TableCell{value: tableValue, cellType: Static})
		}
	}
	step.addInlineTableRow(tableValues)
	return ParseResult{ok: true, warnings: warnings}
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
			lookupCopy.addArgValue(key, &StepArg{value: arg.value, argType: arg.argType, table: arg.table, name: arg.name})
		}
	}
	return lookupCopy
}

func (lookup *ArgLookup) fromDataTableRow(datatable *Table, index int) *ArgLookup {
	dataTableLookup := new(ArgLookup)
	if !datatable.isInitialized() {
		return dataTableLookup
	}
	for _, header := range datatable.headers {
		dataTableLookup.addArgName(header)
		dataTableLookup.addArgValue(header, &StepArg{value: datatable.get(header)[index].value, argType: Static})
	}
	return dataTableLookup
}

//create an empty lookup with only args to resolve dynamic params for steps
func (lookup *ArgLookup) fromDataTable(datatable *Table) *ArgLookup {
	dataTableLookup := new(ArgLookup)
	if !datatable.isInitialized() {
		return dataTableLookup
	}
	for _, header := range datatable.headers {
		dataTableLookup.addArgName(header)
	}
	return dataTableLookup
}

func (warning *Warning) String() string {
	return fmt.Sprintf("line no: %d, %s", warning.lineNo, warning.message)
}

func (scenario Scenario) kind() TokenKind {
	return scenarioKind
}

func (scenario *Scenario) addHeading(heading *Heading) {
	heading.headingType = scenarioHeading
	scenario.heading = heading
}

func (scenario *Scenario) addStep(step *Step) {
	scenario.steps = append(scenario.steps, step)
	scenario.addItem(step)
}

func (scenario *Scenario) addTags(tags *Tags) {
	scenario.tags = tags
	scenario.addItem(tags)
}

func (scenario *Scenario) addComment(comment *Comment) {
	scenario.comments = append(scenario.comments, comment)
	scenario.addItem(comment)
}

func (scenario *Scenario) renameSteps(oldStep Step, newStep Step, orderMap map[int]int) bool {
	isRefactored := false
	for _, step := range scenario.steps {
		isConcept := false
		isRefactored = step.rename(oldStep, newStep, isRefactored, orderMap, &isConcept)
	}
	return isRefactored
}

func (scenario *Scenario) addItem(itemToAdd Item) {
	if scenario.items == nil {
		scenario.items = make([]Item, 0)
	}
	scenario.items = append(scenario.items, itemToAdd)
}

func (scenario *Scenario) latestStep() *Step {
	return scenario.steps[len(scenario.steps)-1]
}

func (heading *Heading) kind() TokenKind {
	return headingKind
}

func (comment *Comment) kind() TokenKind {
	return commentKind
}

func (tags *Tags) kind() TokenKind {
	return tagKind
}

func (step Step) kind() TokenKind {
	return stepKind
}

func (specification *Specification) getSpecItems() []Item {
	specItems := make([]Item, 0)
	for _, item := range specification.items {
		if item.kind() != scenarioKind {
			specItems = append(specItems, item)
		}
	}
	return specItems
}

// Not copying parent as it enters an infinite loop in case of nested concepts. This is because the steps under the concept
// are copied and their parent copying again comes back to copy the same concept.
func (self *Step) getCopy() *Step {
	if !self.isConcept {
		return self
	}
	nestedStepsCopy := make([]*Step, 0)
	for _, nestedStep := range self.conceptSteps {
		nestedStepsCopy = append(nestedStepsCopy, nestedStep.getCopy())
	}

	copiedConceptStep := new(Step)
	*copiedConceptStep = *self
	copiedConceptStep.conceptSteps = nestedStepsCopy
	copiedConceptStep.lookup = *self.lookup.getCopy()
	return copiedConceptStep
}

func (self *Step) copyFrom(another *Step) {
	self.isConcept = another.isConcept

	if another.args == nil {
		self.args = nil
	} else {
		self.args = make([]*StepArg, len(another.args))
		copy(self.args, another.args)
	}

	if another.conceptSteps == nil {
		self.conceptSteps = nil
	} else {
		self.conceptSteps = make([]*Step, len(another.conceptSteps))
		copy(self.conceptSteps, another.conceptSteps)
	}

	if another.fragments == nil {
		self.fragments = nil
	} else {
		self.fragments = make([]*gauge_messages.Fragment, len(another.fragments))
		copy(self.fragments, another.fragments)
	}

	self.lineNo = another.lineNo
	self.lineText = another.lineText
	self.hasInlineTable = another.hasInlineTable
	self.value = another.value
	self.lookup = another.lookup
	self.parent = another.parent
}

func convertToStepText(fragments []*gauge_messages.Fragment) string {
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
	return fmt.Sprintf("[ParseError] %s : %s", result.fileName, result.error.Error())
}
