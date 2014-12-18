package main

import ()

const ERROR_MESSAGE = "No Refactoring Agent Present"

type refactorAgent interface {
	refactor(specs *[]*specification, conceptDictionary *conceptDictionary) (map[*specification]bool, map[string]bool)
}

type renameRefactorer struct {
	oldStep *step
	newStep *step
}

type RefactoringError struct {
	errorMessage string
}

func (error *RefactoringError) Error() string {
	return error.errorMessage
}

func (agent *renameRefactorer) refactor(specs *[]*specification, conceptDictionary *conceptDictionary) (map[*specification]bool, map[string]bool) {
	specsRefactored := make(map[*specification]bool, 0)
	conceptFilesRefactored := make(map[string]bool, 0)
	for _, spec := range *specs {
		specsRefactored[spec] = spec.renameSteps(*agent.oldStep, *agent.newStep)
	}
	for _, concept := range conceptDictionary.conceptsMap {
		conceptFilesRefactored[concept.fileName] = false
	}
	for _, concept := range conceptDictionary.conceptsMap {
		for _, item := range concept.conceptStep.items {
			if item.kind() == stepKind {
				conceptFilesRefactored[concept.fileName] = item.(*step).rename(*agent.oldStep, *agent.newStep, conceptFilesRefactored[concept.fileName])
			}
		}
	}
	return specsRefactored, conceptFilesRefactored
}

func getRefactorAgent(oldStepText, newStepText string) (refactorAgent, error) {
	parser := new(specParser)
	stepTokens, err := parser.generateTokens("* " + oldStepText + "\n" + "*" + newStepText)
	if err != nil {
		return nil, err
	}
	spec := &specification{}
	steps := make([]*step, 0)
	for _, stepToken := range stepTokens {
		step, err := spec.createStepUsingLookup(stepToken, nil)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	if len(stepTokens[0].args) == 0 && len(stepTokens[1].args) == 0 {
		return &renameRefactorer{oldStep: steps[0], newStep: steps[1]}, nil
	}
	return nil, &RefactoringError{errorMessage: ERROR_MESSAGE}
}
