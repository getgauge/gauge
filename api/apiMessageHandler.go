/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package api

import (
	"net"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/conceptExtractor"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/refactor"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"google.golang.org/protobuf/proto"
)

type gaugeAPIMessageHandler struct {
	specInfoGatherer *infoGatherer.SpecInfoGatherer
	Runner           runner.Runner
}

func (handler *gaugeAPIMessageHandler) MessageBytesReceived(bytesRead []byte, connection net.Conn) {
	apiMessage := &gauge_messages.APIMessage{}
	var responseMessage *gauge_messages.APIMessage
	err := proto.Unmarshal(bytesRead, apiMessage)
	if err != nil {
		logger.Errorf(false, "Failed to read API proto message: %s\n", err.Error())
		responseMessage = handler.getErrorMessage(err)
	} else {
		logger.Debugf(false, "Api Request Received: %s", apiMessage)
		messageType := apiMessage.GetMessageType()
		switch messageType {
		case gauge_messages.APIMessage_GetProjectRootRequest:
			responseMessage = handler.projectRootRequestResponse(apiMessage)
		case gauge_messages.APIMessage_GetInstallationRootRequest:
			responseMessage = handler.installationRootRequestResponse(apiMessage)
		case gauge_messages.APIMessage_GetAllStepsRequest:
			responseMessage = handler.getAllStepsRequestResponse(apiMessage)
		case gauge_messages.APIMessage_SpecsRequest:
			responseMessage = handler.getSpecsRequestResponse(apiMessage)
		case gauge_messages.APIMessage_GetStepValueRequest:
			responseMessage = handler.getStepValueRequestResponse(apiMessage)
		case gauge_messages.APIMessage_GetLanguagePluginLibPathRequest:
			responseMessage = handler.getLanguagePluginLibPath(apiMessage)
		case gauge_messages.APIMessage_GetAllConceptsRequest:
			responseMessage = handler.getAllConceptsRequestResponse(apiMessage)
		case gauge_messages.APIMessage_PerformRefactoringRequest:
			responseMessage = handler.performRefactoring(apiMessage)
			handler.performRefresh(responseMessage.PerformRefactoringResponse.FilesChanged)
		case gauge_messages.APIMessage_ExtractConceptRequest:
			responseMessage = handler.extractConcept(apiMessage)
		case gauge_messages.APIMessage_FormatSpecsRequest:
			responseMessage = handler.formatSpecs(apiMessage)
		default:
			responseMessage = handler.createUnsupportedAPIMessageResponse(apiMessage)
		}
	}
	handler.sendMessage(responseMessage, connection)
}

func (handler *gaugeAPIMessageHandler) sendMessage(message *gauge_messages.APIMessage, connection net.Conn) {
	logger.Debugf(false, "Sending API response: %s", message)
	dataBytes, err := proto.Marshal(message)
	if err != nil {
		logger.Errorf(false, "Failed to respond to API request. Could not Marshal response %s\n", err.Error())
	}
	if err := conn.Write(connection, dataBytes); err != nil {
		logger.Errorf(false, "Failed to respond to API request. Could not write response %s\n", err.Error())
	}
}

func (handler *gaugeAPIMessageHandler) projectRootRequestResponse(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	projectRootResponse := &gauge_messages.GetProjectRootResponse{ProjectRoot: config.ProjectRoot}
	return &gauge_messages.APIMessage{MessageType: gauge_messages.APIMessage_GetProjectRootResponse, MessageId: message.MessageId, ProjectRootResponse: projectRootResponse}
}

func (handler *gaugeAPIMessageHandler) installationRootRequestResponse(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	root, err := common.GetInstallationPrefix()
	if err != nil {
		logger.Errorf(false, "Failed to find installation root while responding to API request. %s\n", err.Error())
		root = ""
	}
	installationRootResponse := &gauge_messages.GetInstallationRootResponse{InstallationRoot: root}
	return &gauge_messages.APIMessage{MessageType: gauge_messages.APIMessage_GetInstallationRootResponse, MessageId: message.MessageId, InstallationRootResponse: installationRootResponse}
}

