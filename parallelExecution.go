package main

import "runtime"

type parallelSpecExecution struct {
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
	e.distributeSpecs(numberOfCores())
	return nil
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
