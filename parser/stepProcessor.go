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
	"errors"
	"fmt"
	"strconv"
	"strings"
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

func processStep(parser *SpecParser, token *Token) (*parseError, bool) {

	if len(token.value) == 0 {
		return &parseError{lineNo: token.lineNo, lineText: token.lineText, message: "Step should not be blank"}, true
	}

	stepValue, args, err := processStepText(token.value)
	if err != nil {
		return &parseError{lineNo: token.lineNo, lineText: token.lineText, message: err.Error()}, true
	}

	token.value = stepValue
	token.args = args
	parser.clearState()
	return nil, false
}

func processStepText(text string) (string, []string, error) {
	reservedChars := map[rune]struct{}{'{': {}, '}': {}}
	var stepValue, argText bytes.Buffer

	var args []string

	curBuffer := func(state int) *bytes.Buffer {
		if isInAnyState(state, inQuotes, inDynamicParam) {
			return &argText
		} else {
			return &stepValue
		}
	}

	currentState := inDefault
	lastState := -1

	acceptStaticParam := simpleAcceptor(rune(quotes), rune(quotes), func(int) {
		stepValue.WriteString("{static}")
		args = append(args, argText.String())
		argText.Reset()
	}, inQuotes)

	acceptSpecialDynamicParam := acceptor(rune(dynamicParamStart), rune(dynamicParamEnd), func(currentChar rune, state int) int {
		if currentChar == specialParamIdentifier {
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
			element = getEscapedRuneIfValid(element)
		} else if element == escape {
			lastState = currentState
			currentState = inEscape
			continue
		} else if currentState, inParamBoundary = acceptSpecialDynamicParam(element, currentState); inParamBoundary {
			continue
		} else if currentState, inParamBoundary = acceptStaticParam(element, currentState); inParamBoundary {
			continue
		} else if _, isReservedChar := reservedChars[element]; currentState == inDefault && isReservedChar {
			return "", nil, errors.New(fmt.Sprintf("'%c' is a reserved character and should be escaped", element))
		}

		curBuffer(currentState).WriteRune(element)
	}

	// If it is a valid step, the state should be default when the control reaches here
	if currentState == inQuotes {
		return "", nil, errors.New(fmt.Sprintf("String not terminated"))
	} else if isInState(currentState, inDynamicParam) {
		return "", nil, errors.New(fmt.Sprintf("Dynamic parameter not terminated"))
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
