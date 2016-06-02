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

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
)

type SpecParser struct {
	scanner           *bufio.Scanner
	lineNo            int
	tokens            []*Token
	currentState      int
	processors        map[gauge.TokenKind]func(*SpecParser, *Token) (*ParseError, bool)
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
)

func (parser *SpecParser) initialize() {
	parser.processors = make(map[gauge.TokenKind]func(*SpecParser, *Token) (*ParseError, bool))
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

func (parser *SpecParser) Parse(specText string, conceptDictionary *gauge.ConceptDictionary) (*gauge.Specification, *ParseResult) {
	tokens, errs := parser.GenerateTokens(specText)
	spec, res := parser.CreateSpecification(tokens, conceptDictionary)
	if len(errs) > 0 {
		res.Ok = false
	}
	res.ParseErrors = append(errs, res.ParseErrors...)
	return spec, res
}

func (parser *SpecParser) GenerateTokens(specText string) ([]*Token, []*ParseError) {
	parser.initialize()
	parser.scanner = bufio.NewScanner(strings.NewReader(specText))
	parser.currentState = initial
	var errors []*ParseError
	var newToken *Token
	for line, hasLine := parser.nextLine(); hasLine; line, hasLine = parser.nextLine() {
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) == 0 {
			if newToken != nil && newToken.Kind == gauge.StepKind {
				newToken.Suffix = "\n"
				continue
			}
			newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, LineText: line, Value: "\n"}
		} else if parser.isScenarioHeading(trimmedLine) {
			newToken = &Token{Kind: gauge.ScenarioKind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine[2:])}
		} else if parser.isSpecHeading(trimmedLine) {
			newToken = &Token{Kind: gauge.SpecKind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine[1:])}
		} else if parser.isSpecUnderline(trimmedLine) && (isInState(parser.currentState, commentScope)) {
			newToken = parser.tokens[len(parser.tokens)-1]
			newToken.Kind = gauge.SpecKind
			parser.tokens = append(parser.tokens[:len(parser.tokens)-1])
		} else if parser.isScenarioUnderline(trimmedLine) && (isInState(parser.currentState, commentScope)) {
			newToken = parser.tokens[len(parser.tokens)-1]
			newToken.Kind = gauge.ScenarioKind
			parser.tokens = append(parser.tokens[:len(parser.tokens)-1])
		} else if parser.isStep(trimmedLine) {
			newToken = &Token{Kind: gauge.StepKind, LineNo: parser.lineNo, LineText: strings.TrimSpace(trimmedLine[1:]), Value: strings.TrimSpace(trimmedLine[1:])}
		} else if found, startIndex := parser.checkTag(trimmedLine); found {
			newToken = &Token{Kind: gauge.TagKind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine[startIndex:])}
		} else if parser.isTableRow(trimmedLine) {
			kind := parser.tokenKindBasedOnCurrentState(tableScope, gauge.TableRow, gauge.TableHeader)
			newToken = &Token{Kind: kind, LineNo: parser.lineNo, LineText: line, Value: strings.TrimSpace(trimmedLine)}
		} else if value, found := parser.isDataTable(trimmedLine); found {
			newToken = &Token{Kind: gauge.DataTableKind, LineNo: parser.lineNo, LineText: line, Value: value}
		} else if parser.isTearDown(trimmedLine) {
			newToken = &Token{Kind: gauge.TearDownKind, LineNo: parser.lineNo, LineText: line, Value: trimmedLine}
		} else {
			newToken = &Token{Kind: gauge.CommentKind, LineNo: parser.lineNo, LineText: line, Value: common.TrimTrailingSpace(line)}
		}
		err := parser.accept(newToken)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return parser.tokens, errors
}

