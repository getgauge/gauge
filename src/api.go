package main

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/getgauge/common"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	apiPortEnvVariableName = "GAUGE_API_PORT"
	API_STATIC_PORT        = 8889
)

func findScenarioFiles(fileChan chan<- string) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	walkFn := func(filePath string, info os.FileInfo, err error) error {
		ext := path.Ext(info.Name())
		if strings.ToLower(ext) == ".scn" {
			fileChan <- filePath
		}
		return nil
	}

	filepath.Walk(pwd, walkFn)
	fileChan <- "done"
}

func parseScenarioFiles(fileChan <-chan string) {
	for {
		scenarioFilePath := <-fileChan
		if scenarioFilePath == "done" {
			break
		}

		parser := new(specParser)
		//todo: parse concepts
		scenarioContent, err := common.ReadFileContents(scenarioFilePath)
		if err != nil {
			fmt.Println(err)
		}
		specification, result := parser.parse(scenarioContent, new(conceptDictionary))

		if result.ok {
			availableSteps = append(availableSteps, specification.contexts...)
			for _, scenario := range specification.scenarios {
				availableSteps = append(availableSteps, scenario.steps...)
			}
		} else {
			fmt.Println(result.error.message)
		}

	}
}

func makeListOfAvailableSteps() {
	fileChan := make(chan string)
	go findScenarioFiles(fileChan)
	go parseScenarioFiles(fileChan)
}

func startAPIService() {
	gaugeListener, err := newGaugeListener(apiPortEnvVariableName, API_STATIC_PORT)
	if err != nil {
		fmt.Printf("[Erorr] Failed to start API. %s\n", err.Error())
	}
	gaugeListener.acceptAndHandleMultipleConnections(&GaugeApiMessageHandler{})
}

type GaugeApiMessageHandler struct {
}

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
	if err := writeMessage(conn, responseApiMessage); err != nil {
		fmt.Printf("[Warning] Failed to respond to API request. %s\n", err.Error())
	}
}
