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

type ArgPosition struct {
	index            int
	isRemoved        bool
	previousArgIndex int
}

func createOrderOfArgs(oldStep step, newStep step) map[int]ArgPosition {
	orderMap := make(map[int]ArgPosition)
	isOldStepArgs := true
	otherArgs := newStep.args
	args := oldStep.args
	if len(oldStep.args) < len(newStep.args) {
		args = newStep.args
		otherArgs = oldStep.args
		isOldStepArgs = false
	}
	for i, arg := range args {
		index := SliceIndex(len(args), func(i int) bool { return len(otherArgs) > i && otherArgs[i].String() == arg.String() })
		isRemoved := isOldStepArgs
		if index > -1 {
			isRemoved = false
		}
		orderMap[i] = ArgPosition{index: index, isRemoved: isRemoved, previousArgIndex: i-1}
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
	return &renameRefactorer{oldStep: steps[0], newStep: steps[1]}, nil
}
