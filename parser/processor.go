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
	"strings"

	"github.com/getgauge/gauge/gauge"
)

func processSpec(parser *SpecParser, token *Token) (*ParseError, bool) {
	return nil, false
}

func processTearDown(parser *SpecParser, token *Token) (*ParseError, bool) {
	if len(token.Value) < 3 {
		return &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: "Teardown should have at least three underscore characters"}, true
	}
	return nil, false
}

func processDataTable(parser *SpecParser, token *Token) (*ParseError, bool) {
	if len(strings.TrimSpace(strings.Replace(token.Value, "table:", "", 1))) == 0 {
		return &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: "Table location not specified"}, true
	}
	return nil, false
}

func processScenario(parser *SpecParser, token *Token) (*ParseError, bool) {
	if len(strings.TrimSpace(token.Value)) < 1 {
		return &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: "Scenario heading should have at least one character"}, true
	}
	parser.clearState()
	return nil, false
}

func processComment(parser *SpecParser, token *Token) (*ParseError, bool) {
	parser.clearState()
	addStates(&parser.currentState, commentScope)
	return nil, false
}

func processTag(parser *SpecParser, token *Token) (*ParseError, bool) {
	parser.clearState()
	tokens := splitAndTrimTags(token.Value)

	for _, tagValue := range tokens {
		if len(tagValue) > 0 {
			token.Args = append(token.Args, tagValue)
		}
	}
	return nil, false
}

func processTable(parser *SpecParser, token *Token) (*ParseError, bool) {
	var buffer bytes.Buffer
	shouldEscape := false
	for i, element := range token.Value {
		if i == 0 {
			continue
		}
		if shouldEscape {
			buffer.WriteRune(element)
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
					return &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: "Table header should not be blank"}, true
				}

				if arrayContains(token.Args, trimmedValue) {
					return &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: "Table header cannot have repeated column values"}, true
				}
			}
			token.Args = append(token.Args, trimmedValue)
			buffer.Reset()
		} else {
			buffer.WriteRune(element)
		}

	}

	if !isInState(parser.currentState, tableScope) {
		addStates(&parser.currentState, tableScope)
	} else {
		addStates(&parser.currentState, tableDataScope)
	}

	return nil, false
}

func processConceptStep(spec *gauge.Specification, step *gauge.Step, conceptDictionary *gauge.ConceptDictionary) {
	if conceptFromDictionary := conceptDictionary.Search(step.Value); conceptFromDictionary != nil {
		createConceptStep(spec, conceptFromDictionary.ConceptStep, step)
	}
}

func splitAndTrimTags(tag string) []string {
	listOfTags := strings.Split(tag, ",")
	for i, aTag := range listOfTags {
		listOfTags[i] = strings.TrimSpace(aTag)
	}
	return listOfTags
}
