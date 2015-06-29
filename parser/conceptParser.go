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
	conceptsMap     map[string]*concept
	constructionMap map[string][]*Step
	referenceMap    map[*Step][]*Step
}

type concept struct {
	conceptStep *Step
	fileName    string
}

type ConceptParser struct {
	currentState   int
	currentConcept *Step
}

//concept file can have multiple concept headings
func (parser *ConceptParser) parse(text string) ([]*Step, *parseDetailResult) {
	defer parser.resetState()

	specParser := new(SpecParser)
	tokens, err := specParser.generateTokens(text)
	if err != nil {
		return nil, &parseDetailResult{error: err}
	}
	return parser.createConcepts(tokens)
}

func (parser *ConceptParser) resetState() {
	parser.currentState = initial
	parser.currentConcept = nil
}

func (parser *ConceptParser) createConcepts(tokens []*Token) ([]*Step, *parseDetailResult) {
	parser.currentState = initial
	concepts := make([]*Step, 0)
	var parseDetails *parseDetailResult
	preComments := make([]*Comment, 0)
	addPreComments := false
	for _, token := range tokens {
		if parser.isConceptHeading(token) {
			if isInState(parser.currentState, conceptScope, stepScope) {
				concepts = append(concepts, parser.currentConcept)
			}
			parser.currentConcept, parseDetails = parser.processConceptHeading(token)
			if parseDetails.error != nil {
				return nil, parseDetails
			}
			if addPreComments {
				parser.currentConcept.preComments = preComments
				addPreComments = false
			}
			addStates(&parser.currentState, conceptScope)
		} else if parser.isStep(token) {
			if !isInState(parser.currentState, conceptScope) {
				return nil, &parseDetailResult{error: &parseError{lineNo: token.lineNo, message: "Step is not defined inside a concept heading", lineText: token.lineText}}
			}
			if err := parser.processConceptStep(token); err != nil {
				return nil, &parseDetailResult{error: err}
			}
			addStates(&parser.currentState, stepScope)
		} else if parser.isTableHeader(token) {
			if !isInState(parser.currentState, stepScope) {
				return nil, &parseDetailResult{error: &parseError{lineNo: token.lineNo, message: "Table doesn't belong to any step", lineText: token.lineText}}
			}
			parser.processTableHeader(token)
			addStates(&parser.currentState, tableScope)
		} else if parser.isTableDataRow(token) {
			parser.processTableDataRow(token, &parser.currentConcept.lookup)
		} else {
			comment := &Comment{value: token.value, lineNo: token.lineNo}
			if parser.currentConcept == nil {
				preComments = append(preComments, comment)
				addPreComments = true
				continue
			}
			parser.currentConcept.items = append(parser.currentConcept.items, comment)
		}
	}
	if !isInState(parser.currentState, stepScope) && parser.currentState != initial {
		return nil, &parseDetailResult{error: &parseError{lineNo: parser.currentConcept.lineNo, message: "Concept should have atleast one step", lineText: parser.currentConcept.lineText}}
	}

	if parser.currentConcept != nil {
		concepts = append(concepts, parser.currentConcept)
	}
	return concepts, parseDetails
}

func (parser *ConceptParser) isConceptHeading(token *Token) bool {
	return token.kind == specKind || token.kind == scenarioKind
}

func (parser *ConceptParser) isStep(token *Token) bool {
	return token.kind == stepKind
}

func (parser *ConceptParser) isTableHeader(token *Token) bool {
	return token.kind == tableHeader
}

func (parser *ConceptParser) isTableDataRow(token *Token) bool {
	return token.kind == tableRow
}

