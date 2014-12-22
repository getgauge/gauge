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
	orderMap := createOrderOfArgs(*agent.oldStep, *agent.newStep)
	for _, spec := range *specs {
		specsRefactored[spec] = spec.renameSteps(*agent.oldStep, *agent.newStep, orderMap)
	}
	for _, concept := range conceptDictionary.conceptsMap {
		_, ok := conceptFilesRefactored[concept.fileName]
		conceptFilesRefactored[concept.fileName] = !ok && false || conceptFilesRefactored[concept.fileName]
		for _, item := range concept.conceptStep.items {
			isRefactored := conceptFilesRefactored[concept.fileName]
			conceptFilesRefactored[concept.fileName] = item.kind() == stepKind &&
				item.(*step).rename(*agent.oldStep, *agent.newStep, isRefactored, orderMap) ||
				isRefactored
		}
	}
	return specsRefactored, conceptFilesRefactored
}

func createOrderOfArgs(oldStep step, newStep step) map[int]int {
	orderMap := make(map[int]int)
	for i, arg := range oldStep.args {
		index := SliceIndex(len(oldStep.args), func(i int) bool { return newStep.args[i].String() == arg.String() })
		if index > -1 {
			orderMap[i] = index
		}
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
	if len(stepTokens[0].args) == len(stepTokens[1].args) {
		return &renameRefactorer{oldStep: steps[0], newStep: steps[1]}, nil
	}
	return nil, &RefactoringError{errorMessage: ERROR_MESSAGE}
}
