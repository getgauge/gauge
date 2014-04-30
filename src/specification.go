package main

import (
	"fmt"
	"regexp"
	"strings"
)

type scenario struct {
	heading line
	steps   []*step
	tags    []string
}

type argType int

const (
	static        argType = iota
	dynamic       argType = iota
	tableArg      argType = iota
	specialString argType = iota
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
}

type specification struct {
	heading   line
	scenarios []*scenario
	comments  []*line
	dataTable table
	contexts  []*step
	fileName  string
	tags      []string
}

type line struct {
	value  string
	lineNo int
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
	converters := specParser.initalizeConverters()
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
	//todo move resolution of concepts after processing table headers since it modifies step value
	validationError := specParser.validateSpec(specification)
	if validationError != nil {
		finalResult.ok = false
		finalResult.error = validationError
		return nil, finalResult
	}
	finalResult.ok = true
	return specification, finalResult
}

func (specParser *specParser) initalizeConverters() []func(*token, *int, *specification) parseResult {
	specConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == specKind
	}, func(token *token, spec *specification, state *int) parseResult {
		if spec.heading.value != "" {
			return parseResult{ok: false, error: &parseError{token.lineNo, "Parse error: Multiple spec headings found in same file", token.lineText}}
		}
		spec.heading = line{token.value, token.lineNo}
		addStates(state, specScope)
		return parseResult{ok: true}
	})

	scenarioConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == scenarioKind
	}, func(token *token, spec *specification, state *int) parseResult {
		if spec.heading.value == "" {
			return parseResult{ok: false, error: &parseError{token.lineNo, "Parse error: Scenario should be defined after the spec heading", token.lineText}}
		}
		scenarioHeading := line{token.value, token.lineNo}
		scenario := &scenario{heading: scenarioHeading}
		spec.scenarios = append(spec.scenarios, scenario)
		retainStates(state, specScope)
		addStates(state, scenarioScope)
		return parseResult{ok: true}
	})

	stepConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == stepKind && isInState(*state, scenarioScope)
	}, func(token *token, spec *specification, state *int) parseResult {
		latestScenario := spec.scenarios[len(spec.scenarios)-1]
		err := spec.addStep(token, &latestScenario.steps, specParser.conceptDictionary)
		if err != nil {
			return parseResult{error: err, ok: false}
		}
		retainStates(state, specScope, scenarioScope)
		addStates(state, stepScope)
		return parseResult{ok: true}
	})

	contextConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == stepKind && !isInState(*state, scenarioScope) && isInState(*state, specScope)
	}, func(token *token, spec *specification, state *int) parseResult {
		err := spec.addStep(token, &spec.contexts, specParser.conceptDictionary)
		if err != nil {
			return parseResult{error: err, ok: false}
		}
		retainStates(state, specScope)
		addStates(state, contextScope)
		return parseResult{ok: true}
	})

	commentConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == commentKind
	}, func(token *token, spec *specification, state *int) parseResult {
		commentLine := &line{token.value, token.lineNo}
		spec.comments = append(spec.comments, commentLine)
		retainStates(state, specScope, scenarioScope)
		addStates(state, commentScope)
		return parseResult{ok: true}
	})

	tableHeaderConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == tableHeader && isInState(*state, specScope)
	}, func(token *token, spec *specification, state *int) parseResult {
		if isInState(*state, stepScope) {
			latestScenario := spec.scenarios[len(spec.scenarios)-1]
			latestStep := latestScenario.steps[len(latestScenario.steps)-1]
			addInlineTableHeader(latestStep, token)
		} else if isInState(*state, contextScope) {
			latestContext := spec.contexts[len(spec.contexts)-1]
			addInlineTableHeader(latestContext, token)
		} else if !isInState(*state, scenarioScope) {
			if !spec.dataTable.isInitialized() {
				spec.dataTable.lineNo = token.lineNo
				spec.dataTable.addHeaders(token.args)
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
			latestScenario := spec.scenarios[len(spec.scenarios)-1]
			latestStep := latestScenario.steps[len(latestScenario.steps)-1]
			result = addInlineTableRow(latestStep, token, new(argLookup).fromDataTable(&spec.dataTable))
		} else if isInState(*state, contextScope) {
			latestContext := spec.contexts[len(spec.contexts)-1]
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
		if isInState(*state, scenarioScope) {
			latestScenario := spec.scenarios[len(spec.scenarios)-1]
			latestScenario.tags = token.args
		} else {
			spec.tags = token.args
		}
		return parseResult{ok: true}
	})

	converter := []func(*token, *int, *specification) parseResult{
		specConverter, scenarioConverter, stepConverter, contextConverter, commentConverter, tableHeaderConverter, tableRowConverter, tagConverter,
	}

	return converter
}

