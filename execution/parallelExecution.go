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
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
)

var Strategy string

const Eager string = "eager"
const Lazy string = "lazy"

type parallelExecution struct {
	wg                       sync.WaitGroup
	manifest                 *manifest.Manifest
	specStore                *specStore
	pluginHandler            *plugin.Handler
	currentExecutionInfo     *gauge_messages.ExecutionInfo
	runner                   *runner.TestRunner
	suiteResult              *result.SuiteResult
	numberOfExecutionStreams int
	consoleReporter          reporter.Reporter
	errMaps                  *validationErrMaps
	startTime                time.Time
}

func newParallelExecution(executionInfo *executionInfo) *parallelExecution {
	return &parallelExecution{manifest: executionInfo.manifest, specStore: executionInfo.specStore,
		runner: executionInfo.runner, pluginHandler: executionInfo.pluginHandler,
		numberOfExecutionStreams: executionInfo.numberOfStreams,
		consoleReporter:          executionInfo.consoleReporter, errMaps: executionInfo.errMaps}
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

func (e *parallelExecution) result() *result.SuiteResult {
	return e.suiteResult
}

func (e *parallelExecution) numberOfStreams() int {
	nStreams := e.numberOfExecutionStreams
	size := e.specStore.size()
	if nStreams > size {
		nStreams = size
	}
	return nStreams
}

func (e *parallelExecution) start() {
	e.pluginHandler = plugin.StartPlugins(e.manifest)
	e.startTime = time.Now()
}

func (e *parallelExecution) run() {
	e.start()
	var suiteResults []*result.SuiteResult
	nStreams := e.numberOfStreams()
	logger.Info("Executing in %s parallel streams.", strconv.Itoa(nStreams))
	if isLazy() {
		suiteResults = e.lazyExecution(nStreams)
	} else {
		suiteResults = e.eagerExecution(nStreams)
	}
	e.aggregateResults(suiteResults)
	e.finish()
}

func (e *parallelExecution) eagerExecution(distributions int) []*result.SuiteResult {
	specCollections := filter.DistributeSpecs(e.specStore.specs, distributions)
	suiteResultChannel := make(chan *result.SuiteResult, len(specCollections))
	for i, specCollection := range specCollections {
		go e.startSpecsExecution(specCollection, suiteResultChannel, reporter.NewParallelConsole(i+1))
	}
	var suiteResults []*result.SuiteResult
	for _ = range specCollections {
		suiteResults = append(suiteResults, <-suiteResultChannel)
	}
	return suiteResults
}

func (e *parallelExecution) startSpecsExecution(specCollection *filter.SpecCollection, suiteResults chan *result.SuiteResult, reporter reporter.Reporter) {
	testRunner, err := runner.StartRunnerAndMakeConnection(e.manifest, reporter, make(chan bool))
	if err != nil {
		logger.Errorf("Failed: " + err.Error())
		logger.Debug("Skipping %s specifications", strconv.Itoa(len(specCollection.Specs)))
		suiteResults <- &result.SuiteResult{UnhandledErrors: []error{streamExecError{specsSkipped: specCollection.SpecNames(), message: fmt.Sprintf("Failed to start runner. %s", err.Error())}}}
		return
	}
	e.startSpecsExecutionWithRunner(&specStore{specs: specCollection.Specs}, suiteResults, testRunner, reporter)
}

func (e *parallelExecution) lazyExecution(totalStreams int) []*result.SuiteResult {
	suiteResultChannel := make(chan *result.SuiteResult, e.specStore.size())
	e.wg.Add(totalStreams)
	for i := 0; i < totalStreams; i++ {
		go e.startStream(e.specStore, reporter.NewParallelConsole(i+1), suiteResultChannel)
	}
	e.wg.Wait()
	var suiteResults []*result.SuiteResult
	for i := 0; i < totalStreams; i++ {
		suiteResults = append(suiteResults, <-suiteResultChannel)
	}
	close(suiteResultChannel)
	return suiteResults
}

func (e *parallelExecution) startStream(specStore *specStore, reporter reporter.Reporter, suiteResultChannel chan *result.SuiteResult) {
	defer e.wg.Done()
	testRunner, err := runner.StartRunnerAndMakeConnection(e.manifest, reporter, make(chan bool))
	if err != nil {
		logger.Errorf("Failed to start runner. Reason: %s", err.Error())
		suiteResultChannel <- &result.SuiteResult{UnhandledErrors: []error{fmt.Errorf("Failed to start runner. %s", err.Error())}}
		return
	}
	e.startSpecsExecutionWithRunner(specStore, suiteResultChannel, testRunner, reporter)
}

func (e *parallelExecution) startSpecsExecutionWithRunner(specStore *specStore, suiteResultsChan chan *result.SuiteResult, runner *runner.TestRunner, reporter reporter.Reporter) {
	executionInfo := newExecutionInfo(e.manifest, specStore, runner, e.pluginHandler, reporter, e.errMaps, false)
	simpleExecution := newExecution(executionInfo)
	simpleExecution.run()
	runner.Kill()
	suiteResultsChan <- simpleExecution.result()
}

func (e *parallelExecution) finish() {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{SuiteResult: gauge.ConvertToProtoSuiteResult(e.suiteResult)}}
	e.pluginHandler.NotifyPlugins(message)
	e.pluginHandler.GracefullyKillPlugins()
}

func (e *parallelExecution) aggregateResults(suiteResults []*result.SuiteResult) {
	r := result.NewSuiteResult(ExecuteTags, e.startTime)
	r.SpecsSkippedCount = len(e.errMaps.specErrs)
	for _, result := range suiteResults {
		r.SpecsFailedCount += result.SpecsFailedCount
		r.SpecResults = append(r.SpecResults, result.SpecResults...)
		if result.IsFailed {
			r.IsFailed = true
		}
		if result.PreSuite != nil {
			r.PreSuite = result.PreSuite
		}
		if result.PostSuite != nil {
			r.PostSuite = result.PostSuite
		}
		if result.UnhandledErrors != nil {
			r.UnhandledErrors = append(r.UnhandledErrors, result.UnhandledErrors...)
		}
	}
	r.ExecutionTime = int64(time.Since(e.startTime) / 1e6)
	e.suiteResult = r
}

func isLazy() bool {
	return strings.ToLower(Strategy) == Lazy
}

func isValidStrategy(strategy string) bool {
	strategy = strings.ToLower(strategy)
	return strategy == Lazy || strategy == Eager
}
