/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"bufio"
	"strings"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
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
		return ParseError{FileName: specification.FileName, LineNo: 1, SpanEnd: 1, Message: "Spec does not have any elements"}
	}
	if specification.Heading == nil {
		specification.AddHeading(&gauge.Heading{})
		return ParseError{FileName: specification.FileName, LineNo: 1, SpanEnd: 1, Message: "Spec heading not found"}
	}
	if len(strings.TrimSpace(specification.Heading.Value)) < 1 {
		return ParseError{FileName: specification.FileName, LineNo: specification.Heading.LineNo, SpanEnd: specification.Heading.LineNo, Message: "Spec heading should have at least one character"}
	}

	dataTable := specification.DataTable.Table
	if dataTable.IsInitialized() && dataTable.GetRowCount() == 0 {
		return ParseError{FileName: specification.FileName, LineNo: dataTable.LineNo, SpanEnd: dataTable.LineNo, Message: "Data table should have at least 1 data row"}
	}
	if len(specification.Scenarios) == 0 {
		return ParseError{FileName: specification.FileName, LineNo: specification.Heading.LineNo, SpanEnd: specification.Heading.SpanEnd, Message: "Spec should have at least one scenario"}
	}
	for _, sce := range specification.Scenarios {
		if len(sce.Steps) == 0 {
			return ParseError{FileName: specification.FileName, LineNo: sce.Heading.LineNo, SpanEnd: sce.Heading.SpanEnd, Message: "Scenario should have at least one step"}
		}
	}
	return nil
}

func createStep(spec *gauge.Specification, scn *gauge.Scenario, stepToken *Token) (*gauge.Step, *ParseResult) {
	tables := []*gauge.Table{spec.DataTable.Table}
	if scn != nil {
		tables = append(tables, scn.DataTable.Table)
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
	
	// Handle implicit multiline argument - if we have args but no parameter types
	hasImplicitMultiline := len(stepToken.Args) > 0 && len(argsType) == 0
	
	// Only validate if we don't have an implicit multiline argument
	if !hasImplicitMultiline && argsType != nil && len(argsType) != len(stepToken.Args) {
		return nil, &ParseResult{ParseErrors: []ParseError{ParseError{specFileName, stepToken.LineNo, stepToken.SpanEnd, "Step text should not have '{static}' or '{dynamic}' or '{special}'", stepToken.LineText()}}, Warnings: nil}
	}

	lineText := strings.Join(stepToken.Lines, " ")
	step := &gauge.Step{FileName: specFileName, LineNo: stepToken.LineNo, Value: stepValue, LineText: strings.TrimSpace(lineText), LineSpanEnd: stepToken.SpanEnd}
	arguments := make([]*gauge.StepArg, 0)
	var errors []ParseError
	var warnings []*Warning
	
	// Handle implicit multiline argument
	if hasImplicitMultiline {
		arguments = append(arguments, &gauge.StepArg{ArgType: gauge.SpecialString, Value: stepToken.Args[0]})
	} else {
		// Handle regular arguments
		for i, argType := range argsType {
			argument, parseDetails := createStepArg(stepToken.Args[i], argType, stepToken, lookup, specFileName)
			if parseDetails != nil && len(parseDetails.ParseErrors) > 0 {
				errors = append(errors, parseDetails.ParseErrors...)
			}
			arguments = append(arguments, argument)
			if parseDetails != nil && parseDetails.Warnings != nil {
				warnings = append(warnings, parseDetails.Warnings...)
			}
		}
	}
	
	step.AddArgs(arguments...)

	// Create fragments for implicit multiline argument
	if hasImplicitMultiline {
		// First fragment: the step text
		textFragment := &gauge_messages.Fragment{
			FragmentType: gauge_messages.Fragment_Text,
			Text: step.Value,
		}
		
		// Second fragment: the multiline parameter
		param := &gauge_messages.Parameter{
			ParameterType: gauge_messages.Parameter_Multiline_String,
			Value: stepToken.Args[0],
			Name: "",
		}
		paramFragment := &gauge_messages.Fragment{
			FragmentType: gauge_messages.Fragment_Parameter,
			Parameter: param,
		}
		
		// Replace the fragments array with both fragments
		step.Fragments = []*gauge_messages.Fragment{textFragment, paramFragment}
	}
	
	return step, &ParseResult{ParseErrors: errors, Warnings: warnings}
}