package main

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	"regexp"
	"strings"
)

type scenario struct {
	heading  *heading
	steps    []*step
	comments []*comment
	tags     *tags
	items    []item
}

type argType string

const (
	static                argType = "static"
	dynamic               argType = "dynamic"
	tableArg              argType = "table"
	specialString         argType = "special_string"
	specialTable          argType = "special_table"
	PARAMETER_PLACEHOLDER         = "{}"
)

type stepArg struct {
	name    string
	value   string
	argType argType
	table   table
}

func (stepArg *stepArg) String() string {
	return fmt.Sprintf("{Name: %s,value %s,argType %s,table %v}", stepArg.name, stepArg.value, string(stepArg.argType), stepArg.table)
}

type paramNameValue struct {
	name    string
	stepArg *stepArg
}

func (paramNameValue paramNameValue) String() string {
	return fmt.Sprintf("ParamName: %s, stepArg: %s", paramNameValue.name, paramNameValue.stepArg)
}

type argLookup struct {
	//helps to access the index of an arg at O(1)
	paramIndexMap map[string]int
	paramValue    []paramNameValue
}

func (argLookup argLookup) String() string {
	return fmt.Sprintln(argLookup.paramValue)
}

type step struct {
	lineNo         int
	value          string
	lineText       string
	args           []*stepArg
	isConcept      bool
	lookup         argLookup
	conceptSteps   []*step
	fragments      []*Fragment
	parent         *step
	hasInlineTable bool
	items          []item
	preComments    []*comment
}

func (step *step) getArg(name string) *stepArg {
	if step.parent == nil {
		return step.lookup.getArg(name)
	}
	return step.parent.getArg(step.lookup.getArg(name).value)
}

func (step *step) rename(oldStep step, newStep step, isRefactored bool, orderMap map[int]ArgPosition) bool {
	if strings.TrimSpace(step.value) != strings.TrimSpace(oldStep.value) {
		return isRefactored
	}
	step.value = newStep.value
	args := make([]*stepArg, len(newStep.args))
	for key, value := range orderMap {
		if value.index == -1 {
			args[key] = &stepArg{value: "", argType: static}
		} else {
			args[key] = step.args[value.index]
		}
	}
	step.args = args
	return true
}

func (step *step) deepCopyStepArgs() []*stepArg {
	copiedStepArgs := make([]*stepArg, 0)
	for _, conceptStepArg := range step.args {
		temp := new(stepArg)
		*temp = *conceptStepArg
		copiedStepArgs = append(copiedStepArgs, temp)
	}
	return copiedStepArgs
}

func createStepFromStepRequest(stepReq *ExecuteStepRequest) *step {
	args := createStepArgsFromProtoArguments(stepReq.GetParameters())
	step := &step{value: stepReq.GetParsedStepText(),
		lineText: stepReq.GetActualStepText()}
	step.addArgs(args...)
	return step
}

func createStepArgsFromProtoArguments(parameters []*Parameter) []*stepArg {
	stepArgs := make([]*stepArg, 0)
	for _, parameter := range parameters {
		var arg *stepArg
		switch parameter.GetParameterType() {
		case Parameter_Static:
			arg = &stepArg{argType: static, value: parameter.GetValue(), name: parameter.GetName()}
			break
		case Parameter_Dynamic:
			arg = &stepArg{argType: dynamic, value: parameter.GetValue(), name: parameter.GetName()}
			break
		case Parameter_Special_String:
			arg = &stepArg{argType: specialString, value: parameter.GetValue(), name: parameter.GetName()}
			break
		case Parameter_Table:
			arg = &stepArg{argType: tableArg, table: *(tableFrom(parameter.GetTable())), name: parameter.GetName()}
			break
		case Parameter_Special_Table:
			arg = &stepArg{argType: specialTable, table: *(tableFrom(parameter.GetTable())), name: parameter.GetName()}
			break
		}
		stepArgs = append(stepArgs, arg)
	}
	return stepArgs
}

type specification struct {
	heading   *heading
	scenarios []*scenario
	comments  []*comment
	dataTable table
	contexts  []*step
	fileName  string
	tags      *tags
	items     []item
}

type item interface {
	kind() tokenKind
}

type specFilter interface {
	filter(item) bool
}

type headingType int

