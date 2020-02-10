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
	"os"

	"github.com/getgauge/gauge/config"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type langRunner struct {
	lspID  string
	runner *runner.GrpcRunner
}

var lRunner langRunner

func startRunner() error {
	var err = connectToRunner()
	if err != nil {
		return fmt.Errorf("unable to connect to runner : %s", err.Error())
	}
	return nil
}

func connectToRunner() error {
	logInfo(nil, "Starting language runner")
	outFile, err := util.OpenFile(logger.ActiveLogFile)
	if err != nil {
		return err
	}
	err = os.Setenv("GAUGE_LSP_GRPC", "true")
	if err != nil {
		return err
	}
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		return err
	}

	lRunner.runner, err = runner.StartGrpcRunner(manifest, outFile, outFile, config.IdeRequestTimeout(), false)
	return err
}

func cacheFileOnRunner(uri lsp.DocumentURI, text string, isClosed bool, status gm.CacheFileRequest_FileStatus) error {
	r := &gm.Message{
		MessageType: gm.Message_CacheFileRequest,
		CacheFileRequest: &gm.CacheFileRequest{
			Content:  text,
			FilePath: string(util.ConvertURItoFilePath(uri)),
			IsClosed: false,
			Status:   status,
		},
	}
	_, err := lRunner.runner.ExecuteMessageWithTimeout(r)
	return err
}

func globPatternRequest() (*gm.ImplementationFileGlobPatternResponse, error) {
	implFileGlobPatternRequest := &gm.Message{MessageType: gm.Message_ImplementationFileGlobPatternRequest, ImplementationFileGlobPatternRequest: &gm.ImplementationFileGlobPatternRequest{}}
	response, err := lRunner.runner.ExecuteMessageWithTimeout(implFileGlobPatternRequest)
	if err != nil {
		return nil, err
	}
	return response.GetImplementationFileGlobPatternResponse(), nil
}

func getStepPositionResponse(uri lsp.DocumentURI) (*gm.StepPositionsResponse, error) {
	stepPositionsRequest := &gm.Message{MessageType: gm.Message_StepPositionsRequest, StepPositionsRequest: &gm.StepPositionsRequest{FilePath: util.ConvertURItoFilePath(uri)}}
	response, err := lRunner.runner.ExecuteMessageWithTimeout(stepPositionsRequest)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to runner : %s", err)
	}
	if response.GetStepPositionsResponse().GetError() != "" {
		return nil, fmt.Errorf("error while connecting to runner : %s", response.GetStepPositionsResponse().GetError())
	}
	return response.GetStepPositionsResponse(), nil
}

func getImplementationFileList() (*gm.ImplementationFileListResponse, error) {
	implementationFileListRequest := &gm.Message{MessageType: gm.Message_ImplementationFileListRequest}
	response, err := lRunner.runner.ExecuteMessageWithTimeout(implementationFileListRequest)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to runner : %s", err.Error())
	}
	return response.GetImplementationFileListResponse(), nil
}

func getStepNameResponse(stepValue string) (*gm.StepNameResponse, error) {
	stepNameRequest := &gm.Message{MessageType: gm.Message_StepNameRequest, StepNameRequest: &gm.StepNameRequest{StepValue: stepValue}}
	response, err := lRunner.runner.ExecuteMessageWithTimeout(stepNameRequest)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to runner : %s", err)
	}
	return response.GetStepNameResponse(), nil
}

func putStubImplementation(filePath string, codes []string) (*gm.FileDiff, error) {
	stubImplementationCodeRequest := &gm.Message{
		MessageType: gm.Message_StubImplementationCodeRequest,
		StubImplementationCodeRequest: &gm.StubImplementationCodeRequest{
			ImplementationFilePath: filePath,
			Codes:                  codes,
		},
	}
	response, err := lRunner.runner.ExecuteMessageWithTimeout(stubImplementationCodeRequest)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to runner : %s", err.Error())
	}
	return response.GetFileDiff(), nil
}

func getAllStepsResponse() (*gm.StepNamesResponse, error) {
	getAllStepsRequest := &gm.Message{MessageType: gm.Message_StepNamesRequest, StepNamesRequest: &gm.StepNamesRequest{}}
	response, err := lRunner.runner.ExecuteMessageWithTimeout(getAllStepsRequest)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to runner : %s", err.Error())
	}
	return response.GetStepNamesResponse(), nil
}

func killRunner() error {
	if lRunner.runner != nil {
		return lRunner.runner.Kill()
	}
	return nil
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

func informRunnerCompatibility(ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	if lRunner.lspID != "" {
		return nil
	}
	var params = lsp.ShowMessageParams{
		Type:    lsp.Warning,
		Message: "Current gauge language runner is not compatible with gauge LSP. Some of the editing feature will not work as expected",
	}
	return conn.Notify(ctx, "window/showMessage", params)
}
