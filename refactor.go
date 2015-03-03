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
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
	"path/filepath"
	"strings"
)

type rephraseRefactorer struct {
	oldStep   *step
	newStep   *step
	isConcept bool
}

type refactoringResult struct {
	success            bool
	specsChanged       []string
	conceptsChanged    []string
	runnerFilesChanged []string
	errors             []string
	warnings           []string
}

func performRephraseRefactoring(oldStep, newStep string, ignoreWarnings bool) *refactoringResult {
	if newStep == oldStep {
		if ignoreWarnings {
			return &refactoringResult{success: true}
		}
		return rephraseFailure("Same old step name and new step name.")
	}
	agent, err := getRefactorAgent(oldStep, newStep)

	if err != nil {
		return rephraseFailure(err.Error())
	}

	projectRoot, err := common.GetProjectRoot()
	if err != nil {
		return rephraseFailure(err.Error())
	}

	result := &refactoringResult{success: true, errors: make([]string, 0), warnings: make([]string, 0)}
	specs, specParseResults := findSpecs(filepath.Join(projectRoot, common.SpecsDirectoryName), &conceptDictionary{})
	addErrorsAndWarningsToRefactoringResult(result, specParseResults...)
	if !result.success {
		return result
	}
	conceptDictionary, parseResult := createConceptsDictionary(false)

	addErrorsAndWarningsToRefactoringResult(result, parseResult)
	if !result.success {
		return result
	}

	refactorResult := agent.performRefactoringOn(specs, conceptDictionary, ignoreWarnings)
	refactorResult.warnings = append(refactorResult.warnings, result.warnings...)
	return refactorResult
}

func rephraseFailure(errors ...string) *refactoringResult {
	return &refactoringResult{success: false, errors: errors}
}

func addErrorsAndWarningsToRefactoringResult(refactorResult *refactoringResult, parseResults ...*parseResult) {
	for _, parseResult := range parseResults {
		if !parseResult.ok {
			refactorResult.success = false
			refactorResult.errors = append(refactorResult.errors, parseResult.Error())
		}
		refactorResult.appendWarnings(parseResult.warnings)
	}
}

func (agent *rephraseRefactorer) performRefactoringOn(specs []*specification, conceptDictionary *conceptDictionary, ignore bool) *refactoringResult {
	specsRefactored, conceptFilesRefactored := agent.rephraseInSpecsAndConcepts(&specs, conceptDictionary)
	result := &refactoringResult{success: false, errors: make([]string, 0), warnings: make([]string, 0)}
	if !agent.isConcept {
		runner, connErr := agent.startRunner()
		if connErr != nil {
			result.errors = append(result.errors, connErr.Error())
			return result
		}
		defer runner.kill(getCurrentExecutionLogger())
		stepName, err, warning := agent.getStepNameFromRunner(runner, ignore)
		if err != nil {
			result.errors = append(result.errors, err.Error())
			return result
		}
		if warning == nil {
			runnerFilesChanged, err := agent.requestRunnerForRefactoring(runner, stepName)
			if err != nil {
				result.errors = append(result.errors, fmt.Sprintf("Cannot perform refactoring: %s", err))
				return result
			}
			result.runnerFilesChanged = runnerFilesChanged
		} else {
			result.warnings = append(result.warnings, warning.message)
		}
	}
	specFiles, conceptFiles := writeToConceptAndSpecFiles(specs, conceptDictionary, specsRefactored, conceptFilesRefactored)
	result.specsChanged = specFiles
	result.success = true
	result.conceptsChanged = conceptFiles
	return result
}

