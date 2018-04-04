// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package lang

import (
	"context"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/getgauge/gauge/api"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

type langRunner struct {
	runner    runner.Runner
	killChan  chan bool
	lspID     string
	lspClient *lspRunner
}

type lspRunner struct {
	client gm.LspServiceClient
	conn   *grpc.ClientConn
}

func (r *lspRunner) ExecuteMessageWithTimeout(message *gm.Message) (*gm.Message, error) {
	switch message.MessageType {
	case gm.Message_CacheFileRequest:
		r.client.CacheFile(context.Background(), message.CacheFileRequest)
		return &gm.Message{}, nil
	case gm.Message_StepNamesRequest:
		response, err := r.client.GetStepNames(context.Background(), message.StepNamesRequest)
		return &gm.Message{StepNamesResponse: response}, err
	case gm.Message_StepPositionsRequest:
		response, err := r.client.GetStepPositions(context.Background(), message.StepPositionsRequest)
		return &gm.Message{StepPositionsResponse: response}, err
	case gm.Message_ImplementationFileListRequest:
		response, err := r.client.GetImplementationFiles(context.Background(), &gm.Empty{})
		return &gm.Message{ImplementationFileListResponse: response}, err
	case gm.Message_StubImplementationCodeRequest:
		response, err := r.client.ImplementStub(context.Background(), message.StubImplementationCodeRequest)
		return &gm.Message{FileDiff: response}, err
	case gm.Message_StepValidateRequest:
		response, err := r.client.ValidateStep(context.Background(), message.StepValidateRequest)
		return &gm.Message{MessageType: gm.Message_StepValidateResponse, StepValidateResponse: response}, err
	case gm.Message_RefactorRequest:
		response, err := r.client.Refactor(context.Background(), message.RefactorRequest)
		return &gm.Message{MessageType: gm.Message_RefactorResponse, RefactorResponse: response}, err
	case gm.Message_StepNameRequest:
		response, err := r.client.GetStepName(context.Background(), message.StepNameRequest)
		return &gm.Message{MessageType: gm.Message_StepNameResponse, StepNameResponse: response}, err
	default:
		return nil, nil
	}
}

func (r *lspRunner) ExecuteAndGetStatus(m *gm.Message) *gm.ProtoExecutionResult {
	return nil
}
func (r *lspRunner) IsProcessRunning() bool {
	return false
}
func (r *lspRunner) Kill() error {
	return nil
}
func (r *lspRunner) Connection() net.Conn {
	return nil
}
func (r *lspRunner) IsMultithreaded() bool {
	return false
}
func (r *lspRunner) Pid() int {
	return 0
}

var lRunner langRunner

func startRunner() error {
	lRunner.killChan = make(chan bool)
	var err error
	err = connectToRunner(lRunner.killChan)
	if err != nil {
		return fmt.Errorf("Unable to connect to runner : %s", err.Error())
	}
	return nil
}

func connectToRunner(killChan chan bool) error {
	logInfo(nil, "Starting language runner")
	outFile, err := util.OpenFile(logger.ActiveLogFile)
	if err != nil {
		return err
	}
	os.Setenv("GAUGE_LSP_GRPC", "true")
	lRunner.runner, err = api.ConnectToRunner(killChan, false, outFile)
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:54545"), grpc.WithInsecure())
	if err != nil {
		logDebug(nil, "%s\nSome of the gauge lsp feature will not work as expected. gRPC client not connected.", err.Error())
	}
	client := gm.NewLspServiceClient(conn)
	lRunner.lspClient = client
	return nil
}

func cacheFileOnRunner(uri lsp.DocumentURI, text string, isClosed bool, status gm.CacheFileRequest_FileStatus) error {
	r := &gm.CacheFileRequest{Content: text, FilePath: string(util.ConvertURItoFilePath(uri)), IsClosed: false, Status: status}
	_, err := lRunner.lspClient.CacheFile(context.Background(), r)
	return err
}

func getStepPositionResponse(uri lsp.DocumentURI) (*gm.StepPositionsResponse, error) {
	stepPositionsRequest := &gm.Message{MessageType: gm.Message_StepPositionsRequest, StepPositionsRequest: &gm.StepPositionsRequest{FilePath: util.ConvertURItoFilePath(uri)}}
	response, err := lRunner.lspClient.ExecuteMessageWithTimeout(stepPositionsRequest)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to runner : %s", err)
	}
	if response.GetStepPositionsResponse().GetError() != "" {
		return nil, fmt.Errorf("error while connecting to runner : %s", response.GetStepPositionsResponse().GetError())
	}
	return response.GetStepPositionsResponse(), nil
}

func getImplementationFileList() (*gm.ImplementationFileListResponse, error) {
	implementationFileListRequest := &gm.Message{MessageType: gm.Message_ImplementationFileListRequest}
	response, err := lRunner.lspClient.ExecuteMessageWithTimeout(implementationFileListRequest)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to runner : %s", err.Error())
	}
	return response.GetImplementationFileListResponse(), nil
}

func putStubImplementation(filePath string, codes []string) (*gm.FileDiff, error) {
	stubImplementationCodeRequest := &gm.Message{
		MessageType: gm.Message_StubImplementationCodeRequest,
		StubImplementationCodeRequest: &gm.StubImplementationCodeRequest{
			ImplementationFilePath: filePath,
			Codes: codes,
		},
	}
	response, err := lRunner.lspClient.ExecuteMessageWithTimeout(stubImplementationCodeRequest)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to runner : %s", err.Error())
	}
	return response.GetFileDiff(), nil
}

func getAllStepsResponse() (*gm.StepNamesResponse, error) {
	getAllStepsRequest := &gm.Message{MessageType: gm.Message_StepNamesRequest, StepNamesRequest: &gm.StepNamesRequest{}}
	response, err := lRunner.lspClient.ExecuteMessageWithTimeout(getAllStepsRequest)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to runner : %s", err.Error())
	}
	return response.GetStepNamesResponse(), nil
}

func killRunner() {
	if lRunner.runner != nil {
		lRunner.runner.Kill()
	}
}

func getLanguageIdentifier() (string, error) {
	m, err := manifest.ProjectManifest()
	if err != nil {
		return "", err
	}
	info, err := runner.GetRunnerInfo(m.Language)
	if err != nil {
		return "", err
	}
	return info.LspLangId, nil
}
