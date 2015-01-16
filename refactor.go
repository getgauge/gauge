package main

import (
	"errors"
	"fmt"
	"github.com/getgauge/gauge/config"
	"github.com/golang/protobuf/proto"
	"os"
	"strings"
)

type rephraseRefactorer struct {
	oldStep   *step
	newStep   *step
	isConcept bool
}

func (agent *rephraseRefactorer) refactor(specs *[]*specification, conceptDictionary *conceptDictionary) (map[*specification]bool, map[string]bool) {
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
		step, err := spec.createStepUsingLookup(stepToken, nil)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return &rephraseRefactorer{oldStep: steps[0], newStep: steps[1]}, nil
}

func (agent *rephraseRefactorer) requestRunnerForRefactoring(testRunner *testRunner, stepName string) {
	if agent.isConcept {
		return
	}
	refactorRequest, err := agent.createRefactorRequest(testRunner, stepName)
	if err != nil {
		fmt.Printf("%s", err)
		testRunner.kill()
		os.Exit(1)
	}
	agent.sendRefactorRequest(testRunner, refactorRequest)
	testRunner.kill()
}

func (agent *rephraseRefactorer) startRunner() *testRunner {
	loadGaugeEnvironment()
	startAPIService(0)
	testRunner, err := startRunnerAndMakeConnection(getProjectManifest())
	if err != nil {
		fmt.Printf("Failed to connect to test runner: %s", err)
		os.Exit(1)
	}
	return testRunner
}

func (agent *rephraseRefactorer) sendRefactorRequest(testRunner *testRunner, refactorRequest *Message) {
	response, err := getResponseForMessageWithTimeout(refactorRequest, testRunner.connection, config.RefactorTimeout())
	if err != nil {
		testRunner.kill()
		fmt.Printf("Effects only in spec and concept files: %s", err)
		os.Exit(1)
	} else if !response.GetRefactorResponse().GetSuccess() {
		fmt.Printf("Effects only in spec and concept files: %s", response.GetRefactorResponse().GetError())
		testRunner.kill()
		os.Exit(1)
	}
}

//Todo: Check for inline tables
func (agent *rephraseRefactorer) createRefactorRequest(runner *testRunner, stepName string) (*Message, error) {
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
	return &Message{MessageType: Message_RefactorRequest.Enum(), RefactorRequest: &RefactorRequest{OldStepValue: oldProtoStepValue, NewStepValue: newProtoStepValue, ParamPositions: agent.createParameterPositions(orderMap)}}, nil
}

func (agent *rephraseRefactorer) generateNewStepName(args []string, orderMap map[int]int) string {
	agent.newStep.populateFragments()
	paramIndex := 0
	for _, fragment := range agent.newStep.fragments {
		if fragment.GetFragmentType() == Fragment_Parameter {
			if orderMap[paramIndex] != -1 {
				fragment.GetParameter().Value = proto.String(args[orderMap[paramIndex]])
			}
			paramIndex++
		}
	}
	return convertToStepText(agent.newStep.fragments)
}

func (agent *rephraseRefactorer) getStepNameFromRunner(runner *testRunner) (error, string, bool) {
	stepNameMessage := &Message{MessageType: Message_StepNameRequest.Enum(), StepNameRequest: &GetStepNameRequest{StepValue: proto.String(agent.oldStep.value)}}
	responseMessage, err := getResponseForMessageWithTimeout(stepNameMessage, runner.connection, config.RunnerAPIRequestTimeout())
	if err != nil {
		return err, "", false
	}
	if !(responseMessage.GetStepNameResponse().GetIsStepPresent()) {
		fmt.Println("Step implementation not found: " + agent.oldStep.lineText)
		return nil, "", false
	}
	if responseMessage.GetStepNameResponse().GetHasAlias() {
		return errors.New(fmt.Sprintf("steps with aliases : '%s' cannot be refactored.", strings.Join(responseMessage.GetStepNameResponse().GetStepName(), "', '"))), "", false
	}

	return nil, responseMessage.GetStepNameResponse().GetStepName()[0], true
}

func (agent *rephraseRefactorer) createParameterPositions(orderMap map[int]int) []*ParameterPosition {
	paramPositions := make([]*ParameterPosition, 0)
	for k, v := range orderMap {
		paramPositions = append(paramPositions, &ParameterPosition{NewPosition: proto.Int(k), OldPosition: proto.Int(v)})
	}
	return paramPositions
}

func (agent *rephraseRefactorer) getStepValueFor(step *step, stepName string) (*stepValue, error) {
	return extractStepValueAndParams(stepName, false)
}
