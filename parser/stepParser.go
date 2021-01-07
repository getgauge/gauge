/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/getgauge/gauge/logger"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/gauge"
)

const (
	inDefault      = 1 << iota
	inQuotes       = 1 << iota
	inEscape       = 1 << iota
	inDynamicParam = 1 << iota
	inSpecialParam = 1 << iota
)
const (
	quotes                 = '"'
	escape                 = '\\'
	dynamicParamStart      = '<'
	dynamicParamEnd        = '>'
	specialParamIdentifier = ':'
)

type acceptFn func(rune, int) (int, bool)

// ExtractStepArgsFromToken extracts step args(Static and Dynamic) from the given step token.
func ExtractStepArgsFromToken(stepToken *Token) ([]gauge.StepArg, error) {
	_, argsType := extractStepValueAndParameterTypes(stepToken.Value)
	if argsType != nil && len(argsType) != len(stepToken.Args) {
		return nil, fmt.Errorf("Step text should not have '{static}' or '{dynamic}' or '{special}'")
	}
	var args []gauge.StepArg
	for i, argType := range argsType {
		if gauge.ArgType(argType) == gauge.Static {
			args = append(args, gauge.StepArg{ArgType: gauge.Static, Value: stepToken.Args[i]})
		} else {
			args = append(args, gauge.StepArg{ArgType: gauge.Dynamic, Value: stepToken.Args[i]})
		}
	}
	return args, nil
}

func acceptor(start rune, end rune, onEachChar func(rune, int) int, after func(state int), inState int) acceptFn {

	return func(element rune, currentState int) (int, bool) {
		currentState = onEachChar(element, currentState)
		if element == start {
			if currentState == inDefault {
				return inState, true
			}
		}
		if element == end {
			if currentState&inState != 0 {
				after(currentState)
				return inDefault, true
			}
		}
		return currentState, false
	}
}

func simpleAcceptor(start rune, end rune, after func(int), inState int) acceptFn {
	onEach := func(currentChar rune, state int) int {
		return state
	}
	return acceptor(start, end, onEach, after, inState)
}

func processStep(parser *SpecParser, token *Token) ([]error, bool) {
	if len(token.Value) == 0 {
		return []error{fmt.Errorf("Step should not be blank")}, true
	}

	stepValue, args, err := processStepText(token.Value)
	if err != nil {
		return []error{err}, true
	}

	token.Value = stepValue
	token.Args = args
	parser.clearState()
	return []error{}, false
}

func processStepText(text string) (string, []string, error) {
	reservedChars := map[rune]struct{}{'{': {}, '}': {}}
	var stepValue, argText bytes.Buffer

	var args []string

	curBuffer := func(state int) *bytes.Buffer {
		if isInAnyState(state, inQuotes, inDynamicParam) {
			return &argText
		}
		return &stepValue
	}

	currentState := inDefault
	lastState := -1

	acceptStaticParam := simpleAcceptor(rune(quotes), rune(quotes), func(int) {
		_, err := stepValue.WriteString("{static}")
		if err != nil {
			logger.Errorf(false, "Unable to write `{static}` to step value while parsing : %s", err.Error())
		}
		args = append(args, argText.String())
		argText.Reset()
	}, inQuotes)

	acceptSpecialDynamicParam := acceptor(rune(dynamicParamStart), rune(dynamicParamEnd), func(currentChar rune, state int) int {
		if currentChar == specialParamIdentifier && state == inDynamicParam {
			return state | inSpecialParam
		}
		return state
	}, func(currentState int) {
		if isInState(currentState, inSpecialParam) {
			_, err := stepValue.WriteString("{special}")
			if err != nil {
				logger.Errorf(false, "Unable to write `{special}` to step value while parsing : %s", err.Error())
			}
		} else {
			_, err := stepValue.WriteString("{dynamic}")
			if err != nil {
				logger.Errorf(false, "Unable to write `{special}` to step value while parsing : %s", err.Error())
			}
		}
		args = append(args, argText.String())
		argText.Reset()
	}, inDynamicParam)

	var inParamBoundary bool
	for _, element := range text {
		if currentState == inEscape {
			currentState = lastState
			if _, isReservedChar := reservedChars[element]; currentState == inDefault && !isReservedChar {
				_, err := curBuffer(currentState).WriteRune(escape)
				if err != nil {
					logger.Errorf(false, "Unable to write `\\\\`(escape) to step value while parsing : %s", err.Error())
				}
			} else {
				element = getEscapedRuneIfValid(element)
			}
		} else if element == escape {
			lastState = currentState
			currentState = inEscape
			continue
		} else if currentState, inParamBoundary = acceptSpecialDynamicParam(element, currentState); inParamBoundary {
			continue
		} else if currentState, inParamBoundary = acceptStaticParam(element, currentState); inParamBoundary {
			continue
		} else if _, isReservedChar := reservedChars[element]; currentState == inDefault && isReservedChar {
			return "", nil, fmt.Errorf("'%c' is a reserved character and should be escaped", element)
		}

		_, err := curBuffer(currentState).WriteRune(element)
		if err != nil {
			logger.Errorf(false, "Unable to write `%c` to step value while parsing : %s", element, err.Error())
		}
	}

	// If it is a valid step, the state should be default when the control reaches here
	if currentState == inQuotes {
		return "", nil, fmt.Errorf("String not terminated")
	} else if isInState(currentState, inDynamicParam) {
		return "", nil, fmt.Errorf("Dynamic parameter not terminated")
	}

	return strings.TrimSpace(stepValue.String()), args, nil

}

