package main

import (
	"fmt"
	"regexp"
	"strings"
)

type scenario struct {
	heading line
	steps   []*step
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

type specwarning struct {
	value string
}

type result struct {
	error   *specerror
	warning *specwarning
	ok      bool
}

type cumulativeResult struct {
	error    *specerror
	warnings []*specwarning
	ok       bool
}

func converterFn(predicate func(token token, state *int) bool, apply func(token token, spec *specification, state *int) result) func(token, *int, *specification) result {

	return func(token token, state *int, spec *specification) result {
		if !predicate(token, state) {
			return result{ok: true}
		}
		return apply(token, spec, state)
	}

}

func Specification(tokens []*token) (*specification, cumulativeResult) {

	converters := initalizeConverters()
	specification := &specification{}
	cumulativeResult := &cumulativeResult{}
	state := initial

	for _, token := range tokens {
		for _, converter := range converters {
			result := converter(*token, &state, specification)
			if !result.ok {
				if result.error != nil {
					cumulativeResult.ok = false
					cumulativeResult.error = result.error
					return nil, *cumulativeResult
				}
				if result.warning != nil {
					cumulativeResult.warnings = append(cumulativeResult.warnings, result.warning)
				}
			}
		}
	}
	cumulativeResult.ok = true
	return specification, *cumulativeResult
}

func initalizeConverters() []func(token, *int, *specification) result {
	specConverter := converterFn(func(token token, state *int) bool {
		return token.kind == specKind
	}, func(token token, spec *specification, state *int) result {
		spec.heading = line{token.value, token.lineNo}
		addStates(state, specScope)
		return result{ok: true}
	})

	scenarioConverter := converterFn(func(token token, state *int) bool {
		return token.kind == scenarioKind
	}, func(token token, spec *specification, state *int) result {
		scenarioHeading := line{token.value, token.lineNo}
		scenario := &scenario{heading: scenarioHeading}
		spec.scenarios = append(spec.scenarios, scenario)
		retainStates(state, specScope)
		addStates(state, scenarioScope)
		return result{ok: true}
	})

	stepConverter := converterFn(func(token token, state *int) bool {
		return token.kind == stepKind
	}, func(token token, spec *specification, state *int) result {
		latestScenario := spec.scenarios[len(spec.scenarios)-1]
		currentStep, err := createStep(token)
		if err != nil {
			return result{error: err}
		}
		latestScenario.steps = append(latestScenario.steps, currentStep)
		retainStates(state, specScope, scenarioScope)
		addStates(state, stepScope)
		return result{ok: true}
	})

	contextConverter := converterFn(func(token token, state *int) bool {
		return token.kind == context
	}, func(token token, spec *specification, state *int) result {
		context, err := createStep(token)
		if err != nil {
			return result{error: err}
		}
		spec.contexts = append(spec.contexts, context)
		retainStates(state, specScope)
		addStates(state, contextScope)
		return result{ok: true}
	})

	commentConverter := converterFn(func(token token, state *int) bool {
		return token.kind == commentKind
	}, func(token token, spec *specification, state *int) result {
		commentLine := &line{token.value, token.lineNo}
		spec.comments = append(spec.comments, commentLine)
		retainStates(state, specScope, scenarioScope)
		addStates(state, commentScope)
		return result{ok: true}
	})

	tableHeaderConverter := converterFn(func(token token, state *int) bool {
		return token.kind == tableHeader && isInState(*state, specScope)
	}, func(token token, spec *specification, state *int) result {
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
				return result{ok: false, warning: &specwarning{value}}
			}
		} else {
			value := fmt.Sprintf("table not associated with a step, ignoring table at line no: %d", token.lineNo)
			return result{ok: false, warning: &specwarning{value}}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope)
		addStates(state, tableScope)
		return result{ok: true}
	})

	tableRowConverter := converterFn(func(token token, state *int) bool {
		return token.kind == tableRow && isInState(*state, tableScope)
	}, func(token token, spec *specification, state *int) result {
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
		return result{ok: true}
	})

	converter := []func(token, *int, *specification) result{
		specConverter, scenarioConverter, stepConverter, contextConverter, commentConverter, tableHeaderConverter, tableRowConverter,
	}

	return converter
}

func createStep(stepToken token) (*step, *specerror) {
	step := &step{lineNo: stepToken.lineNo, value: stepToken.value, lineText: strings.TrimSpace(stepToken.lineText)}
	r := regexp.MustCompile("{(dynamic|static|special)}")
	args := r.FindAllStringSubmatch(stepToken.value, -1)
	if args == nil {
		return step, nil
	}
	if len(args) != len(stepToken.args) {
		return nil, &specerror{fmt.Sprintf("Step text should not have '{static}' or '{dynamic}' or '{special}' on line: %d", stepToken.lineNo)}
	}
	step.value = r.ReplaceAllString(step.value, "{}")
	for i, arg := range args {
		step.args = append(step.args, createStepArg(stepToken.args[i], arg[1]))
	}
	return step, nil
}

func createStepArg(argValue string, typeOfArg string) *stepArg {
	var stepArgument *stepArg
	if typeOfArg == "special" {
		stepArgument = new(specialTypeResolver).resolve(argValue)
	} else if typeOfArg == "static" {
		stepArgument = &stepArg{argType: static, value: argValue}
	} else {
		stepArgument = &stepArg{argType: dynamic, value: argValue}
		//todo check dynamic param is resolvable
	}

	return stepArgument
}

type specialTypeResolver struct {
}

func (resolver *specialTypeResolver) resolve(value string) *stepArg {
	return &stepArg{argType: specialString, value: ""}
}
