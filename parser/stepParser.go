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
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
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

var allEscapeChars = map[string]string{`\t`: "\t", `\n`: "\n", `\r`: "\r"}

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
		stepValue.WriteString("{static}")
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
			stepValue.WriteString("{special}")
		} else {
			stepValue.WriteString("{dynamic}")
		}
		args = append(args, argText.String())
		argText.Reset()
	}, inDynamicParam)

	var inParamBoundary bool
	for _, element := range text {
		if currentState == inEscape {
			currentState = lastState
			if _, isReservedChar := reservedChars[element]; currentState == inDefault && !isReservedChar {
				curBuffer(currentState).WriteRune(escape)
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

		curBuffer(currentState).WriteRune(element)
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
	if typeOfArg == "special" {
		resolvedArgValue, err := newSpecialTypeResolver().resolve(argValue)
		if err != nil {
			switch err.(type) {
			case invalidSpecialParamError:
				return treatArgAsDynamic(argValue, token, lookup, fileName)
			default:
				return &gauge.StepArg{ArgType: gauge.Dynamic, Value: argValue, Name: argValue}, &ParseResult{ParseErrors: []ParseError{ParseError{FileName: fileName, LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), LineText: token.LineText}}}
			}
		}
		return resolvedArgValue, nil
	} else if typeOfArg == "static" {
		return &gauge.StepArg{ArgType: gauge.Static, Value: argValue}, nil
	} else {
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
			for _, warn := range result.Warnings {
				parseRes.Warnings = append(parseRes.Warnings, warn)
			}
		}
	}
	return stepArg, parseRes
}

func validateDynamicArg(argValue string, token *Token, lookup *gauge.ArgLookup, fileName string) (*gauge.StepArg, *ParseResult) {
	stepArgument := &gauge.StepArg{ArgType: gauge.Dynamic, Value: argValue, Name: argValue}
	if !isConceptHeader(lookup) && !lookup.ContainsArg(argValue) {
		return stepArgument, &ParseResult{ParseErrors: []ParseError{ParseError{FileName: fileName, LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), LineText: token.LineText}}}
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
				break
			case gauge_messages.Parameter_Dynamic:
				value = fmt.Sprintf("<%s>", fragment.GetParameter().GetValue())
				break
			case gauge_messages.Parameter_Special_String:
				i++
				value = fmt.Sprintf("<%s%d>", "file", i)
				break
			case gauge_messages.Parameter_Special_Table:
				i++
				value = fmt.Sprintf("<%s%d>", "table", i)
				break
			}
		}
		stepText += value
	}
	return stepText
}
