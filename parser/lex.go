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
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/util"
)

const (
	initial             = 1 << iota
	specScope           = 1 << iota
	scenarioScope       = 1 << iota
	commentScope        = 1 << iota
	tableScope          = 1 << iota
	tableSeparatorScope = 1 << iota
	tableDataScope      = 1 << iota
	stepScope           = 1 << iota
	contextScope        = 1 << iota
	tearDownScope       = 1 << iota
	conceptScope        = 1 << iota
	keywordScope        = 1 << iota
	tagsScope           = 1 << iota
	newLineScope        = 1 << iota
)

// Token defines the type of entity identified by the lexer
type Token struct {
	Kind     gauge.TokenKind
	LineNo   int
	LineText string
	Suffix   string
	Args     []string
	Value    string
}

func (parser *SpecParser) initialize() {
	parser.processors = make(map[gauge.TokenKind]func(*SpecParser, *Token) ([]error, bool))
	parser.processors[gauge.SpecKind] = processSpec
	parser.processors[gauge.ScenarioKind] = processScenario
	parser.processors[gauge.CommentKind] = processComment
	parser.processors[gauge.StepKind] = processStep
	parser.processors[gauge.TagKind] = processTag
	parser.processors[gauge.TableHeader] = processTable
	parser.processors[gauge.TableRow] = processTable
	parser.processors[gauge.DataTableKind] = processDataTable
	parser.processors[gauge.TearDownKind] = processTearDown
}

// GenerateTokens generates tokens based on the parsed line.
func (parser *SpecParser) GenerateTokens(specText, fileName string) ([]*Token, []ParseError) {
	parser.initialize()
	parser.scanner = bufio.NewScanner(strings.NewReader(specText))
	parser.currentState = initial
	var errors []ParseError
	var newToken *Token
	var lastTokenErrorCount int
	for line, hasLine, err := parser.nextLine(); hasLine; line, hasLine, err = parser.nextLine() {
		if err != nil {
			errors = append(errors, ParseError{Message: err.Error()})
			return nil, errors
		}
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) == 0 {
			addStates(&parser.currentState, newLineScope)
			if newToken != nil && newToken.Kind == gauge.StepKind {
				newToken.Suffix = "\n"
				continue
			}
			newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, LineText: line, Value: "\n"}
		} else if parser.isScenarioHeading(trimmedLine) {
			newToken = &Token{Kind: gauge.ScenarioKind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine[2:])}
		} else if parser.isSpecHeading(trimmedLine) {
			newToken = &Token{Kind: gauge.SpecKind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine[1:])}
		} else if parser.isSpecUnderline(trimmedLine) {
			if isInState(parser.currentState, commentScope) {
				newToken = parser.tokens[len(parser.tokens)-1]
				newToken.Kind = gauge.SpecKind
				parser.discardLastToken()
			} else {
				newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, LineText: line, Value: common.TrimTrailingSpace(line)}
			}
		} else if parser.isScenarioUnderline(trimmedLine) {
			if isInState(parser.currentState, commentScope) {
				newToken = parser.tokens[len(parser.tokens)-1]
				newToken.Kind = gauge.ScenarioKind
				parser.discardLastToken()
			} else {
				newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, LineText: line, Value: common.TrimTrailingSpace(line)}
			}
		} else if parser.isStep(trimmedLine) {
			newToken = &Token{Kind: gauge.StepKind, LineNo: parser.lineNo, LineText: strings.TrimSpace(trimmedLine[1:]), Value: strings.TrimSpace(trimmedLine[1:])}
		} else if found, startIndex := parser.checkTag(trimmedLine); found || isInState(parser.currentState, tagsScope) {
			if isInState(parser.currentState, tagsScope) {
				startIndex = 0
			}
			if parser.isTagEndingWithComma(trimmedLine) {
				addStates(&parser.currentState, tagsScope)
			} else {
				parser.clearState()
			}
			newToken = &Token{Kind: gauge.TagKind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine[startIndex:])}
		} else if parser.isTableRow(trimmedLine) {
			kind := parser.tokenKindBasedOnCurrentState(tableScope, gauge.TableRow, gauge.TableHeader)
			newToken = &Token{Kind: kind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine)}
		} else if value, found := parser.isDataTable(trimmedLine); found {
			newToken = &Token{Kind: gauge.DataTableKind, LineNo: parser.lineNo, LineText: line, Value: value}
		} else if parser.isTearDown(trimmedLine) {
			newToken = &Token{Kind: gauge.TearDownKind, LineNo: parser.lineNo, LineText: line, Value: trimmedLine}
		} else if shouldAllowMultiLineStep() && newToken != nil && newToken.Kind == gauge.StepKind && !isInState(parser.currentState, newLineScope) {
			v := fmt.Sprintf("%s %s", newToken.LineText, trimmedLine)
			newToken = &Token{Kind: gauge.StepKind, LineNo: newToken.LineNo, LineText: strings.TrimSpace(v), Value: strings.TrimSpace(v)}
			errors = errors[:lastTokenErrorCount]
			parser.discardLastToken()
		} else {
			newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, LineText: line, Value: common.TrimTrailingSpace(line)}
		}
		pErrs := parser.accept(newToken, fileName)
		lastTokenErrorCount = len(pErrs)
		errors = append(errors, pErrs...)
	}
	return parser.tokens, errors
}

func shouldAllowMultiLineStep() bool {
	return util.ConvertToBool(os.Getenv(env.AllowMultilineStep), env.AllowMultilineStep, false)
}

func (parser *SpecParser) tokenKindBasedOnCurrentState(state int, matchingToken gauge.TokenKind, alternateToken gauge.TokenKind) gauge.TokenKind {
	if isInState(parser.currentState, state) {
		return matchingToken
	}
	return alternateToken
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

func (parser *SpecParser) isTagEndingWithComma(text string) bool {
	return strings.HasSuffix(strings.ToLower(text), ",")
}

func (parser *SpecParser) isSpecHeading(text string) bool {
	if len(text) > 1 {
		return text[0] == '#' && text[1] != '#'
	}
	return text[0] == '#'
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
	}
	return text[0] == '*'
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
	if regexp.MustCompile(`^\s*[tT][aA][bB][lL][eE]\s*:(\s*)`).FindIndex([]byte(text)) != nil {
		index := strings.Index(text, ":")
		if index != -1 {
			return "table:" + " " + strings.TrimSpace(strings.SplitAfterN(text, ":", 2)[1]), true
		}
	}
	return "", false
}

//concept header will have dynamic param and should not be resolved through lookup, so passing nil lookup
func isConceptHeader(lookup *gauge.ArgLookup) bool {
	return lookup == nil
}

func (parser *SpecParser) accept(token *Token, fileName string) []ParseError {
	errs, _ := parser.processors[token.Kind](parser, token)
	parser.tokens = append(parser.tokens, token)
	var parseErrs []ParseError
	for _, err := range errs {
		parseErrs = append(parseErrs, ParseError{FileName: fileName, LineNo: token.LineNo, Message: err.Error(), LineText: token.Value})
	}
	return parseErrs
}

func (parser *SpecParser) nextLine() (string, bool, error) {
	scanned := parser.scanner.Scan()
	if scanned {
		parser.lineNo++
		return parser.scanner.Text(), true, nil
	}
	if err := parser.scanner.Err(); err != nil {
		return "", false, err
	}

	return "", false, nil
}

func (parser *SpecParser) clearState() {
	parser.currentState = 0
}

func (parser *SpecParser) discardLastToken() {
	if len(parser.tokens) < 1 {
		return
	}
	parser.tokens = parser.tokens[:len(parser.tokens)-1]
}
