package execution

import (
	"fmt"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/runner"
)

func (e *parallelExecution) executeGrpcMultithreaded() {
	defer close(e.resultChan)
	totalStreams := e.numberOfStreams()
	e.wg.Add(totalStreams)
	r, ok := e.runners[0].(*runner.GrpcRunner)
	if !ok {
		logger.Fatalf(true, "Expected GrpcRunner, but got %T instead. Gauge cannot use this runner.", e.runners[0])
	}
	r.IsExecuting = true
	e.suiteResult = result.NewSuiteResult(ExecuteTags, e.startTime)
	res := initSuiteDataStore(r)
	if res.GetFailed() {
		e.suiteResult.AddUnhandledError(fmt.Errorf("failed to initialize suite datastore. Error: %s", res.GetErrorMessage()))
		return
	}
	e.notifyBeforeSuite()

	for i := 1; i <= totalStreams; i++ {
		go func(stream int) {
			defer e.wg.Done()
			executionInfo := newExecutionInfo(e.specCollection, r, e.pluginHandler, e.errMaps, false, stream)
			se := newSimpleExecution(executionInfo, false, true)
			se.execute()
			e.resultChan <- se.suiteResult
		}(i)
	}

	e.wg.Wait()
	e.notifyAfterSuite()
	r.IsExecuting = false
	if err := r.Kill(); err != nil {
		logger.Infof(true, "unable to kill runner: %s", err.Error())
	}
}

func initSuiteDataStore(r runner.Runner) *gauge_messages.ProtoExecutionResult {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteDataStoreInit,
		SuiteDataStoreInitRequest: &gauge_messages.SuiteDataStoreInitRequest{}}
	return r.ExecuteAndGetStatus(m)
}

func (e *parallelExecution) notifyBeforeSuite() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionStarting,
		ExecutionStartingRequest: &gauge_messages.ExecutionStartingRequest{
			CurrentExecutionInfo: &gauge_messages.ExecutionInfo{},
			Stream:               1},
	}
	e.pluginHandler.NotifyPlugins(m)
	res := e.runners[0].ExecuteAndGetStatus(m)
	e.suiteResult.PreHookMessages = res.Message
	e.suiteResult.PreHookScreenshotFiles = res.ScreenshotFiles
	e.suiteResult.PreHookScreenshots = res.Screenshots
	if res.GetFailed() {
		result.AddPreHook(e.suiteResult, res)
	}
	m.ExecutionStartingRequest.SuiteResult = gauge.ConvertToProtoSuiteResult(e.suiteResult)
	e.pluginHandler.NotifyPlugins(m)
}

func (e *parallelExecution) notifyAfterSuite() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionEnding,
		ExecutionEndingRequest: &gauge_messages.ExecutionEndingRequest{
			CurrentExecutionInfo: &gauge_messages.ExecutionInfo{},
			Stream:               1,
		},
	}
	e.pluginHandler.NotifyPlugins(m)
	res := e.runners[0].ExecuteAndGetStatus(m)
	e.suiteResult.PostHookMessages = res.Message
	e.suiteResult.PostHookScreenshotFiles = res.ScreenshotFiles
	e.suiteResult.PostHookScreenshots = res.Screenshots
	if res.GetFailed() {
		result.AddPostHook(e.suiteResult, res)
	}
	m.ExecutionEndingRequest.SuiteResult = gauge.ConvertToProtoSuiteResult(e.suiteResult)
	e.pluginHandler.NotifyPlugins(m)
}