func (handler *gaugeAPIMessageHandler) getAllStepsRequestResponse(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	steps := handler.specInfoGatherer.Steps(true)
	var stepValueResponses []*gauge_messages.ProtoStepValue
	for _, step := range steps {
		stepValue := parser.CreateStepValue(step)
		stepValueResponses = append(stepValueResponses, gauge.ConvertToProtoStepValue(&stepValue))
	}
	getAllStepsResponse := &gauge_messages.GetAllStepsResponse{AllSteps: stepValueResponses}
	return &gauge_messages.APIMessage{MessageType: gauge_messages.APIMessage_GetAllStepResponse, MessageId: message.MessageId, AllStepsResponse: getAllStepsResponse}
}

func (handler *gaugeAPIMessageHandler) getSpecsRequestResponse(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	getAllSpecsResponse := handler.createSpecsResponseMessageFor(handler.specInfoGatherer.GetAvailableSpecDetails(message.SpecsRequest.Specs))
	return &gauge_messages.APIMessage{MessageType: gauge_messages.APIMessage_SpecsResponse, MessageId: message.MessageId, SpecsResponse: getAllSpecsResponse}
}

func (handler *gaugeAPIMessageHandler) getStepValueRequestResponse(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	request := message.GetStepValueRequest()
	stepText := request.GetStepText()
	hasInlineTable := request.GetHasInlineTable()
	stepValue, err := parser.ExtractStepValueAndParams(stepText, hasInlineTable)

	if err != nil {
		return handler.getErrorResponse(message, err)
	}
	stepValueResponse := &gauge_messages.GetStepValueResponse{StepValue: gauge.ConvertToProtoStepValue(stepValue)}
	return &gauge_messages.APIMessage{MessageType: gauge_messages.APIMessage_GetStepValueResponse, MessageId: message.MessageId, StepValueResponse: stepValueResponse}

}

func (handler *gaugeAPIMessageHandler) getAllConceptsRequestResponse(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	allConceptsResponse := handler.createGetAllConceptsResponseMessageFor(handler.specInfoGatherer.Concepts())
	return &gauge_messages.APIMessage{MessageType: gauge_messages.APIMessage_GetAllConceptsResponse, MessageId: message.MessageId, AllConceptsResponse: allConceptsResponse}
}

func (handler *gaugeAPIMessageHandler) getLanguagePluginLibPath(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	libPathRequest := message.GetLibPathRequest()
	language := libPathRequest.GetLanguage()
	languageInstallDir, err := plugin.GetInstallDir(language, "")
	if err != nil {
		return handler.getErrorMessage(err)
	}
	runnerInfo, err := runner.GetRunnerInfo(language)
	if err != nil {
		return handler.getErrorMessage(err)
	}
	relativeLibPath := runnerInfo.Lib
	libPath := filepath.Join(languageInstallDir, relativeLibPath)
	response := &gauge_messages.GetLanguagePluginLibPathResponse{Path: libPath}
	return &gauge_messages.APIMessage{MessageType: gauge_messages.APIMessage_GetLanguagePluginLibPathResponse, MessageId: message.MessageId, LibPathResponse: response}
}

func (handler *gaugeAPIMessageHandler) getErrorResponse(message *gauge_messages.APIMessage, err error) *gauge_messages.APIMessage {
	errorResponse := &gauge_messages.ErrorResponse{Error: err.Error()}
	return &gauge_messages.APIMessage{MessageType: gauge_messages.APIMessage_ErrorResponse, MessageId: message.MessageId, Error: errorResponse}

}

func (handler *gaugeAPIMessageHandler) getErrorMessage(err error) *gauge_messages.APIMessage {
	id := common.GetUniqueID()
	errorResponse := &gauge_messages.ErrorResponse{Error: err.Error()}
	return &gauge_messages.APIMessage{MessageType: gauge_messages.APIMessage_ErrorResponse, MessageId: id, Error: errorResponse}
}

func (handler *gaugeAPIMessageHandler) createSpecsResponseMessageFor(details []*infoGatherer.SpecDetail) *gauge_messages.SpecsResponse {
	specDetails := make([]*gauge_messages.SpecsResponse_SpecDetail, 0)
	for _, d := range details {
		detail := &gauge_messages.SpecsResponse_SpecDetail{}
		if d.HasSpec() {
			detail.Spec = gauge.ConvertToProtoSpec(d.Spec)
		}
		for _, e := range d.Errs {
			detail.ParseErrors = append(detail.ParseErrors, &gauge_messages.Error{Type: gauge_messages.Error_PARSE_ERROR, Filename: e.FileName, Message: e.Message, LineNumber: int32(e.LineNo)})
		}
		specDetails = append(specDetails, detail)
	}
	return &gauge_messages.SpecsResponse{Details: specDetails}
}

