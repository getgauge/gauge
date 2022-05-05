/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
)

// ExecuteTags holds the tags to filter the execution by
var ExecuteTags = ""
var tableRowsIndexes []int

// SetTableRows is used to limit data driven execution to specific rows
func SetTableRows(tableRows string) {
	tableRowsIndexes = getDataTableRows(tableRows)
}

type simpleExecution struct {
	manifest             *manifest.Manifest
	runner               runner.Runner
	specCollection       *gauge.SpecCollection
	pluginHandler        plugin.Handler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	suiteResult          *result.SuiteResult
	errMaps              *gauge.BuildErrors
	startTime            time.Time
	stream               int
	skipSuiteEvents      bool
}

func newSimpleExecution(executionInfo *executionInfo, combineDataTableSpecs, skipSuiteEvents bool) *simpleExecution {
	if combineDataTableSpecs {
		executionInfo.specs = gauge.NewSpecCollection(executionInfo.specs.Specs(), true)
	}
	ei := &gauge_messages.ExecutionInfo{
		ProjectName:              filepath.Base(config.ProjectRoot),
		NumberOfExecutionStreams: int32(NumberOfExecutionStreams),
		RunnerId:                 int32(executionInfo.stream),
		ExecutionArgs:            gauge.ConvertToProtoExecutionArg(ExecutionArgs),
	}

	return &simpleExecution{
		manifest:             executionInfo.manifest,
		specCollection:       executionInfo.specs,
		runner:               executionInfo.runner,
		pluginHandler:        executionInfo.pluginHandler,
		errMaps:              executionInfo.errMaps,
		stream:               executionInfo.stream,
		skipSuiteEvents:      skipSuiteEvents,
		currentExecutionInfo: ei,
	}
}

func (e *simpleExecution) run() *result.SuiteResult {
	e.start()
	e.execute()
	e.finish()
	return e.suiteResult
}

func (e *simpleExecution) execute() {
	e.suiteResult = result.NewSuiteResult(ExecuteTags, e.startTime)
	defer func() {
		e.suiteResult.UpdateExecTime(e.startTime)
		e.suiteResult.SetSpecsSkippedCount()
	}()
	if !e.skipSuiteEvents {
		logger.Debug(true, "Initialising suite data store.")
		initSuiteDataStoreResult := e.initSuiteDataStore()
		if initSuiteDataStoreResult.GetFailed() {
			e.suiteResult.AddUnhandledError(fmt.Errorf("failed to initialize suite datastore. Error: %s", initSuiteDataStoreResult.GetErrorMessage()))
			return
		}
		e.notifyBeforeSuite()
	}

	if !e.suiteResult.GetFailed() {
		results := e.executeSpecs(e.specCollection)
		e.suiteResult.AddSpecResults(results)
	}

	if !e.skipSuiteEvents {
		e.notifyAfterSuite()
	}
}

func (e *simpleExecution) start() {
	e.startTime = time.Now()
	event.Notify(event.NewExecutionEvent(event.SuiteStart, nil, nil, 0, &gauge_messages.ExecutionInfo{}))
	e.pluginHandler = plugin.StartPlugins(e.manifest)
}

func (e *simpleExecution) finish() {
	e.suiteResult = mergeDataTableSpecResults(e.suiteResult)
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, e.suiteResult, 0, &gauge_messages.ExecutionInfo{}))
	e.notifyExecutionResult()
	e.stopAllPlugins()
}

func (e *simpleExecution) stopAllPlugins() {
	e.notifyExecutionStop()
	if err := e.runner.Kill(); err != nil {
		logger.Errorf(true, "Failed to kill Runner: %s", err.Error())
	}
}

