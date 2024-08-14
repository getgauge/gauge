/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"fmt"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
)

// ConceptParser is used for parsing concepts. Similar, but not the same as a SpecParser
type ConceptParser struct {
	currentState   int
	currentConcept *gauge.Step
}

// Parse Generates token for the given concept file and cretes concepts(array of steps) and parse results.
// concept file can have multiple concept headings.
func (parser *ConceptParser) Parse(text, fileName string) ([]*gauge.Step, *ParseResult) {
	defer parser.resetState()

	specParser := new(SpecParser)
	tokens, errs := specParser.GenerateTokens(text, fileName)
	concepts, res := parser.createConcepts(tokens, fileName)
	return concepts, &ParseResult{ParseErrors: append(errs, res.ParseErrors...), Warnings: res.Warnings}
}

// ParseFile Reads file contents from a give file and parses the file.
func (parser *ConceptParser) ParseFile(file string) ([]*gauge.Step, *ParseResult) {
	fileText, fileReadErr := common.ReadFileContents(file)
	if fileReadErr != nil {
		return nil, &ParseResult{ParseErrors: []ParseError{{Message: fmt.Sprintf("failed to read concept file %s", file)}}}
	}
	return parser.Parse(fileText, file)
}

func (parser *ConceptParser) resetState() {
	parser.currentState = initial
	parser.currentConcept = nil
}

func (parser *ConceptParser) createConcepts(tokens []*Token, fileName string) ([]*gauge.Step, *ParseResult) {
	parser.currentState = initial
	var concepts []*gauge.Step
	parseRes := &ParseResult{ParseErrors: make([]ParseError, 0)}
	var preComments []*gauge.Comment
	addPreComments := false
	for _, token := range tokens {
		if parser.isConceptHeading(token) {
			if isInState(parser.currentState, conceptScope, stepScope) {
				if len(parser.currentConcept.ConceptSteps) < 1 {
					parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: parser.currentConcept.LineNo, SpanEnd: parser.currentConcept.LineSpanEnd, Message: "Concept should have atleast one step", LineText: parser.currentConcept.LineText})
					continue
				}
				concepts = append(concepts, parser.currentConcept)
			}
			var res *ParseResult
			parser.currentConcept, res = parser.processConceptHeading(token, fileName)
			parser.currentState = initial
			if len(res.ParseErrors) > 0 {
				parseRes.ParseErrors = append(parseRes.ParseErrors, res.ParseErrors...)
				continue
			}
			if addPreComments {
				parser.currentConcept.PreComments = preComments
				addPreComments = false
			}
			addStates(&parser.currentState, conceptScope)
		} else if parser.isStep(token) {
			if !isInState(parser.currentState, conceptScope) {
				parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: token.LineNo, SpanEnd: token.SpanEnd, Message: "Step is not defined inside a concept heading", LineText: token.LineText()})
				continue
			}
			if errs := parser.processConceptStep(token, fileName); len(errs) > 0 {
				parseRes.ParseErrors = append(parseRes.ParseErrors, errs...)
			}
			addStates(&parser.currentState, stepScope)
		} else if parser.isTableHeader(token) {
			if !isInState(parser.currentState, stepScope) {
				parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: token.LineNo, SpanEnd: token.SpanEnd, Message: "Table doesn't belong to any step", LineText: token.LineText()})
				continue
			}
			parser.processTableHeader(token)
			addStates(&parser.currentState, tableScope)
		} else if parser.isScenarioHeading(token) {
			parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: token.LineNo, SpanEnd: token.SpanEnd, Message: "Scenario Heading is not allowed in concept file", LineText: token.LineText()})
			continue
		} else if parser.isTableDataRow(token) {
			if areUnderlined(token.Args) && !isInState(parser.currentState, tableSeparatorScope) {
				addStates(&parser.currentState, tableSeparatorScope)
			} else if isInState(parser.currentState, stepScope) {
				parser.processTableDataRow(token, &parser.currentConcept.Lookup, fileName)
			}
		} else {
			retainStates(&parser.currentState, conceptScope)
			addStates(&parser.currentState, commentScope)
			comment := &gauge.Comment{Value: token.Value, LineNo: token.LineNo}
			if parser.currentConcept == nil {
				preComments = append(preComments, comment)
				addPreComments = true
				continue
			}
			parser.currentConcept.Items = append(parser.currentConcept.Items, comment)
		}
	}
	if parser.currentConcept != nil && len(parser.currentConcept.ConceptSteps) < 1 {
		parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: parser.currentConcept.LineNo, SpanEnd: parser.currentConcept.LineSpanEnd, Message: "Concept should have atleast one step", LineText: parser.currentConcept.LineText})
		return nil, parseRes
	}

	if parser.currentConcept != nil {
		concepts = append(concepts, parser.currentConcept)
	}
	return concepts, parseRes
}

