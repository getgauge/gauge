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

package main

import (
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	"path/filepath"
	"strings"
)

type conceptDictionary struct {
	conceptsMap     map[string]*concept
	constructionMap map[string][]*step
	referenceMap    map[*step][]*step
}

type concept struct {
	conceptStep *step
	fileName    string
}

type conceptParser struct {
	currentState   int
	currentConcept *step
}

//concept file can have multiple concept headings
func (parser *conceptParser) parse(text string) ([]*step, *parseDetailResult) {
	defer parser.resetState()

	specParser := new(specParser)
	tokens, err := specParser.generateTokens(text)
	if err != nil {
		return nil, &parseDetailResult{error: err}
	}
	return parser.createConcepts(tokens)
}

func (parser *conceptParser) resetState() {
	parser.currentState = initial
	parser.currentConcept = nil
}

func (parser *conceptParser) createConcepts(tokens []*token) ([]*step, *parseDetailResult) {
	parser.currentState = initial
	concepts := make([]*step, 0)
	var parseDetails *parseDetailResult
	preComments := make([]*comment, 0)
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
			comment := &comment{value: token.value, lineNo: token.lineNo}
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

func (parser *conceptParser) isConceptHeading(token *token) bool {
	return token.kind == specKind || token.kind == scenarioKind
}

func (parser *conceptParser) isStep(token *token) bool {
	return token.kind == stepKind
}

func (parser *conceptParser) isTableHeader(token *token) bool {
	return token.kind == tableHeader
}

func (parser *conceptParser) isTableDataRow(token *token) bool {
	return token.kind == tableRow
}

func (parser *conceptParser) processConceptHeading(token *token) (*step, *parseDetailResult) {
	processStep(new(specParser), token)
	token.lineText = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(token.lineText), "#"))
	var concept *step
	var parseDetails *parseDetailResult
	concept, parseDetails = new(specification).createStepUsingLookup(token, nil)
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

func (parser *conceptParser) processConceptStep(token *token) *parseError {
	processStep(new(specParser), token)
	conceptStep, parseDetails := new(specification).createStepUsingLookup(token, &parser.currentConcept.lookup)
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

func (parser *conceptParser) processTableHeader(token *token) {
	steps := parser.currentConcept.conceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableHeader(currentStep, token)
	items := parser.currentConcept.items
	items[len(items)-1] = currentStep
}

func (parser *conceptParser) processTableDataRow(token *token, argLookup *argLookup) {
	steps := parser.currentConcept.conceptSteps
	currentStep := steps[len(steps)-1]
	addInlineTableRow(currentStep, token, argLookup)
	items := parser.currentConcept.items
	items[len(items)-1] = currentStep
}

func (parser *conceptParser) hasOnlyDynamicParams(step *step) bool {
	for _, arg := range step.args {
		if arg.argType != "dynamic" {
			return false
		}
	}
	return true
}

func (parser *conceptParser) createConceptLookup(concept *step) {
	for _, arg := range concept.args {
		concept.lookup.addArgName(arg.value)
	}
}
func createConceptsDictionary(shouldIgnoreErrors bool) (*conceptDictionary, *parseResult) {
	conceptFiles := util.FindConceptFilesIn(filepath.Join(config.ProjectRoot, common.SpecsDirectoryName))
	conceptsDictionary := newConceptDictionary()
	for _, conceptFile := range conceptFiles {
		if err := addConcepts(conceptFile, conceptsDictionary); err != nil {
			if shouldIgnoreErrors {
				logger.ApiLog.Error("Concept parse failure: %s %s", conceptFile, err)
				continue
			}
			logger.Log.Error(err.Error())
			return nil, &parseResult{error: err, fileName: conceptFile}
		}
	}
	return conceptsDictionary, &parseResult{ok: true}
}

func addConcepts(conceptFile string, conceptDictionary *conceptDictionary) *parseError {
	fileText, fileReadErr := common.ReadFileContents(conceptFile)
	if fileReadErr != nil {
		return &parseError{message: fmt.Sprintf("failed to read concept file %s", conceptFile)}
	}
	concepts, parseResults := new(conceptParser).parse(fileText)
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

func newConceptDictionary() *conceptDictionary {
	return &conceptDictionary{conceptsMap: make(map[string]*concept, 0)}
}

func (conceptDictionary *conceptDictionary) isConcept(step *step) bool {
	_, ok := conceptDictionary.conceptsMap[step.value]
	return ok

}
func (conceptDictionary *conceptDictionary) add(concepts []*step, conceptFile string) *parseError {
	if conceptDictionary.conceptsMap == nil {
		conceptDictionary.conceptsMap = make(map[string]*concept)
	}
	if conceptDictionary.constructionMap == nil {
		conceptDictionary.constructionMap = make(map[string][]*step)
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

func (conceptDictionary *conceptDictionary) search(stepValue string) *concept {
	if concept, ok := conceptDictionary.conceptsMap[stepValue]; ok {
		return concept
	}
	return nil
}

func (conceptDictionary *conceptDictionary) validateConcepts() *parseError {
	for _, concept := range conceptDictionary.conceptsMap {
		err := conceptDictionary.checkCircularReferencing(concept.conceptStep, nil)
		if err != nil {
			err.message = fmt.Sprintf("Circular reference found in concept: \"%s\"\n%s", concept.conceptStep.lineText, err.message)
			return err
		}
	}
	return nil
}

func (conceptDictionary *conceptDictionary) checkCircularReferencing(concept *step, traversedSteps map[string]string) *parseError {
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

func (conceptDictionary *conceptDictionary) replaceNestedConceptSteps(conceptStep *step) {
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
func (conceptDictionary *conceptDictionary) updateStep(step *step) {
	conceptDictionary.constructionMap[step.value] = append(conceptDictionary.constructionMap[step.value], step)
	if !conceptDictionary.constructionMap[step.value][0].isConcept {
		conceptDictionary.constructionMap[step.value] = append(conceptDictionary.constructionMap[step.value], step)
		for _, allSteps := range conceptDictionary.constructionMap[step.value] {
			allSteps.isConcept = step.isConcept
			allSteps.conceptSteps = step.conceptSteps
			allSteps.lookup = step.lookup
		}
	}
}

func (conceptDictionary *conceptDictionary) updateLookupForNestedConcepts() {
	for _, concept := range conceptDictionary.conceptsMap {
		for _, stepInsideConcept := range concept.conceptStep.conceptSteps {
			stepInsideConcept.parent = concept.conceptStep
			if nestedConcept := conceptDictionary.search(stepInsideConcept.value); nestedConcept != nil {
				for i, arg := range nestedConcept.conceptStep.args {
					nestedConcept.conceptStep.lookup.addArgValue(arg.value, &stepArg{argType: stepInsideConcept.args[i].argType, value: stepInsideConcept.args[i].value})
				}
			}
		}
	}
}

func (self *concept) deepCopy() *concept {
	return &concept{fileName: self.fileName, conceptStep: self.conceptStep.getCopy()}
}
