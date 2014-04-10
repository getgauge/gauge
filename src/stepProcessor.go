package main

import (
	"bytes"
	"fmt"
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

func processStep(parser *specParser, token *token) (error, bool) {
	if !isInState(parser.currentState, specScope) {
		return &syntaxError{lineNo: token.lineNo, lineText: token.lineText, message: "Spec heading is not present"}, true
	}

	if len(token.value) == 0 {
		return &syntaxError{lineNo: token.lineNo, lineText: token.lineText, message: "Step should not be blank"}, true
	}
	reservedChars := map[rune]struct{}{'{': {}, '}': {}}
	var stepText, argText bytes.Buffer

	var args []string

	curBuffer := func(state int) *bytes.Buffer {
		if isInAnyState(state, inQuotes, inDynamicParam) {
			return &argText
		} else {
			return &stepText
		}
	}

	currentState := inDefault
	lastState := -1

	acceptStaticParam := simpleAcceptor(rune(quotes), rune(quotes), func(int) {
		stepText.WriteString("{static}")
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
			stepText.WriteString("{special}")
		} else {
			stepText.WriteString("{dynamic}")
		}
		args = append(args, argText.String())
		argText.Reset()
	}, inDynamicParam)

	var inParamBoundary bool
	for _, element := range token.value {
		if currentState == inEscape {
			currentState = lastState
		} else if element == escape {
			lastState = currentState
			currentState = inEscape
			continue
		} else if currentState, inParamBoundary = acceptSpecialDynamicParam(element, currentState); inParamBoundary {
			continue
		} else if currentState, inParamBoundary = acceptStaticParam(element, currentState); inParamBoundary {
			continue
		} else if _, isReservedChar := reservedChars[element]; currentState == inDefault && isReservedChar {
			return &syntaxError{lineNo: token.lineNo, lineText: token.lineText, message: fmt.Sprintf("'%c' is a reserved character and should be escaped", element)}, true
		}

		curBuffer(currentState).WriteRune(element)
	}

	// If it is a valid step, the state should be default when the control reaches here
	if currentState == inQuotes {
		return &syntaxError{lineNo: token.lineNo, lineText: token.lineText, message: "String not terminated"}, true
	} else if isInState(currentState, inDynamicParam) {
		return &syntaxError{lineNo: token.lineNo, lineText: token.lineText, message: "Dynamic parameter not terminated"}, true
	}

	token.value = strings.TrimSpace(stepText.String())
	token.args = args
	retainStates(&parser.currentState, specScope, scenarioScope)
	return nil, false
}
