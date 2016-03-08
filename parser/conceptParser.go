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
	"path/filepath"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
)

type ConceptParser struct {
	currentState   int
	currentConcept *gauge.Step
}

//concept file can have multiple concept headings
func (parser *ConceptParser) Parse(text string) ([]*gauge.Step, *ParseDetailResult) {
	defer parser.resetState()

	specParser := new(SpecParser)
	tokens, err := specParser.GenerateTokens(text)
	if err != nil {
		return nil, &ParseDetailResult{Error: err}
	}
	return parser.createConcepts(tokens)
}

func (parser *ConceptParser) ParseFile(file string) ([]*gauge.Step, *ParseDetailResult) {
	fileText, fileReadErr := common.ReadFileContents(file)
	if fileReadErr != nil {
		return nil, &ParseDetailResult{Error: &ParseError{Message: fmt.Sprintf("failed to read concept file %s", file)}}
	}
	return parser.Parse(fileText)
}

func (parser *ConceptParser) resetState() {
	parser.currentState = initial
	parser.currentConcept = nil
}

func (parser *ConceptParser) createConcepts(tokens []*Token) ([]*gauge.Step, *ParseDetailResult) {
	parser.currentState = initial
	concepts := make([]*gauge.Step, 0)
	var parseDetails *ParseDetailResult
	preComments := make([]*gauge.Comment, 0)
	addPreComments := false
	for _, token := range tokens {
		if parser.isConceptHeading(token) {
			if isInState(parser.currentState, conceptScope, stepScope) {
				concepts = append(concepts, parser.currentConcept)
			}
			parser.currentConcept, parseDetails = parser.processConceptHeading(token)
			if parseDetails.Error != nil {
				return nil, parseDetails
			}
			if addPreComments {
				parser.currentConcept.PreComments = preComments
				addPreComments = false
			}
			addStates(&parser.currentState, conceptScope)
		} else if parser.isStep(token) {
			if !isInState(parser.currentState, conceptScope) {
				return nil, &ParseDetailResult{Error: &ParseError{LineNo: token.LineNo, Message: "Step is not defined inside a concept heading", LineText: token.LineText}}
			}
			if err := parser.processConceptStep(token); err != nil {
				return nil, &ParseDetailResult{Error: err}
			}
			addStates(&parser.currentState, stepScope)
		} else if parser.isTableHeader(token) {
			if !isInState(parser.currentState, stepScope) {
				return nil, &ParseDetailResult{Error: &ParseError{LineNo: token.LineNo, Message: "Table doesn't belong to any step", LineText: token.LineText}}
			}
			parser.processTableHeader(token)
			addStates(&parser.currentState, tableScope)
		} else if parser.isTableDataRow(token) {
			parser.processTableDataRow(token, &parser.currentConcept.Lookup)
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
		return nil, &ParseDetailResult{Error: &ParseError{LineNo: parser.currentConcept.LineNo, Message: "Concept should have atleast one step", LineText: parser.currentConcept.LineText}}
	}

	if parser.currentConcept != nil {
		concepts = append(concepts, parser.currentConcept)
	}
	return concepts, parseDetails
}

func (parser *ConceptParser) isConceptHeading(token *Token) bool {
	return token.Kind == gauge.SpecKind || token.Kind == gauge.ScenarioKind
}

func (parser *ConceptParser) isStep(token *Token) bool {
	return token.Kind == gauge.StepKind
}

func (parser *ConceptParser) isTableHeader(token *Token) bool {
	return token.Kind == gauge.TableHeader
}

func (parser *ConceptParser) isTableDataRow(token *Token) bool {
	return token.Kind == gauge.TableRow
}

func (parser *ConceptParser) processConceptHeading(token *Token) (*gauge.Step, *ParseDetailResult) {
	processStep(new(SpecParser), token)
	token.LineText = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(token.LineText), "#"))
	var concept *gauge.Step
	var parseDetails *ParseDetailResult
	concept, parseDetails = CreateStepUsingLookup(token, nil)
	if parseDetails != nil && parseDetails.Error != nil {
		return nil, parseDetails
	}
	if !parser.hasOnlyDynamicParams(concept) {
		parseDetails.Error = &ParseError{LineNo: token.LineNo, Message: "Concept heading can have only Dynamic Parameters"}
		return nil, parseDetails
	}

	concept.IsConcept = true
	parser.createConceptLookup(concept)
	concept.Items = append(concept.Items, concept)
	return concept, parseDetails
}

