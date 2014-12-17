package main

import ()

const ERROR_MESSAGE = "No Refactoring Agent Present"

type refactorAgent interface {
	refactor(specs *[]*specification, conceptDictionary *conceptDictionary)
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

func (agent *renameRefactorer) refactor(specs *[]*specification, conceptDictionary *conceptDictionary) {
	for _, spec := range *specs {
		spec.renameSteps(*agent.oldStep, *agent.newStep)
	}
	for _, concept := range conceptDictionary.conceptsMap {
		for _, item := range concept.conceptStep.items {
			if item.kind() == stepKind {
				item.(*step).rename(*agent.oldStep, *agent.newStep)
			}
		}
	}
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
