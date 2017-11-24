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
)

type langRunner struct {
	runner   runner.Runner
	killChan chan bool
}

var lRunner langRunner

func startRunner() {
	lRunner.killChan = make(chan bool)
	var err error
	lRunner.runner, err = connectToRunner(lRunner.killChan)
	if err != nil {
		logger.APILog.Infof("Unable to connect to runner : %s", err.Error())
	}
}

func connectToRunner(killChan chan bool) (runner.Runner, error) {
	m, _ := manifest.ProjectManifest()
	if m.Language != "js" {
		return nil, nil
	}
	logger.GaugeLog.Infof("Starting language runner")
	outfile, err := os.OpenFile(logger.GetLogFilePath(logger.GaugeLogFileName), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.APILog.Infof("%s", err.Error())
		return nil, err
	}
	runner, err := api.ConnectToRunner(killChan, false, outfile)
	if err != nil {
		logger.APILog.Infof("%s", err.Error())
		return nil, err
	}
	return runner, nil
}

func cacheFileOnRunner(uri, text string) error {
	cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{Content: text, FilePath: util.ConvertURItoFilePath(uri), IsClosed: false}}
	err := sendMessageToRunner(cacheFileRequest)
	return err
}

func sendMessageToRunner(cacheFileRequest *gm.Message) error {
	err := conn.WriteGaugeMessage(cacheFileRequest, lRunner.runner.Connection())
	if err != nil {
		logger.APILog.Infof("Error while connecting to runner : %s", err.Error())
	}
	return err
}

var GetResponseFromRunner = func(message *gm.Message) (*gm.Message, error) {
	return conn.GetResponseForMessageWithTimeout(message, lRunner.runner.Connection(), config.RunnerConnectionTimeout())
}

func getStepPositionResponse(uri string) (*gm.StepPositionsResponse, error) {
	stepPositionsRequest := &gm.Message{MessageType: gm.Message_StepPositionsRequest, StepPositionsRequest: &gm.StepPositionsRequest{FilePath: util.ConvertURItoFilePath(uri)}}
	response, err := GetResponseFromRunner(stepPositionsRequest)
	if err != nil {
		logger.APILog.Infof("Error while connecting to runner : %s", err.Error())
		return nil, err
	}
	stepPositionsResponse := response.GetStepPositionsResponse()
	if stepPositionsResponse.GetError() != "" {
		return nil, fmt.Errorf("error while connecting to runner : %s", stepPositionsResponse.GetError())
	}
	return stepPositionsResponse, nil
}

func killRunner() {
	lRunner.runner.Kill()
}
