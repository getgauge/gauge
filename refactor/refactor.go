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

/*
   Given old and new step gives the filenames of specification, concepts and files in code changed.

   Refactoring Flow:
	- Refactor specs and concepts in memory
	- Checks if it is a concept or not
	- In case of concept - writes to file and skips the runner
	- If its not a concept (its a step) - need to know the text, so makes a call to runner to get the text(step name)
	- Refactors the text(changes param positions ect) and sends it to runner to refactor implementations.
*/
package refactor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getgauge/gauge/config"
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
	runner    runner.Runner
}

type refactoringResult struct {
	Success            bool
	SpecsChanged       []*gauge_messages.FileChanges
	ConceptsChanged    map[string]string
	RunnerFilesChanged []*gauge_messages.FileChanges
	Errors             []string
	Warnings           []string
}

func (refactoringResult *refactoringResult) String() string {
	result := fmt.Sprintf("Refactoring result from gauge:\n")
	result += fmt.Sprintf("Specs changed        : %s\n", refactoringResult.specFilesChanged())
	result += fmt.Sprintf("Concepts changed     : %s\n", refactoringResult.conceptFilesChanged())
	result += fmt.Sprintf("Source files changed : %s\n", refactoringResult.runnerFilesChanged())
	result += fmt.Sprintf("Warnings             : %s\n", refactoringResult.Warnings)
	return result
}

func (refactoringResult *refactoringResult) appendWarnings(warnings []*parser.Warning) {
	if refactoringResult.Warnings == nil {
		refactoringResult.Warnings = make([]string, 0)
	}
	for _, warning := range warnings {
		refactoringResult.Warnings = append(refactoringResult.Warnings, warning.Message)
	}
}

func (refactoringResult *refactoringResult) AllFilesChanged() []string {
	filesChanged := make([]string, 0)
	filesChanged = append(filesChanged, refactoringResult.specFilesChanged()...)
	filesChanged = append(filesChanged, refactoringResult.conceptFilesChanged()...)
	filesChanged = append(filesChanged, refactoringResult.runnerFilesChanged()...)
	return filesChanged
}

func (refactoringResult *refactoringResult) conceptFilesChanged() []string {
	filesChanged := make([]string, 0)
	for fileName := range refactoringResult.ConceptsChanged {
		filesChanged = append(filesChanged, fileName)
	}
	return filesChanged
}

func (refactoringResult *refactoringResult) specFilesChanged() []string {
	filesChanged := make([]string, 0)
	for _, filesChange := range refactoringResult.SpecsChanged {
		filesChanged = append(filesChanged, filesChange.FileName)
	}
	return filesChanged
}

func (refactoringResult *refactoringResult) runnerFilesChanged() []string {
	filesChanged := make([]string, 0)
	for _, fileChange := range refactoringResult.RunnerFilesChanged {
		filesChanged = append(filesChanged, fileChange.FileName)
	}
	return filesChanged
}

// PerformRephraseRefactoring given an old step and new step refactors specs and concepts in memory and if its a concept writes to file
// else invokes runner to get the step name and refactors the step and sends it to runner to refactor implementation.
func PerformRephraseRefactoring(oldStep, newStep string, startChan *runner.StartChannels, specDirs []string) *refactoringResult {
	defer killRunner(startChan)
	if newStep == oldStep {
		return &refactoringResult{Success: true}
	}
	agent, errs := getRefactorAgent(oldStep, newStep, nil)

	if len(errs) > 0 {
		var messages []string
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		return rephraseFailure(messages...)
	}

	result, specs, conceptDictionary := parseSpecsAndConcepts(specDirs)
	if !result.Success {
		return result
	}

	refactorResult := agent.performRefactoringOn(specs, conceptDictionary, startChan)
	refactorResult.Warnings = append(refactorResult.Warnings, result.Warnings...)
	return refactorResult
}

// GetRefactoringChanges given an old step and new step gives the list of steps that need to be changed to perform refactoring.
// It also provides the changes to be made on the implementation files.
func GetRefactoringChanges(oldStep, newStep string, runner runner.Runner, specDirs []string) *refactoringResult {
	if newStep == oldStep {
		return &refactoringResult{Success: true}
	}
	agent, errs := getRefactorAgent(oldStep, newStep, runner)

	if len(errs) > 0 {
		var messages []string
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		return rephraseFailure(messages...)
	}

	result, specs, conceptDictionary := parseSpecsAndConcepts(specDirs)
	if !result.Success {
		return result
	}

	refactorResult := agent.getRefactoringChangesFor(specs, conceptDictionary)
	refactorResult.Warnings = append(refactorResult.Warnings, result.Warnings...)
	return refactorResult
}

