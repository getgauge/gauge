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

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/getgauge/common"
	"strings"
)

type specParser struct {
	scanner           *bufio.Scanner
	lineNo            int
	tokens            []*token
	currentState      int
	processors        map[tokenKind]func(*specParser, *token) (*parseError, bool)
	conceptDictionary *conceptDictionary
}

type tokenKind int

type token struct {
	kind     tokenKind
	lineNo   int
	lineText string
	args     []string
	value    string
}

type parseError struct {
	lineNo   int
	message  string
	lineText string
}

type parseDetailResult struct {
	error    *parseError
	warnings []*warning
}

func (se *parseError) Error() string {
	return fmt.Sprintf("line no: %d, %s", se.lineNo, se.message)
}

func (token *token) String() string {
	return fmt.Sprintf("kind:%d, lineNo:%d, value:%s, line:%s, args:%s", token.kind, token.lineNo, token.value, token.lineText, token.args)
}

const (
	initial        = 1 << iota
	specScope      = 1 << iota
	scenarioScope  = 1 << iota
	commentScope   = 1 << iota
	tableScope     = 1 << iota
	tableDataScope = 1 << iota
	stepScope      = 1 << iota
	contextScope   = 1 << iota
	conceptScope   = 1 << iota
	keywordScope   = 1 << iota
)

const (
	specKind tokenKind = iota
	tagKind
	scenarioKind
	commentKind
	stepKind
	tableHeader
	tableRow
	headingKind
	tableKind
	dataTableKind
)

func (parser *specParser) initialize() {
	parser.processors = make(map[tokenKind]func(*specParser, *token) (*parseError, bool))
	parser.processors[specKind] = processSpec
	parser.processors[scenarioKind] = processScenario
	parser.processors[commentKind] = processComment
	parser.processors[stepKind] = processStep
	parser.processors[tagKind] = processTag
	parser.processors[tableHeader] = processTable
	parser.processors[tableRow] = processTable
	parser.processors[dataTableKind] = processDataTable
}

func (parser *specParser) parse(specText string, conceptDictionary *conceptDictionary) (*specification, *parseResult) {
	tokens, parseError := parser.generateTokens(specText)
	if parseError != nil {
		return nil, &parseResult{error: parseError, ok: false}
	}
	return parser.createSpecification(tokens, conceptDictionary)
}

func (parser *specParser) generateTokens(specText string) ([]*token, *parseError) {
	parser.initialize()
	parser.scanner = bufio.NewScanner(strings.NewReader(specText))
	parser.currentState = initial
	for line, hasLine := parser.nextLine(); hasLine; line, hasLine = parser.nextLine() {
		trimmedLine := strings.TrimSpace(line)
		var newToken *token
		if len(trimmedLine) == 0 {
			newToken = &token{kind: commentKind, lineNo: parser.lineNo, lineText: line, value: "\n"}
		} else if parser.isScenarioHeading(trimmedLine) {
			newToken = &token{kind: scenarioKind, lineNo: parser.lineNo, lineText: line, value: strings.TrimSpace(trimmedLine[2:])}
		} else if parser.isSpecHeading(trimmedLine) {
			newToken = &token{kind: specKind, lineNo: parser.lineNo, lineText: line, value: strings.TrimSpace(trimmedLine[1:])}
		} else if parser.isSpecUnderline(trimmedLine) && (isInState(parser.currentState, commentScope)) {
			newToken = parser.tokens[len(parser.tokens)-1]
			newToken.kind = specKind
			parser.tokens = append(parser.tokens[:len(parser.tokens)-1])
		} else if parser.isScenarioUnderline(trimmedLine) && (isInState(parser.currentState, commentScope)) {
			newToken = parser.tokens[len(parser.tokens)-1]
			newToken.kind = scenarioKind
			parser.tokens = append(parser.tokens[:len(parser.tokens)-1])
		} else if parser.isStep(trimmedLine) {
			newToken = &token{kind: stepKind, lineNo: parser.lineNo, lineText: strings.TrimSpace(trimmedLine[1:]), value: strings.TrimSpace(trimmedLine[1:])}
		} else if found, startIndex := parser.checkTag(trimmedLine); found {
			newToken = &token{kind: tagKind, lineNo: parser.lineNo, lineText: line, value: strings.TrimSpace(trimmedLine[startIndex:])}
		} else if parser.isTableRow(trimmedLine) {
			kind := parser.tokenKindBasedOnCurrentState(tableScope, tableRow, tableHeader)
			newToken = &token{kind: kind, lineNo: parser.lineNo, lineText: line, value: strings.TrimSpace(trimmedLine)}
		} else if value, found := parser.isDataTable(trimmedLine); found {
			newToken = &token{kind: dataTableKind, lineNo: parser.lineNo, lineText: line, value: value}
		} else {
			newToken = &token{kind: commentKind, lineNo: parser.lineNo, lineText: line, value: common.TrimTrailingSpace(line)}
		}
		error := parser.accept(newToken)
		if error != nil {
			return nil, error
		}
	}
	return parser.tokens, nil

}

func (parser *specParser) tokenKindBasedOnCurrentState(state int, matchingToken tokenKind, alternateToken tokenKind) tokenKind {
	if isInState(parser.currentState, state) {
		return matchingToken
	} else {
		return alternateToken
	}
}

