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
	"fmt"
	"os"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
)

type ConceptParser struct {
	currentState   int
	currentConcept *gauge.Step
}

//concept file can have multiple concept headings
func (parser *ConceptParser) Parse(text, fileName string) ([]*gauge.Step, *ParseResult) {
	defer parser.resetState()

	specParser := new(SpecParser)
	tokens, errs := specParser.GenerateTokens(text, fileName)
	concepts, res := parser.createConcepts(tokens, fileName)
	return concepts, &ParseResult{ParseErrors: append(errs, res.ParseErrors...), Warnings: res.Warnings}
}

func (parser *ConceptParser) ParseFile(file string) ([]*gauge.Step, *ParseResult) {
	fileText, fileReadErr := common.ReadFileContents(file)
	if fileReadErr != nil {
		return nil, &ParseResult{ParseErrors: []*ParseError{&ParseError{Message: fmt.Sprintf("failed to read concept file %s", file)}}}
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
	parseRes := &ParseResult{ParseErrors: make([]*ParseError, 0)}
	var preComments []*gauge.Comment
	addPreComments := false
	for _, token := range tokens {
		if parser.isConceptHeading(token) {
			if isInState(parser.currentState, conceptScope, stepScope) {
				concepts = append(concepts, parser.currentConcept)
			}
			parser.currentConcept, parseRes = parser.processConceptHeading(token, fileName)
			parser.currentState = initial
			if len(parseRes.ParseErrors) > 0 {
				continue
			}
			if addPreComments {
				parser.currentConcept.PreComments = preComments
				addPreComments = false
			}
			addStates(&parser.currentState, conceptScope)
		} else if parser.isStep(token) {
			if !isInState(parser.currentState, conceptScope) {
				parseRes.ParseErrors = append(parseRes.ParseErrors, &ParseError{FileName: fileName, LineNo: token.LineNo, Message: "Step is not defined inside a concept heading", LineText: token.LineText})
				continue
			}
			if errs := parser.processConceptStep(token, fileName); len(errs) > 0 {
				parseRes.ParseErrors = append(parseRes.ParseErrors, errs...)
				continue
			}
			addStates(&parser.currentState, stepScope)
		} else if parser.isTableHeader(token) {
			if !isInState(parser.currentState, stepScope) {
				parseRes.ParseErrors = append(parseRes.ParseErrors, &ParseError{FileName: fileName, LineNo: token.LineNo, Message: "Table doesn't belong to any step", LineText: token.LineText})
				continue
			}
			parser.processTableHeader(token)
			addStates(&parser.currentState, tableScope)
		} else if parser.isScenarioHeading(token) {
			parseRes.ParseErrors = append(parseRes.ParseErrors, &ParseError{FileName: fileName, LineNo: token.LineNo, Message: "Scenario Heading is not allowed in concept file", LineText: token.LineText})
			continue
		} else if parser.isTableDataRow(token) {
			if isInState(parser.currentState, stepScope) {
				parser.processTableDataRow(token, &parser.currentConcept.Lookup, fileName)
			}
		} else {
			comment := &gauge.Comment{Value: token.Value, LineNo: token.LineNo}
			if parser.currentConcept == nil {
				preComments = append(preComments, comment)
				addPreComments = true
				continue
			}
			parser.currentConcept.Items = append(parser.currentConcept.Items, comment)
		}
	}
	if !isInState(parser.currentState, stepScope) && parser.currentState != initial {
		parseRes.ParseErrors = append(parseRes.ParseErrors, &ParseError{FileName: fileName, LineNo: parser.currentConcept.LineNo, Message: "Concept should have atleast one step", LineText: parser.currentConcept.LineText})
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
	token.LineText = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(token.LineText), "#"))
	var concept *gauge.Step
	var parseRes *ParseResult
	concept, parseRes = CreateStepUsingLookup(token, nil, fileName)
	if parseRes != nil && len(parseRes.ParseErrors) > 0 {
		return nil, parseRes
	}
	if !parser.hasOnlyDynamicParams(concept) {
		parseRes.ParseErrors = []*ParseError{&ParseError{FileName: fileName, LineNo: token.LineNo, Message: "Concept heading can have only Dynamic Parameters", LineText: token.LineText}}
		return nil, parseRes
	}

	concept.IsConcept = true
	parser.createConceptLookup(concept)
	concept.Items = append(concept.Items, concept)
	return concept, parseRes
}

