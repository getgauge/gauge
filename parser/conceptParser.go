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
	"strings"

	"bytes"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
)

type ConceptParser struct {
	currentState   int
	currentConcept *gauge.Step
}

// concept file can have multiple concept headings.
// Generates token for the given concept file and cretes concepts(array of steps) and parse results.
func (parser *ConceptParser) Parse(text, fileName string) ([]*gauge.Step, *ParseResult) {
	defer parser.resetState()

	specParser := new(SpecParser)
	tokens, errs := specParser.GenerateTokens(text, fileName)
	concepts, res := parser.createConcepts(tokens, fileName)
	return concepts, &ParseResult{ParseErrors: append(errs, res.ParseErrors...), Warnings: res.Warnings}
}

// Reads file contents from a give file and parses the file.
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
					parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: parser.currentConcept.LineNo, Message: "Concept should have atleast one step", LineText: parser.currentConcept.LineText})
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
				parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: token.LineNo, Message: "Step is not defined inside a concept heading", LineText: token.LineText})
				continue
			}
			if errs := parser.processConceptStep(token, fileName); len(errs) > 0 {
				parseRes.ParseErrors = append(parseRes.ParseErrors, errs...)
			}
			addStates(&parser.currentState, stepScope)
		} else if parser.isTableHeader(token) {
			if !isInState(parser.currentState, stepScope) {
				parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: token.LineNo, Message: "Table doesn't belong to any step", LineText: token.LineText})
				continue
			}
			parser.processTableHeader(token)
			addStates(&parser.currentState, tableScope)
		} else if parser.isScenarioHeading(token) {
			parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: token.LineNo, Message: "Scenario Heading is not allowed in concept file", LineText: token.LineText})
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
		parseRes.ParseErrors = append(parseRes.ParseErrors, ParseError{FileName: fileName, LineNo: parser.currentConcept.LineNo, Message: "Concept should have atleast one step", LineText: parser.currentConcept.LineText})
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
		parseRes.ParseErrors = []ParseError{ParseError{FileName: fileName, LineNo: token.LineNo, Message: "Concept heading can have only Dynamic Parameters", LineText: token.LineText}}
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
	if _, errs := AddConcepts(conceptFiles, conceptsDictionary); len(errs) > 0 {
		for _, err := range errs {
			logger.APILog.Errorf("Concept parse failure: %s %s", conceptFiles[0], err)
		}
		res.ParseErrors = append(res.ParseErrors, errs...)
		res.Ok = false
	}
	vRes := ValidateConcepts(conceptsDictionary)
	if len(vRes.ParseErrors) > 0 {
		res.Ok = false
		res.ParseErrors = append(res.ParseErrors, vRes.ParseErrors...)
	}
	return conceptsDictionary, res
}

func AddConcept(concepts []*gauge.Step, file string, conceptDictionary *gauge.ConceptDictionary) []ParseError {
	var duplicateConcepts map[string][]string = make(map[string][]string)
	for _, conceptStep := range concepts {
		checkForDuplicateConcepts(conceptDictionary, conceptStep, duplicateConcepts, file)
	}
	conceptDictionary.UpdateLookupForNestedConcepts()
	return mergeDuplicateConceptErrors(duplicateConcepts)
}

// AddConcepts parses the given concept file and adds each concept to the concept dictionary.
func AddConcepts(conceptFiles []string, conceptDictionary *gauge.ConceptDictionary) ([]*gauge.Step, []ParseError) {
	var conceptSteps []*gauge.Step
	var parseResults []*ParseResult
	var duplicateConcepts map[string][]string = make(map[string][]string)
	for _, conceptFile := range conceptFiles {
		concepts, parseRes := new(ConceptParser).ParseFile(conceptFile)
		if parseRes != nil && parseRes.Warnings != nil {
			for _, warning := range parseRes.Warnings {
				logger.Warningf(warning.String())
			}
		}
		for _, conceptStep := range concepts {
			checkForDuplicateConcepts(conceptDictionary, conceptStep, duplicateConcepts, conceptFile)
		}
		conceptDictionary.UpdateLookupForNestedConcepts()
		conceptSteps = append(conceptSteps, concepts...)
		parseResults = append(parseResults, parseRes)
	}
	errs := collectAllPArseErrors(parseResults)
	errs = append(errs, mergeDuplicateConceptErrors(duplicateConcepts)...)
	return conceptSteps, errs
}

func mergeDuplicateConceptErrors(duplicateConcepts map[string][]string) []ParseError {
	errs := []ParseError{}
	if len(duplicateConcepts) > 0 {
		for k, v := range duplicateConcepts {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("Duplicate concept definition found => '%s' => at", k))
			for _, value := range v {
				buffer.WriteString(value)
			}
			errs = append(errs, ParseError{Message: buffer.String()})
		}
	}
	return errs
}

func checkForDuplicateConcepts(conceptDictionary *gauge.ConceptDictionary, conceptStep *gauge.Step, duplicateConcepts map[string][]string, conceptFile string) {
	var duplicateConceptStep *gauge.Step
	if _, exists := conceptDictionary.ConceptsMap[conceptStep.Value]; exists {
		duplicateConceptStep = conceptStep
		duplicateConcepts[conceptStep.LineText] = append(duplicateConcepts[conceptStep.LineText], fmt.Sprintf("\n\t%s:%d", conceptFile, conceptStep.LineNo))
	} else {
		conceptDictionary.ConceptsMap[conceptStep.Value] = &gauge.Concept{conceptStep, conceptFile}
	}
	conceptDictionary.ReplaceNestedConceptSteps(conceptStep)
	if duplicateConceptStep != nil {
		conceptInDictionary := conceptDictionary.ConceptsMap[duplicateConceptStep.Value]
		errorInfo := fmt.Sprintf("\n\t%s:%d", conceptInDictionary.FileName, conceptInDictionary.ConceptStep.LineNo)
		if !contains(duplicateConcepts[duplicateConceptStep.LineText], errorInfo) {
			duplicateConcepts[duplicateConceptStep.LineText] = append(duplicateConcepts[duplicateConceptStep.LineText], errorInfo)
		}
	}
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

func collectAllPArseErrors(results []*ParseResult) (errs []ParseError) {
	for _, res := range results {
		errs = append(errs, res.ParseErrors...)
	}
	return
}

func ValidateConcepts(conceptDictionary *gauge.ConceptDictionary) *ParseResult {
	res := &ParseResult{ParseErrors: []ParseError{}}
	var conceptsWithError []*gauge.Concept
	for _, concept := range conceptDictionary.ConceptsMap {
		err := checkCircularReferencing(conceptDictionary, concept.ConceptStep, nil)
		if err != nil {
			delete(conceptDictionary.ConceptsMap, concept.ConceptStep.Value)
			res.ParseErrors = append(res.ParseErrors, err.(ParseError))
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

func checkCircularReferencing(conceptDictionary *gauge.ConceptDictionary, concept *gauge.Step, traversedSteps map[string]string) error {
	if traversedSteps == nil {
		traversedSteps = make(map[string]string, 0)
	}
	con := conceptDictionary.Search(concept.Value)
	if con == nil {
		return nil
	}
	currentConceptFileName := con.FileName
	traversedSteps[concept.Value] = currentConceptFileName
	for _, step := range concept.ConceptSteps {
		if fileName, exists := traversedSteps[step.Value]; exists {
			delete(conceptDictionary.ConceptsMap, step.Value)
			return ParseError{
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
