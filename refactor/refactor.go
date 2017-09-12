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

package refactor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
)

type rephraseRefactorer struct {
	oldStep   *gauge.Step
	newStep   *gauge.Step
	isConcept bool
	startChan *runner.StartChannels
}

type refactoringResult struct {
	Success            bool
	specsChanged       []string
	conceptsChanged    []string
	runnerFilesChanged []string
	Errors             []string
	warnings           []string
}

func (refactoringResult *refactoringResult) String() string {
	result := fmt.Sprintf("Refactoring result from gauge:\n")
	result += fmt.Sprintf("Specs changed        : %s\n", refactoringResult.specsChanged)
	result += fmt.Sprintf("Concepts changed     : %s\n", refactoringResult.conceptsChanged)
	result += fmt.Sprintf("Source files changed : %s\n", refactoringResult.runnerFilesChanged)
	result += fmt.Sprintf("Warnings             : %s\n", refactoringResult.warnings)
	return result
}

func PerformRephraseRefactoring(oldStep, newStep string, startChan *runner.StartChannels, specDirs []string) *refactoringResult {
	defer killRunner(startChan)
	if newStep == oldStep {
		return &refactoringResult{Success: true}
	}
	agent, errs := getRefactorAgent(oldStep, newStep, startChan)

	if len(errs) > 0 {
		var messages []string
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		return rephraseFailure(messages...)
	}

	result := &refactoringResult{Success: true, Errors: make([]string, 0), warnings: make([]string, 0)}

	var specs []*gauge.Specification
	var specParseResults []*parser.ParseResult

	for _, dir := range specDirs {
		specFiles := util.GetSpecFiles(filepath.Join(config.ProjectRoot, dir))
		specSlice, specParseResultsSlice := parser.ParseSpecFiles(specFiles, &gauge.ConceptDictionary{}, gauge.NewBuildErrors())
		specs = append(specs, specSlice...)
		specParseResults = append(specParseResults, specParseResultsSlice...)
	}

	addErrorsAndWarningsToRefactoringResult(result, specParseResults...)
	if !result.Success {
		return result
	}

	conceptDictionary, parseResult := parser.CreateConceptsDictionary()

	addErrorsAndWarningsToRefactoringResult(result, parseResult)
	if !result.Success {
		return result
	}

	refactorResult := agent.performRefactoringOn(specs, conceptDictionary)
	refactorResult.warnings = append(refactorResult.warnings, result.warnings...)
	return refactorResult
}

func killRunner(startChan *runner.StartChannels) {
	startChan.KillChan <- true
}

func rephraseFailure(errors ...string) *refactoringResult {
	return &refactoringResult{Success: false, Errors: errors}
}

func addErrorsAndWarningsToRefactoringResult(refactorResult *refactoringResult, parseResults ...*parser.ParseResult) {
	for _, parseResult := range parseResults {
		if !parseResult.Ok {
			refactorResult.Success = false
			for _, err := range parseResult.Errors() {
				refactorResult.Errors = append(refactorResult.Errors, err)
			}
		}
		refactorResult.appendWarnings(parseResult.Warnings)
	}
}

func (agent *rephraseRefactorer) performRefactoringOn(specs []*gauge.Specification, conceptDictionary *gauge.ConceptDictionary) *refactoringResult {
	specsRefactored, conceptFilesRefactored := agent.rephraseInSpecsAndConcepts(&specs, conceptDictionary)
	result := &refactoringResult{Success: false, Errors: make([]string, 0), warnings: make([]string, 0)}

	var runner runner.Runner
	select {
	case runner = <-agent.startChan.RunnerChan:
	case err := <-agent.startChan.ErrorChan:
		logger.Debugf("Cannot perform refactoring: Unable to connect to runner." + err.Error())
		return result
	}
	if !agent.isConcept {
		stepName, err, warning := agent.getStepNameFromRunner(runner)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			return result
		}
		if warning == nil {
			runnerFilesChanged, err := agent.requestRunnerForRefactoring(runner, stepName)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Cannot perform refactoring: %s", err))
				return result
			}
			result.runnerFilesChanged = runnerFilesChanged
		} else {
			result.warnings = append(result.warnings, warning.Message)
		}
	}
	specFiles, conceptFiles := writeToConceptAndSpecFiles(specs, conceptDictionary, specsRefactored, conceptFilesRefactored)
	result.specsChanged = specFiles
	result.Success = true
	result.conceptsChanged = conceptFiles
	return result
}