func (parser *ConceptParser) processConceptStep(token *Token, fileName string) []*ParseError {
	processStep(new(SpecParser), token)
	conceptStep, parseRes := CreateStepUsingLookup(token, &parser.currentConcept.Lookup, fileName)
	if parseRes != nil && len(parseRes.ParseErrors) > 0 {
		return parseRes.ParseErrors
	}
	parser.currentConcept.ConceptSteps = append(parser.currentConcept.ConceptSteps, conceptStep)
	parser.currentConcept.Items = append(parser.currentConcept.Items, conceptStep)
	return nil
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

func CreateConceptsDictionary() (*gauge.ConceptDictionary, *ParseResult) {
	cptFilesMap := make(map[string]bool, 0)
	for _, cpt := range util.GetConceptFiles() {
		cptFilesMap[cpt] = true
	}
	var conceptFiles []string
	for cpt := range cptFilesMap {
		conceptFiles = append(conceptFiles, cpt)
	}
	conceptsDictionary := gauge.NewConceptDictionary()
	res := &ParseResult{Ok: true}
	for _, conceptFile := range conceptFiles {
		if errs := AddConcepts(conceptFile, conceptsDictionary); len(errs) > 0 {
			for _, err := range errs {
				logger.APILog.Error("Concept parse failure: %s %s", conceptFile, err)
			}
			res.ParseErrors = append(res.ParseErrors, errs...)
			res.Ok = false
		}
	}
	vRes := validateConcepts(conceptsDictionary)
	if len(vRes.ParseErrors) > 0 {
		for _, err := range res.ParseErrors {
			logger.Errorf(err.Error())
		}
		for _, err := range vRes.ParseErrors {
			logger.Errorf("%s:%s", vRes.FileName, err.Error())
		}
		os.Exit(1)
	}
	return conceptsDictionary, res
}

func AddConcepts(conceptFile string, conceptDictionary *gauge.ConceptDictionary) []*ParseError {
	concepts, parseRes := new(ConceptParser).ParseFile(conceptFile)
	if parseRes != nil && parseRes.Warnings != nil {
		for _, warning := range parseRes.Warnings {
			logger.Warning(warning.String())
		}
	}
	if parseRes != nil && len(parseRes.ParseErrors) > 0 {
		return parseRes.ParseErrors
	}
	var errs []*ParseError
	for _, conceptStep := range concepts {
		if _, exists := conceptDictionary.ConceptsMap[conceptStep.Value]; exists {
			errs = append(errs, &ParseError{FileName: conceptFile, LineNo: conceptStep.LineNo, Message: "Duplicate concept definition found", LineText: conceptStep.LineText})
		}
		conceptDictionary.ReplaceNestedConceptSteps(conceptStep)
		conceptDictionary.ConceptsMap[conceptStep.Value] = &gauge.Concept{conceptStep, conceptFile}
	}
	conceptDictionary.UpdateLookupForNestedConcepts()
	return errs
}

func validateConcepts(conceptDictionary *gauge.ConceptDictionary) *ParseResult {
	for _, concept := range conceptDictionary.ConceptsMap {
		err := checkCircularReferencing(conceptDictionary, concept.ConceptStep, nil)
		if err != nil {
			return &ParseResult{ParseErrors: []*ParseError{err}, FileName: concept.FileName}
		}
	}
	return &ParseResult{ParseErrors: []*ParseError{}}
}

func checkCircularReferencing(conceptDictionary *gauge.ConceptDictionary, concept *gauge.Step, traversedSteps map[string]string) *ParseError {
	if traversedSteps == nil {
		traversedSteps = make(map[string]string, 0)
	}
	currentConceptFileName := conceptDictionary.Search(concept.Value).FileName
	traversedSteps[concept.Value] = currentConceptFileName
	for _, step := range concept.ConceptSteps {
		if fileName, exists := traversedSteps[step.Value]; exists {
			return &ParseError{
				FileName: fileName,
				LineText: step.LineText,
				LineNo:   concept.LineNo,
				Message:  fmt.Sprintf("Circular reference found in concept. \"%s\" => %s:%d", concept.LineText, fileName, step.LineNo),
			}
		}
		if step.IsConcept {
			if err := checkCircularReferencing(conceptDictionary, step, traversedSteps); err != nil {
				return err
			}
		}
	}
	delete(traversedSteps, concept.Value)
	return nil
}