func (parser *ConceptParser) processConceptStep(token *Token) *ParseError {
	processStep(new(SpecParser), token)
	conceptStep, parseDetails := CreateStepUsingLookup(token, &parser.currentConcept.Lookup)
	if parseDetails != nil && parseDetails.Error != nil {
		return parseDetails.Error
	}
	if conceptStep.Value == parser.currentConcept.Value {
		return &ParseError{LineNo: conceptStep.LineNo, Message: "Cyclic dependancy found. Step is calling concept again.", LineText: conceptStep.LineText}

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

func (parser *ConceptParser) processTableDataRow(token *Token, argLookup *gauge.ArgLookup) {
	steps := parser.currentConcept.ConceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableRow(currentStep, token, argLookup)
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
func CreateConceptsDictionary(shouldIgnoreErrors bool, dirs []string) (*gauge.ConceptDictionary, *ParseResult) {
	var conceptFiles []string
	if len(dirs) < 1 {
		conceptFiles = util.FindConceptFilesIn(filepath.Join(config.ProjectRoot, common.SpecsDirectoryName))
	} else {
		for _, dir := range dirs {
			conceptFiles = append(conceptFiles, util.FindConceptFilesIn(filepath.Join(config.ProjectRoot, dir))...)
		}
	}
	conceptsDictionary := gauge.NewConceptDictionary()
	for _, conceptFile := range conceptFiles {
		if err := AddConcepts(conceptFile, conceptsDictionary); err != nil {
			if shouldIgnoreErrors {
				logger.APILog.Error("Concept parse failure: %s %s", conceptFile, err)
				continue
			}
			logger.Errorf(err.Error())
			return nil, &ParseResult{ParseError: err, FileName: conceptFile}
		}
	}
	return conceptsDictionary, &ParseResult{Ok: true}
}

func AddConcepts(conceptFile string, conceptDictionary *gauge.ConceptDictionary) *ParseError {
	concepts, parseResults := new(ConceptParser).ParseFile(conceptFile)
	if parseResults != nil && parseResults.Warnings != nil {
		for _, warning := range parseResults.Warnings {
			logger.Warning(warning.String())
		}
	}
	if parseResults != nil && parseResults.Error != nil {
		return parseResults.Error
	}
	for _, conceptStep := range concepts {
		if _, exists := conceptDictionary.ConceptsMap[conceptStep.Value]; exists {
			return &ParseError{Message: "Duplicate concept definition found", LineNo: conceptStep.LineNo, LineText: conceptStep.LineText}
		}
		conceptDictionary.ReplaceNestedConceptSteps(conceptStep)
		conceptDictionary.ConceptsMap[conceptStep.Value] = &gauge.Concept{conceptStep, conceptFile}
	}
	conceptDictionary.UpdateLookupForNestedConcepts()
	return validateConcepts(conceptDictionary)
}

func validateConcepts(conceptDictionary *gauge.ConceptDictionary) *ParseError {
	for _, concept := range conceptDictionary.ConceptsMap {
		err := checkCircularReferencing(conceptDictionary, concept.ConceptStep, nil)
		if err != nil {
			err.Message = fmt.Sprintf("Circular reference found in concept: \"%s\"\n%s", concept.ConceptStep.LineText, err.Message)
			return err
		}
	}
	return nil
}

func checkCircularReferencing(conceptDictionary *gauge.ConceptDictionary, concept *gauge.Step, traversedSteps map[string]string) *ParseError {
	if traversedSteps == nil {
		traversedSteps = make(map[string]string, 0)
	}
	currentConceptFileName := conceptDictionary.Search(concept.Value).FileName
	traversedSteps[concept.Value] = currentConceptFileName
	for _, step := range concept.ConceptSteps {
		if fileName, exists := traversedSteps[step.Value]; exists {
			return &ParseError{LineNo: step.LineNo,
				Message: fmt.Sprintf("%s: The concept \"%s\" references a higher concept > %s: \"%s\"", currentConceptFileName, concept.LineText, fileName, step.LineText),
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