func (parser *ConceptParser) isConceptHeading(token *Token) bool {
	return token.Kind == gauge.SpecKind
}

func (parser *ConceptParser) isStep(token *Token) bool {
	return token.Kind == gauge.StepKind
}

func (parser *ConceptParser) isScenarioHeading(token *Token) bool {
	return token.Kind == gauge.ScenarioKind
}

func (parser *ConceptParser) isTableHeader(token *Token) bool {
	return token.Kind == gauge.TableHeader
}

func (parser *ConceptParser) isTableDataRow(token *Token) bool {
	return token.Kind == gauge.TableRow
}

func (parser *ConceptParser) processConceptHeading(token *Token, fileName string) (*gauge.Step, *ParseResult) {
	processStep(new(SpecParser), token)
	token.Lines[0] = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(token.Lines[0]), "#"))
	var concept *gauge.Step
	var parseRes *ParseResult
	concept, parseRes = CreateStepUsingLookup(token, nil, fileName)
	if parseRes != nil && len(parseRes.ParseErrors) > 0 {
		return nil, parseRes
	}
	if !parser.hasOnlyDynamicParams(concept) {
		parseRes.ParseErrors = []ParseError{ParseError{FileName: fileName, LineNo: token.LineNo, SpanEnd: token.SpanEnd, Message: "Concept heading can have only Dynamic Parameters", LineText: token.LineText()}}
		return nil, parseRes
	}

	concept.IsConcept = true
	parser.createConceptLookup(concept)
	concept.Items = append(concept.Items, concept)
	return concept, parseRes
}

func (parser *ConceptParser) processConceptStep(token *Token, fileName string) []ParseError {
	processStep(new(SpecParser), token)
	conceptStep, parseRes := CreateStepUsingLookup(token, &parser.currentConcept.Lookup, fileName)
	if conceptStep != nil {
		conceptStep.Suffix = token.Suffix
		parser.currentConcept.ConceptSteps = append(parser.currentConcept.ConceptSteps, conceptStep)
		parser.currentConcept.Items = append(parser.currentConcept.Items, conceptStep)
	}
	return parseRes.ParseErrors
}

func (parser *ConceptParser) processTableHeader(token *Token) {
	steps := parser.currentConcept.ConceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableHeader(currentStep, token)
	items := parser.currentConcept.Items
	items[len(items)-1] = currentStep
}

func (parser *ConceptParser) processTableDataRow(token *Token, argLookup *gauge.ArgLookup, fileName string) {
	steps := parser.currentConcept.ConceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableRow(currentStep, token, argLookup, fileName)
	items := parser.currentConcept.Items
	items[len(items)-1] = currentStep
}

func (parser *ConceptParser) hasOnlyDynamicParams(step *gauge.Step) bool {
	for _, arg := range step.Args {
		if arg.ArgType != gauge.Dynamic {
			return false
		}
	}
	return true
}

func (parser *ConceptParser) createConceptLookup(concept *gauge.Step) {
	for _, arg := range concept.Args {
		concept.Lookup.AddArgName(arg.Value)
	}
}

// CreateConceptsDictionary generates a ConceptDictionary which is map of concept text to concept. ConceptDictionary is used to search for a concept.
func CreateConceptsDictionary() (*gauge.ConceptDictionary, *ParseResult, error) {
	cptFilesMap := make(map[string]bool)
	for _, cpt := range util.GetConceptFiles() {
		cptFilesMap[cpt] = true
	}
	var conceptFiles []string
	for cpt := range cptFilesMap {
		conceptFiles = append(conceptFiles, cpt)
	}
	conceptsDictionary := gauge.NewConceptDictionary()
	res := &ParseResult{Ok: true}
	if _, errs, e := AddConcepts(conceptFiles, conceptsDictionary); len(errs) > 0 {
		if e != nil {
			return nil, nil, e
		}
		for _, err := range errs {
			logger.Errorf(false, "Concept parse failure: %s %s", conceptFiles[0], err)
		}
		res.ParseErrors = append(res.ParseErrors, errs...)
		res.Ok = false
	}
	vRes := ValidateConcepts(conceptsDictionary)
	if len(vRes.ParseErrors) > 0 {
		res.Ok = false
		res.ParseErrors = append(res.ParseErrors, vRes.ParseErrors...)
	}
	return conceptsDictionary, res, nil
}