func (e *simpleExecution) executeSpecs(sc *gauge.SpecCollection) (results []*result.SpecResult) {
	for sc.HasNext() {
		specs := sc.Next()
		var preHookFailures, postHookFailures []*gauge_messages.ProtoHookFailure
		var specResults []*result.SpecResult
		var before, after = true, false
		for i, spec := range specs {
			if i == len(specs)-1 {
				after = true
			}
			res := newSpecExecutor(spec, e.runner, e.pluginHandler, e.errMaps, e.stream).execute(before, preHookFailures == nil, after)
			before = false
			specResults = append(specResults, res)
			preHookFailures = append(preHookFailures, res.GetPreHook()...)
			postHookFailures = append(postHookFailures, res.GetPostHook()...)
			res.ProtoSpec.PreHookFailures, res.ProtoSpec.PostHookFailures = []*gauge_messages.ProtoHookFailure{}, []*gauge_messages.ProtoHookFailure{}
		}
		for _, res := range specResults {
			for _, preHook := range preHookFailures {
				res.AddPreHook(&gauge_messages.ProtoHookFailure{
					StackTrace:            preHook.StackTrace,
					ErrorMessage:          preHook.ErrorMessage,
					FailureScreenshotFile: preHook.FailureScreenshotFile,
					TableRowIndex:         preHook.TableRowIndex,
				})
			}
			for _, postHook := range postHookFailures {
				res.AddPostHook(&gauge_messages.ProtoHookFailure{
					StackTrace:            postHook.StackTrace,
					ErrorMessage:          postHook.ErrorMessage,
					FailureScreenshotFile: postHook.FailureScreenshotFile,
					TableRowIndex:         postHook.TableRowIndex,
				})
			}
			results = append(results, res)
		}
	}
	return results
}

func (e *simpleExecution) notifyBeforeSuite() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionStarting,
		ExecutionStartingRequest: &gauge_messages.ExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo, Stream: int32(e.stream)}}
	res := e.executeHook(m)
	e.suiteResult.PreHookMessages = res.Message
	e.suiteResult.PreHookScreenshotFiles = res.ScreenshotFiles
	if res.GetFailed() {
		handleHookFailure(e.suiteResult, res, result.AddPreHook)
	}
	m.ExecutionStartingRequest.SuiteResult = gauge.ConvertToProtoSuiteResult(e.suiteResult)
	e.pluginHandler.NotifyPlugins(m)
}

func (e *simpleExecution) notifyAfterSuite() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecutionEnding,
		ExecutionEndingRequest: &gauge_messages.ExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo, Stream: int32(e.stream)}}
	res := e.executeHook(m)
	e.suiteResult.PostHookMessages = res.Message
	e.suiteResult.PostHookScreenshotFiles = res.ScreenshotFiles
	if res.GetFailed() {
		handleHookFailure(e.suiteResult, res, result.AddPostHook)
	}
	m.ExecutionEndingRequest.SuiteResult = gauge.ConvertToProtoSuiteResult(e.suiteResult)
	e.pluginHandler.NotifyPlugins(m)
}

func (e *simpleExecution) initSuiteDataStore() *(gauge_messages.ProtoExecutionResult) {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteDataStoreInit,
		SuiteDataStoreInitRequest: &gauge_messages.SuiteDataStoreInitRequest{Stream: int32(e.stream)}}
	return e.runner.ExecuteAndGetStatus(m)
}

func (e *simpleExecution) executeHook(m *gauge_messages.Message) *(gauge_messages.ProtoExecutionResult) {
	e.pluginHandler.NotifyPlugins(m)
	return e.runner.ExecuteAndGetStatus(m)
}

func (e *simpleExecution) notifyExecutionResult() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteExecutionResult,
		SuiteExecutionResult: &gauge_messages.SuiteExecutionResult{SuiteResult: gauge.ConvertToProtoSuiteResult(e.suiteResult)}}
	e.pluginHandler.NotifyPlugins(m)
}

func (e *simpleExecution) notifyExecutionStop() {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_KillProcessRequest,
		KillProcessRequest: &gauge_messages.KillProcessRequest{}}
	e.pluginHandler.NotifyPlugins(m)
	e.pluginHandler.GracefullyKillPlugins()
}

func handleHookFailure(hookResult result.Result, execResult *gauge_messages.ProtoExecutionResult, f func(result.Result, *gauge_messages.ProtoExecutionResult)) {
	f(hookResult, execResult)
}
