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
	"os"

	"fmt"

	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

type langRunner struct {
	runner   runner.Runner
	killChan chan bool
	lspID    string
}

var lRunner langRunner

func startRunner() error {
	lRunner.killChan = make(chan bool)
	var err error
	lRunner.runner, err = connectToRunner(lRunner.killChan)
	if err != nil {
		return fmt.Errorf("Unable to connect to runner : %s", err.Error())
	}
	return nil
}

func connectToRunner(killChan chan bool) (runner.Runner, error) {
	logger.Infof(false, "Starting language runner")
	outfile, err := os.OpenFile(logger.GetLogFile(logger.GaugeLogFileName), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	runner, err := api.ConnectToRunner(killChan, false, outfile)
	if err != nil {
		return nil, err
	}
	return runner, nil
}

func cacheFileOnRunner(uri lsp.DocumentURI, text string) error {
	cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{Content: text, FilePath: string(util.ConvertURItoFilePath(uri)), IsClosed: false}}
	return sendMessageToRunner(cacheFileRequest)
}

func sendMessageToRunner(cacheFileRequest *gm.Message) error {
	err := conn.WriteGaugeMessage(cacheFileRequest, lRunner.runner.Connection())
	if err != nil {
		return fmt.Errorf("Error while connecting to runner : %v", err)
	}
	return nil
}

var GetResponseFromRunner = func(message *gm.Message) (*gm.Message, error) {
	if lRunner.runner == nil {
		return nil, fmt.Errorf("Error while connecting to runner")
	}
	return conn.GetResponseForMessageWithTimeout(message, lRunner.runner.Connection(), config.RunnerRequestTimeout())
}

func getStepPositionResponse(uri lsp.DocumentURI) (*gm.StepPositionsResponse, error) {
	stepPositionsRequest := &gm.Message{MessageType: gm.Message_StepPositionsRequest, StepPositionsRequest: &gm.StepPositionsRequest{FilePath: string(util.ConvertURItoFilePath(uri))}}
	response, err := GetResponseFromRunner(stepPositionsRequest)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to runner : %s", err)
	}
	stepPositionsResponse := response.GetStepPositionsResponse()
	if stepPositionsResponse.GetError() != "" {
		return nil, fmt.Errorf("error while connecting to runner : %s", stepPositionsResponse.GetError())
	}
	return stepPositionsResponse, nil
}

func getImplementationFileList() (*gm.ImplementationFileListResponse, error) {
	implementationFileListRequest := &gm.Message{MessageType: gm.Message_ImplementationFileListRequest, ImplementationFileListRequest: &gm.ImplementationFileListRequest{}}
	response, err := GetResponseFromRunner(implementationFileListRequest)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to runner : %s", err.Error())
	}
	implementationFileListResponse := response.GetImplementationFileListResponse()
	return implementationFileListResponse, nil
}

func putStubImplementation(filePath string, codes []string) (*gm.FileChanges, error) {
	stubImplementationCodeRequest := &gm.Message{MessageType: gm.Message_StubImplementationCodeRequest, StubImplementationCodeRequest: &gm.StubImplementationCodeRequest{ImplementationFilePath: filePath, Codes: codes}}
	response, err := GetResponseFromRunner(stubImplementationCodeRequest)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to runner : %s", err.Error())
	}
	return response.GetFileChanges(), nil
}

func getAllStepsResponse() (*gm.StepNamesResponse, error) {
	getAllStepsRequest := &gm.Message{MessageType: gm.Message_StepNamesRequest, StepNamesRequest: &gm.StepNamesRequest{}}
	response, err := GetResponseFromRunner(getAllStepsRequest)
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
