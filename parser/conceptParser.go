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
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	"path/filepath"
	"strings"
)

type ConceptDictionary struct {
	ConceptsMap     map[string]*Concept
	constructionMap map[string][]*Step
	referenceMap    map[*Step][]*Step
}

type Concept struct {
	ConceptStep *Step
	FileName    string
}

type ConceptParser struct {
	currentState   int
	currentConcept *Step
}

//concept file can have multiple concept headings
func (parser *ConceptParser) Parse(text string) ([]*Step, *ParseDetailResult) {
	defer parser.resetState()

	specParser := new(SpecParser)
	tokens, err := specParser.GenerateTokens(text)
	if err != nil {
		return nil, &ParseDetailResult{Error: err}
	}
	return parser.createConcepts(tokens)
}

func (parser *ConceptParser) ParseFile(file string) ([]*Step, *ParseDetailResult) {
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

func (parser *ConceptParser) createConcepts(tokens []*Token) ([]*Step, *ParseDetailResult) {
	parser.currentState = initial
	concepts := make([]*Step, 0)
	var parseDetails *ParseDetailResult
	preComments := make([]*Comment, 0)
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
			comment := &Comment{Value: token.Value, LineNo: token.LineNo}
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
	return token.Kind == SpecKind || token.Kind == ScenarioKind
}

func (parser *ConceptParser) isStep(token *Token) bool {
	return token.Kind == StepKind
}

func (parser *ConceptParser) isTableHeader(token *Token) bool {
	return token.Kind == TableHeader
}

func (parser *ConceptParser) isTableDataRow(token *Token) bool {
	return token.Kind == TableRow
}

func (parser *ConceptParser) processConceptHeading(token *Token) (*Step, *ParseDetailResult) {
	processStep(new(SpecParser), token)
	token.LineText = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(token.LineText), "#"))
	var concept *Step
	var parseDetails *ParseDetailResult
	concept, parseDetails = new(Specification).CreateStepUsingLookup(token, nil)
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
	conceptStep, parseDetails := new(Specification).CreateStepUsingLookup(token, &parser.currentConcept.Lookup)
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

func (parser *ConceptParser) processTableDataRow(token *Token, argLookup *ArgLookup) {
	steps := parser.currentConcept.ConceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableRow(currentStep, token, argLookup)
	items := parser.currentConcept.Items
	items[len(items)-1] = currentStep
}

func (parser *ConceptParser) hasOnlyDynamicParams(step *Step) bool {
	for _, arg := range step.Args {
		if arg.ArgType != Dynamic {
			return false
		}
	}
	return true
}

func (parser *ConceptParser) createConceptLookup(concept *Step) {
	for _, arg := range concept.Args {
		concept.Lookup.addArgName(arg.Value)
	}
}
func CreateConceptsDictionary(shouldIgnoreErrors bool) (*ConceptDictionary, *ParseResult) {
	conceptFiles := util.FindConceptFilesIn(filepath.Join(config.ProjectRoot, common.SpecsDirectoryName))
	conceptsDictionary := NewConceptDictionary()
	for _, conceptFile := range conceptFiles {
		if err := AddConcepts(conceptFile, conceptsDictionary); err != nil {
			if shouldIgnoreErrors {
				logger.ApiLog.Error("Concept parse failure: %s %s", conceptFile, err)
				continue
			}
			logger.Log.Error(err.Error())
			return nil, &ParseResult{ParseError: err, FileName: conceptFile}
		}
	}
	return conceptsDictionary, &ParseResult{Ok: true}
}

func AddConcepts(conceptFile string, conceptDictionary *ConceptDictionary) *ParseError {
	concepts, parseResults := new(ConceptParser).ParseFile(conceptFile)
	if parseResults != nil && parseResults.Warnings != nil {
		for _, warning := range parseResults.Warnings {
			logger.Log.Warning(warning.String())
		}
	}
	if parseResults != nil && parseResults.Error != nil {
		return parseResults.Error
	}
	return conceptDictionary.Add(concepts, conceptFile)
}

func NewConceptDictionary() *ConceptDictionary {
	return &ConceptDictionary{ConceptsMap: make(map[string]*Concept, 0)}
}