func parseSpecsAndConcepts(specDirs []string) (*refactoringResult, []*gauge.Specification, *gauge.ConceptDictionary) {
	result := &refactoringResult{Success: true, Errors: make([]string, 0), Warnings: make([]string, 0)}

	var specs []*gauge.Specification
	var specParseResults []*parser.ParseResult

	for _, dir := range specDirs {
		specFiles := util.GetSpecFiles([]string{filepath.Join(config.ProjectRoot, dir)})
		specSlice, specParseResultsSlice := parser.ParseSpecFiles(specFiles, &gauge.ConceptDictionary{}, gauge.NewBuildErrors())
		specs = append(specs, specSlice...)
		specParseResults = append(specParseResults, specParseResultsSlice...)
	}

	addErrorsAndWarningsToRefactoringResult(result, specParseResults...)
	if !result.Success {
		return result, nil, nil
	}

	conceptDictionary, parseResult, err := parser.CreateConceptsDictionary()
	if err != nil {
		return rephraseFailure(err.Error()), nil, nil
	}
	addErrorsAndWarningsToRefactoringResult(result, parseResult)
	return result, specs, conceptDictionary
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

func (agent *rephraseRefactorer) performRefactoringOn(specs []*gauge.Specification, conceptDictionary *gauge.ConceptDictionary, startChan *runner.StartChannels) *refactoringResult {
	specsRefactored, conceptFilesRefactored := agent.rephraseInSpecsAndConcepts(&specs, conceptDictionary)
	var runner runner.Runner
	select {
	case runner = <-startChan.RunnerChan:
	case err := <-startChan.ErrorChan:
		logger.Debugf(true, "Cannot perform refactoring: Unable to connect to runner."+err.Error())
		return &refactoringResult{Success: false, Errors: make([]string, 0), Warnings: make([]string, 0)}
	}
	agent.runner = runner
	result := agent.refactorStepImplementations(true)
	if !result.Success {
		return result
	}
	result.SpecsChanged, result.ConceptsChanged = getFileChanges(specs, conceptDictionary, specsRefactored, conceptFilesRefactored)
	writeFileChangesToDisk(result)
	return result
}

func writeFileChangesToDisk(result *refactoringResult) {
	for _, fileChange := range result.SpecsChanged {
		util.SaveFile(fileChange.FileName, fileChange.FileContent, true)
	}
	for fileName, text := range result.ConceptsChanged {
		util.SaveFile(fileName, text, true)
	}
}

func (agent *rephraseRefactorer) getRefactoringChangesFor(specs []*gauge.Specification, conceptDictionary *gauge.ConceptDictionary) *refactoringResult {
	specsRefactored, conceptFilesRefactored := agent.rephraseInSpecsAndConcepts(&specs, conceptDictionary)
	result := agent.refactorStepImplementations(false)
	if !result.Success {
		return result
	}
	result.SpecsChanged, result.ConceptsChanged = getFileChanges(specs, conceptDictionary, specsRefactored, conceptFilesRefactored)
	return result
}

func (agent *rephraseRefactorer) refactorStepImplementations(shouldSaveChanges bool) *refactoringResult {
	result := &refactoringResult{Success: false, Errors: make([]string, 0), Warnings: make([]string, 0)}
	if !agent.isConcept {
		stepName, err, warning := agent.getStepNameFromRunner(agent.runner)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			return result
		}
		if warning == nil {
			runnerFilesChanged, err := agent.requestRunnerForRefactoring(agent.runner, stepName, shouldSaveChanges)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Cannot perform refactoring: %s", err))
				return result
			}
			result.RunnerFilesChanged = runnerFilesChanged
		} else {
			result.Warnings = append(result.Warnings, warning.Message)
		}
	}
	result.Success = true
	return result
}

