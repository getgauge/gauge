/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
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
	runners                  []runner.Runner
	suiteResult              *result.SuiteResult
	numberOfExecutionStreams int
	tagsToFilter             string
	errMaps                  *gauge.BuildErrors
	startTime                time.Time
	resultChan               chan *result.SuiteResult
}

func newParallelExecution(e *executionInfo) *parallelExecution {
	return &parallelExecution{
		manifest:                 e.manifest,
		specCollection:           e.specs,
		runners:                  []runner.Runner{e.runner},
		pluginHandler:            e.pluginHandler,
		numberOfExecutionStreams: e.numberOfStreams,
		tagsToFilter:             e.tagsToFilter,
		errMaps:                  e.errMaps,
		resultChan:               make(chan *result.SuiteResult),
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
	event.Notify(event.NewExecutionEvent(event.SuiteStart, nil, nil, 0, &gauge_messages.ExecutionInfo{}))
	e.pluginHandler = plugin.StartPlugins(e.manifest)
}

func (e *parallelExecution) startRunnersForRemainingStreams() {
	totalStreams := e.numberOfStreams()
	rChan := make(chan runner.Runner, totalStreams-1)
	for i := 2; i <= totalStreams; i++ {
		go func(stream int) {
			r, err := e.startRunner(e.specCollection, stream)
			if len(err) > 0 {
				e.resultChan <- &result.SuiteResult{UnhandledErrors: err}
				return
			}
			rChan <- r
		}(i)
	}
	for i := 1; i < totalStreams; i++ {
		e.runners = append(e.runners, <-rChan)
	}
}

func (e *parallelExecution) run() *result.SuiteResult {
	e.start()
	var res []*result.SuiteResult
	if env.AllowFilteredParallelExecution() && e.tagsToFilter != "" {
		parallesSpecs, serialSpecs := filter.FilterSpecForParallelRun(e.specCollection.Specs(), e.tagsToFilter)
		if Verbose {
			logger.Infof(true, "Applied tags '%s' to filter specs for parallel execution", e.tagsToFilter)
			logger.Infof(true, "No of specs to be executed in serial : %d", len(serialSpecs))
			logger.Infof(true, "No of specs to be executed in parallel : %d", len(parallesSpecs))
		}
		if len(serialSpecs) > 0 {
			logger.Infof(true, "Executing %d specs in serial.", len(serialSpecs))
			e.specCollection = gauge.NewSpecCollection(parallesSpecs, false)
			res = append(res, e.executeSpecsInSerial(gauge.NewSpecCollection(serialSpecs, true)))
		}
	}

	if e.specCollection.Size() > 0 {
		logger.Infof(true, "Executing in %d parallel streams.", e.numberOfStreams())
		// skipcq CRT-A0013
		if e.isMultithreaded() {
			logger.Debugf(true, "Using multithreading for parallel execution.")
			if e.runners[0].Info().GRPCSupport {
				go e.executeGrpcMultithreaded()
			} else {
				go e.executeLegacyMultithreaded()
			}
		} else if isLazy() {
			go e.executeLazily()
		} else {
			go e.executeEagerly()
		}

		for r := range e.resultChan {
			res = append(res, r)
		}
	} else {
		logger.Infof(true, "No specs remains to execute in parallel.")
	}
	e.aggregateResults(res)
	e.finish()
	return e.suiteResult
}

func (e *parallelExecution) executeLazily() {
	defer close(e.resultChan)
	e.wg.Add(e.numberOfStreams())
	e.startRunnersForRemainingStreams()

	for i := 1; i <= len(e.runners); i++ {
		go func(stream int) {
			defer e.wg.Done()
			e.startSpecsExecutionWithRunner(e.specCollection, e.runners[stream-1], stream)
		}(i)
	}
	e.wg.Wait()
}

func (e *parallelExecution) executeLegacyMultithreaded() {
	defer close(e.resultChan)
	totalStreams := e.numberOfStreams()
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
	_ = os.Setenv(gaugeAPIPortsEnv, strings.Join(ports, ","))
	writer := logger.NewLogWriter(e.manifest.Language, true, 0)
	r, err := runner.StartLegacyRunner(e.manifest, "0", writer, make(chan bool), false)
	if err != nil {
		logger.Fatalf(true, "failed to start runner. %s", err.Error())
	}
	for i := 0; i < totalStreams; i++ {
		connection, err := handlers[i].AcceptConnection(config.RunnerConnectionTimeout(), make(chan error))
		if err != nil {
			logger.Error(true, err.Error())
		}
		crapRunner := &runner.MultithreadedRunner{}
		crapRunner.SetConnection(connection)
		go e.startMultithreaded(crapRunner, e.resultChan, i+1)
	}
	e.wg.Wait()
	err = r.Cmd.Process.Kill()
	if err != nil {
		logger.Infof(true, "unable to kill runner: %s", err.Error())
	}
}

func (e *parallelExecution) startMultithreaded(r runner.Runner, resChan chan *result.SuiteResult, stream int) {
	defer e.wg.Done()
	e.startSpecsExecutionWithRunner(e.specCollection, r, stream)
}

func (e *parallelExecution) executeEagerly() {
	defer close(e.resultChan)
	distributions := e.numberOfStreams()
	specs := filter.DistributeSpecs(e.specCollection.Specs(), distributions)
	e.wg.Add(distributions)
	e.startRunnersForRemainingStreams()

	for i, s := range specs {
		i, s := i, s
		go func(j int) {
			defer e.wg.Done()
			e.startSpecsExecutionWithRunner(s, e.runners[j], j+1)
		}(i)
	}
	e.wg.Wait()
}

func (e *parallelExecution) startRunner(s *gauge.SpecCollection, stream int) (runner.Runner, []error) {
	if os.Getenv("GAUGE_CUSTOM_BUILD_PATH") == "" {
		_ = os.Setenv("GAUGE_CUSTOM_BUILD_PATH", path.Join(os.Getenv("GAUGE_PROJECT_ROOT"), "gauge_bin"))
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

func (e *parallelExecution) startSpecsExecutionWithRunner(s *gauge.SpecCollection, runner runner.Runner, stream int) {
	executionInfo := newExecutionInfo(s, runner, e.pluginHandler, e.errMaps, false, stream)
	se := newSimpleExecution(executionInfo, false, false)
	se.execute()
	err := runner.Kill()
	if err != nil {
		logger.Errorf(true, "Failed to kill runner. %s", err.Error())
	}
	e.resultChan <- se.suiteResult
}

func (e *parallelExecution) executeSpecsInSerial(s *gauge.SpecCollection) *result.SuiteResult {
	runner, err := e.startRunner(s, 1)
	if err != nil {
		return &result.SuiteResult{UnhandledErrors: err}
	}
	executionInfo := newExecutionInfo(s, runner, e.pluginHandler, e.errMaps, false, 1)
	se := newSimpleExecution(executionInfo, false, false)
	se.execute()
	er := runner.Kill()
	if er != nil {
		logger.Errorf(true, "Failed to kill runner. %s", er.Error())
	}

	return se.suiteResult
}

func (e *parallelExecution) finish() {
	e.suiteResult = mergeDataTableSpecResults(e.suiteResult)
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, e.suiteResult, 0, &gauge_messages.ExecutionInfo{}))
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
	if e.suiteResult != nil {
		r.PreHookMessages = e.suiteResult.PreHookMessages
		r.PreHookScreenshotFiles = e.suiteResult.PreHookScreenshotFiles
		r.PostHookMessages = e.suiteResult.PostHookMessages
		r.PostHookScreenshotFiles = e.suiteResult.PostHookScreenshotFiles
	}
	for _, suiteResult := range suiteResults {
		r.SpecsFailedCount += suiteResult.SpecsFailedCount
		r.SpecResults = append(r.SpecResults, suiteResult.SpecResults...)
		r.PreHookMessages = append(r.PreHookMessages, suiteResult.PreHookMessages...)
		r.PreHookScreenshotFiles = append(r.PreHookScreenshotFiles, suiteResult.PreHookScreenshotFiles...)
		r.PostHookMessages = append(r.PostHookMessages, suiteResult.PostHookMessages...)
		r.PostHookScreenshotFiles = append(r.PostHookScreenshotFiles, suiteResult.PostHookScreenshotFiles...)
		if suiteResult.IsFailed {
			r.IsFailed = true
		}
		if suiteResult.PreSuite != nil {
			r.PreSuite = suiteResult.PreSuite
		}
		if suiteResult.PostSuite != nil {
			r.PostSuite = suiteResult.PostSuite
		}
		if suiteResult.UnhandledErrors != nil {
			r.UnhandledErrors = append(r.UnhandledErrors, suiteResult.UnhandledErrors...)
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
	if !e.runners[0].IsMultithreaded() {
		logger.Warningf(true, "Runner doesn't support mutithreading, using multiprocess parallel execution.")
		return false
	}
	return true
}
