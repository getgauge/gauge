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

func findStepsInSpecFiles(specFiles []string) {
	parser := new(specParser)
	for _, file := range specFiles {
		scenarioContent, err := common.ReadFileContents(file)
		if err != nil {
			continue
		}
		specification, result := parser.parse(scenarioContent, new(conceptDictionary))

		if result.ok {
			availableStepNames = append(availableStepNames, getStepValues(specification.contexts)...)
			for _, scenario := range specification.scenarios {
				availableStepNames = append(availableStepNames, getStepValues(scenario.steps)...)
			}
		}
	}
}

func getStepValues(steps []*step) []string {
	stepValues := make([]string, 0)
	for _, step := range steps {
		stepValues = append(stepValues, step.value)
	}
	return stepValues
}

func makeListOfAvailableSteps() {
	specFiles := findSpecsFilesIn(common.SpecsDirectoryName)
	go findStepsInSpecFiles(specFiles)
}

func startAPIService() {
	gaugeListener, err := newGaugeListener(apiPortEnvVariableName, API_STATIC_PORT)
	if err != nil {
		fmt.Printf("[Erorr] Failed to start API. %s\n", err.Error())
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
	getAllStepsResponse := &GetAllStepsResponse{Steps: availableStepNames}
	responseApiMessage := &APIMessage{MessageType: APIMessage_GetProjectRootResponse.Enum(), MessageId: message.MessageId, AllStepsResponse: getAllStepsResponse}
	handler.sendMessage(responseApiMessage, conn)
}

func (handler *GaugeApiMessageHandler) sendMessage(message *APIMessage, conn net.Conn) {
	if err := writeMessage(conn, message); err != nil {
		fmt.Printf("[Warning] Failed to respond to API request. %s\n", err.Error())
	}
}
