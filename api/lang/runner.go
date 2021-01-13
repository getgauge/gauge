/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"fmt"
	"os"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

type langRunner struct {
	lspID  string
	runner *runner.GrpcRunner
}

var lRunner langRunner

var recommendedExtensions = map[string]string{
	"java":   "vscjava.vscode-java-pack",
	"dotnet": "ms-dotnettools.csharp",
	"python": "ms-python.python",
	"ruby":   "rebornix.ruby",
}

func startRunner() error {
	var err = connectToRunner()
	if err != nil {
		var installMessage = ""
		m, e := manifest.ProjectManifest()
		if e == nil {
			installMessage = fmt.Sprintf(" Install '%s' extension for code insights.", recommendedExtensions[m.Language])
		}
		errStr := "Gauge could not initialize.%s For more information see" +
			"[Problems](command:workbench.actions.view.problems), check logs." +
			"[Troubleshooting](https://docs.gauge.org/troubleshooting.html?language=javascript&ide=vscode#gauge-could-not-initialize-for-more-information-see-problems)"
		return fmt.Errorf(errStr, installMessage)
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
