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
	"strings"

	"github.com/getgauge/gauge/gauge"
)

// SpecParser is responsible for parsing a Specification. It delegates to respective processors composed sub-entities
type SpecParser struct {
	scanner           *bufio.Scanner
	lineNo            int
	tokens            []*Token
	currentState      int
	processors        map[gauge.TokenKind]func(*SpecParser, *Token) ([]error, bool)
	conceptDictionary *gauge.ConceptDictionary
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

func createStep(spec *gauge.Specification, scn *gauge.Scenario, stepToken *Token) (*gauge.Step, *ParseResult) {
	tables := []*gauge.Table{&spec.DataTable.Table}
	if scn != nil {
		tables = append(tables, &scn.DataTable.Table)
	}
	dataTableLookup := new(gauge.ArgLookup).FromDataTables(tables...)
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
