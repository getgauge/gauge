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

func makeListOfAvailableSteps() {
	addStepValuesToAvailableSteps(getStepsFromRunner())
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
	id := common.GetUniqueId()
	message := &Message{MessageId: &id, MessageType: Message_KillProcessRequest.Enum(),
		KillProcessRequest: &KillProcessRequest{}}

	return writeMessage(connection, message)
}

func createGetStepValueRequest() *Message {
	return &Message{MessageType: Message_StepNamesRequest.Enum(), StepNamesRequest: &StepNamesRequest{}}
}

func startAPIService() {
	gaugeListener, err := newGaugeListener(apiPortEnvVariableName, API_STATIC_PORT)
	if err != nil {
		fmt.Printf("[Error] Failed to start API. %s\n", err.Error())
	}
	gaugeListener.acceptConnections(&GaugeApiMessageHandler{})
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
		case APIMessage_GetAllSpecsRequest:
			handler.respondToGetAllSpecsRequest(apiMessage, conn)
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

func (handler *GaugeApiMessageHandler) createGetAllSpecsResponseMessageFor(specs []*specification) *GetAllSpecsResponse {
	protoSpecs := make([]*ProtoSpec, 0)
	for _, spec := range specs {
		protoSpecs = append(protoSpecs, convertToProtoSpec(spec))
	}
	return &GetAllSpecsResponse{Specs: protoSpecs}
}

func convertToProtoSpec(spec *specification) *ProtoSpec {
	protoItems := make([]*ProtoItem, 0)
	for _, item := range spec.items {
		protoItems = append(protoItems, convertToProtoItem(item))
	}
	return &ProtoSpec{Items: protoItems}
}

func convertToProtoItem(item item) *ProtoItem {
	switch item.kind() {
	case headingKind:
		return convertToProtoHeadingItem(item.(*heading))
	case scenarioKind:
		return convertToProtoScenarioItem(item.(*scenario))
	case stepKind:
		return convertToProtoStepItem(item.(*step))
	case commentKind:
		return convertToProtoCommentItem(item.(*comment))
	case tagKind:
		return convertToProtoTagsItem(item.(*tags))
	case tableKind:
		return convertToProtoTableItem(item.(*table))
	}
	return nil
}

func convertToProtoHeadingItem(heading *heading) *ProtoItem {
	var protoHeadingType ProtoHeading_HeadingType
	headingType := heading.headingType
	if headingType == specHeading {
		protoHeadingType = ProtoHeading_Spec
	} else {
		protoHeadingType = ProtoHeading_Scenario
	}
	return &ProtoItem{ItemType: ProtoItem_Heading.Enum(), Heading: &ProtoHeading{HeadingType: protoHeadingType.Enum(), Text: proto.String(heading.value)}}
}

func convertToProtoStepItem(step *step) *ProtoItem {
	if step.isConcept {
		return convertToProtoConcept(step)
	}
	return &ProtoItem{ItemType: ProtoItem_Step.Enum(), Step: convertToProtoStep(step)}
}

func convertToProtoScenarioItem(scenario *scenario) *ProtoItem {
	scenarioItems := make([]*ProtoItem, 0)
	for _, item := range scenario.items {
		scenarioItems = append(scenarioItems, convertToProtoItem(item))
	}
	protoScenario := &ProtoScenario{ScenarioItems: scenarioItems}
	return &ProtoItem{ItemType: ProtoItem_Scenario.Enum(), Scenario: protoScenario}
}

func convertToProtoConcept(concept *step) *ProtoItem {
	protoConcept := &ProtoConcept{ConceptStep: convertToProtoStep(concept), Steps: convertToProtoSteps(concept.conceptSteps)}
	protoConceptItem := &ProtoItem{ItemType: ProtoItem_Concept.Enum(), Concept: protoConcept}
	return protoConceptItem
}

func convertToProtoStep(step *step) *ProtoStep {
	return &ProtoStep{Text: proto.String(step.lineText), Parameters: convertToProtoParameters(step.args)}
}

func convertToProtoSteps(steps []*step) []*ProtoStep {
	protoSteps := make([]*ProtoStep, 0)
	for _, step := range steps {
		protoSteps = append(protoSteps, convertToProtoStep(step))
	}
	return protoSteps
}

func convertToProtoCommentItem(comment *comment) *ProtoItem {
	return &ProtoItem{ItemType: ProtoItem_Comment.Enum(), Comment: &ProtoComment{Text: proto.String(comment.value)}}
}

func convertToProtoTagsItem(tags *tags) *ProtoItem {
	return &ProtoItem{ItemType: ProtoItem_Tags.Enum(), Tags: &ProtoTags{Tags: tags.values}}
}

func convertToProtoTableItem(table *table) *ProtoItem {
	return &ProtoItem{ItemType: ProtoItem_Table.Enum(), Table: convertToProtoTableParam(table)}
}

func convertToProtoParameters(args []*stepArg) []*Parameter {
	params := make([]*Parameter, 0)
	for _, arg := range args {
		params = append(params, convertToProtoParameter(arg))
	}
	return params
}

func convertToProtoParameter(arg *stepArg) *Parameter {
	switch arg.argType {
	case static:
		return &Parameter{ParameterType: Parameter_Static.Enum(), Value: proto.String(arg.value)}
	case dynamic:
		return &Parameter{ParameterType: Parameter_Dynamic.Enum(), Value: proto.String(arg.value)}
	case tableArg:
		return &Parameter{ParameterType: Parameter_Table.Enum(), Table: convertToProtoTableParam(&arg.table)}
	}
	return nil
}

func convertToProtoTableParam(table *table) *ProtoTableParam {
	protoTableParam := &ProtoTableParam{Rows: make([]*ProtoTableRow, 0)}
	protoTableParam.Headers = &ProtoTableRow{Cells: table.headers}
	for _, row := range table.getRows() {
		protoTableParam.Rows = append(protoTableParam.Rows, &ProtoTableRow{Cells: row})
	}
	return protoTableParam
}

func (handler *GaugeApiMessageHandler) sendMessage(message *APIMessage, conn net.Conn) {
	if err := writeMessage(conn, message); err != nil {
		fmt.Printf("[Warning] Failed to respond to API request. %s\n", err.Error())
	}
}
