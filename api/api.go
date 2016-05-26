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

package api

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
)

// StartAPI calls StartAPIService and returns the channels
func StartAPI() *runner.StartChannels {
	startChan := &runner.StartChannels{RunnerChan: make(chan runner.Runner), ErrorChan: make(chan error), KillChan: make(chan bool)}
	sig := &infoGatherer.SpecInfoGatherer{}
	go startAPIService(0, startChan, sig)
	return startChan
}

// StartAPIService starts the Gauge API service
func startAPIService(port int, startChannels *runner.StartChannels, sig *infoGatherer.SpecInfoGatherer) {
	apiHandler := &gaugeAPIMessageHandler{specInfoGatherer: sig}
	gaugeConnectionHandler, err := conn.NewGaugeConnectionHandler(port, apiHandler)
	if err != nil {
		startChannels.ErrorChan <- fmt.Errorf("Connection error. %s", err.Error())
		return
	}
	if port == 0 {
		if err := common.SetEnvVariable(common.APIPortEnvVariableName, strconv.Itoa(gaugeConnectionHandler.ConnectionPortNumber())); err != nil {
			startChannels.ErrorChan <- fmt.Errorf("Failed to set Env variable %s. %s", common.APIPortEnvVariableName, err.Error())
			return
		}
	}
	go gaugeConnectionHandler.HandleMultipleConnections()

	runner, err := connectToRunner(startChannels.KillChan)
	if err != nil {
		startChannels.ErrorChan <- err
		return
	}
	startChannels.RunnerChan <- runner
}

func connectToRunner(killChannel chan bool) (runner.Runner, error) {
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		return nil, err
	}

	runner, connErr := runner.StartRunnerAndMakeConnection(manifest, reporter.Current(), killChannel)
	if connErr != nil {
		return nil, connErr
	}

	return runner, nil
}

func runAPIServiceIndefinitely(port int, specDirs []string) {
	startChan := &runner.StartChannels{RunnerChan: make(chan runner.Runner), ErrorChan: make(chan error), KillChan: make(chan bool)}

	sig := &infoGatherer.SpecInfoGatherer{SpecDirs: specDirs}
	sig.MakeListOfAvailableSteps()
	go startAPIService(port, startChan, sig)
	go checkParentIsAlive(startChan)

	for {
		select {
		case runner := <-startChan.RunnerChan:
			logger.Info("Got a kill message. Killing runner.")
			runner.Kill()
		case err := <-startChan.ErrorChan:
			logger.Fatalf("Killing Gauge daemon. %v", err.Error())
		}
	}
}

func checkParentIsAlive(startChannels *runner.StartChannels) {
	parentProcessID := os.Getppid()
	for {
		if !util.IsProcessRunning(parentProcessID) {
			startChannels.ErrorChan <- fmt.Errorf("Parent process with pid %d has terminated.", parentProcessID)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// RunInBackground runs Gauge in daemonized mode on the given apiPort
func RunInBackground(apiPort string, specDirs []string) {
	var port int
	var err error
	if apiPort != "" {
		port, err = strconv.Atoi(apiPort)
		if err != nil {
			logger.Fatalf(fmt.Sprintf("Invalid port number: %s", apiPort))
		}
		os.Setenv(common.APIPortEnvVariableName, apiPort)
	} else {
		port, err = conn.GetPortFromEnvironmentVariable(common.APIPortEnvVariableName)
		if err != nil {
			logger.Fatalf(fmt.Sprintf("Failed to start API Service. %s \n", err.Error()))
		}
	}
	runAPIServiceIndefinitely(port, specDirs)
}

func Start(specsDir []string) *conn.GaugeConnectionHandler {
	sig := &infoGatherer.SpecInfoGatherer{SpecDirs: specsDir}
	sig.MakeListOfAvailableSteps()
	apiHandler := &gaugeAPIMessageHandler{specInfoGatherer: sig}
	gaugeConnectionHandler, err := conn.NewGaugeConnectionHandler(0, apiHandler)
	if err != nil {
		logger.Fatalf(err.Error())
	}
	errChan := make(chan error)
	go gaugeConnectionHandler.AcceptConnection(config.RunnerConnectionTimeout(), errChan)
	go func() {
		e := <-errChan
		logger.Fatalf(e.Error())
	}()
	return gaugeConnectionHandler
}