func (handler *gaugeAPIMessageHandler) createGetAllConceptsResponseMessageFor(conceptInfos []*gauge_messages.ConceptInfo) *gauge_messages.GetAllConceptsResponse {
	return &gauge_messages.GetAllConceptsResponse{Concepts: conceptInfos}
}

func (handler *gaugeAPIMessageHandler) performRefactoring(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	refactoringRequest := message.PerformRefactoringRequest
	response := &gauge_messages.PerformRefactoringResponse{}
	c := make(chan bool)
	runner, err := ConnectToRunner(c, false)
	defer func() {
		err := runner.Kill()
		if err != nil {
			logger.Errorf(true, "failed to kill runner with pid: %d", runner.Pid())
		}
	}()
	if err != nil {
		response.Success = false
		response.Errors = []string{err.Error()}
		return &gauge_messages.APIMessage{MessageId: message.MessageId, MessageType: gauge_messages.APIMessage_PerformRefactoringResponse, PerformRefactoringResponse: response}
	}
	refactoringResult := refactor.GetRefactoringChanges(refactoringRequest.GetOldStep(), refactoringRequest.GetNewStep(), runner, handler.specInfoGatherer.SpecDirs, true)
	refactoringResult.WriteToDisk()
	if refactoringResult.Success {
		logger.Infof(false, "%s", refactoringResult.String())
	} else {
		logger.Errorf(false, "Refactoring response from gauge. Errors : %s", refactoringResult.Errors)
	}
	response.Success = refactoringResult.Success
	response.Errors = refactoringResult.Errors
	response.FilesChanged = refactoringResult.AllFilesChanged()
	return &gauge_messages.APIMessage{MessageId: message.MessageId, MessageType: gauge_messages.APIMessage_PerformRefactoringResponse, PerformRefactoringResponse: response}
}

func (handler *gaugeAPIMessageHandler) performRefresh(files []string) {
	for _, file := range files {
		if util.IsConcept(file) {
			handler.specInfoGatherer.OnConceptFileModify(file)
		}
	}
	for _, file := range files {
		if util.IsSpec(file) {
			handler.specInfoGatherer.OnSpecFileModify(file)
		}
	}
}

func (handler *gaugeAPIMessageHandler) extractConcept(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	request := message.GetExtractConceptRequest()
	success, err, filesChanged := conceptExtractor.ExtractConcept(request.GetConceptName(), request.GetSteps(), request.GetConceptFileName(), request.GetChangeAcrossProject(), request.GetSelectedTextInfo())
	response := &gauge_messages.ExtractConceptResponse{IsSuccess: success, Error: err.Error(), FilesChanged: filesChanged}
	return &gauge_messages.APIMessage{MessageId: message.MessageId, MessageType: gauge_messages.APIMessage_ExtractConceptResponse, ExtractConceptResponse: response}
}

func (handler *gaugeAPIMessageHandler) formatSpecs(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	request := message.GetFormatSpecsRequest()
	results := formatter.FormatSpecFiles(request.GetSpecs()...)
	var warnings []string
	var errors []string
	for _, result := range results {
		if result.ParseErrors != nil {
			for _, err := range result.ParseErrors {
				errors = append(errors, err.Error())
			}
		}
		if result.Warnings != nil {
			var warningTexts []string
			for _, warning := range result.Warnings {
				warningTexts = append(warningTexts, warning.String())
			}
			warnings = append(warnings, warningTexts...)
		}
	}
	formatResponse := &gauge_messages.FormatSpecsResponse{Errors: errors, Warnings: warnings}
	return &gauge_messages.APIMessage{MessageId: message.MessageId, MessageType: gauge_messages.APIMessage_FormatSpecsResponse, FormatSpecsResponse: formatResponse}
}

func (handler *gaugeAPIMessageHandler) createUnsupportedAPIMessageResponse(message *gauge_messages.APIMessage) *gauge_messages.APIMessage {
	return &gauge_messages.APIMessage{MessageId: message.MessageId,
		MessageType:                   gauge_messages.APIMessage_UnsupportedApiMessageResponse,
		UnsupportedApiMessageResponse: &gauge_messages.UnsupportedApiMessageResponse{}}
}