func (parser *SpecParser) tokenKindBasedOnCurrentState(state int, matchingToken gauge.TokenKind, alternateToken gauge.TokenKind) gauge.TokenKind {
	if isInState(parser.currentState, state) {
		return matchingToken
	} else {
		return alternateToken
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

//concept header will have dynamic param and should not be resolved through lookup, so passing nil lookup
func isConceptHeader(lookup *gauge.ArgLookup) bool {
	return lookup == nil
}

func (parser *SpecParser) accept(token *Token) *ParseError {
	error, _ := parser.processors[token.Kind](parser, token)
	parser.tokens = append(parser.tokens, token)
	return error
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

func (parser *SpecParser) CreateSpecification(tokens []*Token, conceptDictionary *gauge.ConceptDictionary) (*gauge.Specification, *ParseResult) {
	parser.conceptDictionary = conceptDictionary
	converters := parser.initializeConverters()
	specification := &gauge.Specification{}
	finalResult := &ParseResult{ParseErrors: make([]*ParseError, 0), Ok: true}
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
	specification.ProcessConceptStepsFrom(conceptDictionary)
	validationError := parser.validateSpec(specification)
	if validationError != nil {
		finalResult.Ok = false
		finalResult.ParseErrors = append([]*ParseError{validationError}, finalResult.ParseErrors...)
	}
	return specification, finalResult
}

func (parser *SpecParser) initializeConverters() []func(*Token, *int, *gauge.Specification) ParseResult {
	specConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.SpecKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		if spec.Heading != nil {
			return ParseResult{Ok: false, ParseErrors: []*ParseError{&ParseError{token.LineNo, "Parse error: Multiple spec headings found in same file", token.LineText}}}
		}

		spec.AddHeading(&gauge.Heading{LineNo: token.LineNo, Value: token.Value})
		addStates(state, specScope)
		return ParseResult{Ok: true}
	})

	scenarioConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.ScenarioKind
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		if spec.Heading == nil {
			return ParseResult{Ok: false, ParseErrors: []*ParseError{&ParseError{token.LineNo, "Parse error: Scenario should be defined after the spec heading", token.LineText}}}
		}
		for _, scenario := range spec.Scenarios {
			if strings.ToLower(scenario.Heading.Value) == strings.ToLower(token.Value) {
				return ParseResult{Ok: false, ParseErrors: []*ParseError{&ParseError{token.LineNo, "Parse error: Duplicate scenario definition '" + scenario.Heading.Value + "' found in the same specification", token.LineText}}}
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
		if parseDetails != nil && len(parseDetails.Errors) > 0 {
			return ParseResult{ParseErrors: parseDetails.Errors, Ok: false, Warnings: parseDetails.Warnings}
		}
		latestScenario.AddStep(stepToAdd)
		retainStates(state, specScope, scenarioScope)
		addStates(state, stepScope)
		if parseDetails.Warnings != nil {
			return ParseResult{Ok: false, Warnings: parseDetails.Warnings}
		}
		return ParseResult{Ok: true, Warnings: parseDetails.Warnings}
	})

	contextConverter := converterFn(func(token *Token, state *int) bool {
		return token.Kind == gauge.StepKind && !isInState(*state, scenarioScope) && isInState(*state, specScope) && !isInState(*state, tearDownScope)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		stepToAdd, parseDetails := createStep(spec, token)
		if parseDetails != nil && len(parseDetails.Errors) > 0 {
			return ParseResult{ParseErrors: parseDetails.Errors, Ok: false, Warnings: parseDetails.Warnings}
		}
		spec.AddContext(stepToAdd)
		retainStates(state, specScope)
		addStates(state, contextScope)
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
		if parseDetails != nil && len(parseDetails.Errors) > 0 {
			return ParseResult{ParseErrors: parseDetails.Errors, Ok: false, Warnings: parseDetails.Warnings}
		}
		spec.TearDownSteps = append(spec.TearDownSteps, stepToAdd)
		spec.AddItem(stepToAdd)
		retainStates(state, specScope, tearDownScope)

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
			e := &ParseError{LineNo: parser.lineNo, LineText: token.Value, Message: fmt.Sprintf("Could not resolve table from %s", token.LineText)}
			return ParseResult{ParseErrors: []*ParseError{e}, Ok: false}
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
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{value, token.LineNo}}}
		} else {
			value := "Data table not associated with spec"
			spec.AddComment(&gauge.Comment{token.LineText, token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{value, token.LineNo}}}
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
				return ParseResult{Ok: false, Warnings: []*Warning{&Warning{value, token.LineNo}}}
			}
		} else {
			value := "Table not associated with a step, ignoring table"
			spec.LatestScenario().AddComment(&gauge.Comment{token.LineText, token.LineNo})
			return ParseResult{Ok: false, Warnings: []*Warning{&Warning{value, token.LineNo}}}
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
			result = addInlineTableRow(latestStep, token, new(gauge.ArgLookup).FromDataTable(&spec.DataTable.Table))
		} else if isInState(*state, contextScope) {
			latestContext := spec.LatestContext()
			result = addInlineTableRow(latestContext, token, new(gauge.ArgLookup).FromDataTable(&spec.DataTable.Table))
		} else if isInState(*state, tearDownScope) {
			if len(spec.TearDownSteps) > 0 {
				latestTeardown := spec.LatestTeardown()
				result = addInlineTableRow(latestTeardown, token, new(gauge.ArgLookup).FromDataTable(&spec.DataTable.Table))
			} else {
				spec.AddComment(&gauge.Comment{token.LineText, token.LineNo})
			}
		} else {
			//todo validate datatable rows also
			spec.DataTable.Table.AddRowValues(token.Args)
			result = ParseResult{Ok: true}
		}
		retainStates(state, specScope, scenarioScope, stepScope, contextScope, tearDownScope, tableScope, tableSeparatorScope)
		return result
	})

	tagConverter := converterFn(func(token *Token, state *int) bool {
		return (token.Kind == gauge.TagKind)
	}, func(token *Token, spec *gauge.Specification, state *int) ParseResult {
		tags := &gauge.Tags{Values: token.Args}
		if isInState(*state, scenarioScope) {
			if spec.LatestScenario().NTags() != 0 {
				return ParseResult{Ok: false, ParseErrors: []*ParseError{&ParseError{LineNo: token.LineNo, Message: "Tags can be defined only once per scenario", LineText: token.LineText}}}
			}
			spec.LatestScenario().AddTags(tags)
		} else {
			if spec.NTags() != 0 {
				return ParseResult{Ok: false, ParseErrors: []*ParseError{&ParseError{LineNo: token.LineNo, Message: "Tags can be defined only once per specification", LineText: token.LineText}}}
			}
			spec.AddTags(tags)
		}
		return ParseResult{Ok: true}
	})

	converter := []func(*Token, *int, *gauge.Specification) ParseResult{
		specConverter, scenarioConverter, stepConverter, contextConverter, commentConverter, tableHeaderConverter, tableRowConverter, tagConverter, keywordConverter, tearDownConverter, tearDownStepConverter,
	}

	return converter
}

