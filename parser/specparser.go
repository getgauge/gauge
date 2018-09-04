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
	"regexp"
	"strings"

	"github.com/getgauge/gauge/util"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
)

type SpecParser struct {
	scanner           *bufio.Scanner
	lineNo            int
	tokens            []*Token
	currentState      int
	processors        map[gauge.TokenKind]func(*SpecParser, *Token) ([]error, bool)
	conceptDictionary *gauge.ConceptDictionary
}

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

// Parse generates tokens for the given spec text and creates the specification.
func (parser *SpecParser) Parse(specText string, conceptDictionary *gauge.ConceptDictionary, specFile string) (*gauge.Specification, *ParseResult, error) {
	tokens, errs := parser.GenerateTokens(specText, specFile)
	spec, res, err := parser.CreateSpecification(tokens, conceptDictionary, specFile)
	if err != nil {
		return nil, nil, err
	}
	res.FileName = specFile
	if len(errs) > 0 {
		res.Ok = false
	}
	res.ParseErrors = append(errs, res.ParseErrors...)
	return spec, res, nil
}

// ParseSpecText without validating and replacing concepts.
func (parser *SpecParser) ParseSpecText(specText string, specFile string) (*gauge.Specification, *ParseResult) {
	tokens, errs := parser.GenerateTokens(specText, specFile)
	spec, res := parser.createSpecification(tokens, specFile)
	res.FileName = specFile
	if len(errs) > 0 {
		res.Ok = false
	}
	res.ParseErrors = append(errs, res.ParseErrors...)
	return spec, res
}

// Generates tokens based on the parsed line.
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
		} else if parser.isSpecUnderline(trimmedLine) && isInState(parser.currentState, commentScope) {
			newToken = parser.tokens[len(parser.tokens)-1]
			newToken.Kind = gauge.SpecKind
			parser.discardLastToken()
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
		} else if newToken != nil && newToken.Kind == gauge.StepKind && !isInState(parser.currentState, newLineScope) {
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
	} else {
		return text[0] == '#'
	}
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

// CreateSpecification creates specification from the given set of tokens.
func (parser *SpecParser) CreateSpecification(tokens []*Token, conceptDictionary *gauge.ConceptDictionary, specFile string) (*gauge.Specification, *ParseResult, error) {
	parser.conceptDictionary = conceptDictionary
	specification, finalResult := parser.createSpecification(tokens, specFile)
	if err := specification.ProcessConceptStepsFrom(conceptDictionary); err != nil {
		return nil, nil, err
	}
	err := parser.validateSpec(specification)
	if err != nil {
		finalResult.Ok = false
		finalResult.ParseErrors = append([]ParseError{err.(ParseError)}, finalResult.ParseErrors...)
	}
	return specification, finalResult, nil
}

func (parser *SpecParser) createSpecification(tokens []*Token, specFile string) (*gauge.Specification, *ParseResult) {
	finalResult := &ParseResult{ParseErrors: make([]ParseError, 0), Ok: true}
	converters := parser.initializeConverters()
	specification := &gauge.Specification{FileName: specFile}
	state := initial
	for _, token := range tokens {
		for _, converter := range converters {
			result := converter(token, &state, specification)
			if !result.Ok {
				if result.ParseErrors != nil {
					finalResult.Ok = false
					finalResult.ParseErrors = append(finalResult.ParseErrors, result.ParseErrors...)
				}
			}
			if result.Warnings != nil {
				if finalResult.Warnings == nil {
					finalResult.Warnings = make([]*Warning, 0)
				}
				finalResult.Warnings = append(finalResult.Warnings, result.Warnings...)
			}
		}
	}
	if len(specification.Scenarios) > 0 {
		specification.LatestScenario().Span.End = tokens[len(tokens)-1].LineNo
	}
	return specification, finalResult
}