func (agent *rephraseRefactorer) rephraseInSpecsAndConcepts(specs *[]*gauge.Specification, conceptDictionary *gauge.ConceptDictionary) (map[*gauge.Specification]bool, map[string]bool) {
	specsRefactored := make(map[*gauge.Specification]bool, 0)
	conceptFilesRefactored := make(map[string]bool, 0)
	orderMap := agent.createOrderOfArgs()
	for _, spec := range *specs {
		specsRefactored[spec] = spec.RenameSteps(*agent.oldStep, *agent.newStep, orderMap)
	}
	isConcept := false
	for _, concept := range conceptDictionary.ConceptsMap {
		_, ok := conceptFilesRefactored[concept.FileName]
		conceptFilesRefactored[concept.FileName] = !ok && false || conceptFilesRefactored[concept.FileName]
		for _, item := range concept.ConceptStep.Items {
			isRefactored := conceptFilesRefactored[concept.FileName]
			conceptFilesRefactored[concept.FileName] = item.Kind() == gauge.StepKind &&
				item.(*gauge.Step).Rename(*agent.oldStep, *agent.newStep, isRefactored, orderMap, &isConcept) ||
				isRefactored
		}
	}
	agent.isConcept = isConcept
	return specsRefactored, conceptFilesRefactored
}

