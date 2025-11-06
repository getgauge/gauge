/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
	"path/filepath"
	"strings"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
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
	ConceptsChanged    []*gauge_messages.FileChanges
	RunnerFilesChanged []*gauge_messages.FileChanges
	Errors             []string
	Warnings           []string
}

func (res *refactoringResult) String() string {
	o := `Refactoring result from gauge:
Specs changed        : %s
Concepts changed     : %s
Source files changed : %s
Warnings             : %s
`
	return fmt.Sprintf(o, res.specFilesChanged(), res.conceptFilesChanged(), res.runnerFilesChanged(), res.Warnings)
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
	for _, fileChange := range refactoringResult.ConceptsChanged {
		filesChanged = append(filesChanged, fileChange.FileName)
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

func (refactoringResult *refactoringResult) WriteToDisk() {
	if !refactoringResult.Success {
		return
	}
	// fileChange.FileContent need not be deprecated. To save the refactored file, it is much simpler and less error prone
	// to replace the file with new content, rather than parsing again and replacing specific lines.
	for _, fileChange := range refactoringResult.SpecsChanged {
		util.SaveFile(fileChange.FileName, fileChange.FileContent, true) //nolint:staticcheck
	}
	for _, fileChange := range refactoringResult.ConceptsChanged {
		util.SaveFile(fileChange.FileName, fileChange.FileContent, true) //nolint:staticcheck
	}
}

// GetRefactoringChanges given an old step and new step gives the list of steps that need to be changed to perform refactoring.
// It also provides the changes to be made on the implementation files.
func GetRefactoringChanges(oldStep, newStep string, r runner.Runner, specDirs []string, saveToDisk bool) *refactoringResult {
	if newStep == oldStep {
		return &refactoringResult{Success: true}
	}
	agent, errs := getRefactorAgent(oldStep, newStep, r)

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

	refactorResult := agent.getRefactoringChangesFor(specs, conceptDictionary, saveToDisk)
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

func rephraseFailure(errs ...string) *refactoringResult {
	return &refactoringResult{Success: false, Errors: errs}
}

func addErrorsAndWarningsToRefactoringResult(refactorResult *refactoringResult, parseResults ...*parser.ParseResult) {
	for _, parseResult := range parseResults {
		if !parseResult.Ok {
			refactorResult.Success = false
			refactorResult.Errors = append(refactorResult.Errors, parseResult.Errors()...)
		}
		refactorResult.appendWarnings(parseResult.Warnings)
	}
}

func (agent *rephraseRefactorer) getRefactoringChangesFor(specs []*gauge.Specification, conceptDictionary *gauge.ConceptDictionary, saveToDisk bool) *refactoringResult {
	specsRefactored, conceptFilesRefactored := agent.rephraseInSpecsAndConcepts(&specs, conceptDictionary)
	result := agent.refactorStepImplementations(saveToDisk)
	if !result.Success {
		return result
	}
	result.SpecsChanged, result.ConceptsChanged = getFileChanges(specs, conceptDictionary, specsRefactored, conceptFilesRefactored)
	return result
}

func (agent *rephraseRefactorer) refactorStepImplementations(shouldSaveChanges bool) *refactoringResult {
	result := &refactoringResult{Success: false, Errors: make([]string, 0), Warnings: make([]string, 0)}
	if !agent.isConcept {
		stepName, warning, err := agent.getStepNameFromRunner(agent.runner)
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

func (agent *rephraseRefactorer) rephraseInSpecsAndConcepts(specs *[]*gauge.Specification, conceptDictionary *gauge.ConceptDictionary) (map[*gauge.Specification][]*gauge.StepDiff, map[string][]*gauge.StepDiff) {
	specsRefactored := make(map[*gauge.Specification][]*gauge.StepDiff)
	conceptsRefactored := make(map[string][]*gauge.StepDiff)
	orderMap := agent.createOrderOfArgs()
	for _, spec := range *specs {
		diffs, isRefactored := spec.RenameSteps(agent.oldStep, agent.newStep, orderMap)
		if isRefactored {
			specsRefactored[spec] = diffs
		}
	}
	isConcept := false
	for _, concept := range conceptDictionary.ConceptsMap {
		isRefactored := false
		for _, item := range concept.ConceptStep.Items {
			if item.Kind() == gauge.StepKind {
				diff, isRefactored := item.(*gauge.Step).Rename(agent.oldStep, agent.newStep, isRefactored, orderMap, &isConcept)
				if isRefactored {
					conceptsRefactored[concept.FileName] = append(conceptsRefactored[concept.FileName], diff)
				}
			}
		}
	}
	agent.isConcept = isConcept
	return specsRefactored, conceptsRefactored
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

func getRefactorAgent(oldStepText, newStepText string, r runner.Runner) (*rephraseRefactorer, []parser.ParseError) {
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
	return &rephraseRefactorer{oldStep: steps[0], newStep: steps[1], runner: r}, []parser.ParseError{}
}

func (agent *rephraseRefactorer) requestRunnerForRefactoring(testRunner runner.Runner, stepName string, shouldSaveChanges bool) ([]*gauge_messages.FileChanges, error) {
	refactorRequest, err := agent.createRefactorRequest(stepName, shouldSaveChanges)
	if err != nil {
		return nil, err
	}
	refactorResponse := agent.sendRefactorRequest(testRunner, refactorRequest)
	var runnerError error
	if !refactorResponse.GetSuccess() {
		logger.Errorf(false, "Refactoring error response from runner: %v", refactorResponse.GetError())
		runnerError = errors.New(refactorResponse.GetError())
	}
	if len(refactorResponse.GetFileChanges()) == 0 {
		for _, file := range refactorResponse.GetFilesChanged() {
			refactorResponse.FileChanges = append(refactorResponse.FileChanges, &gauge_messages.FileChanges{FileName: file})
		}
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

// Todo: Check for inline tables
func (agent *rephraseRefactorer) createRefactorRequest(stepName string, shouldSaveChanges bool) (*gauge_messages.Message, error) {
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

func (agent *rephraseRefactorer) getStepNameFromRunner(r runner.Runner) (string, *parser.Warning, error) {
	stepNameMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_StepNameRequest, StepNameRequest: &gauge_messages.StepNameRequest{StepValue: agent.oldStep.Value}}
	responseMessage, err := r.ExecuteMessageWithTimeout(stepNameMessage)

	if err != nil {
		return "", nil, err
	}
	if !(responseMessage.GetStepNameResponse().GetIsStepPresent()) {
		return "", &parser.Warning{Message: fmt.Sprintf("Step implementation not found: %s", agent.oldStep.LineText)}, nil
	}
	if responseMessage.GetStepNameResponse().GetHasAlias() {
		return "", nil, fmt.Errorf("steps with aliases : '%s' cannot be refactored", strings.Join(responseMessage.GetStepNameResponse().GetStepName(), "', '"))
	}
	if responseMessage.GetStepNameResponse().GetIsExternal() {
		return "", nil, fmt.Errorf("external step: Cannot refactor '%s' is in external project or library", strings.Join(responseMessage.GetStepNameResponse().GetStepName(), "', '"))
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
		newtext := strings.TrimSpace(formatter.FormatStep(diff.NewStep))
		if diff.IsConcept && !diff.OldStep.InConcept() {
			newtext = strings.ReplaceAll(newtext, "*", "#")
		}
		oldFragments := util.GetLinesFromText(strings.TrimSpace(formatter.FormatStep(&diff.OldStep)))
		d := &gauge_messages.TextDiff{
			Span: &gauge_messages.Span{
				Start:     int64(diff.OldStep.LineNo),
				StartChar: 0,
				End:       int64(diff.OldStep.LineNo + len(oldFragments) - 1),
				EndChar:   int64(len(oldFragments[len(oldFragments)-1])),
			},
			Content: newtext,
		}
		textDiffs = append(textDiffs, d)
	}
	return textDiffs
}

func getFileChanges(specs []*gauge.Specification, conceptDictionary *gauge.ConceptDictionary, specsRefactored map[*gauge.Specification][]*gauge.StepDiff, conceptsRefactored map[string][]*gauge.StepDiff) ([]*gauge_messages.FileChanges, []*gauge_messages.FileChanges) {
	specFiles := []*gauge_messages.FileChanges{}
	conceptFiles := []*gauge_messages.FileChanges{}
	for _, spec := range specs {
		if stepDiffs, ok := specsRefactored[spec]; ok {
			formatted := formatter.FormatSpecification(spec)
			specFiles = append(specFiles, &gauge_messages.FileChanges{FileName: spec.FileName, FileContent: formatted, Diffs: createDiffs(stepDiffs)})
		}
	}
	conceptMap := formatter.FormatConcepts(conceptDictionary)
	for file, diffs := range conceptsRefactored {
		conceptFiles = append(conceptFiles, &gauge_messages.FileChanges{FileName: file, FileContent: conceptMap[file], Diffs: createDiffs(diffs)})
	}
	return specFiles, conceptFiles
}