func (parser *SpecParser) initializeConverters() []func(*Token, *int, *gauge.Specification) ParseResult {
	specConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.SpecKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		if spec.Heading != nil {
			return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{spec.FileName, token.LineNo, "Multiple spec headings found in same file", token.LineText}}}
		}

		spec.AddHeading(&gauge.Heading{LineNo: token.LineNo, Value: token.Value})
		addStates(state, specScope)
		return ParseResult{Ok: true}
	})

	scenarioConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.ScenarioKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		if spec.Heading == nil {
			return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{spec.FileName, token.LineNo, "Scenario should be defined after the spec heading", token.LineText}}}
		}
		for _, scenario := range spec.Scenarios {
			if strings.ToLower(scenario.Heading.Value) == strings.ToLower(token.Value) {
				return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{spec.FileName, token.LineNo, "Duplicate scenario definition '" + scenario.Heading.Value + "' found in the same specification", token.LineText}}}
			}
		}
		scenario := &gauge.Scenario{Span: &gauge.Span{Start: token.LineNo, End: token.LineNo}}
		if len(spec.Scenarios) > 0 {
			spec.LatestScenario().Span.End = token.LineNo - 1
		}
		scenario.AddHeading(&gauge.Heading{Value: token.Value, LineNo: token.LineNo})
		spec.AddScenario(scenario)

		retainStates(state, specScope)
		addStates(state, scenarioScope)
		return ParseResult{Ok: true}
	})

	stepConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.StepKind && isInState(*state, scenarioScope)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		latestScenario := spec.LatestScenario()
		stepToAdd, parseDetails := createStep(spec, token)
		if stepToAdd == nil {
			return ParseResult{ParseErrors: parseDetails.ParseErrors, Ok: false, Warnings: parseDetails.Warnings}
		}
		latestScenario.AddStep(stepToAdd)
		retainStates(state, specScope, scenarioScope)
		addStates(state, stepScope)
		if parseDetails != nil && len(parseDetails.ParseErrors) > 0 {
			return ParseResult{ParseErrors: parseDetails.ParseErrors, Ok: false, Warnings: parseDetails.Warnings}
		}
		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	contextConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.StepKind && !isInState(*state, scenarioScope) && isInState(*state, specScope) && !isInState(*state, tearDownScope)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		stepToAdd, parseDetails := createStep(spec, token)
		if stepToAdd == nil {
			return ParseResult{ParseErrors: parseDetails.ParseErrors, Ok: false, Warnings: parseDetails.Warnings}
		}
		spec.AddContext(stepToAdd)
		retainStates(state, specScope)
		addStates(state, contextScope)
		if parseDetails != nil && len(parseDetails.ParseErrors) > 0 {
			parseDetails.Ok = false
			return *parseDetails
		}
		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	tearDownConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.TearDownKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		retainStates(state, specScope)
		addStates(state, tearDownScope)
		spec.AddItem(&gauge.TearDown{LineNo: token.LineNo, Value: token.Value})
		return ParseResult{Ok: true}
	})

	tearDownStepConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.StepKind && isInState(*state, tearDownScope)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		stepToAdd, parseDetails := createStep(spec, token)
		if stepToAdd == nil {
			return ParseResult{ParseErrors: parseDetails.ParseErrors, Ok: false, Warnings: parseDetails.Warnings}
		}
		spec.TearDownSteps = append(spec.TearDownSteps, stepToAdd)
		spec.AddItem(stepToAdd)
		retainStates(state, specScope, tearDownScope)
		if parseDetails != nil && len(parseDetails.ParseErrors) > 0 {
			parseDetails.Ok = false
			return *parseDetails
		}
		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	commentConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.CommentKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		comment := &gauge.Comment{token.Value, token.LineNo}
		if isInState(*state, scenarioScope) {
			spec.LatestScenario().AddComment(comment)
		} else {
			spec.AddComment(comment)
		}
		retainStates(state, specScope, scenarioScope, tearDownScope)
		addStates(state, commentScope)
		return ParseResult{Ok: true}
	})

	keywordConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.DataTableKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		resolvedArg, err := newSpecialTypeResolver().resolve(token.Value)
		if resolvedArg == nil || err != nil {
			e := ParseError{FileName: spec.FileName, LineNo: token.LineNo, LineText: token.LineText, Message: fmt.Sprintf("Could not resolve table from %s", token.LineText)}
			return ParseResult{ParseErrors: []ParseError{e}, Ok: false}
		}
		if isInState(*state, specScope) && !spec.DataTable.IsInitialized() {
			externalTable := &gauge.DataTable{}
			externalTable.Table = resolvedArg.Table
			externalTable.LineNo = token.LineNo
			externalTable.Value = token.Value
			externalTable.IsExternal = true
			spec.AddExternalDataTable(externalTable)
		} else if isInState(*state, specScope) && spec.DataTable.IsInitialized() {
			value := "Multiple data table present, ignoring table"
			spec.AddComment(&gauge.Comment{token.LineText, token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{spec.FileName, token.LineNo, value}}}
		} else {
			value := "Data table not associated with spec"
			spec.AddComment(&gauge.Comment{token.LineText, token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{spec.FileName, token.LineNo, value}}}
		}
		retainStates(state, specScope)
		addStates(state, keywordScope)
		return ParseResult{Ok: true}
	})

	tableHeaderConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.TableHeader && isInState(*state, specScope)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		if isInState(*state, stepScope) {
			latestScenario := spec.LatestScenario()
			latestStep := latestScenario.LatestStep()
			addInlineTableHeader(latestStep, token)
		} else if isInState(*state, contextScope) {
			latestContext := spec.LatestContext()
			addInlineTableHeader(latestContext, token)
		} else if isInState(*state, tearDownScope) {
			if len(spec.TearDownSteps) > 0 {
				latestTeardown := spec.LatestTeardown()
				addInlineTableHeader(latestTeardown, token)
			} else {
				spec.AddComment(&gauge.Comment{token.LineText, token.LineNo})
			}
		} else if !isInState(*state, scenarioScope) {
			if !spec.DataTable.Table.IsInitialized() {
				dataTable := &gauge.Table{}
				dataTable.LineNo = token.LineNo
				dataTable.AddHeaders(token.Args)
				spec.AddDataTable(dataTable)
			} else {
				value := "Multiple data table present, ignoring table"
				spec.AddComment(&gauge.Comment{token.LineText, token.LineNo})
				return ParseResult{Ok: false, Warnings: []*Warning{&Warning{spec.FileName, token.LineNo, value}}}
			}
		} else {
			value := "Table not associated with a step, ignoring table"
			spec.LatestScenario().AddComment(&gauge.Comment{token.LineText, token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{spec.FileName, token.LineNo, value}}}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tearDownScope)
		addStates(state, tableScope)
		return ParseResult{Ok: true}
	})

	tableRowConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.TableRow
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		var result ParseResult
		//When table is to be treated as a comment
		if !isInState(*state, tableScope) {
			if isInState(*state, scenarioScope) {
				spec.LatestScenario().AddComment(&gauge.Comment{token.LineText, token.LineNo})
			} else {
				spec.AddComment(&gauge.Comment{token.LineText, token.LineNo})
			}
		} else if areUnderlined(token.Args) && !isInState(*state, tableSeparatorScope) {
			retainStates(state, specScope, scenarioScope, stepScope, contextScope, tearDownScope, tableScope)
			addStates(state, tableSeparatorScope)
			// skip table separator
			result = ParseResult{Ok: true}
		} else if isInState(*state, stepScope) {
			latestScenario := spec.LatestScenario()
			latestStep := latestScenario.LatestStep()
			result = addInlineTableRow(latestStep, token, new(gauge.ArgLookup).FromDataTable(&spec.DataTable.Table), spec.FileName)
		} else if isInState(*state, contextScope) {
			latestContext := spec.LatestContext()
			result = addInlineTableRow(latestContext, token, new(gauge.ArgLookup).FromDataTable(&spec.DataTable.Table), spec.FileName)
		} else if isInState(*state, tearDownScope) {
			if len(spec.TearDownSteps) > 0 {
				latestTeardown := spec.LatestTeardown()
				result = addInlineTableRow(latestTeardown, token, new(gauge.ArgLookup).FromDataTable(&spec.DataTable.Table), spec.FileName)
			} else {
				spec.AddComment(&gauge.Comment{token.LineText, token.LineNo})
			}
		} else {
			tableValues, warnings, err := validateTableRows(token, new(gauge.ArgLookup).FromDataTable(&spec.DataTable.Table), spec.FileName)
			if len(err) > 0 {
				result = ParseResult{Ok: false, Warnings: warnings, ParseErrors: err}
			} else {
				spec.DataTable.Table.AddRowValues(tableValues)
				result = ParseResult{Ok: true, Warnings: warnings}
			}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tearDownScope, tableScope, tableSeparatorScope)
		return result
	})

	tagConverter := converterFn(func(token *Token, state *int) bool {
		return (token.Kind == gauge.TagKind)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		tags := &gauge.Tags{RawValues: [][]string{token.Args}}
		if isInState(*state, scenarioScope) {
			if isInState(*state, tagsScope) {
				spec.LatestScenario().Tags.Add(tags.RawValues[0])
			} else {
				if spec.LatestScenario().NTags() != 0 {
					return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{FileName: spec.FileName, LineNo: token.LineNo, Message: "Tags can be defined only once per scenario", LineText: token.LineText}}}
				}
				spec.LatestScenario().AddTags(tags)
			}
		} else {
			if isInState(*state, tagsScope) {
				spec.Tags.Add(tags.RawValues[0])
			} else {
				if spec.NTags() != 0 {
					return ParseResult{Ok: false, ParseErrors: []ParseError{ParseError{FileName: spec.FileName, LineNo: token.LineNo, Message: "Tags can be defined only once per specification", LineText: token.LineText}}}
				}
				spec.AddTags(tags)
			}
		}
		addStates(state, tagsScope)
		return ParseResult{Ok: true}
	})

	converter := []func(*Token, *int, *gauge.Specification) ParseResult{
		specConverter, scenarioConverter, stepConverter, contextConverter, commentConverter, tableHeaderConverter, tableRowConverter, tagConverter, keywordConverter, tearDownConverter, tearDownStepConverter,
	}

	return converter
}

