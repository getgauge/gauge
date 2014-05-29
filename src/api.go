package main

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/getgauge/common"
	"log"
	"net"
)

const (
	apiPortEnvVariableName = "GAUGE_API_PORT"
	API_STATIC_PORT        = 8889
)

var availableStepsMap = make(map[string]bool)

func makeListOfAvailableSteps() {
	addStepValuesToAvailableSteps(getStepsFromRunner())
	specFiles := findSpecsFilesIn(common.SpecsDirectoryName)
	findStepsInSpecFiles(specFiles)
}

func findStepsInSpecFiles(specFiles []string) {
	parser := new(specParser)
	for _, file := range specFiles {
		scenarioContent, err := common.ReadFileContents(file)
		if err != nil {
			continue
		}
		specification, result := parser.parse(scenarioContent, new(conceptDictionary))

		if result.ok {
			addStepsToAvailableSteps(specification.contexts)
			for _, scenario := range specification.scenarios {
				addStepsToAvailableSteps(scenario.steps)
			}
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

func getStepsFromRunner() []string {
	steps := make([]string, 0)
	runnerConnection, connErr := startRunnerAndMakeConnection(getProjectManifest())
	if connErr == nil {
		message, err := getResponse(runnerConnection, createGetStepValueRequest())
		if err == nil {
			allStepsResponse := message.GetStepNamesResponse()
			steps = append(steps, allStepsResponse.GetSteps()...)
		}
		killRunner(runnerConnection)
	}
	return steps

}

func killRunner(connection net.Conn) error {
	message := &Message{MessageType: Message_KillProcessRequest.Enum(),
		KillProcessRequest: &KillProcessRequest{}}

	_, err := getResponse(connection, message)
	return err
}

func createGetStepValueRequest() *Message {
	return &Message{MessageType: Message_StepNamesRequest.Enum(), StepNamesRequest: &StepNamesRequest{}}
}

func startAPIService() {
	gaugeListener, err := newGaugeListener(apiPortEnvVariableName, API_STATIC_PORT)
	if err != nil {
		fmt.Printf("[Error] Failed to start API. %s\n", err.Error())
	}
	gaugeListener.acceptAndHandleMultipleConnections(&GaugeApiMessageHandler{})
}

type GaugeApiMessageHandler struct{}

func (handler *GaugeApiMessageHandler) messageReceived(bytesRead []byte, conn net.Conn) {
	apiMessage := &APIMessage{}
	err := proto.Unmarshal(bytesRead, apiMessage)
	if err != nil {
		log.Printf("[Warning] Failed to read proto message: %s\n", err.Error())
	} else {
		messageType := apiMessage.GetMessageType()
		switch messageType {
		case APIMessage_GetProjectRootRequest:
			handler.respondToProjectRootRequest(apiMessage, conn)
			break
		case APIMessage_GetAllStepsRequest:
			handler.respondToGetAllStepsRequest(apiMessage, conn)
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
	responseApiMessage := &APIMessage{MessageType: APIMessage_GetProjectRootResponse.Enum(), MessageId: message.MessageId, AllStepsResponse: getAllStepsResponse}
	handler.sendMessage(responseApiMessage, conn)
}

func (handler *GaugeApiMessageHandler) sendMessage(message *APIMessage, conn net.Conn) {
	if err := writeMessage(conn, message); err != nil {
		fmt.Printf("[Warning] Failed to respond to API request. %s\n", err.Error())
	}
}
