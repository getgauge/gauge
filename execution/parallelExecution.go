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

	"os"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
)

var Strategy string

const Eager string = "eager"
const Lazy string = "lazy"
const enableMultithreadingEnv = "enable_multithreading"

type parallelExecution struct {
	wg                       sync.WaitGroup
	manifest                 *manifest.Manifest
	specCollection           *gauge.SpecCollection
	pluginHandler            plugin.Handler
	currentExecutionInfo     *gauge_messages.ExecutionInfo
	runner                   runner.Runner
	suiteResult              *result.SuiteResult
	numberOfExecutionStreams int
	errMaps                  *gauge.BuildErrors
	startTime                time.Time
}

func newParallelExecution(e *executionInfo) *parallelExecution {
	return &parallelExecution{
		manifest:                 e.manifest,
		specCollection:           e.specs,
		runner:                   e.runner,
		pluginHandler:            e.pluginHandler,
		numberOfExecutionStreams: e.numberOfStreams,
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
	event.Notify(event.NewExecutionEvent(event.SuiteStart, nil, nil, 0, gauge_messages.ExecutionInfo{}))
	e.pluginHandler = plugin.StartPlugins(e.manifest)
}

func (e *parallelExecution) run() *result.SuiteResult {
	e.start()

	nStreams := e.numberOfStreams()
	logger.Info("Executing in %s parallel streams.", strconv.Itoa(nStreams))

	resChan := make(chan *result.SuiteResult)
	if e.isMultithreaded() {
		logger.Debug("Using multithreading for parallel execution.")
		go e.executeMultithreaded(nStreams, resChan)
	} else if isLazy() {
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
	return e.suiteResult
}

func (e *parallelExecution) executeLazily(totalStreams int, resChan chan *result.SuiteResult) {
	e.wg.Add(totalStreams)
	for i := 0; i < totalStreams; i++ {
		go e.startStream(e.specCollection, resChan, i+1)
	}
	e.wg.Wait()
	close(resChan)
}

func (e *parallelExecution) executeMultithreaded(totalStreams int, resChan chan *result.SuiteResult) {
	e.wg.Add(totalStreams)
	handlers := make([]*conn.GaugeConnectionHandler, 0)
	var ports []string
	for i := 0; i < totalStreams; i++ {
		port, err := conn.GetPortFromEnvironmentVariable(common.GaugePortEnvName)
		if err != nil {
			port = 0
		}
		handler, err := conn.NewGaugeConnectionHandler(port, nil)
		if err != nil {
			fmt.Println(err)
		}
		ports = append(ports, strconv.Itoa(handler.ConnectionPortNumber()))
		handlers = append(handlers, handler)
	}
	os.Setenv("GAUGE_API_PORTS", strings.Join(ports, ","))
	r, err := runner.StartRunner(e.manifest, "0", reporter.ParallelReporter(0), make(chan bool), false)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < totalStreams; i++ {
		connection, err := handlers[i].AcceptConnection(config.RunnerConnectionTimeout(), make(chan error))
		if err != nil {
			fmt.Println(err)
		}
		crapRunner := &runner.MultithreadedRunner{}
		crapRunner.SetConnection(connection)
		go e.startMultithreaded(crapRunner, resChan, i+1)
	}
	e.wg.Wait()
	r.Cmd.Process.Kill()
	close(resChan)
}

func (e *parallelExecution) startMultithreaded(r runner.Runner, resChan chan *result.SuiteResult, stream int) {
	defer e.wg.Done()
	e.startSpecsExecutionWithRunner(e.specCollection, resChan, r, stream)
}

func (e *parallelExecution) executeEagerly(distributions int, resChan chan *result.SuiteResult) {
	specs := filter.DistributeSpecs(e.specCollection.Specs(), distributions)
	e.wg.Add(distributions)
	for i, s := range specs {
		go e.startSpecsExecution(s, resChan, i+1)
	}
	e.wg.Wait()
	close(resChan)
}

func (e *parallelExecution) startStream(s *gauge.SpecCollection, resChan chan *result.SuiteResult, stream int) {
	defer e.wg.Done()
	runner, err := runner.Start(e.manifest, reporter.ParallelReporter(stream), make(chan bool), false)
	if err != nil {
		logger.Errorf("Failed to start runner. %s", err.Error())
		resChan <- &result.SuiteResult{UnhandledErrors: []error{fmt.Errorf("Failed to start runner. %s", err.Error())}}
		return
	}
	e.startSpecsExecutionWithRunner(s, resChan, runner, stream)
}

func (e *parallelExecution) startSpecsExecution(s *gauge.SpecCollection, resChan chan *result.SuiteResult, stream int) {
	defer e.wg.Done()
	runner, err := runner.Start(e.manifest, reporter.ParallelReporter(stream), make(chan bool), false)
	if err != nil {
		logger.Errorf("Failed to start runner. %s", err.Error())
		logger.Debug("Skipping %d specifications", s.Size())
		resChan <- &result.SuiteResult{UnhandledErrors: []error{streamExecError{specsSkipped: s.SpecNames(), message: fmt.Sprintf("Failed to start runner. %s", err.Error())}}}
		return
	}
	e.startSpecsExecutionWithRunner(s, resChan, runner, stream)
}

func (e *parallelExecution) startSpecsExecutionWithRunner(s *gauge.SpecCollection, resChan chan *result.SuiteResult, runner runner.Runner, stream int) {
	executionInfo := newExecutionInfo(s, runner, e.pluginHandler, e.errMaps, false, stream)
	se := newSimpleExecution(executionInfo, false)
	se.execute()
	runner.Kill()
	resChan <- se.suiteResult
}

func (e *parallelExecution) finish() {
	e.suiteResult = mergeDataTableSpecResults(e.suiteResult)
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, e.suiteResult, 0, gauge_messages.ExecutionInfo{}))
	message := &gauge_messages.Message{
		MessageType: gauge_messages.Message_SuiteExecutionResult,
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{
			SuiteResult: gauge.ConvertToProtoSuiteResult(e.suiteResult),
		},
	}
	e.pluginHandler.NotifyPlugins(message)
	e.pluginHandler.GracefullyKillPlugins()
}

func (e *parallelExecution) aggregateResults(suiteResults []*result.SuiteResult) {
	r := result.NewSuiteResult(ExecuteTags, e.startTime)
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
	e.suiteResult.SetSpecsSkippedCount()
}

func isLazy() bool {
	return strings.ToLower(Strategy) == Lazy
}

func isValidStrategy(strategy string) bool {
	strategy = strings.ToLower(strategy)
	return strategy == Lazy || strategy == Eager
}

func (e *parallelExecution) isMultithreaded() bool {
	value := util.ConvertToBool(os.Getenv(enableMultithreadingEnv), enableMultithreadingEnv, false)
	if !value {
		return false
	}
	if !e.runner.IsMultithreaded() {
		logger.Warning("Runner doesn't support mutithreading, using multiprocess parallel execution.")
		return false
	}
	return true
}