func (parser *SpecParser) validateSpec(specification *gauge.Specification) error {
	if len(specification.Items) == 0 {
		specification.AddHeading(&gauge.Heading{})
		return ParseError{FileName: specification.FileName, LineNo: 1, Message: "Spec does not have any elements"}
	}
	if specification.Heading == nil {
		specification.AddHeading(&gauge.Heading{})
		return ParseError{FileName: specification.FileName, LineNo: 1, Message: "Spec heading not found"}
	}
	if len(strings.TrimSpace(specification.Heading.Value)) < 1 {
		return ParseError{FileName: specification.FileName, LineNo: specification.Heading.LineNo, Message: "Spec heading should have at least one character"}
	}

	dataTable := specification.DataTable.Table
	if dataTable.IsInitialized() && dataTable.GetRowCount() == 0 {
		return ParseError{FileName: specification.FileName, LineNo: dataTable.LineNo, Message: "Data table should have at least 1 data row"}
	}
	if len(specification.Scenarios) == 0 {
		return ParseError{FileName: specification.FileName, LineNo: specification.Heading.LineNo, Message: "Spec should have atleast one scenario"}
	}
	for _, sce := range specification.Scenarios {
		if len(sce.Steps) == 0 {
			return ParseError{FileName: specification.FileName, LineNo: sce.Heading.LineNo, Message: "Scenario should have atleast one step"}
		}
	}
	return nil
}

