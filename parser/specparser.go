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
	"bufio"
	"bytes"
	"fmt"
	"github.com/getgauge/common"
	"strings"
)

type SpecParser struct {
	scanner           *bufio.Scanner
	lineNo            int
	tokens            []*Token
	currentState      int
	processors        map[TokenKind]func(*SpecParser, *Token) (*ParseError, bool)
	conceptDictionary *ConceptDictionary
}

type TokenKind int

type Token struct {
	Kind     TokenKind
	LineNo   int
	LineText string
	Args     []string
	Value    string
}

type ParseError struct {
	LineNo   int
	Message  string
	LineText string
}

type ParseDetailResult struct {
	Error    *ParseError
	Warnings []*Warning
}

func (se *ParseError) Error() string {
	return fmt.Sprintf("line no: %d, %s", se.LineNo, se.Message)
}

func (token *Token) String() string {
	return fmt.Sprintf("kind:%d, lineNo:%d, value:%s, line:%s, args:%s", token.Kind, token.LineNo, token.Value, token.LineText, token.Args)
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
	tearDownScope  = 1 << iota
	conceptScope   = 1 << iota
	keywordScope   = 1 << iota
)

const (
	SpecKind TokenKind = iota
	TagKind
	ScenarioKind
	CommentKind
	StepKind
	TableHeader
	TableRow
	HeadingKind
	TableKind
	DataTableKind
	TearDownKind
)

func (parser *SpecParser) initialize() {
	parser.processors = make(map[TokenKind]func(*SpecParser, *Token) (*ParseError, bool))
	parser.processors[SpecKind] = processSpec
	parser.processors[ScenarioKind] = processScenario
	parser.processors[CommentKind] = processComment
	parser.processors[StepKind] = processStep
	parser.processors[TagKind] = processTag
	parser.processors[TableHeader] = processTable
	parser.processors[TableRow] = processTable
	parser.processors[DataTableKind] = processDataTable
	parser.processors[TearDownKind] = processTearDown
}

func (parser *SpecParser) Parse(specText string, conceptDictionary *ConceptDictionary) (*Specification, *ParseResult) {
	tokens, parseError := parser.GenerateTokens(specText)
	if parseError != nil {
		return nil, &ParseResult{ParseError: parseError, Ok: false}
	}
	return parser.CreateSpecification(tokens, conceptDictionary)
}

func (parser *SpecParser) GenerateTokens(specText string) ([]*Token, *ParseError) {
	parser.initialize()
	parser.scanner = bufio.NewScanner(strings.NewReader(specText))
	parser.currentState = initial
	for line, hasLine := parser.nextLine(); hasLine; line, hasLine = parser.nextLine() {
		trimmedLine := strings.TrimSpace(line)
		var newToken *Token
		if len(trimmedLine) == 0 {
			newToken = &Token{Kind: CommentKind, LineNo: parser.lineNo, LineText: line, Value: "\n"}
		} else if parser.isScenarioHeading(trimmedLine) {
			newToken = &Token{Kind: ScenarioKind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine[2:])}
		} else if parser.isSpecHeading(trimmedLine) {
			newToken = &Token{Kind: SpecKind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine[1:])}
		} else if parser.isSpecUnderline(trimmedLine) && (isInState(parser.currentState, commentScope)) {
			newToken = parser.tokens[len(parser.tokens)-1]
			newToken.Kind = SpecKind
			parser.tokens = append(parser.tokens[:len(parser.tokens)-1])
		} else if parser.isScenarioUnderline(trimmedLine) && (isInState(parser.currentState, commentScope)) {
			newToken = parser.tokens[len(parser.tokens)-1]
			newToken.Kind = ScenarioKind
			parser.tokens = append(parser.tokens[:len(parser.tokens)-1])
		} else if parser.isStep(trimmedLine) {
			newToken = &Token{Kind: StepKind, LineNo: parser.lineNo, LineText: strings.TrimSpace(trimmedLine[1:]), Value: strings.TrimSpace(trimmedLine[1:])}
		} else if found, startIndex := parser.checkTag(trimmedLine); found {
			newToken = &Token{Kind: TagKind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine[startIndex:])}
		} else if parser.isTableRow(trimmedLine) {
			kind := parser.tokenKindBasedOnCurrentState(tableScope, TableRow, TableHeader)
			newToken = &Token{Kind: kind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine)}
		} else if value, found := parser.isDataTable(trimmedLine); found {
			newToken = &Token{Kind: DataTableKind, LineNo: parser.lineNo, LineText: line, Value: value}
		} else if parser.isTearDown(trimmedLine) {
			newToken = &Token{Kind: TearDownKind, LineNo: parser.lineNo, LineText: line, Value: trimmedLine}
		} else {
			newToken = &Token{Kind: CommentKind, LineNo: parser.lineNo, LineText: line, Value: common.TrimTrailingSpace(line)}
		}
		error := parser.accept(newToken)
		if error != nil {
			return nil, error
		}
	}
	return parser.tokens, nil

}