func (parser *ConceptParser) processConceptHeading(token *Token) (*Step, *parseDetailResult) {
	processStep(new(SpecParser), token)
	token.lineText = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(token.lineText), "#"))
	var concept *Step
	var parseDetails *parseDetailResult
	concept, parseDetails = new(Specification).createStepUsingLookup(token, nil)
	if parseDetails != nil && parseDetails.error != nil {
		return nil, parseDetails
	}
	if !parser.hasOnlyDynamicParams(concept) {
		parseDetails.error = &parseError{lineNo: token.lineNo, message: "Concept heading can have only Dynamic Parameters"}
		return nil, parseDetails
	}

	concept.isConcept = true
	parser.createConceptLookup(concept)
	concept.items = append(concept.items, concept)
	return concept, parseDetails
}

func (parser *ConceptParser) processConceptStep(token *Token) *parseError {
	processStep(new(SpecParser), token)
	conceptStep, parseDetails := new(Specification).createStepUsingLookup(token, &parser.currentConcept.lookup)
	if parseDetails != nil && parseDetails.error != nil {
		return parseDetails.error
	}
	if conceptStep.value == parser.currentConcept.value {
		return &parseError{lineNo: conceptStep.lineNo, message: "Cyclic dependancy found. Step is calling concept again.", lineText: conceptStep.lineText}

	}
	parser.currentConcept.conceptSteps = append(parser.currentConcept.conceptSteps, conceptStep)
	parser.currentConcept.items = append(parser.currentConcept.items, conceptStep)
	return nil
}

func (parser *ConceptParser) processTableHeader(token *Token) {
	steps := parser.currentConcept.conceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableHeader(currentStep, token)
	items := parser.currentConcept.items
	items[len(items)-1] = currentStep
}

func (parser *ConceptParser) processTableDataRow(token *Token, argLookup *ArgLookup) {
	steps := parser.currentConcept.conceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableRow(currentStep, token, argLookup)
	items := parser.currentConcept.items
	items[len(items)-1] = currentStep
}

func (parser *ConceptParser) hasOnlyDynamicParams(step *Step) bool {
	for _, arg := range step.args {
		if arg.argType != Dynamic {
			return false
		}
	}
	return true
}

func (parser *ConceptParser) createConceptLookup(concept *Step) {
	for _, arg := range concept.args {
		concept.lookup.addArgName(arg.value)
	}
}
func createConceptsDictionary(shouldIgnoreErrors bool) (*ConceptDictionary, *ParseResult) {
	conceptFiles := util.FindConceptFilesIn(filepath.Join(config.ProjectRoot, common.SpecsDirectoryName))
	conceptsDictionary := newConceptDictionary()
	for _, conceptFile := range conceptFiles {
		if err := addConcepts(conceptFile, conceptsDictionary); err != nil {
			if shouldIgnoreErrors {
				logger.ApiLog.Error("Concept parse failure: %s %s", conceptFile, err)
				continue
			}
			logger.Log.Error(err.Error())
			return nil, &ParseResult{error: err, fileName: conceptFile}
		}
	}
	return conceptsDictionary, &ParseResult{ok: true}
}

func addConcepts(conceptFile string, conceptDictionary *ConceptDictionary) *parseError {
	fileText, fileReadErr := common.ReadFileContents(conceptFile)
	if fileReadErr != nil {
		return &parseError{message: fmt.Sprintf("failed to read concept file %s", conceptFile)}
	}
	concepts, parseResults := new(ConceptParser).parse(fileText)
	if parseResults != nil && parseResults.warnings != nil {
		for _, warning := range parseResults.warnings {
			logger.Log.Warning(warning.String())
		}
	}
	if parseResults != nil && parseResults.error != nil {
		return parseResults.error
	}
	return conceptDictionary.add(concepts, conceptFile)
}

func newConceptDictionary() *ConceptDictionary {
	return &ConceptDictionary{conceptsMap: make(map[string]*concept, 0)}
}

