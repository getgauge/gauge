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
	specialString argType = iota
	specialTable  argType = iota
)

type stepArg struct {
	value   string
	argType argType
	table   table
}

type step struct {
	lineNo      int
	value       string
	lineText    string
	args        []*stepArg
	inlineTable table
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

type specerror struct {
	message string
}

func (specerror *specerror) Error() string {
	return specerror.message
}

type parseResult struct {
	error    *specerror
	warnings []string
	ok       bool
	specFile string
}

func converterFn(predicate func(token *token, state *int) bool, apply func(token *token, spec *specification, state *int) parseResult) func(*token, *int, *specification) parseResult {

	return func(token *token, state *int, spec *specification) parseResult {
		if !predicate(token, state) {
			return parseResult{ok: true}
		}
		return apply(token, spec, state)
	}

}

func (specParser *specParser) createSpecification(tokens []*token) (*specification, *parseResult) {

	converters := initalizeConverters()
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
						finalResult.warnings = make([]string, 0)
					}
					finalResult.warnings = append(finalResult.warnings, result.warnings...)
				}
			}
		}
	}
	finalResult.ok = true
	return specification, finalResult
}

func initalizeConverters() []func(*token, *int, *specification) parseResult {
	specConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == specKind
	}, func(token *token, spec *specification, state *int) parseResult {
		spec.heading = line{token.value, token.lineNo}
		addStates(state, specScope)
		return parseResult{ok: true}
	})

	scenarioConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == scenarioKind
	}, func(token *token, spec *specification, state *int) parseResult {
		scenarioHeading := line{token.value, token.lineNo}
		scenario := &scenario{heading: scenarioHeading}
		spec.scenarios = append(spec.scenarios, scenario)
		retainStates(state, specScope)
		addStates(state, scenarioScope)
		return parseResult{ok: true}
	})

	stepConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == stepKind
	}, func(token *token, spec *specification, state *int) parseResult {
		latestScenario := spec.scenarios[len(spec.scenarios)-1]
		err := spec.addStep(token, &latestScenario.steps)
		if err != nil {
			return parseResult{error: err, ok: false}
		}
		retainStates(state, specScope, scenarioScope)
		addStates(state, stepScope)
		return parseResult{ok: true}
	})

	contextConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == context
	}, func(token *token, spec *specification, state *int) parseResult {
		err := spec.addStep(token, &spec.contexts)
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
			latestStep.inlineTable.addHeaders(token.args)
		} else if isInState(*state, contextScope) {
			spec.contexts[len(spec.contexts)-1].inlineTable.addHeaders(token.args)
		} else if !isInState(*state, scenarioScope) {
			if !spec.dataTable.isInitialized() {
				spec.dataTable.addHeaders(token.args)
			} else {
				value := fmt.Sprintf("multiple data table present, ignoring table at line no: %d", token.lineNo)
				return parseResult{ok: false, warnings: []string{value}}
			}
		} else {
			value := fmt.Sprintf("table not associated with a step, ignoring table at line no: %d", token.lineNo)
			return parseResult{ok: false, warnings: []string{value}}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope)
		addStates(state, tableScope)
		return parseResult{ok: true}
	})

	tableRowConverter := converterFn(func(token *token, state *int) bool {
		return token.kind == tableRow && isInState(*state, tableScope)
	}, func(token *token, spec *specification, state *int) parseResult {
		if isInState(*state, stepScope) {
			latestScenario := spec.scenarios[len(spec.scenarios)-1]
			latestStep := latestScenario.steps[len(latestScenario.steps)-1]
			latestStep.inlineTable.addRowValues(token.args)
		} else if isInState(*state, contextScope) {
			spec.contexts[len(spec.contexts)-1].inlineTable.addRowValues(token.args)
		} else {
			spec.dataTable.addRowValues(token.args)
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tableScope)
		return parseResult{ok: true}
	})

	tagConverter := converterFn(func(token *token, state *int) bool {
		return (token.kind == specTag) || (token.kind == scenarioTag)
	}, func(token *token, spec *specification, state *int) parseResult {
		if token.kind == specTag {
			spec.tags = token.args
		} else if token.kind == scenarioTag {
			latestScenario := spec.scenarios[len(spec.scenarios)-1]
			latestScenario.tags = token.args
		}
		return parseResult{ok: true}
	})

	converter := []func(*token, *int, *specification) parseResult{
		specConverter, scenarioConverter, stepConverter, contextConverter, commentConverter, tableHeaderConverter, tableRowConverter, tagConverter,
	}

	return converter
}

func (spec *specification) addStep(stepToken *token, addTo *[]*step) *specerror {
	step := &step{lineNo: stepToken.lineNo, value: stepToken.value, lineText: strings.TrimSpace(stepToken.lineText)}
	r := regexp.MustCompile("{(dynamic|static|special)}")

	args := r.FindAllStringSubmatch(stepToken.value, -1)

	if args == nil {
		*addTo = append(*addTo, step)
		return nil
	}
	if len(args) != len(stepToken.args) {
		return &specerror{fmt.Sprintf("Step text should not have '{static}' or '{dynamic}' or '{special}' on line: %d", stepToken.lineNo)}
	}
	step.value = r.ReplaceAllString(step.value, "{}")
	for i, arg := range args {
		arg, err := spec.createStepArg(stepToken.args[i], arg[1], stepToken)
		if err != nil {
			return err
		}
		step.args = append(step.args, arg)
	}

	*addTo = append(*addTo, step)
	return nil
}

func (spec *specification) createStepArg(argValue string, typeOfArg string, token *token) (*stepArg, *specerror) {
	var stepArgument *stepArg
	if typeOfArg == "special" {
		return new(specialTypeResolver).resolve(argValue), nil
	} else if typeOfArg == "static" {
		return &stepArg{argType: static, value: argValue}, nil
	} else {
		if !spec.dataTable.isInitialized() {
			return nil, &specerror{fmt.Sprintf("No data table found for dynamic paramter <%s> : %s lineNo: %d", argValue, token.lineText, token.lineNo)}
		} else if !spec.dataTable.headerExists(argValue) {
			return nil, &specerror{fmt.Sprintf("No data table column found for dynamic paramter <%s> : %s lineNo: %d", argValue, token.lineText, token.lineNo)}
		}
		stepArgument = &stepArg{argType: dynamic, value: argValue}
		return stepArgument, nil
	}

}

type specialTypeResolver struct {
}

func (resolver *specialTypeResolver) resolve(value string) *stepArg {
	return &stepArg{argType: specialString, value: ""}
}