func (agent *rephraseRefactorer) rephraseInSpecsAndConcepts(specs *[]*gauge.Specification, conceptDictionary *gauge.ConceptDictionary) (map[*gauge.Specification][]*gauge.StepDiff, map[string]bool) {
	specsRefactored := make(map[*gauge.Specification][]*gauge.StepDiff, 0)
	conceptFilesRefactored := make(map[string]bool, 0)
	orderMap := agent.createOrderOfArgs()
	for _, spec := range *specs {
		diffs, isRefactored := spec.RenameSteps(*agent.oldStep, *agent.newStep, orderMap)
		if isRefactored {
			specsRefactored[spec] = diffs
		}
	}
	isConcept := false
	for _, concept := range conceptDictionary.ConceptsMap {
		_, ok := conceptFilesRefactored[concept.FileName]
		conceptFilesRefactored[concept.FileName] = !ok && false || conceptFilesRefactored[concept.FileName]
		for _, item := range concept.ConceptStep.Items {
			isRefactored := conceptFilesRefactored[concept.FileName]
			if item.Kind() == gauge.StepKind {
				if _, ok := item.(*gauge.Step).Rename(*agent.oldStep, *agent.newStep, isRefactored, orderMap, &isConcept); ok {
					conceptFilesRefactored[concept.FileName] = true
				}
			}
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

// SliceIndex gives the index of the args.
func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func getRefactorAgent(oldStepText, newStepText string, runner runner.Runner) (*rephraseRefactorer, []parser.ParseError) {
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
	return &rephraseRefactorer{oldStep: steps[0], newStep: steps[1], runner: runner}, []parser.ParseError{}
}

func (agent *rephraseRefactorer) requestRunnerForRefactoring(testRunner runner.Runner, stepName string, shouldSaveChanges bool) ([]*gauge_messages.FileChanges, error) {
	refactorRequest, err := agent.createRefactorRequest(testRunner, stepName, shouldSaveChanges)
	if err != nil {
		return nil, err
	}
	refactorResponse := agent.sendRefactorRequest(testRunner, refactorRequest)
	var runnerError error
	if !refactorResponse.GetSuccess() {
		logger.Errorf(false, "Refactoring error response from runner: %v", refactorResponse.GetError())
		runnerError = errors.New(refactorResponse.GetError())
	}
	return refactorResponse.GetFileChanges(), runnerError
}

func (agent *rephraseRefactorer) sendRefactorRequest(testRunner runner.Runner, refactorRequest *gauge_messages.Message) *gauge_messages.RefactorResponse {
	response, err := testRunner.ExecuteMessageWithTimeout(refactorRequest)
	if err != nil {
		return &gauge_messages.RefactorResponse{Success: false, Error: err.Error()}
	}
	return response.GetRefactorResponse()
}

//Todo: Check for inline tables
func (agent *rephraseRefactorer) createRefactorRequest(runner runner.Runner, stepName string, shouldSaveChanges bool) (*gauge_messages.Message, error) {
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
	return &gauge_messages.Message{MessageType: gauge_messages.Message_RefactorRequest,
		RefactorRequest: &gauge_messages.RefactorRequest{
			OldStepValue:   oldProtoStepValue,
			NewStepValue:   newProtoStepValue,
			ParamPositions: agent.createParameterPositions(orderMap),
			SaveChanges:    shouldSaveChanges,
		},
	}, nil
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
	responseMessage, err := runner.ExecuteMessageWithTimeout(stepNameMessage)
	if err != nil {
		return "", err, nil
	}
	if !(responseMessage.GetStepNameResponse().GetIsStepPresent()) {
		return "", nil, &parser.Warning{Message: fmt.Sprintf("Step implementation not found: %s", agent.oldStep.LineText)}
	}
	if responseMessage.GetStepNameResponse().GetHasAlias() {
		return "", fmt.Errorf("steps with aliases : '%s' cannot be refactored", strings.Join(responseMessage.GetStepNameResponse().GetStepName(), "', '")), nil
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

func createDiffs(diffs []*gauge.StepDiff) []*gauge_messages.TextDiff {
	textDiffs := []*gauge_messages.TextDiff{}
	for _, diff := range diffs {
		newtext := formatter.FormatStep(diff.NewStep)
		oldtext := formatter.FormatStep(&diff.OldStep)
		d := &gauge_messages.TextDiff{
			Span: &gauge_messages.Span{
				Start:     int64(diff.OldStep.LineNo),
				StartChar: 0,
				End:       int64(diff.OldStep.LineNo),
				EndChar:   int64(len(oldtext)),
			},
			Content: newtext,
		}
		textDiffs = append(textDiffs, d)
	}
	return textDiffs
}

func getFileChanges(specs []*gauge.Specification, conceptDictionary *gauge.ConceptDictionary, specsRefactored map[*gauge.Specification][]*gauge.StepDiff, conceptFilesRefactored map[string]bool) ([]*gauge_messages.FileChanges, map[string]string) {
	specFiles := []*gauge_messages.FileChanges{}
	conceptFiles := make(map[string]string, 0)
	for _, spec := range specs {
		if stepDiffs, ok := specsRefactored[spec]; ok {
			formatted := formatter.FormatSpecification(spec)
			specFiles = append(specFiles, &gauge_messages.FileChanges{FileName: spec.FileName, FileContent: formatted, Diffs: createDiffs(stepDiffs)})
		}
	}
	conceptMap := formatter.FormatConcepts(conceptDictionary)
	for fileName, text := range conceptMap {
		if conceptFilesRefactored[fileName] {
			conceptFiles[fileName] = text
		}
	}
	return specFiles, conceptFiles
}

func printRefactoringSummary(refactoringResult *refactoringResult) {
	exitCode := 0
	if !refactoringResult.Success {
		exitCode = 1
		for _, err := range refactoringResult.Errors {
			logger.Errorf(true, "%s \n", err)
		}
	}
	for _, warning := range refactoringResult.Warnings {
		logger.Warningf(true, "%s \n", warning)
	}
	logger.Infof(true, "%d specifications changed.\n", len(refactoringResult.specFilesChanged()))
	logger.Infof(true, "%d concepts changed.\n", len(refactoringResult.conceptFilesChanged()))
	logger.Infof(true, "%d files in code changed.\n", len(refactoringResult.RunnerFilesChanged))
	os.Exit(exitCode)
}

// RefactorSteps performs rephrase refactoring and prints the refactoring summary which includes errors and warnings generated during refactoring and
// files changed during refactoring : specification files, concept files and the implementation files changed.
func RefactorSteps(oldStep, newStep string, startChan *runner.StartChannels, specDirs []string) {
	refactoringResult := PerformRephraseRefactoring(oldStep, newStep, startChan, specDirs)
	printRefactoringSummary(refactoringResult)
}