func (conceptDictionary *ConceptDictionary) isConcept(step *Step) bool {
	_, ok := conceptDictionary.conceptsMap[step.value]
	return ok

}
func (conceptDictionary *ConceptDictionary) add(concepts []*Step, conceptFile string) *parseError {
	if conceptDictionary.conceptsMap == nil {
		conceptDictionary.conceptsMap = make(map[string]*concept)
	}
	if conceptDictionary.constructionMap == nil {
		conceptDictionary.constructionMap = make(map[string][]*Step)
	}
	for _, conceptStep := range concepts {
		if _, exists := conceptDictionary.conceptsMap[conceptStep.value]; exists {
			return &parseError{message: "Duplicate concept definition found", lineNo: conceptStep.lineNo, lineText: conceptStep.lineText}
		}
		conceptDictionary.replaceNestedConceptSteps(conceptStep)
		conceptDictionary.conceptsMap[conceptStep.value] = &concept{conceptStep, conceptFile}
	}
	conceptDictionary.updateLookupForNestedConcepts()
	return conceptDictionary.validateConcepts()
}

func (conceptDictionary *ConceptDictionary) search(stepValue string) *concept {
	if concept, ok := conceptDictionary.conceptsMap[stepValue]; ok {
		return concept
	}
	return nil
}

func (conceptDictionary *ConceptDictionary) validateConcepts() *parseError {
	for _, concept := range conceptDictionary.conceptsMap {
		err := conceptDictionary.checkCircularReferencing(concept.conceptStep, nil)
		if err != nil {
			err.message = fmt.Sprintf("Circular reference found in concept: \"%s\"\n%s", concept.conceptStep.lineText, err.message)
			return err
		}
	}
	return nil
}

func (conceptDictionary *ConceptDictionary) checkCircularReferencing(concept *Step, traversedSteps map[string]string) *parseError {
	if traversedSteps == nil {
		traversedSteps = make(map[string]string, 0)
	}
	currentConceptFileName := conceptDictionary.search(concept.value).fileName
	traversedSteps[concept.value] = currentConceptFileName
	for _, step := range concept.conceptSteps {
		if fileName, exists := traversedSteps[step.value]; exists {
			return &parseError{lineNo: step.lineNo,
				message: fmt.Sprintf("%s: The concept \"%s\" references a higher concept -> %s: \"%s\"", currentConceptFileName, concept.lineText, fileName, step.lineText),
			}

		}
		if step.isConcept {
			if err := conceptDictionary.checkCircularReferencing(step, traversedSteps); err != nil {
				return err
			}
		}
	}
	delete(traversedSteps, concept.value)
	return nil
}

func (conceptDictionary *ConceptDictionary) replaceNestedConceptSteps(conceptStep *Step) {
	conceptDictionary.updateStep(conceptStep)
	for i, stepInsideConcept := range conceptStep.conceptSteps {
		if nestedConcept := conceptDictionary.search(stepInsideConcept.value); nestedConcept != nil {
			//replace step with actual concept
			conceptStep.conceptSteps[i].conceptSteps = nestedConcept.conceptStep.conceptSteps
			conceptStep.conceptSteps[i].isConcept = nestedConcept.conceptStep.isConcept
			conceptStep.conceptSteps[i].lookup = nestedConcept.conceptStep.lookup
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
			allSteps.Lookup = step.Lookup
		}
	}
}

func (conceptDictionary *ConceptDictionary) updateLookupForNestedConcepts() {
	for _, concept := range conceptDictionary.conceptsMap {
		for _, stepInsideConcept := range concept.conceptStep.conceptSteps {
			stepInsideConcept.parent = concept.conceptStep
			if nestedConcept := conceptDictionary.search(stepInsideConcept.value); nestedConcept != nil {
				for i, arg := range nestedConcept.conceptStep.args {
					nestedConcept.conceptStep.lookup.addArgValue(arg.value, &StepArg{argType: stepInsideConcept.args[i].argType, value: stepInsideConcept.args[i].value})
				}
			}
		}
	}
}

func (self *concept) deepCopy() *concept {
	return &concept{fileName: self.fileName, conceptStep: self.conceptStep.getCopy()}
}