func (parser *specParser) isSpecHeading(text string) bool {
	if len(text) > 1 {
		return text[0] == '#' && text[1] != '#'
	} else {
		return text[0] == '#'
	}
}

func (parser *specParser) checkTag(text string) (bool, int) {
	lowerCased := strings.ToLower
	tagColon := "tags:"
	tagSpaceColon := "tags :"
	if tagStartIndex := strings.Index(lowerCased(text), tagColon); tagStartIndex == 0 {
		return true, len(tagColon)
	} else if tagStartIndex := strings.Index(lowerCased(text), tagSpaceColon); tagStartIndex == 0 {
		return true, len(tagSpaceColon)
	}
	return false, -1
}

func (parser *specParser) isScenarioHeading(text string) bool {
	if len(text) > 2 {
		return text[0] == '#' && text[1] == '#' && text[2] != '#'
	} else if len(text) == 2 {
		return text[0] == '#' && text[1] == '#'
	}
	return false
}

func (parser *specParser) isStep(text string) bool {
	if len(text) > 1 {
		return text[0] == '*' && text[1] != '*'
	} else {
		return text[0] == '*'
	}
}

func (parser *specParser) isScenarioUnderline(text string) bool {
	return isUnderline(text, rune('-'))
}

func (parser *specParser) isTableRow(text string) bool {
	return text[0] == '|' && text[len(text)-1] == '|'
}

func (parser *specParser) isSpecUnderline(text string) bool {
	return isUnderline(text, rune('='))

}

func (parser *specParser) isDataTable(text string) (string, bool) {
	lowerCased := strings.ToLower
	tableColon := "table:"
	tableSpaceColon := "table :"
	if strings.HasPrefix(lowerCased(text), tableColon) {
		return tableColon + " " + strings.TrimSpace(strings.Replace(lowerCased(text), tableColon, "", 1)), true
	} else if strings.HasPrefix(lowerCased(text), tableSpaceColon) {
		return tableColon + " " + strings.TrimSpace(strings.Replace(lowerCased(text), tableSpaceColon, "", 1)), true
	}
	return "", false
}

func (parser *specParser) accept(token *token) *parseError {
	error, shouldSkip := parser.processors[token.kind](parser, token)
	if error != nil {
		return error
	}
	if shouldSkip {
		return nil
	}
	parser.tokens = append(parser.tokens, token)
	return nil
}

func processSpec(parser *specParser, token *token) (*parseError, bool) {
	if len(token.value) < 1 {
		return &parseError{lineNo: parser.lineNo, lineText: token.value, message: "Spec heading should have at least one character"}, true
	}
	return nil, false
}

func processDataTable(parser *specParser, token *token) (*parseError, bool) {
	if len(strings.TrimSpace(strings.Replace(token.value, "table:", "", 1))) == 0 {
		return &parseError{lineNo: parser.lineNo, lineText: token.value, message: "Table location not specified"}, true
	}
	resolvedArg, err := newSpecialTypeResolver().resolve(token.value)
	if resolvedArg == nil || err != nil {
		return &parseError{lineNo: parser.lineNo, lineText: token.value, message: fmt.Sprintf("Could not resolve table from %s", token.lineText)}, true
	}
	return nil, false
}

func processScenario(parser *specParser, token *token) (*parseError, bool) {
	if len(token.value) < 1 {
		return &parseError{lineNo: parser.lineNo, lineText: token.value, message: "Scenario heading should have at least one character"}, true
	}
	parser.clearState()
	return nil, false
}

func processComment(parser *specParser, token *token) (*parseError, bool) {
	parser.clearState()
	addStates(&parser.currentState, commentScope)
	return nil, false
}

func processTag(parser *specParser, token *token) (*parseError, bool) {
	parser.clearState()
	tokens := splitAndTrimTags(token.value)

	for _, tagValue := range tokens {
		if len(tagValue) > 0 {
			token.args = append(token.args, tagValue)
		}
	}
	return nil, false
}

func splitAndTrimTags(tag string) []string {
	listOfTags := strings.Split(tag, ",")
	for i, aTag := range listOfTags {
		listOfTags[i] = strings.TrimSpace(aTag)
	}
	return listOfTags
}

func processTable(parser *specParser, token *token) (*parseError, bool) {

	var buffer bytes.Buffer
	shouldEscape := false
	for i, element := range token.value {
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

			if token.kind == tableHeader {
				if len(trimmedValue) == 0 {
					return &parseError{lineNo: parser.lineNo, lineText: token.value, message: "Table header should not be blank"}, true
				}

				if arrayContains(token.args, trimmedValue) {
					return &parseError{lineNo: parser.lineNo, lineText: token.value, message: "Table header cannot have repeated column values"}, true
				}
			}
			token.args = append(token.args, trimmedValue)
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

func (parser *specParser) nextLine() (string, bool) {
	scanned := parser.scanner.Scan()
	if scanned {
		parser.lineNo++
		return parser.scanner.Text(), true
	}
	if err := parser.scanner.Err(); err != nil {
		panic(err)
	}

	return "", false
}

func (parser *specParser) clearState() {
	parser.currentState = 0
}