func (conceptDictionary *ConceptDictionary) isConcept(step *Step) bool {
	_, ok := conceptDictionary.ConceptsMap[step.Value]
	return ok

}
func (conceptDictionary *ConceptDictionary) Add(concepts []*Step, conceptFile string) *ParseError {
	if conceptDictionary.ConceptsMap == nil {
		conceptDictionary.ConceptsMap = make(map[string]*Concept)
	}
	if conceptDictionary.constructionMap == nil {
		conceptDictionary.constructionMap = make(map[string][]*Step)
	}
	for _, conceptStep := range concepts {
		if _, exists := conceptDictionary.ConceptsMap[conceptStep.Value]; exists {
			return &ParseError{Message: "Duplicate concept definition found", LineNo: conceptStep.LineNo, LineText: conceptStep.LineText}
		}
		conceptDictionary.replaceNestedConceptSteps(conceptStep)
		conceptDictionary.ConceptsMap[conceptStep.Value] = &Concept{conceptStep, conceptFile}
	}
	conceptDictionary.updateLookupForNestedConcepts()
	return conceptDictionary.validateConcepts()
}

func (conceptDictionary *ConceptDictionary) search(stepValue string) *Concept {
	if concept, ok := conceptDictionary.ConceptsMap[stepValue]; ok {
		return concept
	}
	return nil
}

func (conceptDictionary *ConceptDictionary) validateConcepts() *ParseError {
	for _, concept := range conceptDictionary.ConceptsMap {
		err := conceptDictionary.checkCircularReferencing(concept.ConceptStep, nil)
		if err != nil {
			err.Message = fmt.Sprintf("Circular reference found in concept: \"%s\"\n%s", concept.ConceptStep.LineText, err.Message)
			return err
		}
	}
	return nil
}

func (conceptDictionary *ConceptDictionary) checkCircularReferencing(concept *Step, traversedSteps map[string]string) *ParseError {
	if traversedSteps == nil {
		traversedSteps = make(map[string]string, 0)
	}
	currentConceptFileName := conceptDictionary.search(concept.Value).FileName
	traversedSteps[concept.Value] = currentConceptFileName
	for _, step := range concept.ConceptSteps {
		if fileName, exists := traversedSteps[step.Value]; exists {
			return &ParseError{LineNo: step.LineNo,
				Message: fmt.Sprintf("%s: The concept \"%s\" references a higher concept -> %s: \"%s\"", currentConceptFileName, concept.LineText, fileName, step.LineText),
			}

		}
		if step.IsConcept {
			if err := conceptDictionary.checkCircularReferencing(step, traversedSteps); err != nil {
				return err
			}
		}
	}
	delete(traversedSteps, concept.Value)
	return nil
}

func (conceptDictionary *ConceptDictionary) replaceNestedConceptSteps(conceptStep *Step) {
	conceptDictionary.updateStep(conceptStep)
	for i, stepInsideConcept := range conceptStep.ConceptSteps {
		if nestedConcept := conceptDictionary.search(stepInsideConcept.Value); nestedConcept != nil {
			//replace step with actual concept
			conceptStep.ConceptSteps[i].ConceptSteps = nestedConcept.ConceptStep.ConceptSteps
			conceptStep.ConceptSteps[i].IsConcept = nestedConcept.ConceptStep.IsConcept
			conceptStep.ConceptSteps[i].Lookup = *nestedConcept.ConceptStep.Lookup.getCopy()
		} else {
			conceptDictionary.updateStep(stepInsideConcept)
		}
	}
}

//mutates the step with concept steps so that anyone who is referencing the step will now refer a concept
func (conceptDictionary *ConceptDictionary) updateStep(step *Step) {
	conceptDictionary.constructionMap[step.Value] = append(conceptDictionary.constructionMap[step.Value], step)
	if !conceptDictionary.constructionMap[step.Value][0].IsConcept {
		conceptDictionary.constructionMap[step.Value] = append(conceptDictionary.constructionMap[step.Value], step)
		for _, allSteps := range conceptDictionary.constructionMap[step.Value] {
			allSteps.IsConcept = step.IsConcept
			allSteps.ConceptSteps = step.ConceptSteps
			allSteps.Lookup = *step.Lookup.getCopy()
		}
	}
}

func (conceptDictionary *ConceptDictionary) updateLookupForNestedConcepts() {
	for _, concept := range conceptDictionary.ConceptsMap {
		for _, stepInsideConcept := range concept.ConceptStep.ConceptSteps {
			stepInsideConcept.Parent = concept.ConceptStep
			if nestedConcept := conceptDictionary.search(stepInsideConcept.Value); nestedConcept != nil {
				for i, arg := range nestedConcept.ConceptStep.Args {
					stepInsideConcept.Lookup.addArgValue(arg.Value, &StepArg{ArgType: stepInsideConcept.Args[i].ArgType, Value: stepInsideConcept.Args[i].Value})
				}
			}
		}
	}
}

func (self *Concept) deepCopy() *Concept {
	return &Concept{FileName: self.FileName, ConceptStep: self.ConceptStep.getCopy()}
}