func getEscapedRuneIfValid(element rune) rune {
	allEscapeChars := map[string]rune{"t": '\t', "n": '\n'}
	elementToStr, err := strconv.Unquote(strconv.QuoteRune(element))
	if err != nil {
		return element
	}
	for key, val := range allEscapeChars {
		if key == elementToStr {
			return val
		}
	}
	return element
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
	return r.ReplaceAllString(stepTokenValue, gauge.ParameterPlaceholder), argsType
}

func createStepArg(argValue string, typeOfArg string, token *Token, lookup *gauge.ArgLookup, fileName string) (*gauge.StepArg, *ParseResult) {
	switch typeOfArg {
	case "special":
		resolvedArgValue, err := newSpecialTypeResolver().resolve(argValue)
		if err != nil {
			switch err.(type) {
			case invalidSpecialParamError:
				return treatArgAsDynamic(argValue, token, lookup, fileName)
			default:
				return &gauge.StepArg{ArgType: gauge.Dynamic, Value: argValue, Name: argValue}, &ParseResult{ParseErrors: []ParseError{ParseError{FileName: fileName, LineNo: token.LineNo, SpanEnd: token.SpanEnd, Message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), LineText: token.LineText()}}}
			}
		}
		return resolvedArgValue, nil
	case "static":
		return &gauge.StepArg{ArgType: gauge.Static, Value: argValue}, nil
	default:
		return validateDynamicArg(argValue, token, lookup, fileName)
	}
}

func treatArgAsDynamic(argValue string, token *Token, lookup *gauge.ArgLookup, fileName string) (*gauge.StepArg, *ParseResult) {
	parseRes := &ParseResult{Warnings: []*Warning{&Warning{FileName: fileName, LineNo: token.LineNo, Message: fmt.Sprintf("Could not resolve special param type <%s>. Treating it as dynamic param.", argValue)}}}
	stepArg, result := validateDynamicArg(argValue, token, lookup, fileName)
	if result != nil {
		if len(result.ParseErrors) > 0 {
			parseRes.ParseErrors = result.ParseErrors
		}
		if result.Warnings != nil {
			parseRes.Warnings = append(parseRes.Warnings, result.Warnings...)
		}
	}
	return stepArg, parseRes
}

func validateDynamicArg(argValue string, token *Token, lookup *gauge.ArgLookup, fileName string) (*gauge.StepArg, *ParseResult) {
	stepArgument := &gauge.StepArg{ArgType: gauge.Dynamic, Value: argValue, Name: argValue}
	if !isConceptHeader(lookup) && !lookup.ContainsArg(argValue) {
		return stepArgument, &ParseResult{ParseErrors: []ParseError{ParseError{FileName: fileName, LineNo: token.LineNo, SpanEnd: token.SpanEnd, Message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), LineText: token.LineText()}}}
	}

	return stepArgument, nil
}

// ConvertToStepText accumulates fragments of a step, (ex. parameters) and returns the step text
// used to generate the annotation text in a step implementation
func ConvertToStepText(fragments []*gauge_messages.Fragment) string {
	stepText := ""
	var i int
	for _, fragment := range fragments {
		value := ""
		if fragment.GetFragmentType() == gauge_messages.Fragment_Text {
			value = fragment.GetText()
		} else {
			switch fragment.GetParameter().GetParameterType() {
			case gauge_messages.Parameter_Static:
				value = fmt.Sprintf("\"%s\"", fragment.GetParameter().GetValue())
			case gauge_messages.Parameter_Dynamic:
				value = fmt.Sprintf("<%s>", fragment.GetParameter().GetValue())
			case gauge_messages.Parameter_Special_String:
				i++
				value = fmt.Sprintf("<%s%d>", "file", i)
			case gauge_messages.Parameter_Special_Table:
				i++
				value = fmt.Sprintf("<%s%d>", "table", i)
			}
		}
		stepText += value
	}
	return stepText
}
