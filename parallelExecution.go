package main

import (
	"fmt"
	"github.com/getgauge/gauge/gauge_messages"
	"runtime"
)

type parallelSpecExecution struct {
	manifest             *manifest
	runner               *testRunner
	specifications       []*specification
	pluginHandler        *pluginHandler
	currentExecutionInfo *gauge_messages.ExecutionInfo
	aggregateResult      *suiteResult
}

type specCollection struct {
	specs []*specification
}

func (e *parallelSpecExecution) start() *suiteResult {
	specCollections := e.distributeSpecs(numberOfCores())
	suiteResultChannel := make(chan *suiteResult, len(specCollections))

	for _, specCollection := range specCollections {
		go e.startSpecsExecution(specCollection, suiteResultChannel)
	}

	suiteResults := make([]*suiteResult, 0)
	for _, _ = range specCollections {
		suiteResults = append(suiteResults, <-suiteResultChannel)
	}

	e.aggregateResult = e.aggregateResults(suiteResults)
	return e.aggregateResult
}

func (e *parallelSpecExecution) startSpecsExecution(specCollection *specCollection, suiteResults chan *suiteResult) {
	runner, err := startRunnerAndMakeConnection(e.manifest)
	if err != nil {
		fmt.Println("Failed: " + err.Error())
		suiteResults <- &suiteResult{}
		return
	}
	execution := newExecution(e.manifest, specCollection.specs, runner, e.pluginHandler, false)
	result := execution.start()
	runner.kill()
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
	}
	return aggregateResult
}

func numberOfCores() int {
	return runtime.NumCPU()
}