func (parser *SpecParser) tokenKindBasedOnCurrentState(state int, matchingToken TokenKind, alternateToken TokenKind) TokenKind {
	if isInState(parser.currentState, state) {
		return matchingToken
	} else {
		return alternateToken
	}
}

func (parser *SpecParser) isSpecHeading(text string) bool {
	if len(text) > 1 {
		return text[0] == '#' && text[1] != '#'
	} else {
		return text[0] == '#'
	}
}

func (parser *SpecParser) checkTag(text string) (bool, int) {
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

func (parser *SpecParser) isScenarioHeading(text string) bool {
	if len(text) > 2 {
		return text[0] == '#' && text[1] == '#' && text[2] != '#'
	} else if len(text) == 2 {
		return text[0] == '#' && text[1] == '#'
	}
	return false
}

func (parser *SpecParser) isStep(text string) bool {
	if len(text) > 1 {
		return text[0] == '*' && text[1] != '*'
	} else {
		return text[0] == '*'
	}
}

func (parser *SpecParser) isScenarioUnderline(text string) bool {
	return isUnderline(text, rune('-'))
}

func (parser *SpecParser) isTableRow(text string) bool {
	return text[0] == '|' && text[len(text)-1] == '|'
}

func (parser *SpecParser) isTearDown(text string) bool {
	return isUnderline(text, rune('_'))
}

func (parser *SpecParser) isSpecUnderline(text string) bool {
	return isUnderline(text, rune('='))

}

func (parser *SpecParser) isDataTable(text string) (string, bool) {
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

func (parser *SpecParser) accept(token *Token) *ParseError {
	error, shouldSkip := parser.processors[token.Kind](parser, token)
	if error != nil {
		return error
	}
	if shouldSkip {
		return nil
	}
	parser.tokens = append(parser.tokens, token)
	return nil
}

func processSpec(parser *SpecParser, token *Token) (*ParseError, bool) {
	if len(token.Value) < 1 {
		return &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: "Spec heading should have at least one character"}, true
	}
	return nil, false
}

func processTearDown(parser *SpecParser, token *Token) (*ParseError, bool) {
	if len(token.Value) < 3 {
		return &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: "Teardown should have at least three character"}, true
	}
	return nil, false
}

func processDataTable(parser *SpecParser, token *Token) (*ParseError, bool) {
	if len(strings.TrimSpace(strings.Replace(token.Value, "table:", "", 1))) == 0 {
		return &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: "Table location not specified"}, true
	}
	resolvedArg, err := newSpecialTypeResolver().resolve(token.Value)
	if resolvedArg == nil || err != nil {
		return &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: fmt.Sprintf("Could not resolve table from %s", token.LineText)}, true
	}
	return nil, false
}

func processScenario(parser *SpecParser, token *Token) (*ParseError, bool) {
	if len(token.Value) < 1 {
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

func splitAndTrimTags(tag string) []string {
	listOfTags := strings.Split(tag, ",")
	for i, aTag := range listOfTags {
		listOfTags[i] = strings.TrimSpace(aTag)
	}
	return listOfTags
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

			if token.Kind == TableHeader {
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

func (parser *SpecParser) nextLine() (string, bool) {
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

func (parser *SpecParser) clearState() {
	parser.currentState = 0
}