func (agent *rephraseRefactorer) rephraseInSpecsAndConcepts(specs *[]*specification, conceptDictionary *conceptDictionary) (map[*specification]bool, map[string]bool) {
	specsRefactored := make(map[*specification]bool, 0)
	conceptFilesRefactored := make(map[string]bool, 0)
	orderMap := agent.createOrderOfArgs()
	for _, spec := range *specs {
		specsRefactored[spec] = spec.renameSteps(*agent.oldStep, *agent.newStep, orderMap)
	}
	isConcept := false
	for _, concept := range conceptDictionary.conceptsMap {
		_, ok := conceptFilesRefactored[concept.fileName]
		conceptFilesRefactored[concept.fileName] = !ok && false || conceptFilesRefactored[concept.fileName]
		for _, item := range concept.conceptStep.items {
			isRefactored := conceptFilesRefactored[concept.fileName]
			conceptFilesRefactored[concept.fileName] = item.kind() == stepKind &&
				item.(*step).rename(*agent.oldStep, *agent.newStep, isRefactored, orderMap, &isConcept) ||
				isRefactored
		}
	}
	agent.isConcept = isConcept
	return specsRefactored, conceptFilesRefactored
}

func (agent *rephraseRefactorer) createOrderOfArgs() map[int]int {
	orderMap := make(map[int]int, len(agent.newStep.args))
	for i, arg := range agent.newStep.args {
		orderMap[i] = SliceIndex(len(agent.oldStep.args), func(i int) bool { return agent.oldStep.args[i].String() == arg.String() })
	}
	return orderMap
}

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func getRefactorAgent(oldStepText, newStepText string) (*rephraseRefactorer, error) {
	parser := new(specParser)
	stepTokens, err := parser.generateTokens("* " + oldStepText + "\n" + "*" + newStepText)
	if err != nil {
		return nil, err
	}
	spec := &specification{}
	steps := make([]*step, 0)
	for _, stepToken := range stepTokens {
		step, parseDetails := spec.createStepUsingLookup(stepToken, nil)
		if parseDetails != nil && parseDetails.error != nil {
			return nil, parseDetails.error
		}
		steps = append(steps, step)
	}
	return &rephraseRefactorer{oldStep: steps[0], newStep: steps[1]}, nil
}

func (agent *rephraseRefactorer) requestRunnerForRefactoring(testRunner *testRunner, stepName string) ([]string, error) {
	refactorRequest, err := agent.createRefactorRequest(testRunner, stepName)
	if err != nil {
		return nil, err
	}
	refactorResponse := agent.sendRefactorRequest(testRunner, refactorRequest)
	var runnerError error
	if !refactorResponse.GetSuccess() {
		apiLog.Error("Refactoring error response from runner: %v", refactorResponse.GetError())
		runnerError = errors.New(refactorResponse.GetError())
	}
	return refactorResponse.GetFilesChanged(), runnerError
}

func (agent *rephraseRefactorer) startRunner() (*testRunner, error) {
	loadGaugeEnvironment()
	startAPIService(0)
	testRunner, err := startRunnerAndMakeConnection(getProjectManifest(getCurrentExecutionLogger()), getCurrentExecutionLogger())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to connect to test runner: %s", err))
	}
	return testRunner, nil
}

func (agent *rephraseRefactorer) sendRefactorRequest(testRunner *testRunner, refactorRequest *gauge_messages.Message) *gauge_messages.RefactorResponse {
	response, err := getResponseForMessageWithTimeout(refactorRequest, testRunner.connection, config.RefactorTimeout())
	if err != nil {
		return &gauge_messages.RefactorResponse{Success: proto.Bool(false), Error: proto.String(err.Error())}
	}
	return response.GetRefactorResponse()
}

//Todo: Check for inline tables
func (agent *rephraseRefactorer) createRefactorRequest(runner *testRunner, stepName string) (*gauge_messages.Message, error) {
	oldStepValue, err := agent.getStepValueFor(agent.oldStep, stepName)
	if err != nil {
		return nil, err
	}
	orderMap := agent.createOrderOfArgs()
	newStepName := agent.generateNewStepName(oldStepValue.args, orderMap)
	newStepValue, err := extractStepValueAndParams(newStepName, false)
	if err != nil {
		return nil, err
	}
	oldProtoStepValue := convertToProtoStepValue(oldStepValue)
	newProtoStepValue := convertToProtoStepValue(newStepValue)
	return &gauge_messages.Message{MessageType: gauge_messages.Message_RefactorRequest.Enum(), RefactorRequest: &gauge_messages.RefactorRequest{OldStepValue: oldProtoStepValue, NewStepValue: newProtoStepValue, ParamPositions: agent.createParameterPositions(orderMap)}}, nil
}

