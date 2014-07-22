package main

import (
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"log"
	"strconv"
	"sync"
)

const (
	apiPortEnvVariableName = "GAUGE_API_PORT"
	API_STATIC_PORT        = 8889
)

func makeListOfAvailableSteps(runner *testRunner) {
	addStepValuesToAvailableSteps(getStepsFromRunner(runner))
	specFiles := findSpecsFilesIn(common.SpecsDirectoryName)
	dictionary, _ := createConceptsDictionary(true)
	availableSpecs = parseSpecFiles(specFiles, dictionary)
	findAvailableStepsInSpecs(availableSpecs)
}

func parseSpecFiles(specFiles []string, dictionary *conceptDictionary) []*specification {
	specs := make([]*specification, 0)
	for _, file := range specFiles {
		specContent, err := common.ReadFileContents(file)
		if err != nil {
			continue
		}
		parser := new(specParser)
		specification, result := parser.parse(specContent, dictionary)

		if result.ok {
			specs = append(specs, specification)
		}
	}
	return specs
}

func findAvailableStepsInSpecs(specs []*specification) {
	for _, spec := range specs {
		addStepsToAvailableSteps(spec.contexts)
		for _, scenario := range spec.scenarios {
			addStepsToAvailableSteps(scenario.steps)
		}
	}
}

func addStepsToAvailableSteps(steps []*step) {
	for _, step := range steps {
		if _, ok := availableStepsMap[step.value]; !ok {
			availableStepsMap[step.value] = true
		}
	}
}

func addStepValuesToAvailableSteps(stepValues []string) {
	for _, step := range stepValues {
		addToAvailableSteps(step)
	}
}

func addToAvailableSteps(step string) {
	if _, ok := availableStepsMap[step]; !ok {
		availableStepsMap[step] = true
	}
}

func getAvailableStepNames() []string {
	stepNames := make([]string, 0)
	for stepName, _ := range availableStepsMap {
		stepNames = append(stepNames, stepName)
	}
	return stepNames
}

func getStepsFromRunner(runner *testRunner) []string {
	steps := make([]string, 0)
	if runner == nil {
		runner, connErr := startRunnerAndMakeConnection(getProjectManifest())
		if connErr == nil {
			steps = append(steps, requestForSteps(runner)...)
			runner.kill()
		}
	} else {
		steps = append(steps, requestForSteps(runner)...)
	}
	return steps
}

func requestForSteps(runner *testRunner) []string {
	message, err := getResponseForGaugeMessage(createGetStepNamesRequest(), runner.connectionHandler)
	if err == nil {
		allStepsResponse := message.GetStepNamesResponse()
		return allStepsResponse.GetSteps()
	}
	return make([]string, 0)
}

func createGetStepNamesRequest() *Message {
	return &Message{MessageType: Message_StepNamesRequest.Enum(), StepNamesRequest: &StepNamesRequest{}}
}

func startAPIService(port int) error {
	gaugeConnectionHandler, err := newGaugeConnectionHandler(port, new(gaugeApiMessageHandler))
	if err != nil {
		return err
	}
	if port == 0 {
		if err := common.SetEnvVariable(apiPortEnvVariableName, strconv.Itoa(gaugeConnectionHandler.connectionPortNumber())); err != nil {
			return errors.New(fmt.Sprintf("Failed to set Env variable %s. %s", apiPortEnvVariableName, err.Error()))
		}
	}
	go gaugeConnectionHandler.handleMultipleConnections()
	return nil
}

func runAPIServiceIndefinitely(port int, wg *sync.WaitGroup) {
	wg.Add(1)
	startAPIService(port)
}

type gaugeApiMessageHandler struct{}

func (handler *gaugeApiMessageHandler) messageBytesReceived(bytesRead []byte, connectionHandler *gaugeConnectionHandler) {
	apiMessage := &APIMessage{}
	var responseMessage *APIMessage
	err := proto.Unmarshal(bytesRead, apiMessage)
	if err != nil {
		log.Printf("[Warning] Failed to read API proto message: %s\n", err.Error())
		responseMessage = handler.getErrorMessage(err)
	} else {
		messageType := apiMessage.GetMessageType()
		switch messageType {
		case APIMessage_GetProjectRootRequest:
			responseMessage = handler.projectRootRequestResponse(apiMessage)
			break
		case APIMessage_GetInstallationRootRequest:
			responseMessage = handler.installationRootRequestResponse(apiMessage)
			break
		case APIMessage_GetAllStepsRequest:
			responseMessage = handler.getAllStepsRequestResponse(apiMessage)
			break
		case APIMessage_GetAllSpecsRequest:
			responseMessage = handler.getAllSpecsRequestResponse(apiMessage)
			break
		case APIMessage_GetStepValueRequest:
			responseMessage = handler.getStepValueRequestResponse(apiMessage)
			break
		}
	}
	handler.sendMessage(responseMessage, connectionHandler)
}

