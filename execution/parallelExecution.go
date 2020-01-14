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
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
)

// Strategy for execution, can be either 'Eager' or 'Lazy'
var Strategy string

// Eager is a parallelization strategy for execution. In this case tests are distributed before execution, thus making them an equal number based distribution.
const Eager string = "eager"

// Lazy is a parallelization strategy for execution. In this case tests assignment will be dynamic during execution, i.e. assign the next spec in line to the stream that has completed it’s previous execution and is waiting for more work.
const Lazy string = "lazy"

const (
	gaugeAPIPortsEnv            = "GAUGE_API_PORTS"
	gaugeParallelStreamCountEnv = "GAUGE_PARALLEL_STREAMS_COUNT"
)

type parallelExecution struct {
	wg                       sync.WaitGroup
	manifest                 *manifest.Manifest
	specCollection           *gauge.SpecCollection
	pluginHandler            plugin.Handler
	runner                   runner.Runner
	suiteResult              *result.SuiteResult
	numberOfExecutionStreams int
	tagsToFilter             string
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
		tagsToFilter:             e.tagsToFilter,
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
	var res []*result.SuiteResult
	if env.AllowFilteredParallelExecution() && e.tagsToFilter != "" {
		p, s := filter.FilterSpecForParallelRun(e.specCollection.Specs(), e.tagsToFilter)
		if Verbose {
			printAdditionalExecutionInfo(p, s, e.tagsToFilter)
		}
		if len(s) > 0 {
			logger.Infof(true, "Executing %d specs in serial.", len(s))
			e.specCollection = gauge.NewSpecCollection(p, false)
			res = append(res, e.executeSpecsInSerial(gauge.NewSpecCollection(s, true)))
		}
	}

	nStreams := e.numberOfStreams()
	logger.Infof(true, "Executing in %s parallel streams.", strconv.Itoa(nStreams))
	resChan := make(chan *result.SuiteResult)

	if e.isMultithreaded() {
		logger.Debugf(true, "Using multithreading for parallel execution.")
		if e.runner.Info().GRPCSupport {
			go e.executeGrpcMultithreaded(nStreams, resChan)
		} else {
			go e.executeLegacyMultithreaded(nStreams, resChan)
		}
	} else if isLazy() {
		go e.executeLazily(nStreams, resChan)
	} else {
		go e.executeEagerly(nStreams, resChan)
	}

	for r := range resChan {
		res = append(res, r)
	}
	e.aggregateResults(res)

	e.finish()
	return e.suiteResult
}

func printAdditionalExecutionInfo(p []*gauge.Specification, s []*gauge.Specification, tags string) {
	logger.Infof(true, "Applied tags '%s' to filter specs for parallel execution", tags)
	logger.Infof(true, "No of specs to be executed in serial : %d", len(s))
	logger.Infof(true, "No of specs to be executed in parallel : %d", len(p))
}

func (e *parallelExecution) executeLazily(totalStreams int, resChan chan *result.SuiteResult) {
	e.wg.Add(totalStreams)
	for i := 0; i < totalStreams; i++ {
		go e.startStream(e.specCollection, resChan, i+1)
	}
	e.wg.Wait()
	close(resChan)
}

func (e *parallelExecution) executeGrpcMultithreaded(totalStreams int, resChan chan *result.SuiteResult) {
	e.wg.Add(totalStreams)
	os.Setenv(gaugeParallelStreamCountEnv, strconv.Itoa(totalStreams))
	r, err := runner.StartGrpcRunner(e.manifest, os.Stdout, os.Stderr, config.RunnerRequestTimeout(), true)
	r.IsExecuting = true
	if err != nil {
		logger.Fatalf(true, "failed to create handler. %s", err.Error())
	}
	for i := 0; i < totalStreams; i++ {
		go e.startMultithreaded(r, resChan, i+1)
	}
	e.wg.Wait()
	r.IsExecuting = false
	if err = r.Kill(); err != nil {
		logger.Infof(true, "unable to kill runner: %s", err.Error())
	}
	close(resChan)
}

