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

package execution

import (
	"errors"
	"fmt"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Strategy string

const EAGER string = "eager"
const LAZY string = "lazy"

type parallelSpecExecution struct {
	wg                       sync.WaitGroup
	manifest                 *manifest.Manifest
	specifications           []*parser.Specification
	pluginHandler            *plugin.PluginHandler
	currentExecutionInfo     *gauge_messages.ExecutionInfo
	runner                   *runner.TestRunner
	aggregateResult          *result.SuiteResult
	numberOfExecutionStreams int
	logger                   *logger.GaugeLogger
	errMaps                  *validationErrMaps
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

type parallelInfo struct {
	inParallel      bool
	numberOfStreams int
}

func (self *parallelInfo) isValid() bool {
	if self.numberOfStreams < 1 {
		logger.Log.Error("Invalid input(%s) to --n flag.", strconv.Itoa(self.numberOfStreams))
		return false
	}
	currentStrategy := strings.ToLower(Strategy)
	if currentStrategy != LAZY && currentStrategy != EAGER {
		logger.Log.Error("Invalid input(%s) to --strategy flag.", Strategy)
		return false
	}
	return true
}

func isLazy() bool {
	return strings.ToLower(Strategy) == LAZY
}

func (e *parallelSpecExecution) getNumberOfStreams() int {
	nStreams := e.numberOfExecutionStreams
	if nStreams > len(e.specifications) {
		nStreams = len(e.specifications)
	}
	return nStreams
}

func (e *parallelSpecExecution) start() *result.SuiteResult {
	suiteResults := make([]*result.SuiteResult, 0)
	nStreams := e.getNumberOfStreams()
	e.logger.Info("Executing in %s parallel streams.", strconv.Itoa(nStreams))

	startTime := time.Now()
	if isLazy() {
		suiteResults = e.lazyExecution(nStreams)
	} else {
		suiteResults = e.eagerExecution(nStreams)
	}

	e.aggregateResult = e.aggregateResults(suiteResults)
	e.aggregateResult.Timestamp = startTime.Format(config.LayoutForTimeStamp)
	e.aggregateResult.ProjectName = filepath.Base(config.ProjectRoot)
	e.aggregateResult.Environment = env.CurrentEnv
	e.aggregateResult.Tags = ExecuteTags
	e.aggregateResult.ExecutionTime = int64(time.Since(startTime) / 1e6)
	return e.aggregateResult
}

func (e *parallelSpecExecution) eagerExecution(distributions int) []*result.SuiteResult {
	specCollections := filter.DistributeSpecs(e.specifications, distributions)
	suiteResultChannel := make(chan *result.SuiteResult, len(specCollections))
	for i, specCollection := range specCollections {
		go e.startSpecsExecution(specCollection, suiteResultChannel, nil, logger.NewParallelLogger(i+1))
	}
	suiteResults := make([]*result.SuiteResult, 0)
	for _, _ = range specCollections {
		suiteResults = append(suiteResults, <-suiteResultChannel)
	}
	return suiteResults
}

func (e *parallelSpecExecution) startSpecsExecution(specCollection *filter.SpecCollection, suiteResults chan *result.SuiteResult, testRunner *runner.TestRunner, log *logger.GaugeLogger) {
	var err error
	testRunner, err = runner.StartRunnerAndMakeConnection(e.manifest, log, make(chan bool))
	if err != nil {
		e.logger.Error("Failed: " + err.Error())
		e.logger.Debug("Skipping %s specifications", strconv.Itoa(len(specCollection.Specs)))
		suiteResults <- &result.SuiteResult{UnhandledErrors: []error{streamExecError{specsSkipped: specCollection.SpecNames(), message: fmt.Sprintf("Failed to start runner. %s", err.Error())}}}
		return
	}
	e.startSpecsExecutionWithRunner(specCollection, suiteResults, testRunner, log)
}

func (e *parallelSpecExecution) lazyExecution(totalStreams int) []*result.SuiteResult {
	allSpecs := &specList{}
	for _, spec := range e.specifications {
		allSpecs.specs = append(allSpecs.specs, spec)
	}
	suiteResultChannel := make(chan *result.SuiteResult, len(e.specifications))
	e.wg.Add(totalStreams)
	for i := 0; i < totalStreams; i++ {
		go e.startStream(allSpecs, logger.NewParallelLogger(i+1), suiteResultChannel)
	}
	e.wg.Wait()
	suiteResults := make([]*result.SuiteResult, 0)
	for i := 0; i < totalStreams; i++ {
		suiteResults = append(suiteResults, <-suiteResultChannel)
	}
	close(suiteResultChannel)
	return suiteResults
}

func (e *parallelSpecExecution) startStream(specs *specList, log *logger.GaugeLogger, suiteResultChannel chan *result.SuiteResult) {
	defer e.wg.Done()
	testRunner, err := runner.StartRunnerAndMakeConnection(e.manifest, log, make(chan bool))
	if err != nil {
		log.Error("Failed to start runner. Reason: %s", err.Error())
		suiteResultChannel <- &result.SuiteResult{UnhandledErrors: []error{errors.New(fmt.Sprintf("Failed to start runner. %s", err.Error()))}}
		return
	}
	simpleExecution := newSimpleExecution(&executionInfo{e.manifest, make([]*parser.Specification, 0), testRunner, e.pluginHandler, nil, log, e.errMaps})
	e.aggregateResult = simpleExecution.executeStream(specs)
	suiteResultChannel <- e.aggregateResult
}

func (e *parallelSpecExecution) startSpecsExecutionWithRunner(specCollection *filter.SpecCollection, suiteResults chan *result.SuiteResult, runner *runner.TestRunner, logger *logger.GaugeLogger) {
	execution := newExecution(&executionInfo{e.manifest, specCollection.Specs, runner, e.pluginHandler, &parallelInfo{inParallel: false}, logger, e.errMaps})
	result := execution.start()
	runner.Kill()
	suiteResults <- result
}

func (e *parallelSpecExecution) finish() {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{SuiteResult: parser.ConvertToProtoSuiteResult(e.aggregateResult)}}
	e.pluginHandler.NotifyPlugins(message)
	e.pluginHandler.GracefullyKillPlugins()
}

func (e *parallelSpecExecution) aggregateResults(suiteResults []*result.SuiteResult) *result.SuiteResult {
	aggregateResult := &result.SuiteResult{IsFailed: false, SpecResults: make([]*result.SpecResult, 0)}
	for _, result := range suiteResults {
		aggregateResult.ExecutionTime += result.ExecutionTime
		aggregateResult.SpecsFailedCount += result.SpecsFailedCount
		aggregateResult.SpecResults = append(aggregateResult.SpecResults, result.SpecResults...)
		if result.IsFailed {
			aggregateResult.IsFailed = true
		}
		if result.PreSuite != nil {
			aggregateResult.PreSuite = result.PreSuite
		}
		if result.PostSuite != nil {
			aggregateResult.PostSuite = result.PostSuite
		}
		if result.UnhandledErrors != nil {
			aggregateResult.UnhandledErrors = append(aggregateResult.UnhandledErrors, result.UnhandledErrors...)
		}
	}
	return aggregateResult
}

type specList struct {
	mutex sync.Mutex
	specs []*parser.Specification
}

func (s *specList) isEmpty() bool {
	s.mutex.Lock()
	if len(s.specs) == 0 {
		s.mutex.Unlock()
		return true
	}
	s.mutex.Unlock()
	return false
}

func (s *specList) getSpec() *parser.Specification {
	s.mutex.Lock()
	var spec *parser.Specification
	spec = s.specs[:1][0]
	s.specs = s.specs[1:]
	s.mutex.Unlock()
	return spec
}
