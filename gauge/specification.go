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
	special               argType = "special"
	PARAMETER_PLACEHOLDER         = "{}"
)

type stepArg struct {
	value   string
	argType argType
	table   table
}

type paramNameValue struct {
	name    string
	stepArg *stepArg
}

type argLookup struct {
	paramIndexMap map[string]int
	paramValue    []paramNameValue
}

type step struct {
	lineNo       int
	value        string
	lineText     string
	args         []*stepArg
	isConcept    bool
	lookup       argLookup
	conceptSteps []*step
	fragments    []*Fragment
}

func createStepFromStepRequest(stepReq *ExecuteStepRequest) *step {
	args := createStepArgsFromProtoArguments(stepReq.Args)
	return &step{value: stepReq.GetParsedStepText(),
		lineText: stepReq.GetActualStepText(), args: args}
}

func createStepArgsFromProtoArguments(arguments []*Argument) []*stepArg {
	args := make([]*stepArg, 0)
	for _, arguments := range arguments {
		var a *stepArg
		if arguments.GetType() == "table" {
			a = &stepArg{value: arguments.GetValue(), argType: tableArg, table: *(tableFrom(arguments.GetTable()))}
		} else {
			a = &stepArg{value: arguments.GetValue(), argType: static}
		}
		args = append(args, a)
	}
	return args
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
	var argument *stepArg
	var err *parseError
	for i, argType := range argsType {
		argument, err = spec.createStepArg(stepToken.args[i], argType, stepToken, lookup)
		if err != nil {
			return nil, err
		}
		step.args = append(step.args, argument)
	}
	step.populateFragments()
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

func (specification *specification) createConceptStep(conceptFromDictionary *step, originalStep *step) {
	lookup := conceptFromDictionary.lookup.getCopy()
	originalStep.isConcept = true
	originalStep.conceptSteps = conceptFromDictionary.conceptSteps
	specification.populateConceptLookup(lookup, conceptFromDictionary.args, originalStep.args)
	originalStep.lookup = *lookup
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
	specification.addItem(tags)
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

func (step *step) populateFragments() {
	r := regexp.MustCompile(PARAMETER_PLACEHOLDER)
	/*
		enter {} and {} bar
		returns
		[[6 8] [13 15]]
	*/
	argSplitIndices := r.FindAllStringSubmatchIndex(step.value, -1)
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

func (spec *specification) populateConceptLookup(lookup *argLookup, conceptArgs []*stepArg, stepArgs []*stepArg) {
	for i, arg := range stepArgs {
		lookup.addArgValue(conceptArgs[i].value, &stepArg{value: arg.value, argType: arg.argType, table: arg.table})
	}
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
		stepArgument = &stepArg{argType: dynamic, value: argValue}
		return stepArgument, nil
	}
}

//Step value is modified when inline table is found to account for the new parameter by appending {}
//todo validate headers for dynamic
func addInlineTableHeader(step *step, token *token) {
	tableArg := &stepArg{argType: tableArg}
	tableArg.table.addHeaders(token.args)
	step.args = append(step.args, tableArg)
	step.value = fmt.Sprintf("%s %s", step.value, PARAMETER_PLACEHOLDER)
}

func addInlineTableRow(step *step, token *token, argLookup *argLookup) parseResult {
	dynamicArgMatcher := regexp.MustCompile("^<(.*)>$")
	tableArg := step.args[len(step.args)-1]
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
	tableArg.table.addRows(tableValues)
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
			lookupCopy.addArgValue(key, &stepArg{value: arg.value, argType: arg.argType, table: arg.table})
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
	scenario.addItem(tags)
}

func (scenario *scenario) addComment(comment *comment) {
	scenario.comments = append(scenario.comments, comment)
	scenario.addItem(comment)
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