func (spec *specification) addStep(stepToken *token, addTo *[]*step, conceptDictionary *conceptDictionary) *parseError {
	var stepToAdd *step
	var err *parseError
	stepValue, _ := spec.extractStepValueAndParameterTypes(stepToken.value)
	if conceptFromDictionary := conceptDictionary.search(stepValue); conceptFromDictionary != nil {
		stepToAdd, err = spec.createConceptStep(conceptFromDictionary.conceptStep, stepToken)
	} else {
		dataTableLookup := new(argLookup).fromDataTable(&spec.dataTable)
		stepToAdd, err = spec.createStep(stepToken, dataTableLookup)
	}
	if err != nil {
		return err
	}
	*addTo = append(*addTo, stepToAdd)
	return nil
}

func (spec *specification) createConceptStep(conceptFromDictionary *step, stepToken *token) (*step, *parseError) {
	lookup := conceptFromDictionary.lookup.getCopy()
	conceptStep, err := spec.createStep(stepToken, nil)
	conceptStep.isConcept = true
	if err != nil {
		return nil, err
	}
	conceptStep.conceptSteps = conceptFromDictionary.conceptSteps
	spec.populateConceptLookup(lookup, conceptFromDictionary.args, conceptStep.args)
	conceptStep.lookup = *lookup
	return conceptStep, nil
}

func (spec *specification) createStep(stepToken *token, lookup *argLookup) (*step, *parseError) {
	stepValue, argsType := spec.extractStepValueAndParameterTypes(stepToken.value)
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
	return step, nil
}

func (specParser *specParser) validateSpec(specification *specification) *parseError {
	dataTable := specification.dataTable
	if dataTable.isInitialized() && dataTable.getRowCount() == 0 {
		return &parseError{lineNo: dataTable.lineNo, message: "Data table should have at least 1 data row"}
	}
	return nil
}

func (spec *specification) extractStepValueAndParameterTypes(stepTokenValue string) (string, []string) {
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
	return r.ReplaceAllString(stepTokenValue, "{}"), argsType
}

func (spec *specification) populateConceptLookup(lookup *argLookup, conceptArgs []*stepArg, stepArgs []*stepArg) {
	for i, arg := range stepArgs {
		lookup.addArgValue(conceptArgs[i].value, &stepArg{value: arg.value, argType: arg.argType, table: arg.table})
	}
}

func (spec *specification) createStepArg(argValue string, typeOfArg string, token *token, lookup *argLookup) (*stepArg, *parseError) {
	var stepArgument *stepArg
	if typeOfArg == "special" {
		return new(specialTypeResolver).resolve(argValue), nil
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
	step.value = fmt.Sprintf("%s {}", step.value)
}

func addInlineTableRow(step *step, token *token, argLookup *argLookup) parseResult {
	dynamicArgMatcher := regexp.MustCompile("<(.*)>")
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

type specialTypeResolver struct {
}

func (resolver *specialTypeResolver) resolve(value string) *stepArg {
	return &stepArg{argType: specialString, value: ""}
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
