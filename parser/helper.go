/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import "strconv"

func isInState(currentState int, statesToCheck ...int) bool {
	var mask int
	for _, value := range statesToCheck {
		mask |= value
	}
	return (mask & currentState) != 0
}

func isInAnyState(currentState int, statesToCheck ...int) bool {
	for _, value := range statesToCheck {
		if (currentState & value) != 0 {
			return true
		}
	}
	return false
}

func retainStates(currentState *int, statesToKeep ...int) {
	var mask int
	for _, value := range statesToKeep {
		mask |= value
	}
	*currentState = mask & *currentState
}

func addStates(currentState *int, states ...int) {
	var mask int
	for _, value := range states {
		mask |= value
	}
	*currentState = mask | *currentState
}

func isUnderline(text string, underlineChar rune) bool {
	if len(text) == 0 || rune(text[0]) != underlineChar {
		return false
	}
	for _, value := range text {
		if rune(value) != underlineChar {
			return false
		}
	}
	return true
}

func areUnderlined(values []string) bool {
	if len(values) == 0 {
		return false
	}
	isValuesNonEmpty := false
	for _, value := range values {
		if len(value) == 0 {
			continue
		}
		isValuesNonEmpty = true
		if !isUnderline(value, rune('-')) {
			return false
		}
	}
	return isValuesNonEmpty
}

func arrayContains(array []string, toFind string) bool {
	for _, value := range array {
		if value == toFind {
			return true
		}
	}
	return false
}

// GetUnescapedString uses the go escape sequences to escape control characters and non printable characters.
func GetUnescapedString(string1 string) string {
	unescaped := strconv.Quote(string1)
	return unescaped[1 : len(unescaped)-1]
}
