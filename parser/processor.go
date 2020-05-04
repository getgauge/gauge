/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/getgauge/gauge/gauge"
)

func processSpec(parser *SpecParser, token *Token) ([]error, bool) {
	return []error{}, false
}

func processTearDown(parser *SpecParser, token *Token) ([]error, bool) {
	if len(token.Value) < 3 {
		return []error{fmt.Errorf("Teardown should have at least three underscore characters")}, true
	}
	return []error{}, false
}

func processDataTable(parser *SpecParser, token *Token) ([]error, bool) {
	if len(strings.TrimSpace(strings.Replace(token.Value, "table:", "", 1))) == 0 {
		return []error{fmt.Errorf("Table location not specified")}, true
	}
	return []error{}, false
}

func processScenario(parser *SpecParser, token *Token) ([]error, bool) {
	if len(strings.TrimSpace(token.Value)) < 1 {
		return []error{fmt.Errorf("Scenario heading should have at least one character")}, true
	}
	parser.clearState()
	return []error{}, false
}

func processComment(parser *SpecParser, token *Token) ([]error, bool) {
	parser.clearState()
	addStates(&parser.currentState, commentScope)
	return []error{}, false
}

func processTag(parser *SpecParser, token *Token) ([]error, bool) {
	if isInState(parser.currentState, tagsScope) {
		retainStates(&parser.currentState, tagsScope)
	} else {
		parser.clearState()
	}
	tokens := splitAndTrimTags(token.Value)

	for _, tagValue := range tokens {
		if len(tagValue) > 0 {
			token.Args = append(token.Args, tagValue)
		}
	}
	return []error{}, false
}

func processTable(parser *SpecParser, token *Token) ([]error, bool) {
	var buffer bytes.Buffer
	shouldEscape := false
	var errs []error
	for i, element := range token.Value {
		if i == 0 {
			continue
		}
		if shouldEscape {
			_, err := buffer.WriteRune(element)
			if err != nil {
				errs = append(errs, err)
			}
			shouldEscape = false
			continue
		}
		if element == '\\' {
			shouldEscape = true
			continue
		} else if element == '|' {
			trimmedValue := strings.TrimSpace(buffer.String())

			if token.Kind == gauge.TableHeader {
				if len(trimmedValue) == 0 {
					errs = append(errs, fmt.Errorf("Table header should not be blank"))
				} else if arrayContains(token.Args, trimmedValue) {
					errs = append(errs, fmt.Errorf("Table header cannot have repeated column values"))
				}
			}
			token.Args = append(token.Args, trimmedValue)
			buffer.Reset()
		} else {
			_, err := buffer.WriteRune(element)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if !isInState(parser.currentState, tableScope) {
		addStates(&parser.currentState, tableScope)
	} else {
		addStates(&parser.currentState, tableDataScope)
	}

	return errs, false
}

func splitAndTrimTags(tag string) []string {
	listOfTags := strings.Split(tag, ",")
	for i, aTag := range listOfTags {
		listOfTags[i] = strings.TrimSpace(aTag)
	}
	return listOfTags
}
