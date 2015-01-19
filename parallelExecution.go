package main

import (
	"runtime"
	"sync"
)

type parallelSpecExecution struct {
	manifest             *manifest
	runner               *testRunner
	specifications       []*specification
	pluginHandler        *pluginHandler
	currentExecutionInfo *ExecutionInfo
	suiteResult          *suiteResult
}

type specCollection struct {
	specs []*specification
}

func (e *parallelSpecExecution) start() *suiteResult {
	specCollections := e.distributeSpecs(numberOfCores())
	suiteResultChannel := make(chan *suiteResult)
	errChannel := make(chan error)
	errors := make([]error, 0)
	suiteResults := make([]*suiteResult, 0)
	var wg sync.WaitGroup
	for _, specCollection := range specCollections {
		wg.Add(1)
		go e.startSpecsExecution(specCollection, suiteResultChannel, errChannel, &wg)
	}
	wg.Wait()

	for err := range errChannel {
		errors = append(errors, err)
	}
	for result := range suiteResultChannel {
		suiteResults = append(suiteResults, result)
	}
	close(errChannel)
	close(suiteResultChannel)
	return suiteResults[0]
}

func (e *parallelSpecExecution) startSpecsExecution(specCollection *specCollection, suiteResults chan *suiteResult, errChannel chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	runner, err := startRunnerAndMakeConnection(e.manifest)
	if err != nil {
		errChannel <- err
		return
	}
	execution := newExecution(e.manifest, specCollection.specs, runner, e.pluginHandler, false)
	suiteResults <- execution.start()
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

func numberOfCores() int {
	return runtime.NumCPU()
}
