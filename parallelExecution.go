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

package main

import (
	"fmt"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"runtime"
	"strconv"
	"time"
)

type parallelSpecExecution struct {
	manifest                 *manifest
	specifications           []*specification
	pluginHandler            *pluginHandler
	currentExecutionInfo     *gauge_messages.ExecutionInfo
	runner                   *testRunner
	aggregateResult          *suiteResult
	numberOfExecutionStreams int
	writer                   executionLogger
}

type streamExecError struct {
	specsSkipped []string
	message      string
}

func (s streamExecError) Error() string {
	var specNames string
	for _, spec := range s.specsSkipped {
		specNames += fmt.Sprintf("%s\n", spec)
	}
	return fmt.Sprintf("The following specifications could not be executed:\n%sReason : %s.", specNames, s.message)
}

func (s streamExecError) numberOfSpecsSkipped() int {
	return len(s.specsSkipped)
}

type specCollection struct {
	specs []*specification
}

type parallelInfo struct {
	inParallel      bool
	numberOfStreams int
}

func (self *parallelInfo) isValid() bool {
	if self.numberOfStreams < 1 {
		logger.Log.Error("Invalid input(%s) to --n flag", strconv.Itoa(self.numberOfStreams))
		return false
	}
	return true
}

func (e *parallelSpecExecution) start() *suiteResult {
	startTime := time.Now()

	specCollections := e.distributeSpecs(e.numberOfExecutionStreams)
	suiteResultChannel := make(chan *suiteResult, len(specCollections))
	for i, specCollection := range specCollections {
		go e.startSpecsExecution(specCollection, suiteResultChannel, nil, newParallelExecutionConsoleWriter(i+1))
	}
	e.writer.Info("Executing in %s parallel streams.", strconv.Itoa(len(specCollections)))
	suiteResults := make([]*suiteResult, 0)
	for _, _ = range specCollections {
		suiteResults = append(suiteResults, <-suiteResultChannel)
	}

	e.aggregateResult = e.aggregateResults(suiteResults)
	e.aggregateResult.executionTime = int64(time.Since(startTime) / 1e6)
	return e.aggregateResult
}

func (e *parallelSpecExecution) startSpecsExecution(specCollection *specCollection, suiteResults chan *suiteResult, runner *testRunner, writer executionLogger) {
	var err error
	runner, err = startRunnerAndMakeConnection(e.manifest, writer)
	if err != nil {
		e.writer.Error("Failed: " + err.Error())
		e.writer.Debug("Skipping %s specifications", strconv.Itoa(len(specCollection.specs)))
		suiteResults <- &suiteResult{unhandledErrors: []error{streamExecError{specsSkipped: specCollection.specNames(), message: fmt.Sprintf("Failed to start runner. %s", err.Error())}}}
		return
	}
	e.startSpecsExecutionWithRunner(specCollection, suiteResults, runner, writer)
}

func (e *parallelSpecExecution) startSpecsExecutionWithRunner(specCollection *specCollection, suiteResults chan *suiteResult, runner *testRunner, writer executionLogger) {
	execution := newExecution(e.manifest, specCollection.specs, runner, e.pluginHandler, &parallelInfo{inParallel: false}, writer)
	result := execution.start()
	runner.kill(e.writer)
	suiteResults <- result
}

func (e *parallelSpecExecution) distributeSpecs(distributions int) []*specCollection {
	if distributions > len(e.specifications) {
		distributions = len(e.specifications)
	}
	specCollections := make([]*specCollection, distributions)
	for i := 0; i < len(e.specifications); i++ {
		mod := i % distributions
		if specCollections[mod] == nil {
			specCollections[mod] = &specCollection{specs: make([]*specification, 0)}
		}
		specCollections[mod].specs = append(specCollections[mod].specs, e.specifications[i])
	}
	return specCollections
}

func (e *parallelSpecExecution) finish() {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{SuiteResult: convertToProtoSuiteResult(e.aggregateResult)}}
	e.pluginHandler.notifyPlugins(message)
	e.pluginHandler.gracefullyKillPlugins()
}

func (e *parallelSpecExecution) aggregateResults(suiteResults []*suiteResult) *suiteResult {
	aggregateResult := &suiteResult{isFailed: false, specResults: make([]*specResult, 0)}
	for _, result := range suiteResults {
		aggregateResult.executionTime += result.executionTime
		aggregateResult.specsFailedCount += result.specsFailedCount
		aggregateResult.specResults = append(aggregateResult.specResults, result.specResults...)
		if result.isFailed {
			aggregateResult.isFailed = true
		}
		if result.preSuite != nil {
			aggregateResult.preSuite = result.preSuite
		}
		if result.postSuite != nil {
			aggregateResult.postSuite = result.postSuite
		}
		if result.unhandledErrors != nil {
			aggregateResult.unhandledErrors = append(aggregateResult.unhandledErrors, result.unhandledErrors...)
		}
	}
	return aggregateResult
}

func numberOfCores() int {
	return runtime.NumCPU()
}

func (s *specCollection) specNames() []string {
	specNames := make([]string, 0)
	for _, spec := range s.specs {
		specNames = append(specNames, spec.fileName)
	}
	return specNames
}
