/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/gauge"
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
	Kind    gauge.TokenKind
	LineNo  int
	Suffix  string
	Args    []string
	Value   string
	Lines   []string
	SpanEnd int
}

func (t *Token) LineText() string {
	return strings.Join(t.Lines, " ")
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
// GenerateTokens gets tokens based on the parsed line.
func (parser *SpecParser) GenerateTokens(specText, fileName string) ([]*Token, []ParseError) {
    parser.initialize()
    parser.scanner = bufio.NewScanner(strings.NewReader(specText))
    parser.currentState = initial
    var errors []ParseError
    var newToken *Token
    var lastTokenErrorCount int
    
    // Store lines for multiline detection
    var allLines []string
    scanner := bufio.NewScanner(strings.NewReader(specText))
    for scanner.Scan() {
        allLines = append(allLines, scanner.Text())
    }
    
    lineIndex := 0
    parser.scanner = bufio.NewScanner(strings.NewReader(specText))
    
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
                lineIndex++
                continue
            }
            newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, Lines: []string{line}, Value: "\n", SpanEnd: parser.lineNo}
        } else if parser.isScenarioHeading(trimmedLine) {
            newToken = &Token{Kind: gauge.ScenarioKind, LineNo: parser.lineNo, Lines: []string{line}, Value: strings.TrimSpace(trimmedLine[2:]), SpanEnd: parser.lineNo}
        } else if parser.isSpecHeading(trimmedLine) {
            newToken = &Token{Kind: gauge.SpecKind, LineNo: parser.lineNo, Lines: []string{line}, Value: strings.TrimSpace(trimmedLine[1:]), SpanEnd: parser.lineNo}
        } else if parser.isSpecUnderline(trimmedLine) {
            if isInState(parser.currentState, commentScope) {
                newToken = parser.tokens[len(parser.tokens)-1]
                newToken.Kind = gauge.SpecKind
                newToken.SpanEnd = parser.lineNo
                parser.discardLastToken()
            } else {
                newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, Lines: []string{line}, Value: common.TrimTrailingSpace(line), SpanEnd: parser.lineNo}
            }
        } else if parser.isScenarioUnderline(trimmedLine) {
            if isInState(parser.currentState, commentScope) {
                newToken = parser.tokens[len(parser.tokens)-1]
                newToken.Kind = gauge.ScenarioKind
                newToken.SpanEnd = parser.lineNo
                parser.discardLastToken()
            } else {
                newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, Lines: []string{line}, Value: common.TrimTrailingSpace(line), SpanEnd: parser.lineNo}
            }
        } else if parser.isStep(trimmedLine) {
            stepLine := strings.TrimSpace(trimmedLine[1:])
            stepToken := &Token{Kind: gauge.StepKind, LineNo: parser.lineNo, Lines: []string{stepLine}, Value: stepLine, SpanEnd: parser.lineNo}

            // Check for multiline strings - use the pre-stored lines
            if lineIndex+1 < len(allLines) {
                nextLine := allLines[lineIndex+1]
                nextTrimmed := strings.TrimSpace(nextLine)
                if nextTrimmed == `"""` {
                    if content, found, consumedLines := parser.extractMultilineContent(allLines, lineIndex+1); found {
                        stepToken.Args = []string{content}
                        // Advance the scanner past the multiline content
                        for i := 0; i < consumedLines; i++ {
                            parser.nextLine()
                            lineIndex++
                        }
                    }
                }
            }
            
            newToken = stepToken
        } else if found, startIndex := parser.checkTag(trimmedLine); found || isInState(parser.currentState, tagsScope) {
            if isInState(parser.currentState, tagsScope) {
                startIndex = 0
            }
            if parser.isTagEndingWithComma(trimmedLine) {
                addStates(&parser.currentState, tagsScope)
            } else {
                parser.clearState()
            }
            newToken = &Token{Kind: gauge.TagKind, LineNo: parser.lineNo, Lines: []string{line}, Value: strings.TrimSpace(trimmedLine[startIndex:]), SpanEnd: parser.lineNo}
        } else if parser.isTableRow(trimmedLine) {
            kind := parser.tokenKindBasedOnCurrentState(tableScope, gauge.TableRow, gauge.TableHeader)
            newToken = &Token{Kind: kind, LineNo: parser.lineNo, Lines: []string{line}, Value: strings.TrimSpace(trimmedLine), SpanEnd: parser.lineNo}
        } else if value, found := parser.isDataTable(trimmedLine); found {
            newToken = &Token{Kind: gauge.DataTableKind, LineNo: parser.lineNo, Lines: []string{line}, Value: value, SpanEnd: parser.lineNo}
        } else if parser.isTearDown(trimmedLine) {
            newToken = &Token{Kind: gauge.TearDownKind, LineNo: parser.lineNo, Lines: []string{line}, Value: trimmedLine, SpanEnd: parser.lineNo}
        } else if env.AllowMultiLineStep() && newToken != nil && newToken.Kind == gauge.StepKind && !isInState(parser.currentState, newLineScope) {
			v := strings.TrimSpace(fmt.Sprintf("%s %s", newToken.LineText(), line))
			
			// For classic multiline steps, just merge the text but DON'T set Args
			// Args will be populated later by processStep parameter processing
			newToken = parser.tokens[len(parser.tokens)-1]
			newToken.Value = v
			newToken.Lines = append(newToken.Lines, line)
			newToken.SpanEnd = parser.lineNo
			newToken.Args = []string{} // Ensure Args is empty for parameter processing
			errors = errors[:lastTokenErrorCount]
			parser.discardLastToken()
        } else {
            newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, Lines: []string{line}, Value: common.TrimTrailingSpace(line), SpanEnd: parser.lineNo}
        }
        pErrs := parser.accept(newToken, fileName)
        lastTokenErrorCount = len(pErrs)
        errors = append(errors, pErrs...)
        lineIndex++
    }
    return parser.tokens, errors
}

// extractMultilineContent extracts content between """ delimiters
func (parser *SpecParser) extractMultilineContent(lines []string, startIndex int) (string, bool, int) {
	var content []string
	consumedLines := 1 // Start by consuming the opening """

	for i := startIndex + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == `"""` {
			return strings.Join(content, "\n"), true, consumedLines + 1 // +1 for closing """
		}
		content = append(content, lines[i])
		consumedLines++
	}

	return "", false, consumedLines
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
		parseErrs = append(parseErrs, ParseError{FileName: fileName, LineNo: token.LineNo, Message: err.Error(), LineText: token.Value, SpanEnd: token.SpanEnd})
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

func (parser *SpecParser) tokenKindBasedOnCurrentState(state int, matchingToken gauge.TokenKind, alternateToken gauge.TokenKind) gauge.TokenKind {
	if isInState(parser.currentState, state) {
		return matchingToken
	}
	return alternateToken
}