func (e *parallelExecution) executeLegacyMultithreaded(totalStreams int, resChan chan *result.SuiteResult) {
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
			logger.Errorf(true, "failed to create handler. %s", err.Error())
		}
		ports = append(ports, strconv.Itoa(handler.ConnectionPortNumber()))
		handlers = append(handlers, handler)
	}
	os.Setenv(gaugeAPIPortsEnv, strings.Join(ports, ","))
	writer := logger.NewLogWriter(e.manifest.Language, true, 0)
	r, err := runner.StartLegacyRunner(e.manifest, "0", writer, make(chan bool), false)
	if err != nil {
		logger.Fatalf(true, "failed to start runner. %s", err.Error())
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
	err = r.Cmd.Process.Kill()
	if err != nil {
		logger.Infof(true, "unable to kill runner: %s", err.Error())
	}
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
		go e.startStream(s, resChan, i+1)
	}
	e.wg.Wait()
	close(resChan)
}

func (e *parallelExecution) startStream(s *gauge.SpecCollection, resChan chan *result.SuiteResult, stream int) {
	defer e.wg.Done()
	runner, err := e.startRunner(s, stream)
	if len(err) > 0 {
		resChan <- &result.SuiteResult{UnhandledErrors: err}
		return
	}
	e.startSpecsExecutionWithRunner(s, resChan, runner, stream)
}

func (e *parallelExecution) startRunner(s *gauge.SpecCollection, stream int) (runner.Runner, []error) {
	if os.Getenv("GAUGE_CUSTOM_BUILD_PATH") == "" {
		os.Setenv("GAUGE_CUSTOM_BUILD_PATH", path.Join(os.Getenv("GAUGE_PROJECT_ROOT"), "gauge_bin"))
	}
	runner, err := runner.Start(e.manifest, stream, make(chan bool), false)
	if err != nil {
		logger.Errorf(true, "Failed to start runner. %s", err.Error())
		logger.Debugf(true, "Skipping %d specifications", s.Size())
		if isLazy() {
			return nil, []error{fmt.Errorf("Failed to start runner. %s", err.Error())}
		}
		return nil, []error{streamExecError{specsSkipped: s.SpecNames(), message: fmt.Sprintf("Failed to start runner. %s", err.Error())}}
	}
	return runner, nil
}

func (e *parallelExecution) startSpecsExecutionWithRunner(s *gauge.SpecCollection, resChan chan *result.SuiteResult, runner runner.Runner, stream int) {
	executionInfo := newExecutionInfo(s, runner, e.pluginHandler, e.errMaps, false, stream)
	se := newSimpleExecution(executionInfo, false)
	se.execute()
	err := runner.Kill()
	if err != nil {
		logger.Errorf(true, "Failed to kill runner. %s", err.Error())
	}
	resChan <- se.suiteResult
}

func (e *parallelExecution) executeSpecsInSerial(s *gauge.SpecCollection) *result.SuiteResult {
	runner, err := e.startRunner(s, 1)
	if err != nil {
		return &result.SuiteResult{UnhandledErrors: err}
	}
	executionInfo := newExecutionInfo(s, runner, e.pluginHandler, e.errMaps, false, 1)
	se := newSimpleExecution(executionInfo, false)
	se.execute()
	er := runner.Kill()
	if er != nil {
		logger.Errorf(true, "Failed to kill runner. %s", er.Error())
	}

	return se.suiteResult
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
	if err := e.runner.Kill(); err != nil {
		logger.Errorf(true, "Failed to kill Runner: %s", err.Error())
	}
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
	if !env.EnableMultiThreadedExecution() {
		return false
	}
	if !e.runner.IsMultithreaded() {
		logger.Warningf(true, "Runner doesn't support mutithreading, using multiprocess parallel execution.")
		return false
	}
	return true
}