const (
	specHeading     = 0
	scenarioHeading = 1
)

type heading struct {
	value       string
	lineNo      int
	headingType headingType
}

type comment struct {
	value  string
	lineNo int
}

type tags struct {
	values []string
}

type warning struct {
	message string
	lineNo  int
}

type parseResult struct {
	error    *parseError
	warnings []*warning
	ok       bool
	fileName string
}

func converterFn(predicate func(token *token, state *int) bool, apply func(token *token, spec *specification, state *int) parseResult) func(*token, *int, *specification) parseResult {

	return func(token *token, state *int, spec *specification) parseResult {
		if !predicate(token, state) {
			return parseResult{ok: true}
		}
		return apply(token, spec, state)
	}

}

func (specParser *specParser) createSpecification(tokens []*token, conceptDictionary *conceptDictionary) (*specification, *parseResult) {
	specParser.conceptDictionary = conceptDictionary
	converters := specParser.initializeConverters()
	specification := &specification{}
	finalResult := &parseResult{}
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
				if result.warnings != nil {
					if finalResult.warnings == nil {
						finalResult.warnings = make([]*warning, 0)
					}
					finalResult.warnings = append(finalResult.warnings, result.warnings...)
				}
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

func (specParser *specParser) initializeConverters() []func(*token, *int, *specification) parseResult {
	specConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == specKind
	}, func(token *token, spec *specification, state *int) parseResult {
		if spec.heading != nil {
			return parseResult{ok: false, error: &parseError{token.lineNo, "Parse error: Multiple spec headings found in same file", token.lineText}}
		}

		spec.addHeading(&heading{lineNo: token.lineNo, value: token.value})
		addStates(state, specScope)
		return parseResult{ok: true}
	})

	scenarioConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == scenarioKind
	}, func(token *token, spec *specification, state *int) parseResult {
		if spec.heading == nil {
			return parseResult{ok: false, error: &parseError{token.lineNo, "Parse error: Scenario should be defined after the spec heading", token.lineText}}
		}
		for _, scenario := range spec.scenarios {
			if strings.ToLower(scenario.heading.value) == strings.ToLower(token.value) {
				return parseResult{ok: false, error: &parseError{token.lineNo, "Parse error: Duplicate scenario definitions are not allowed in the same specification", token.lineText}}
			}
		}
		scenario := &scenario{}
		scenario.addHeading(&heading{value: token.value, lineNo: token.lineNo})
		spec.addScenario(scenario)

		retainStates(state, specScope)
		addStates(state, scenarioScope)
		return parseResult{ok: true}
	})

	stepConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == stepKind && isInState(*state, scenarioScope)
	}, func(token *token, spec *specification, state *int) parseResult {
		latestScenario := spec.latestScenario()
		stepToAdd, err := spec.createStep(token)
		if err != nil {
			return parseResult{error: err, ok: false}
		}
		latestScenario.addStep(stepToAdd)
		retainStates(state, specScope, scenarioScope)
		addStates(state, stepScope)
		return parseResult{ok: true}
	})

	contextConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == stepKind && !isInState(*state, scenarioScope) && isInState(*state, specScope)
	}, func(token *token, spec *specification, state *int) parseResult {
		stepToAdd, err := spec.createStep(token)
		if err != nil {
			return parseResult{error: err, ok: false}
		}
		spec.addContext(stepToAdd)
		retainStates(state, specScope)
		addStates(state, contextScope)
		return parseResult{ok: true}
	})

	commentConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == commentKind
	}, func(token *token, spec *specification, state *int) parseResult {
		comment := &comment{token.value, token.lineNo}
		if isInState(*state, scenarioScope) {
			spec.latestScenario().addComment(comment)
		} else {
			spec.addComment(comment)
		}
		retainStates(state, specScope, scenarioScope)
		addStates(state, commentScope)
		return parseResult{ok: true}
	})

	tableHeaderConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == tableHeader && isInState(*state, specScope)
	}, func(token *token, spec *specification, state *int) parseResult {
		if isInState(*state, stepScope) {
			latestScenario := spec.latestScenario()
			latestStep := latestScenario.latestStep()
			addInlineTableHeader(latestStep, token)
		} else if isInState(*state, contextScope) {
			latestContext := spec.latestContext()
			addInlineTableHeader(latestContext, token)
		} else if !isInState(*state, scenarioScope) {
			if !spec.dataTable.isInitialized() {
				dataTable := &table{}
				dataTable.lineNo = token.lineNo
				dataTable.addHeaders(token.args)
				spec.addDataTable(dataTable)
			} else {
				value := "Multiple data table present, ignoring table"
				return parseResult{ok: false, warnings: []*warning{&warning{value, token.lineNo}}}
			}
		} else {
			value := "Table not associated with a step, ignoring table"
			return parseResult{ok: false, warnings: []*warning{&warning{value, token.lineNo}}}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope)
		addStates(state, tableScope)
		return parseResult{ok: true}
	})

	tableRowConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == tableRow && isInState(*state, tableScope)
	}, func(token *token, spec *specification, state *int) parseResult {
		var result parseResult
		if isInState(*state, stepScope) {
			latestScenario := spec.latestScenario()
			latestStep := latestScenario.latestStep()
			result = addInlineTableRow(latestStep, token, new(argLookup).fromDataTable(&spec.dataTable))
		} else if isInState(*state, contextScope) {
			latestContext := spec.latestContext()
			result = addInlineTableRow(latestContext, token, new(argLookup).fromDataTable(&spec.dataTable))
		} else {
			//todo validate datatable rows also
			spec.dataTable.addRowValues(token.args)
			result = parseResult{ok: true}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tableScope)
		return result
	})

	tagConverter := converterFn(func(token *token, state *int) bool {
		return (token.kind == tagKind)
	}, func(token *token, spec *specification, state *int) parseResult {
		tags := &tags{values: token.args}
		if isInState(*state, scenarioScope) {
			spec.latestScenario().addTags(tags)
		} else {
			spec.addTags(tags)
		}
		return parseResult{ok: true}
	})

	converter := []func(*token, *int, *specification) parseResult{
		specConverter, scenarioConverter, stepConverter, contextConverter, commentConverter, tableHeaderConverter, tableRowConverter, tagConverter,
	}

	return converter
}