func converterFn(predicate func(token *Token, state *int) bool, apply func(token *Token, spec *gauge.Specification, state *int) ParseResult) func(*Token, *int, *gauge.Specification) ParseResult {
	return func(token *Token, state *int, spec *gauge.Specification) ParseResult {
		if !predicate(token, state) {
			return ParseResult{Ok: true}
		}
		return apply(token, spec, state)
	}
}

func createStep(spec *gauge.Specification, stepToken *Token) (*gauge.Step, *ParseResult) {
	dataTableLookup := new(gauge.ArgLookup).FromDataTable(&spec.DataTable.Table)
	stepToAdd, parseDetails := CreateStepUsingLookup(stepToken, dataTableLookup, spec.FileName)
	if stepToAdd != nil {
		stepToAdd.Suffix = stepToken.Suffix
	}
	return stepToAdd, parseDetails
}

// CreateStepUsingLookup generates gauge steps from step token and args lookup.
func CreateStepUsingLookup(stepToken *Token, lookup *gauge.ArgLookup, specFileName string) (*gauge.Step, *ParseResult) {
	stepValue, argsType := extractStepValueAndParameterTypes(stepToken.Value)
	if argsType != nil && len(argsType) != len(stepToken.Args) {
		return nil, &ParseResult{ParseErrors: []ParseError{ParseError{specFileName, stepToken.LineNo, "Step text should not have '{static}' or '{dynamic}' or '{special}'", stepToken.LineText}}, Warnings: nil}
	}
	step := &gauge.Step{FileName: specFileName, LineNo: stepToken.LineNo, Value: stepValue, LineText: strings.TrimSpace(stepToken.LineText)}
	arguments := make([]*gauge.StepArg, 0)
	var errors []ParseError
	var warnings []*Warning
	for i, argType := range argsType {
		argument, parseDetails := createStepArg(stepToken.Args[i], argType, stepToken, lookup, specFileName)
		if parseDetails != nil && len(parseDetails.ParseErrors) > 0 {
			errors = append(errors, parseDetails.ParseErrors...)
		}
		arguments = append(arguments, argument)
		if parseDetails != nil && parseDetails.Warnings != nil {
			for _, warn := range parseDetails.Warnings {
				warnings = append(warnings, warn)
			}
		}
	}
	step.AddArgs(arguments...)
	return step, &ParseResult{ParseErrors: errors, Warnings: warnings}
}

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