func (agent *rephraseRefactorer) generateNewStepName(args []string, orderMap map[int]int) string {
	agent.newStep.populateFragments()
	paramIndex := 0
	for _, fragment := range agent.newStep.fragments {
		if fragment.GetFragmentType() == gauge_messages.Fragment_Parameter {
			if orderMap[paramIndex] != -1 {
				fragment.GetParameter().Value = proto.String(args[orderMap[paramIndex]])
			}
			paramIndex++
		}
	}
	return convertToStepText(agent.newStep.fragments)
}

func (agent *rephraseRefactorer) getStepNameFromRunner(runner *testRunner, ignore bool) (string, error, *warning) {
	stepNameMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_StepNameRequest.Enum(), StepNameRequest: &gauge_messages.StepNameRequest{StepValue: proto.String(agent.oldStep.value)}}
	responseMessage, err := getResponseForMessageWithTimeout(stepNameMessage, runner.connection, config.RunnerRequestTimeout())
	if err != nil {
		return "", err, nil
	}
	if !(responseMessage.GetStepNameResponse().GetIsStepPresent()) {
		return "", nil, &warning{message: fmt.Sprintf("Step implementation not found: %s", agent.oldStep.lineText)}
	}
	if responseMessage.GetStepNameResponse().GetHasAlias() {
		return "", errors.New(fmt.Sprintf("steps with aliases : '%s' cannot be refactored.", strings.Join(responseMessage.GetStepNameResponse().GetStepName(), "', '"))), nil
	}
	return responseMessage.GetStepNameResponse().GetStepName()[0], nil, nil
}

func (agent *rephraseRefactorer) createParameterPositions(orderMap map[int]int) []*gauge_messages.ParameterPosition {
	paramPositions := make([]*gauge_messages.ParameterPosition, 0)
	for k, v := range orderMap {
		paramPositions = append(paramPositions, &gauge_messages.ParameterPosition{NewPosition: proto.Int(k), OldPosition: proto.Int(v)})
	}
	return paramPositions
}

func (agent *rephraseRefactorer) getStepValueFor(step *step, stepName string) (*stepValue, error) {
	return extractStepValueAndParams(stepName, false)
}

func writeToConceptAndSpecFiles(specs []*specification, conceptDictionary *conceptDictionary, specsRefactored map[*specification]bool, conceptFilesRefactored map[string]bool) ([]string, []string) {
	specFiles := make([]string, 0)
	conceptFiles := make([]string, 0)
	for _, spec := range specs {
		if specsRefactored[spec] {
			specFiles = append(specFiles, spec.fileName)
			formatted := formatSpecification(spec)
			saveFile(spec.fileName, formatted, true)
		}
	}
	conceptMap := formatConcepts(conceptDictionary)
	for fileName, concept := range conceptMap {
		if conceptFilesRefactored[fileName] {
			conceptFiles = append(conceptFiles, fileName)
			saveFile(fileName, concept, true)
		}
	}
	return specFiles, conceptFiles
}

func (refactoringResult *refactoringResult) appendWarnings(warnings []*warning) {
	if refactoringResult.warnings == nil {
		refactoringResult.warnings = make([]string, 0)
	}
	for _, warning := range warnings {
		refactoringResult.warnings = append(refactoringResult.warnings, warning.message)
	}
}

func (refactoringResult *refactoringResult) allFilesChanges() []string {
	filesChanged := make([]string, 0)
	filesChanged = append(filesChanged, refactoringResult.specsChanged...)
	filesChanged = append(filesChanged, refactoringResult.conceptsChanged...)
	filesChanged = append(filesChanged, refactoringResult.runnerFilesChanged...)
	return filesChanged

}