func (spec *specification) createStep(stepToken *token) (*step, *parseError) {
	dataTableLookup := new(argLookup).fromDataTable(&spec.dataTable)
	stepToAdd, err := spec.createStepUsingLookup(stepToken, dataTableLookup)
	if err != nil {
		return nil, err
	}
	return stepToAdd, nil
}

func (spec *specification) createStepUsingLookup(stepToken *token, lookup *argLookup) (*step, *parseError) {
	stepValue, argsType := extractStepValueAndParameterTypes(stepToken.value)
	if argsType != nil && len(argsType) != len(stepToken.args) {
		return nil, &parseError{stepToken.lineNo, "Step text should not have '{static}' or '{dynamic}' or '{special}'", stepToken.lineText}
	}
	step := &step{lineNo: stepToken.lineNo, value: stepValue, lineText: strings.TrimSpace(stepToken.lineText)}
	arguments := make([]*stepArg, 0)
	for i, argType := range argsType {
		argument, err := spec.createStepArg(stepToken.args[i], argType, stepToken, lookup)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, argument)
	}
	step.addArgs(arguments...)
	return step, nil
}

func (specification *specification) processConceptStepsFrom(conceptDictionary *conceptDictionary) {
	for _, step := range specification.contexts {
		specification.processConceptStep(step, conceptDictionary)
	}
	for _, scenario := range specification.scenarios {
		for _, step := range scenario.steps {
			specification.processConceptStep(step, conceptDictionary)
		}
	}
}

func (specification *specification) processConceptStep(step *step, conceptDictionary *conceptDictionary) {
	if conceptFromDictionary := conceptDictionary.search(step.value); conceptFromDictionary != nil {
		specification.createConceptStep(conceptFromDictionary.conceptStep, step)
	}
}