//Step value is modified when inline table is found to account for the new parameter by appending {}
//todo validate headers for dynamic
func addInlineTableHeader(step *gauge.Step, token *Token) {
	step.Value = fmt.Sprintf("%s %s", step.Value, gauge.ParameterPlaceholder)
	step.HasInlineTable = true
	step.AddInlineTableHeaders(token.Args)
}

func addInlineTableRow(step *gauge.Step, token *Token, argLookup *gauge.ArgLookup, fileName string) ParseResult {
	tableValues, warnings, err := validateTableRows(token, argLookup, fileName)
	if len(err) > 0 {
		return ParseResult{Ok: false, Warnings: warnings, ParseErrors: err}
	}
	step.AddInlineTableRow(tableValues)
	return ParseResult{Ok: true, Warnings: warnings}
}

func validateTableRows(token *Token, argLookup *gauge.ArgLookup, fileName string) ([]gauge.TableCell, []*Warning, []ParseError) {
	dynamicArgMatcher := regexp.MustCompile("^<(.*)>$")
	specialArgMatcher := regexp.MustCompile("^<(file:.*)>$")
	tableValues := make([]gauge.TableCell, 0)
	warnings := make([]*Warning, 0)
	error := make([]ParseError, 0)
	for _, tableValue := range token.Args {
		if specialArgMatcher.MatchString(tableValue) {
			match := specialArgMatcher.FindAllStringSubmatch(tableValue, -1)
			param := match[0][1]
			file := strings.TrimSpace(strings.TrimPrefix(param, "file:"))
			tableValues = append(tableValues, gauge.TableCell{Value: param, CellType: gauge.SpecialString})
			if _, err := util.GetFileContents(file); err != nil {
				error = append(error, ParseError{FileName: fileName, LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic param <%s> could not be resolved, Missing file: %s", param, file), LineText: token.LineText})
			}
		} else if dynamicArgMatcher.MatchString(tableValue) {
			match := dynamicArgMatcher.FindAllStringSubmatch(tableValue, -1)
			param := match[0][1]
			if !argLookup.ContainsArg(param) {
				tableValues = append(tableValues, gauge.TableCell{Value: tableValue, CellType: gauge.Static})
				warnings = append(warnings, &Warning{FileName: fileName, LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic param <%s> could not be resolved, Treating it as static param", param)})
			} else {
				tableValues = append(tableValues, gauge.TableCell{Value: param, CellType: gauge.Dynamic})
			}
		} else {
			tableValues = append(tableValues, gauge.TableCell{Value: tableValue, CellType: gauge.Static})
		}
	}
	return tableValues, warnings, error
}

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

type Token struct {
	Kind     gauge.TokenKind
	LineNo   int
	LineText string
	Suffix   string
	Args     []string
	Value    string
}

type ParseError struct {
	FileName string
	LineNo   int
	Message  string
	LineText string
}

// Error prints error with filename, line number, error message and step text.
func (se ParseError) Error() string {
	if se.LineNo == 0 && se.FileName == "" {
		return fmt.Sprintf("%s", se.Message)
	}
	return fmt.Sprintf("%s:%d %s => '%s'", se.FileName, se.LineNo, se.Message, se.LineText)
}

func (token *Token) String() string {
	return fmt.Sprintf("kind:%d, lineNo:%d, value:%s, line:%s, args:%s", token.Kind, token.LineNo, token.Value, token.LineText, token.Args)
}

type ParseResult struct {
	ParseErrors []ParseError
	Warnings    []*Warning
	Ok          bool
	FileName    string
}

// Errors Prints parse errors and critical errors.
func (result *ParseResult) Errors() (errors []string) {
	for _, err := range result.ParseErrors {
		errors = append(errors, fmt.Sprintf("[ParseError] %s", err.Error()))
	}
	return
}

type Warning struct {
	FileName string
	LineNo   int
	Message  string
}

func (warning *Warning) String() string {
	return fmt.Sprintf("%s:%d %s", warning.FileName, warning.LineNo, warning.Message)
}
