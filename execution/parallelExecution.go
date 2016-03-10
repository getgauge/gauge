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

	nStreams := e.numberOfStreams()
	logger.Info("Executing in %s parallel streams.", strconv.Itoa(nStreams))

	resChan := make(chan *result.SuiteResult)
	if isLazy() {
		go e.executeLazily(nStreams, resChan)
	} else {
		go e.executeEagerly(nStreams, resChan)
	}

	var res []*result.SuiteResult
	for r := range resChan {
		res = append(res, r)
	}
	e.aggregateResults(res)

	e.finish()
}
func (e *parallelExecution) executeLazily(totalStreams int, resChan chan *result.SuiteResult) {
	e.wg.Add(totalStreams)
	for i := 0; i < totalStreams; i++ {
		go e.startStream(e.specCollection, reporter.NewParallelConsole(i+1), resChan)
	}
	e.wg.Wait()
	close(resChan)
}

func (e *parallelExecution) executeEagerly(distributions int, resChan chan *result.SuiteResult) {
	specs := filter.DistributeSpecs(e.specCollection.Specs(), distributions)
	e.wg.Add(distributions)
	for i, s := range specs {
		go e.startSpecsExecution(s, reporter.NewParallelConsole(i+1), resChan)
	}
	e.wg.Wait()
	close(resChan)
}

func (e *parallelExecution) startStream(s *gauge.SpecCollection, reporter reporter.Reporter, resChan chan *result.SuiteResult) {
	defer e.wg.Done()
	runner, err := runner.StartRunnerAndMakeConnection(e.manifest, reporter, make(chan bool))
	if err != nil {
		logger.Errorf("Failed to start runner. %s", err.Error())
		resChan <- &result.SuiteResult{UnhandledErrors: []error{fmt.Errorf("Failed to start runner. %s", err.Error())}}
		return
	}
	e.startSpecsExecutionWithRunner(s, resChan, runner, reporter)
}

func (e *parallelExecution) startSpecsExecution(s *gauge.SpecCollection, reporter reporter.Reporter, resChan chan *result.SuiteResult) {
	defer e.wg.Done()
	runner, err := runner.StartRunnerAndMakeConnection(e.manifest, reporter, make(chan bool))
	if err != nil {
		logger.Errorf("Failed to start runner. %s", err.Error())
		logger.Debug("Skipping %d specifications", s.Size())
		resChan <- &result.SuiteResult{UnhandledErrors: []error{streamExecError{specsSkipped: s.SpecNames(), message: fmt.Sprintf("Failed to start runner. %s", err.Error())}}}
		return
	}
	e.startSpecsExecutionWithRunner(s, resChan, runner, reporter)
}

func (e *parallelExecution) startSpecsExecutionWithRunner(s *gauge.SpecCollection, resChan chan *result.SuiteResult, runner *runner.TestRunner, reporter reporter.Reporter) {
	executionInfo := newExecutionInfo(s, runner, e.pluginHandler, reporter, e.errMaps, false)
	se := newSimpleExecution(executionInfo)
	se.execute()
	runner.Kill()
	resChan <- se.result()
}

func (e *parallelExecution) finish() {
	message := &gauge_messages.Message{
		MessageType: gauge_messages.Message_SuiteExecutionResult.Enum(),
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{
			SuiteResult: gauge.ConvertToProtoSuiteResult(e.suiteResult),
		},
	}
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