func (parser *SpecParser) validateSpec(specification *gauge.Specification) *ParseError {
	if len(specification.Items) == 0 {
		return &ParseError{LineNo: 1, Message: "Spec does not have any elements"}
	}
	if specification.Heading == nil {
		return &ParseError{LineNo: 1, Message: "Spec heading not found"}
	}
	if len(strings.TrimSpace(specification.Heading.Value)) < 1 {
		return &ParseError{LineNo: specification.Heading.LineNo, Message: "Spec heading should have at least one character"}
	}

	dataTable := specification.DataTable.Table
	if dataTable.IsInitialized() && dataTable.GetRowCount() == 0 {
		return &ParseError{LineNo: dataTable.LineNo, Message: "Data table should have at least 1 data row"}
	}
	return nil
}

func CreateStepFromStepRequest(stepReq *gauge_messages.ExecuteStepRequest) *gauge.Step {
	args := createStepArgsFromProtoArguments(stepReq.GetParameters())
	step := &gauge.Step{Value: stepReq.GetParsedStepText(),
		LineText: stepReq.GetActualStepText()}
	step.AddArgs(args...)
	return step
}

func createStepArgsFromProtoArguments(parameters []*gauge_messages.Parameter) []*gauge.StepArg {
	stepArgs := make([]*gauge.StepArg, 0)
	for _, parameter := range parameters {
		var arg *gauge.StepArg
		switch parameter.GetParameterType() {
		case gauge_messages.Parameter_Static:
			arg = &gauge.StepArg{ArgType: gauge.Static, Value: parameter.GetValue(), Name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Dynamic:
			arg = &gauge.StepArg{ArgType: gauge.Dynamic, Value: parameter.GetValue(), Name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Special_String:
			arg = &gauge.StepArg{ArgType: gauge.SpecialString, Value: parameter.GetValue(), Name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Table:
			arg = &gauge.StepArg{ArgType: gauge.TableArg, Table: *(TableFrom(parameter.GetTable())), Name: parameter.GetName()}
			break
		case gauge_messages.Parameter_Special_Table:
			arg = &gauge.StepArg{ArgType: gauge.SpecialTable, Table: *(TableFrom(parameter.GetTable())), Name: parameter.GetName()}
			break
		}
		stepArgs = append(stepArgs, arg)
	}
	return stepArgs
}

func converterFn(predicate func(token *Token, state *int) bool, apply func(token *Token, spec *gauge.Specification, state *int) ParseResult) func(*Token, *int, *gauge.Specification) ParseResult {
	return func(token *Token, state *int, spec *gauge.Specification) ParseResult {
		if !predicate(token, state) {
			return ParseResult{Ok: true}
		}
		return apply(token, spec, state)
	}
}

func createStep(spec *gauge.Specification, stepToken *Token) (*gauge.Step, *ParseDetailResult) {
	dataTableLookup := new(gauge.ArgLookup).FromDataTable(&spec.DataTable.Table)
	stepToAdd, parseDetails := CreateStepUsingLookup(stepToken, dataTableLookup)

	if parseDetails != nil && len(parseDetails.Errors) > 0 {
		return nil, parseDetails
	}
	stepToAdd.Suffix = stepToken.Suffix
	return stepToAdd, parseDetails
}

func CreateStepUsingLookup(stepToken *Token, lookup *gauge.ArgLookup) (*gauge.Step, *ParseDetailResult) {
	stepValue, argsType := extractStepValueAndParameterTypes(stepToken.Value)
	if argsType != nil && len(argsType) != len(stepToken.Args) {
		return nil, &ParseDetailResult{Errors: []*ParseError{&ParseError{stepToken.LineNo, "Step text should not have '{static}' or '{dynamic}' or '{special}'", stepToken.LineText}}, Warnings: nil}
	}
	step := &gauge.Step{LineNo: stepToken.LineNo, Value: stepValue, LineText: strings.TrimSpace(stepToken.LineText)}
	arguments := make([]*gauge.StepArg, 0)
	var warnings []*Warning
	for i, argType := range argsType {
		argument, parseDetails := createStepArg(stepToken.Args[i], argType, stepToken, lookup)
		if parseDetails != nil && len(parseDetails.Errors) > 0 {
			return nil, parseDetails
		}
		arguments = append(arguments, argument)
		if parseDetails != nil && parseDetails.Warnings != nil {
			for _, warn := range parseDetails.Warnings {
				warnings = append(warnings, warn)
			}
		}
	}
	step.AddArgs(arguments...)
	return step, &ParseDetailResult{Warnings: warnings}
}

func createConceptStep(spec *gauge.Specification, concept *gauge.Step, originalStep *gauge.Step) {
	stepCopy := concept.GetCopy()
	originalArgs := originalStep.Args
	originalStep.CopyFrom(stepCopy)
	originalStep.Args = originalArgs

	// set parent of all concept steps to be the current concept (referred as originalStep here)
	// this is used to fetch from parent's lookup when nested
	for _, conceptStep := range originalStep.ConceptSteps {
		conceptStep.Parent = originalStep
	}

	spec.PopulateConceptLookup(&originalStep.Lookup, concept.Args, originalStep.Args)
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

func createStepArg(argValue string, typeOfArg string, token *Token, lookup *gauge.ArgLookup) (*gauge.StepArg, *ParseDetailResult) {
	if typeOfArg == "special" {
		resolvedArgValue, err := newSpecialTypeResolver().resolve(argValue)
		if err != nil {
			switch err.(type) {
			case invalidSpecialParamError:
				return treatArgAsDynamic(argValue, token, lookup)
			default:
				return nil, &ParseDetailResult{Errors: []*ParseError{&ParseError{LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), LineText: token.LineText}}}
			}
		}
		return resolvedArgValue, nil
	} else if typeOfArg == "static" {
		return &gauge.StepArg{ArgType: gauge.Static, Value: argValue}, nil
	} else {
		return validateDynamicArg(argValue, token, lookup)
	}
}

func treatArgAsDynamic(argValue string, token *Token, lookup *gauge.ArgLookup) (*gauge.StepArg, *ParseDetailResult) {
	parseDetailRes := &ParseDetailResult{Warnings: []*Warning{&Warning{LineNo: token.LineNo, Message: fmt.Sprintf("Could not resolve special param type <%s>. Treating it as dynamic param.", argValue)}}}
	stepArg, result := validateDynamicArg(argValue, token, lookup)
	if result != nil {
		if len(result.Errors) > 0 {
			parseDetailRes.Errors = result.Errors
		}
		if result.Warnings != nil {
			for _, warn := range result.Warnings {
				parseDetailRes.Warnings = append(parseDetailRes.Warnings, warn)
			}
		}
	}
	return stepArg, parseDetailRes
}

func validateDynamicArg(argValue string, token *Token, lookup *gauge.ArgLookup) (*gauge.StepArg, *ParseDetailResult) {
	if !isConceptHeader(lookup) && !lookup.ContainsArg(argValue) {
		return nil, &ParseDetailResult{Errors: []*ParseError{&ParseError{LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic parameter <%s> could not be resolved", argValue), LineText: token.LineText}}}
	}
	stepArgument := &gauge.StepArg{ArgType: gauge.Dynamic, Value: argValue, Name: argValue}
	return stepArgument, nil
}

//Step value is modified when inline table is found to account for the new parameter by appending {}
//todo validate headers for dynamic
func addInlineTableHeader(step *gauge.Step, token *Token) {
	step.Value = fmt.Sprintf("%s %s", step.Value, gauge.ParameterPlaceholder)
	step.HasInlineTable = true
	step.AddInlineTableHeaders(token.Args)
}

func addInlineTableRow(step *gauge.Step, token *Token, argLookup *gauge.ArgLookup) ParseResult {
	dynamicArgMatcher := regexp.MustCompile("^<(.*)>$")
	tableValues := make([]gauge.TableCell, 0)
	warnings := make([]*Warning, 0)
	for _, tableValue := range token.Args {
		if dynamicArgMatcher.MatchString(tableValue) {
			match := dynamicArgMatcher.FindAllStringSubmatch(tableValue, -1)
			param := match[0][1]
			if !argLookup.ContainsArg(param) {
				tableValues = append(tableValues, gauge.TableCell{Value: tableValue, CellType: gauge.Static})
				warnings = append(warnings, &Warning{LineNo: token.LineNo, Message: fmt.Sprintf("Dynamic param <%s> could not be resolved, Treating it as static param", param)})
			} else {
				tableValues = append(tableValues, gauge.TableCell{Value: param, CellType: gauge.Dynamic})
			}
		} else {
			tableValues = append(tableValues, gauge.TableCell{Value: tableValue, CellType: gauge.Static})
		}
	}
	step.AddInlineTableRow(tableValues)
	return ParseResult{Ok: true, Warnings: warnings}
}

func ConvertToStepText(fragments []*gauge_messages.Fragment) string {
	stepText := ""
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
	LineNo   int
	Message  string
	LineText string
}

type ParseDetailResult struct {
	Errors   []*ParseError
	Warnings []*Warning
}

func (se *ParseError) Error() string {
	return fmt.Sprintf("%d: %s => '%s'", se.LineNo, se.Message, se.LineText)
}

func (token *Token) String() string {
	return fmt.Sprintf("kind:%d, lineNo:%d, value:%s, line:%s, args:%s", token.Kind, token.LineNo, token.Value, token.LineText, token.Args)
}

type ParseResult struct {
	ParseErrors []*ParseError
	Warnings    []*Warning
	Ok          bool
	FileName    string
}

func (result *ParseResult) Errors() []string {
	var errors []string
	for _, err := range result.ParseErrors {
		errors = append(errors, fmt.Sprintf("[ParseError] %s:%s", result.FileName, err.Error()))
	}
	return errors
}

type Warning struct {
	Message string
	LineNo  int
}

func (warning *Warning) String() string {
	return fmt.Sprintf("line no: %d, %s", warning.LineNo, warning.Message)
}
