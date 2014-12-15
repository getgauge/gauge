package main

import ()

const ERROR_MESSAGE = "No Refactoring Agent Present"

type refactorAgent interface {
	refactor(specs *[]*specification)
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

func (agent *renameRefactorer) refactor(specs *[]*specification) {
	for _, spec := range *specs {
		spec.renameSteps(*agent.oldStep, *agent.newStep)
	}
}

func getRefactorAgent(oldStepText, newStepText string) (refactorAgent, error) {
	parser := new(specParser)
	stepTokens, err := parser.generateTokens("* " + oldStepText + "\n" + "*" + newStepText)
	if err != nil {
		return nil, err
	}
	spec := &specification{}

	oldStep, err := spec.createStepUsingLookup(stepTokens[0], nil)
	if err != nil {
		return nil, err
	}
	newStep, err := spec.createStepUsingLookup(stepTokens[1], nil)
	if err != nil {
		return nil, err
	}
	if len(stepTokens[0].args) == 0 && len(stepTokens[1].args) == 0 {
		return &renameRefactorer{oldStep: oldStep, newStep: newStep}, nil
	}
	return nil, &RefactoringError{errorMessage: ERROR_MESSAGE}
}
