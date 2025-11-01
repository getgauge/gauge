/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
)

type StartChannels struct {
	// this will hold the runner
	RunnerChan chan runner.Runner
	// this will hold the error while creating runner
	ErrorChan chan error
	// this holds a flag based on which the runner is terminated
	KillChan chan bool
}

// StartAPI calls StartAPIService and returns the channels
func StartAPI(debug bool) *StartChannels {
	startChan := &StartChannels{RunnerChan: make(chan runner.Runner), ErrorChan: make(chan error), KillChan: make(chan bool)}
	sig := &infoGatherer.SpecInfoGatherer{}
	go startAPIService(0, startChan, sig, debug)
	return startChan
}

// StartAPIService starts the Gauge API service
func startAPIService(port int, startChannels *StartChannels, sig *infoGatherer.SpecInfoGatherer, debug bool) {
	startAPIServiceWithoutRunner(port, startChannels, sig)

	runner, err := ConnectToRunner(startChannels.KillChan, debug)
	if err != nil {
		startChannels.ErrorChan <- err
		return
	}
	startChannels.RunnerChan <- runner
}

func startAPIServiceWithoutRunner(port int, startChannels *StartChannels, sig *infoGatherer.SpecInfoGatherer) {
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
}

func ConnectToRunner(killChannel chan bool, debug bool) (runner.Runner, error) {
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		return nil, err
	}
	runner, connErr := runner.Start(manifest, 0, killChannel, debug)
	if connErr != nil {
		return nil, connErr
	}

	return runner, nil
}

func runAPIServiceIndefinitely(port int, specDirs []string) {
	startChan := &StartChannels{RunnerChan: make(chan runner.Runner), ErrorChan: make(chan error), KillChan: make(chan bool)}

	sig := &infoGatherer.SpecInfoGatherer{SpecDirs: specDirs}
	sig.Init()
	go startAPIServiceWithoutRunner(port, startChan, sig)
	go checkParentIsAlive(startChan)

	logger.Infof(true, "Gauge daemon initialized and listening on port: %d", port)

	for {
		select {
		case runner := <-startChan.RunnerChan:
			logger.Infof(true, "Got a kill message. Killing runner.")
			err := runner.Kill()
			if err != nil {
				logger.Errorf(true, "Unable to kill runner with PID %d. %s", runner.Pid(), err.Error())
			}
		case err := <-startChan.ErrorChan:
			logger.Fatalf(true, "Killing Gauge daemon. %v", err.Error())
		}
	}
}

func checkParentIsAlive(startChannels *StartChannels) {
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
			logger.Fatalf(true, "Invalid port number: %s", apiPort)
		}
		_ = os.Setenv(common.APIPortEnvVariableName, apiPort)
	} else {
		port, err = conn.GetPortFromEnvironmentVariable(common.APIPortEnvVariableName)
		if err != nil {
			logger.Fatalf(true, "Failed to start API Service. %s \n", err.Error())
		}
	}
	runAPIServiceIndefinitely(port, specDirs)
}

func Start(specsDir []string) *conn.GaugeConnectionHandler {
	sig := &infoGatherer.SpecInfoGatherer{SpecDirs: specsDir}
	sig.Init()
	apiHandler := &gaugeAPIMessageHandler{specInfoGatherer: sig}
	gaugeConnectionHandler, err := conn.NewGaugeConnectionHandler(0, apiHandler)
	if err != nil {
		logger.Fatal(true, err.Error())
	}
	errChan := make(chan error)
	go func() {
		_, err := gaugeConnectionHandler.AcceptConnection(config.RunnerConnectionTimeout(), errChan)
		if err != nil {
			logger.Fatal(true, err.Error())
		}
	}()
	go func() {
		e := <-errChan
		logger.Fatal(true, e.Error())
	}()
	return gaugeConnectionHandler
}