func (handler *gaugeApiMessageHandler) sendMessage(message *APIMessage, connectionHandler *gaugeConnectionHandler) {
	dataBytes, err := proto.Marshal(message)
	if err != nil {
		fmt.Printf("[Warning] Failed to respond to API request. Could not Marshal response %s\n", err.Error())
	}
	if err := connectionHandler.write(dataBytes); err != nil {
		fmt.Printf("[Warning] Failed to respond to API request. Could not write response %s\n", err.Error())
	}
}

func (handler *gaugeApiMessageHandler) projectRootRequestResponse(message *APIMessage) *APIMessage {
	root, err := common.GetProjectRoot()
	if err != nil {
		fmt.Printf("[Warning] Failed to find project root while responding to API request. %s\n", err.Error())
		root = ""
	}
	projectRootResponse := &GetProjectRootResponse{ProjectRoot: proto.String(root)}
	return &APIMessage{MessageType: APIMessage_GetProjectRootResponse.Enum(), MessageId: message.MessageId, ProjectRootResponse: projectRootResponse}

}

func (handler *gaugeApiMessageHandler) installationRootRequestResponse(message *APIMessage) *APIMessage {
	root, err := common.GetInstallationPrefix()
	if err != nil {
		fmt.Printf("[Warning] Failed to find installation root while responding to API request. %s\n", err.Error())
		root = ""
	}
	installationRootResponse := &GetInstallationRootResponse{InstallationRoot: proto.String(root)}
	return &APIMessage{MessageType: APIMessage_GetInstallationRootResponse.Enum(), MessageId: message.MessageId, InstallationRootResponse: installationRootResponse}
}

func (handler *gaugeApiMessageHandler) getAllStepsRequestResponse(message *APIMessage) *APIMessage {
	getAllStepsResponse := &GetAllStepsResponse{Steps: getAvailableStepNames()}
	return &APIMessage{MessageType: APIMessage_GetAllStepResponse.Enum(), MessageId: message.MessageId, AllStepsResponse: getAllStepsResponse}
}

func (handler *gaugeApiMessageHandler) getAllSpecsRequestResponse(message *APIMessage) *APIMessage {
	getAllSpecsResponse := handler.createGetAllSpecsResponseMessageFor(availableSpecs)
	return &APIMessage{MessageType: APIMessage_GetAllSpecsResponse.Enum(), MessageId: message.MessageId, AllSpecsResponse: getAllSpecsResponse}
}

func (handler *gaugeApiMessageHandler) getStepValueRequestResponse(message *APIMessage) *APIMessage {
	stepText := message.GetStepValueRequest().GetStepText()
	stepValue, params, err := extractStepValueAndParams(stepText)
	if err != nil {
		return handler.getErrorResponse(message, err)
	}
	stepValueResponse := &GetStepValueResponse{StepValue: proto.String(stepValue), Parameters: params}
	return &APIMessage{MessageType: APIMessage_GetStepValueResponse.Enum(), MessageId: message.MessageId, StepValueResponse: stepValueResponse}

}

func (handler *gaugeApiMessageHandler) getErrorResponse(message *APIMessage, err error) *APIMessage {
	errorResponse := &ErrorResponse{Error: proto.String(err.Error())}
	return &APIMessage{MessageType: APIMessage_ErrorResponse.Enum(), MessageId: message.MessageId, Error: errorResponse}

}

func (handler *gaugeApiMessageHandler) getErrorMessage(err error) *APIMessage {
	id := common.GetUniqueId()
	errorResponse := &ErrorResponse{Error: proto.String(err.Error())}
	return &APIMessage{MessageType: APIMessage_ErrorResponse.Enum(), MessageId: &id, Error: errorResponse}
}

func (handler *gaugeApiMessageHandler) createGetAllSpecsResponseMessageFor(specs []*specification) *GetAllSpecsResponse {
	protoSpecs := make([]*ProtoSpec, 0)
	for _, spec := range specs {
		protoSpecs = append(protoSpecs, convertToProtoSpec(spec))
	}
	return &GetAllSpecsResponse{Specs: protoSpecs}
}

func extractStepValueAndParams(stepText string) (string, []string, error) {
	stepValueWithPlaceHolders, args, err := processStepText(stepText)
	if err != nil {
		return "", nil, err
	}
	stepValue, _ := extractStepValueAndParameterTypes(stepValueWithPlaceHolders)
	return stepValue, args, nil

}