func (agent *rephraseRefactorer) createOrderOfArgs() map[int]int {
	orderMap := make(map[int]int, len(agent.newStep.Args))
	for i, arg := range agent.newStep.Args {
		orderMap[i] = SliceIndex(len(agent.oldStep.Args), func(i int) bool { return agent.oldStep.Args[i].String() == arg.String() })
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

func getRefactorAgent(oldStepText, newStepText string, startChan *runner.StartChannels) (*rephraseRefactorer, []parser.ParseError) {
	specParser := new(parser.SpecParser)
	stepTokens, errs := specParser.GenerateTokens("* "+oldStepText+"\n"+"*"+newStepText, "")
	if len(errs) > 0 {
		return nil, errs
	}

	steps := make([]*gauge.Step, 0)
	for _, stepToken := range stepTokens {
		step, parseRes := parser.CreateStepUsingLookup(stepToken, nil, "")
		if parseRes != nil && len(parseRes.ParseErrors) > 0 {
			return nil, parseRes.ParseErrors
		}
		steps = append(steps, step)
	}
	return &rephraseRefactorer{oldStep: steps[0], newStep: steps[1], startChan: startChan}, []parser.ParseError{}
}

func (agent *rephraseRefactorer) requestRunnerForRefactoring(testRunner runner.Runner, stepName string) ([]string, error) {
	refactorRequest, err := agent.createRefactorRequest(testRunner, stepName)
	if err != nil {
		return nil, err
	}
	refactorResponse := agent.sendRefactorRequest(testRunner, refactorRequest)
	var runnerError error
	if !refactorResponse.GetSuccess() {
		logger.APILog.Errorf("Refactoring error response from runner: %v", refactorResponse.GetError())
		runnerError = errors.New(refactorResponse.GetError())
	}
	return refactorResponse.GetFilesChanged(), runnerError
}

func (agent *rephraseRefactorer) sendRefactorRequest(testRunner runner.Runner, refactorRequest *gauge_messages.Message) *gauge_messages.RefactorResponse {
	response, err := conn.GetResponseForMessageWithTimeout(refactorRequest, testRunner.Connection(), config.RefactorTimeout())
	if err != nil {
		return &gauge_messages.RefactorResponse{Success: false, Error: err.Error()}
	}
	return response.GetRefactorResponse()
}

//Todo: Check for inline tables
func (agent *rephraseRefactorer) createRefactorRequest(runner runner.Runner, stepName string) (*gauge_messages.Message, error) {
	oldStepValue, err := agent.getStepValueFor(agent.oldStep, stepName)
	if err != nil {
		return nil, err
	}
	orderMap := agent.createOrderOfArgs()
	newStepName := agent.generateNewStepName(oldStepValue.Args, orderMap)
	newStepValue, err := parser.ExtractStepValueAndParams(newStepName, false)
	if err != nil {
		return nil, err
	}
	oldProtoStepValue := gauge.ConvertToProtoStepValue(oldStepValue)
	newProtoStepValue := gauge.ConvertToProtoStepValue(newStepValue)
	return &gauge_messages.Message{MessageType: gauge_messages.Message_RefactorRequest, RefactorRequest: &gauge_messages.RefactorRequest{OldStepValue: oldProtoStepValue, NewStepValue: newProtoStepValue, ParamPositions: agent.createParameterPositions(orderMap)}}, nil
}

func (agent *rephraseRefactorer) generateNewStepName(args []string, orderMap map[int]int) string {
	agent.newStep.PopulateFragments()
	paramIndex := 0
	for _, fragment := range agent.newStep.Fragments {
		if fragment.GetFragmentType() == gauge_messages.Fragment_Parameter {
			if orderMap[paramIndex] != -1 {
				fragment.GetParameter().Value = args[orderMap[paramIndex]]
			}
			paramIndex++
		}
	}
	return parser.ConvertToStepText(agent.newStep.Fragments)
}

func (agent *rephraseRefactorer) getStepNameFromRunner(runner runner.Runner) (string, error, *parser.Warning) {
	stepNameMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_StepNameRequest, StepNameRequest: &gauge_messages.StepNameRequest{StepValue: agent.oldStep.Value}}
	responseMessage, err := conn.GetResponseForMessageWithTimeout(stepNameMessage, runner.Connection(), config.RunnerRequestTimeout())
	if err != nil {
		return "", err, nil
	}
	if !(responseMessage.GetStepNameResponse().GetIsStepPresent()) {
		return "", nil, &parser.Warning{Message: fmt.Sprintf("Step implementation not found: %s", agent.oldStep.LineText)}
	}
	if responseMessage.GetStepNameResponse().GetHasAlias() {
		return "", fmt.Errorf("steps with aliases : '%s' cannot be refactored.", strings.Join(responseMessage.GetStepNameResponse().GetStepName(), "', '")), nil
	}
	return responseMessage.GetStepNameResponse().GetStepName()[0], nil, nil
}

func (agent *rephraseRefactorer) createParameterPositions(orderMap map[int]int) []*gauge_messages.ParameterPosition {
	paramPositions := make([]*gauge_messages.ParameterPosition, 0)
	for k, v := range orderMap {
		paramPositions = append(paramPositions, &gauge_messages.ParameterPosition{NewPosition: int32(k), OldPosition: int32(v)})
	}
	return paramPositions
}

func (agent *rephraseRefactorer) getStepValueFor(step *gauge.Step, stepName string) (*gauge.StepValue, error) {
	return parser.ExtractStepValueAndParams(stepName, false)
}

func writeToConceptAndSpecFiles(specs []*gauge.Specification, conceptDictionary *gauge.ConceptDictionary, specsRefactored map[*gauge.Specification]bool, conceptFilesRefactored map[string]bool) ([]string, []string) {
	specFiles := make([]string, 0)
	conceptFiles := make([]string, 0)
	for _, spec := range specs {
		if specsRefactored[spec] {
			specFiles = append(specFiles, spec.FileName)
			formatted := formatter.FormatSpecification(spec)
			util.SaveFile(spec.FileName, formatted, true)
		}
	}
	conceptMap := formatter.FormatConcepts(conceptDictionary)
	for fileName, concept := range conceptMap {
		if conceptFilesRefactored[fileName] {
			conceptFiles = append(conceptFiles, fileName)
			util.SaveFile(fileName, concept, true)
		}
	}
	return specFiles, conceptFiles
}

func (refactoringResult *refactoringResult) appendWarnings(warnings []*parser.Warning) {
	if refactoringResult.warnings == nil {
		refactoringResult.warnings = make([]string, 0)
	}
	for _, warning := range warnings {
		refactoringResult.warnings = append(refactoringResult.warnings, warning.Message)
	}
}

func (refactoringResult *refactoringResult) AllFilesChanges() []string {
	filesChanged := make([]string, 0)
	filesChanged = append(filesChanged, refactoringResult.specsChanged...)
	filesChanged = append(filesChanged, refactoringResult.conceptsChanged...)
	filesChanged = append(filesChanged, refactoringResult.runnerFilesChanged...)
	return filesChanged
}

func printRefactoringSummary(refactoringResult *refactoringResult) {
	exitCode := 0
	if !refactoringResult.Success {
		exitCode = 1
		for _, err := range refactoringResult.Errors {
			logger.Errorf("%s \n", err)
		}
	}
	for _, warning := range refactoringResult.warnings {
		logger.Warningf("%s \n", warning)
	}
	logger.Infof("%d specifications changed.\n", len(refactoringResult.specsChanged))
	logger.Infof("%d concepts changed.\n", len(refactoringResult.conceptsChanged))
	logger.Infof("%d files in code changed.\n", len(refactoringResult.runnerFilesChanged))
	os.Exit(exitCode)
}

func RefactorSteps(oldStep, newStep string, startChan *runner.StartChannels, specDirs []string) {
	refactoringResult := PerformRephraseRefactoring(oldStep, newStep, startChan, specDirs)
	printRefactoringSummary(refactoringResult)
}
