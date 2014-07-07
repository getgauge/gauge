package main

import (
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"log"
	"net"
	"sync"
)

const (
	apiPortEnvVariableName = "GAUGE_API_PORT"
	API_STATIC_PORT        = 8889
)

func makeListOfAvailableSteps(runnerConn net.Conn) {
	addStepValuesToAvailableSteps(getStepsFromRunner(runnerConn))
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

func getStepsFromRunner(runnerConnection net.Conn) []string {
	steps := make([]string, 0)
	if runnerConnection == nil {
		var connErr error
		runnerConnection, connErr = startRunnerAndMakeConnection(getProjectManifest())
		if connErr == nil {
			steps = append(steps, requestForSteps(runnerConnection)...)
			killRunner(runnerConnection)
		}
	} else {
		steps = append(steps, requestForSteps(runnerConnection)...)
	}
	return steps
}

func requestForSteps(connection net.Conn) []string {
	message, err := getResponse(connection, createGetStepNamesRequest())
	if err == nil {
		allStepsResponse := message.GetStepNamesResponse()
		return allStepsResponse.GetSteps()
	}
	return make([]string, 0)
}

func killRunner(connection net.Conn) error {
	id := common.GetUniqueId()
	message := &Message{MessageId: &id, MessageType: Message_KillProcessRequest.Enum(),
		KillProcessRequest: &KillProcessRequest{}}

	return writeMessage(connection, message)
}

func createGetStepNamesRequest() *Message {
	return &Message{MessageType: Message_StepNamesRequest.Enum(), StepNamesRequest: &StepNamesRequest{}}
}

func startAPIService(port int) error {
	gaugeListener, err := newGaugeListener(port)
	if err != nil {
		return err
	}
	if port == 0 {
		if err := common.SetEnvVariable(apiPortEnvVariableName, gaugeListener.portNumber()); err != nil {
			return errors.New(fmt.Sprintf("Failed to set Env variable %s. %s", apiPortEnvVariableName, err.Error()))
		}
	}
	go gaugeListener.acceptConnections(&GaugeApiMessageHandler{})
	return nil
}

func runAPIServiceIndefinitely(port int, wg *sync.WaitGroup) {
	wg.Add(1)
	startAPIService(port)
}

type GaugeApiMessageHandler struct{}

func (handler *GaugeApiMessageHandler) messageReceived(bytesRead []byte, conn net.Conn) {
	apiMessage := &APIMessage{}
	err := proto.Unmarshal(bytesRead, apiMessage)
	if err != nil {
		log.Printf("[Warning] Failed to read API proto message: %s\n", err.Error())
		handler.sendErrorMessage(err, conn)
	} else {
		messageType := apiMessage.GetMessageType()
		switch messageType {
		case APIMessage_GetProjectRootRequest:
			handler.respondToProjectRootRequest(apiMessage, conn)
			break
		case APIMessage_GetAllStepsRequest:
			handler.respondToGetAllStepsRequest(apiMessage, conn)
			break
		case APIMessage_GetAllSpecsRequest:
			handler.respondToGetAllSpecsRequest(apiMessage, conn)
			break
		case APIMessage_GetStepValueRequest:
			handler.respondToGetStepValueRequest(apiMessage, conn)
			break
		}
	}
}

func (handler *GaugeApiMessageHandler) respondToProjectRootRequest(message *APIMessage, conn net.Conn) {
	root, err := common.GetProjectRoot()
	if err != nil {
		fmt.Printf("[Warning] Failed to find project root while responding to API request. %s\n", err.Error())
		root = ""
	}
	projectRootResponse := &GetProjectRootResponse{ProjectRoot: proto.String(root)}
	responseApiMessage := &APIMessage{MessageType: APIMessage_GetProjectRootResponse.Enum(), MessageId: message.MessageId, ProjectRootResponse: projectRootResponse}
	handler.sendMessage(responseApiMessage, conn)
}

func (handler *GaugeApiMessageHandler) respondToGetAllStepsRequest(message *APIMessage, conn net.Conn) {
	getAllStepsResponse := &GetAllStepsResponse{Steps: getAvailableStepNames()}
	responseApiMessage := &APIMessage{MessageType: APIMessage_GetAllStepResponse.Enum(), MessageId: message.MessageId, AllStepsResponse: getAllStepsResponse}
	handler.sendMessage(responseApiMessage, conn)
}

func (handler *GaugeApiMessageHandler) respondToGetAllSpecsRequest(message *APIMessage, conn net.Conn) {
	getAllSpecsResponse := handler.createGetAllSpecsResponseMessageFor(availableSpecs)
	responseApiMessage := &APIMessage{MessageType: APIMessage_GetAllSpecsResponse.Enum(), MessageId: message.MessageId, AllSpecsResponse: getAllSpecsResponse}
	handler.sendMessage(responseApiMessage, conn)
}

func (handler *GaugeApiMessageHandler) respondToGetStepValueRequest(message *APIMessage, conn net.Conn) {
	stepText := message.GetStepValueRequest().GetStepText()
	stepValue, params, err := extractStepValueAndParams(stepText)
	if err != nil {
		handler.sendErrorResponse(err, message, conn)
		return
	}
	stepValueResponse := &GetStepValueResponse{StepValue: proto.String(stepValue), Parameters: params}
	responseApiMessage := &APIMessage{MessageType: APIMessage_GetStepValueResponse.Enum(), MessageId: message.MessageId, StepValueResponse: stepValueResponse}
	handler.sendMessage(responseApiMessage, conn)
}

func (handler *GaugeApiMessageHandler) sendErrorResponse(err error, message *APIMessage, conn net.Conn) {
	handler.sendErrorResponseWithId(err, message.MessageId, conn)
}

func (handler *GaugeApiMessageHandler) sendErrorMessage(err error, conn net.Conn) {
	id := common.GetUniqueId()
	handler.sendErrorResponseWithId(err, &id, conn)
}

func (handler *GaugeApiMessageHandler) sendErrorResponseWithId(err error, id *int64, conn net.Conn) {
	errorResponse := &ErrorResponse{Error: proto.String(err.Error())}
	responseApiMessage := &APIMessage{MessageType: APIMessage_ErrorResponse.Enum(), MessageId: id, Error: errorResponse}
	handler.sendMessage(responseApiMessage, conn)
}

func (handler *GaugeApiMessageHandler) createGetAllSpecsResponseMessageFor(specs []*specification) *GetAllSpecsResponse {
	protoSpecs := make([]*ProtoSpec, 0)
	for _, spec := range specs {
		protoSpecs = append(protoSpecs, convertToProtoSpec(spec))
	}
	return &GetAllSpecsResponse{Specs: protoSpecs}
}

func (handler *GaugeApiMessageHandler) sendMessage(message *APIMessage, conn net.Conn) {
	if err := writeMessage(conn, message); err != nil {
		fmt.Printf("[Warning] Failed to respond to API request. %s\n", err.Error())
	}
}

func extractStepValueAndParams(stepText string) (string, []string, error) {
	stepValueWithPlaceHolders, args, err := processStepText(stepText)
	if err != nil {
		return "", nil, err
	}
	stepValue, _ := extractStepValueAndParameterTypes(stepValueWithPlaceHolders)
	return stepValue, args, nil

}