// AddConcept adds the concept in the ConceptDictionary.
func AddConcept(concepts []*gauge.Step, file string, conceptDictionary *gauge.ConceptDictionary) ([]ParseError, error) {
	parseErrors := make([]ParseError, 0)
	for _, conceptStep := range concepts {
		if dupConcept, exists := conceptDictionary.ConceptsMap[conceptStep.Value]; exists {
			parseErrors = append(parseErrors, ParseError{
				FileName: file,
				LineNo:   conceptStep.LineNo,
				SpanEnd:  conceptStep.LineSpanEnd,
				Message:  "Duplicate concept definition found",
				LineText: conceptStep.LineText,
			},
				ParseError{
					FileName: dupConcept.FileName,
					LineNo:   dupConcept.ConceptStep.LineNo,
					SpanEnd:  conceptStep.LineSpanEnd,
					Message:  "Duplicate concept definition found",
					LineText: dupConcept.ConceptStep.LineText,
				})
		}
		conceptDictionary.ConceptsMap[conceptStep.Value] = &gauge.Concept{ConceptStep: conceptStep, FileName: file}
		if err := conceptDictionary.ReplaceNestedConceptSteps(conceptStep); err != nil {
			return nil, err
		}
	}
	err := conceptDictionary.UpdateLookupForNestedConcepts()
	return parseErrors, err
}

// AddConcepts parses the given concept file and adds each concept to the concept dictionary.
func AddConcepts(conceptFiles []string, conceptDictionary *gauge.ConceptDictionary) ([]*gauge.Step, []ParseError, error) {
	var conceptSteps []*gauge.Step
	var parseResults []*ParseResult
	for _, conceptFile := range conceptFiles {
		concepts, parseRes := new(ConceptParser).ParseFile(conceptFile)
		if parseRes != nil && parseRes.Warnings != nil {
			for _, warning := range parseRes.Warnings {
				logger.Warning(true, warning.String())
			}
		}
		parseErrors, err := AddConcept(concepts, conceptFile, conceptDictionary)
		if err != nil {
			return nil, nil, err
		}
		parseRes.ParseErrors = append(parseRes.ParseErrors, parseErrors...)
		conceptSteps = append(conceptSteps, concepts...)
		parseResults = append(parseResults, parseRes)
	}
	errs := collectAllParseErrors(parseResults)
	return conceptSteps, errs, nil
}

func collectAllParseErrors(results []*ParseResult) (errs []ParseError) {
	for _, res := range results {
		errs = append(errs, res.ParseErrors...)
	}
	return
}

// ValidateConcepts ensures that there are no circular references within
func ValidateConcepts(conceptDictionary *gauge.ConceptDictionary) *ParseResult {
	res := &ParseResult{ParseErrors: []ParseError{}}
	var conceptsWithError []*gauge.Concept
	for _, concept := range conceptDictionary.ConceptsMap {
		errs := checkCircularReferencing(conceptDictionary, concept.ConceptStep, nil)
		if errs != nil {
			delete(conceptDictionary.ConceptsMap, concept.ConceptStep.Value)
			res.ParseErrors = append(res.ParseErrors, errs...)
			conceptsWithError = append(conceptsWithError, concept)
		}
	}
	for _, con := range conceptsWithError {
		removeAllReferences(conceptDictionary, con)
	}
	return res
}

func removeAllReferences(conceptDictionary *gauge.ConceptDictionary, concept *gauge.Concept) {
	for _, cpt := range conceptDictionary.ConceptsMap {
		var nestedSteps []*gauge.Step
		for _, con := range cpt.ConceptStep.ConceptSteps {
			if con.Value != concept.ConceptStep.Value {
				nestedSteps = append(nestedSteps, con)
			}
		}
		cpt.ConceptStep.ConceptSteps = nestedSteps
	}
}

func checkCircularReferencing(conceptDictionary *gauge.ConceptDictionary, concept *gauge.Step, traversedSteps map[string]string) []ParseError {
	if traversedSteps == nil {
		traversedSteps = make(map[string]string)
	}
	con := conceptDictionary.Search(concept.Value)
	if con == nil {
		return nil
	}
	currentConceptFileName := con.FileName
	traversedSteps[concept.Value] = currentConceptFileName
	for _, step := range concept.ConceptSteps {
		if _, exists := traversedSteps[step.Value]; exists {
			conceptDictionary.Remove(concept.Value)
			return []ParseError{
				{
					FileName: step.FileName,
					LineText: step.LineText,
					LineNo:   step.LineNo,
					SpanEnd:  step.LineSpanEnd,
					Message:  fmt.Sprintf("Circular reference found in concept. \"%s\" => %s:%d", concept.LineText, concept.FileName, concept.LineNo),
				},
				{
					FileName: concept.FileName,
					LineText: concept.LineText,
					LineNo:   concept.LineNo,
					SpanEnd:  step.LineSpanEnd,
					Message:  fmt.Sprintf("Circular reference found in concept. \"%s\" => %s:%d", step.LineText, step.FileName, step.LineNo),
				},
			}
		}
		if step.IsConcept {
			if errs := checkCircularReferencing(conceptDictionary, step, traversedSteps); errs != nil {
				conceptDictionary.Remove(concept.Value)
				return errs
			}
		}
	}
	delete(traversedSteps, concept.Value)
	return nil
}
