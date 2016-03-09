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
	"github.com/getgauge/gauge/validation"
)

var Strategy string

const Eager string = "eager"
const Lazy string = "lazy"

type parallelExecution struct {
	wg                       sync.WaitGroup
	manifest                 *manifest.Manifest
	specCollection           *gauge.SpecCollection
	pluginHandler            *plugin.Handler
	currentExecutionInfo     *gauge_messages.ExecutionInfo
	runner                   *runner.TestRunner
	suiteResult              *result.SuiteResult
	numberOfExecutionStreams int
	consoleReporter          reporter.Reporter
	errMaps                  *validation.ValidationErrMaps
	startTime                time.Time
}

func newParallelExecution(e *executionInfo) *parallelExecution {
	return &parallelExecution{
		manifest:                 e.manifest,
		specCollection:           e.specs,
		runner:                   e.runner,
		pluginHandler:            e.pluginHandler,
		numberOfExecutionStreams: e.numberOfStreams,
		consoleReporter:          e.consoleReporter,
		errMaps:                  e.errMaps,
	}
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

func (e *parallelExecution) result() *result.SuiteResult {
	return e.suiteResult
}

func (e *parallelExecution) numberOfStreams() int {
	nStreams := e.numberOfExecutionStreams
	size := e.specCollection.Size()
	if nStreams > size {
		nStreams = size
	}
	return nStreams
}

func (e *parallelExecution) start() {
	e.startTime = time.Now()
	e.pluginHandler = plugin.StartPlugins(e.manifest)
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
	specs := filter.DistributeSpecs(e.specCollection.Specs(), distributions)
	schan := make(chan *result.SuiteResult, len(specs))
	for i, s := range specs {
		go e.startSpecsExecution(s, schan, reporter.NewParallelConsole(i+1))
	}
	var suiteResults []*result.SuiteResult
	for _ = range specs {
		suiteResults = append(suiteResults, <-schan)
	}
	return suiteResults
}

func (e *parallelExecution) startSpecsExecution(s *gauge.SpecCollection, suiteResults chan *result.SuiteResult, reporter reporter.Reporter) {
	testRunner, err := runner.StartRunnerAndMakeConnection(e.manifest, reporter, make(chan bool))
	if err != nil {
		logger.Errorf("Failed: " + err.Error())
		logger.Debug("Skipping %d specifications", s.Size())
		suiteResults <- &result.SuiteResult{UnhandledErrors: []error{streamExecError{specsSkipped: s.SpecNames(), message: fmt.Sprintf("Failed to start runner. %s", err.Error())}}}
		return
	}
	e.startSpecsExecutionWithRunner(s, suiteResults, testRunner, reporter)
}

func (e *parallelExecution) lazyExecution(totalStreams int) []*result.SuiteResult {
	suiteResultChannel := make(chan *result.SuiteResult, e.specCollection.Size())
	e.wg.Add(totalStreams)
	for i := 0; i < totalStreams; i++ {
		go e.startStream(e.specCollection, reporter.NewParallelConsole(i+1), suiteResultChannel)
	}
	e.wg.Wait()
	var suiteResults []*result.SuiteResult
	for i := 0; i < totalStreams; i++ {
		suiteResults = append(suiteResults, <-suiteResultChannel)
	}
	close(suiteResultChannel)
	return suiteResults
}

func (e *parallelExecution) startStream(s *gauge.SpecCollection, reporter reporter.Reporter, suiteResultChannel chan *result.SuiteResult) {
	defer e.wg.Done()
	testRunner, err := runner.StartRunnerAndMakeConnection(e.manifest, reporter, make(chan bool))
	if err != nil {
		logger.Errorf("Failed to start runner. Reason: %s", err.Error())
		suiteResultChannel <- &result.SuiteResult{UnhandledErrors: []error{fmt.Errorf("Failed to start runner. %s", err.Error())}}
		return
	}
	e.startSpecsExecutionWithRunner(s, suiteResultChannel, testRunner, reporter)
}

func (e *parallelExecution) startSpecsExecutionWithRunner(s *gauge.SpecCollection, suiteResultsChan chan *result.SuiteResult, runner *runner.TestRunner, reporter reporter.Reporter) {
	executionInfo := newExecutionInfo(s, runner, e.pluginHandler, reporter, e.errMaps, false)
	se := newSimpleExecution(executionInfo)
	se.execute()
	runner.Kill()
	suiteResultsChan <- se.result()
}

func (e *parallelExecution) finish() {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{SuiteResult: gauge.ConvertToProtoSuiteResult(e.suiteResult)}}
	e.pluginHandler.NotifyPlugins(message)
	e.pluginHandler.GracefullyKillPlugins()
}

func (e *parallelExecution) aggregateResults(suiteResults []*result.SuiteResult) {
	r := result.NewSuiteResult(ExecuteTags, e.startTime)
	r.SpecsSkippedCount = len(e.errMaps.SpecErrs)
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