func (specification *specification) createConceptStep(concept *step, originalStep *step) {
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

func (specification *specification) addItem(itemToAdd item) {
	if specification.items == nil {
		specification.items = make([]item, 0)
	}

	specification.items = append(specification.items, itemToAdd)
}

func (specification *specification) addHeading(heading *heading) {
	heading.headingType = specHeading
	specification.heading = heading
}

func (specification *specification) addScenario(scenario *scenario) {
	specification.scenarios = append(specification.scenarios, scenario)
	specification.addItem(scenario)
}

func (specification *specification) addContext(contextStep *step) {
	specification.contexts = append(specification.contexts, contextStep)
	specification.addItem(contextStep)
}

func (specification *specification) addComment(comment *comment) {
	specification.comments = append(specification.comments, comment)
	specification.addItem(comment)
}

func (specification *specification) addDataTable(table *table) {
	specification.dataTable = *table
	specification.addItem(table)
}

func (specification *specification) addTags(tags *tags) {
	specification.tags = tags
}

func (specification *specification) latestScenario() *scenario {
	return specification.scenarios[len(specification.scenarios)-1]
}

func (specification *specification) latestContext() *step {
	return specification.contexts[len(specification.contexts)-1]
}

func (specParser *specParser) validateSpec(specification *specification) *parseError {
	if len(specification.items) == 0 {
		return &parseError{lineNo: 1, message: "Spec does not have any elements"}
	}
	if specification.heading == nil {
		return &parseError{lineNo: 1, message: "Spec heading not found"}
	}
	dataTable := specification.dataTable
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

func (step *step) addArgs(args ...*stepArg) {
	step.args = append(step.args, args...)
	step.populateFragments()
}

func (step *step) addInlineTableHeaders(headers []string) {
	tableArg := &stepArg{argType: tableArg}
	tableArg.table.addHeaders(headers)
	step.addArgs(tableArg)
}

func (step *step) addInlineTableRow(row []tableCell) {
	lastArg := step.args[len(step.args)-1]
	lastArg.table.addRows(row)
	step.populateFragments()
}

func (step *step) populateFragments() {
	r := regexp.MustCompile(PARAMETER_PLACEHOLDER)
	/*
		enter {} and {} bar
		returns
		[[6 8] [13 15]]
	*/
	argSplitIndices := r.FindAllStringSubmatchIndex(step.value, -1)
	step.fragments = make([]*Fragment, 0)
	if len(step.args) == 0 {
		step.fragments = append(step.fragments, &Fragment{FragmentType: Fragment_Text.Enum(), Text: proto.String(step.value)})
		return
	}

	textStartIndex := 0
	for argIndex, argIndices := range argSplitIndices {
		if textStartIndex < argIndices[0] {
			step.fragments = append(step.fragments, &Fragment{FragmentType: Fragment_Text.Enum(), Text: proto.String(step.value[textStartIndex:argIndices[0]])})
		}
		parameter := convertToProtoParameter(step.args[argIndex])
		step.fragments = append(step.fragments, &Fragment{FragmentType: Fragment_Parameter.Enum(), Parameter: parameter})
		textStartIndex = argIndices[1]
	}
	if textStartIndex < len(step.value) {
		step.fragments = append(step.fragments, &Fragment{FragmentType: Fragment_Text.Enum(), Text: proto.String(step.value[textStartIndex:len(step.value)])})
	}

}

func (spec *specification) filter(filter specFilter) {
	for i := 0; i < len(spec.items); i++ {
		if filter.filter(spec.items[i]) {
			spec.removeItem(i)
			i--
		}
	}
}

func (spec *specification) removeItem(itemIndex int) {
	item := spec.items[itemIndex]
	if len(spec.items)-1 == itemIndex {
		spec.items = spec.items[:itemIndex]
	} else if 0 == itemIndex {
		spec.items = spec.items[itemIndex+1:]
	} else {
		spec.items = append(spec.items[:itemIndex], spec.items[itemIndex+1:]...)
	}
	if item.kind() == scenarioKind {
		spec.removeScenario(item.(*scenario))
	}
}

func (spec *specification) removeScenario(scenario *scenario) {
	index := getIndexFor(scenario, spec.scenarios)
	if len(spec.scenarios)-1 == index {
		spec.scenarios = spec.scenarios[:index]
	} else if index == 0 {
		spec.scenarios = spec.scenarios[index+1:]
	} else {
		spec.scenarios = append(spec.scenarios[:index], spec.scenarios[index+1:]...)
	}
}

func (spec *specification) populateConceptLookup(lookup *argLookup, conceptArgs []*stepArg, stepArgs []*stepArg) {
	for i, arg := range stepArgs {
		lookup.addArgValue(conceptArgs[i].value, &stepArg{value: arg.value, argType: arg.argType, table: arg.table, name: arg.name})
	}
}

func (spec *specification) renameSteps(oldStep step, newStep step, orderMap map[int]ArgPosition) bool {
	isRefactored := false
	for _, step := range spec.contexts {
		isRefactored = step.rename(oldStep, newStep, isRefactored, orderMap)
	}
	for _, scenario := range spec.scenarios {
		refactor := scenario.renameSteps(oldStep, newStep, orderMap)
		if refactor {
			isRefactored = refactor
		}
	}
	return isRefactored
}

func (spec *specification) createStepArg(argValue string, typeOfArg string, token *token, lookup *argLookup) (*stepArg, *parseError) {
	var stepArgument *stepArg
	if typeOfArg == "special" {
		resolvedArgValue, err := newSpecialTypeResolver().resolve(argValue)
		if err != nil {
			return nil, &parseError{lineNo: token.lineNo, message: err.Error(), lineText: token.lineText}
		}
		return resolvedArgValue, nil
	} else if typeOfArg == "static" {
		return &stepArg{argType: static, value: argValue}, nil
	} else {
		if !isConceptHeader(lookup) && !lookup.containsArg(argValue) {
			return nil, &parseError{lineNo: token.lineNo, message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), lineText: token.lineText}
		}
		stepArgument = &stepArg{argType: dynamic, value: argValue, name: argValue}
		return stepArgument, nil
	}
}

//Step value is modified when inline table is found to account for the new parameter by appending {}
//todo validate headers for dynamic
func addInlineTableHeader(step *step, token *token) {
	step.value = fmt.Sprintf("%s %s", step.value, PARAMETER_PLACEHOLDER)
	step.hasInlineTable = true
	step.addInlineTableHeaders(token.args)

}

func addInlineTableRow(step *step, token *token, argLookup *argLookup) parseResult {
	dynamicArgMatcher := regexp.MustCompile("^<(.*)>$")
	tableValues := make([]tableCell, 0)
	for _, tableValue := range token.args {
		if dynamicArgMatcher.MatchString(tableValue) {
			match := dynamicArgMatcher.FindAllStringSubmatch(tableValue, -1)
			param := match[0][1]
			if !argLookup.containsArg(param) {
				return parseResult{ok: false, error: &parseError{lineNo: token.lineNo, message: fmt.Sprintf("Dynamic param <%s> could not be resolved", param), lineText: token.lineText}}
			}
			tableValues = append(tableValues, tableCell{value: param, cellType: dynamic})
		} else {
			tableValues = append(tableValues, tableCell{value: tableValue, cellType: static})
		}
	}
	step.addInlineTableRow(tableValues)
	return parseResult{ok: true}
}

//concept header will have dynamic param and should not be resolved through lookup, so passing nil lookup
func isConceptHeader(lookup *argLookup) bool {
	return lookup == nil
}

func (lookup *argLookup) addArgName(argName string) {
	if lookup.paramIndexMap == nil {
		lookup.paramIndexMap = make(map[string]int)
		lookup.paramValue = make([]paramNameValue, 0)
	}
	lookup.paramIndexMap[argName] = len(lookup.paramValue)
	lookup.paramValue = append(lookup.paramValue, paramNameValue{name: argName})
}

func (lookup *argLookup) addArgValue(param string, stepArg *stepArg) {
	paramIndex, ok := lookup.paramIndexMap[param]
	if !ok {
		panic(fmt.Sprintf("Accessing an invalid parameter (%s)", param))
	}
	lookup.paramValue[paramIndex].stepArg = stepArg
}

func (lookup *argLookup) containsArg(param string) bool {
	_, ok := lookup.paramIndexMap[param]
	return ok
}

func (lookup *argLookup) getArg(param string) *stepArg {
	paramIndex, ok := lookup.paramIndexMap[param]
	if !ok {
		panic(fmt.Sprintf("Accessing an invalid parameter (%s)", param))
	}
	return lookup.paramValue[paramIndex].stepArg
}

func (lookup *argLookup) getCopy() *argLookup {
	lookupCopy := new(argLookup)
	for key, _ := range lookup.paramIndexMap {
		lookupCopy.addArgName(key)
		arg := lookup.getArg(key)
		if arg != nil {
			lookupCopy.addArgValue(key, &stepArg{value: arg.value, argType: arg.argType, table: arg.table, name: arg.name})
		}
	}
	return lookupCopy
}

func (lookup *argLookup) fromDataTableRow(datatable *table, index int) *argLookup {
	dataTableLookup := new(argLookup)
	if !datatable.isInitialized() {
		return dataTableLookup
	}
	for _, header := range datatable.headers {
		dataTableLookup.addArgName(header)
		dataTableLookup.addArgValue(header, &stepArg{value: datatable.get(header)[index].value, argType: static})
	}
	return dataTableLookup
}

//create an empty lookup with only args to resolve dynamic params for steps
func (lookup *argLookup) fromDataTable(datatable *table) *argLookup {
	dataTableLookup := new(argLookup)
	if !datatable.isInitialized() {
		return dataTableLookup
	}
	for _, header := range datatable.headers {
		dataTableLookup.addArgName(header)
	}
	return dataTableLookup
}

func (warning *warning) String() string {
	return fmt.Sprintf("line no: %d, %s", warning.lineNo, warning.message)
}

func (scenario scenario) kind() tokenKind {
	return scenarioKind
}

func (scenario *scenario) addHeading(heading *heading) {
	heading.headingType = scenarioHeading
	scenario.heading = heading
}

func (scenario *scenario) addStep(step *step) {
	scenario.steps = append(scenario.steps, step)
	scenario.addItem(step)
}

func (scenario *scenario) addTags(tags *tags) {
	scenario.tags = tags
}

func (scenario *scenario) addComment(comment *comment) {
	scenario.comments = append(scenario.comments, comment)
	scenario.addItem(comment)
}

func (scenario *scenario) renameSteps(oldStep step, newStep step, orderMap map[int]ArgPosition) bool {
	isRefactored := false
	for _, step := range scenario.steps {
		isRefactored = step.rename(oldStep, newStep, isRefactored, orderMap)
	}
	return isRefactored
}

func (scenario *scenario) addItem(itemToAdd item) {
	if scenario.items == nil {
		scenario.items = make([]item, 0)
	}
	scenario.items = append(scenario.items, itemToAdd)
}

func (scenario *scenario) latestStep() *step {
	return scenario.steps[len(scenario.steps)-1]
}

func (heading *heading) kind() tokenKind {
	return headingKind
}

func (comment *comment) kind() tokenKind {
	return commentKind
}

func (tags *tags) kind() tokenKind {
	return tagKind
}

func (step step) kind() tokenKind {
	return stepKind
}

func (specification *specification) getSpecItems() []item {
	specItems := make([]item, 0)
	for _, item := range specification.items {
		if item.kind() != scenarioKind {
			specItems = append(specItems, item)
		}
	}
	return specItems
}

// Not copying parent as it enters an infinite loop in case of nested concepts. This is because the steps under the concept
// are copied and their parent copying again comes back to copy the same concept.
func (self *step) getCopy() *step {
	if !self.isConcept {
		return self
	}
	nestedStepsCopy := make([]*step, 0)
	for _, nestedStep := range self.conceptSteps {
		nestedStepsCopy = append(nestedStepsCopy, nestedStep.getCopy())
	}

	copiedConceptStep := new(step)
	*copiedConceptStep = *self
	copiedConceptStep.conceptSteps = nestedStepsCopy
	copiedConceptStep.lookup = *self.lookup.getCopy()
	return copiedConceptStep
}

func (self *step) copyFrom(another *step) {
	self.isConcept = another.isConcept

	if another.args == nil {
		self.args = nil
	} else {
		self.args = make([]*stepArg, len(another.args))
		copy(self.args, another.args)
	}

	if another.conceptSteps == nil {
		self.conceptSteps = nil
	} else {
		self.conceptSteps = make([]*step, len(another.conceptSteps))
		copy(self.conceptSteps, another.conceptSteps)
	}

	if another.fragments == nil {
		self.fragments = nil
	} else {
		self.fragments = make([]*Fragment, len(another.fragments))
		copy(self.fragments, another.fragments)
	}

	self.lineNo = another.lineNo
	self.lineText = another.lineText
	self.hasInlineTable = another.hasInlineTable
	self.value = another.value
	self.lookup = another.lookup
	self.parent = another.parent
}
