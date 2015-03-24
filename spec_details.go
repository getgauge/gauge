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

package main

import (
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
	"strconv"
	"sync"
	"time"
)

type specInfoGatherer struct {
	availableSpecs    []*specification
	availableStepsMap map[string]*stepValue
	stepsFromRunner   []string
	specStepMapCache  map[string][]*step
	conceptInfos      []*gauge_messages.ConceptInfo
	mutex             sync.Mutex
}

func (specInfoGatherer *specInfoGatherer) makeListOfAvailableSteps(runner *testRunner) {
	specInfoGatherer.availableStepsMap = make(map[string]*stepValue)
	specInfoGatherer.specStepMapCache = make(map[string][]*step)
	specInfoGatherer.stepsFromRunner = specInfoGatherer.getStepsFromRunner(runner)
	specInfoGatherer.addStepValuesToAvailableSteps(specInfoGatherer.stepsFromRunner)
	newSpecStepMap, conceptInfos := specInfoGatherer.getAllStepsFromSpecs()
	specInfoGatherer.conceptInfos = conceptInfos
	specInfoGatherer.addStepsToAvailableSteps(newSpecStepMap)

	conceptStepsMap := specInfoGatherer.getAllStepsFromConcepts()
	specInfoGatherer.addStepsToAvailableSteps(conceptStepsMap)

	go specInfoGatherer.refreshSteps(config.ApiRefreshInterval())
}

func (specInfoGatherer *specInfoGatherer) getAllStepsFromSpecs() (map[string][]*step, []*gauge_messages.ConceptInfo) {
	dictionary, _ := createConceptsDictionary(true)
	specInfoGatherer.availableSpecs = specInfoGatherer.findSpecs(common.SpecsDirectoryName, dictionary)
	return specInfoGatherer.findAvailableStepsInSpecs(specInfoGatherer.availableSpecs), specInfoGatherer.createConceptInfos(dictionary)
}

func (specInfoGatherer *specInfoGatherer) getAllStepsFromConcepts() map[string][]*step {
	allStepsInConcepts := make(map[string][]*step, 0)
	conceptFiles := findConceptFiles()
	for _, conceptFile := range conceptFiles {
		fileText, fileReadErr := common.ReadFileContents(conceptFile)
		if fileReadErr != nil {
			apiLog.Error("failed to read concept file %s", conceptFile)
			continue
		}
		concepts, err := new(conceptParser).parse(fileText)
		if err != nil {
			apiLog.Error("Concept Parse failure: %s: line no: %s, %s", conceptFile, strconv.Itoa(err.lineNo), err.message)
			continue
		}
		conceptSteps := make([]*step, 0)
		for _, concept := range concepts {
			for _, conceptStep := range concept.conceptSteps {
				conceptSteps = append(conceptSteps, conceptStep)
			}
		}
		allStepsInConcepts[conceptFile] = conceptSteps
	}
	return allStepsInConcepts
}
func (specInfoGatherer *specInfoGatherer) createConceptInfos(dictionary *conceptDictionary) []*gauge_messages.ConceptInfo {
	conceptInfos := make([]*gauge_messages.ConceptInfo, 0)
	for _, concept := range dictionary.conceptsMap {
		stepValue, err := extractStepValueAndParams(concept.conceptStep.lineText, concept.conceptStep.hasInlineTable)
		if err != nil {
			continue
		}
		conceptInfos = append(conceptInfos, &gauge_messages.ConceptInfo{StepValue: convertToProtoStepValue(stepValue), Filepath: proto.String(concept.fileName), LineNumber: proto.Int(concept.conceptStep.lineNo)})
	}
	return conceptInfos
}

func (specInfoGatherer *specInfoGatherer) refreshSteps(seconds time.Duration) {
	for {
		time.Sleep(seconds)
		specInfoGatherer.mutex.Lock()
		specInfoGatherer.availableStepsMap = make(map[string]*stepValue, 0)
		specInfoGatherer.addStepValuesToAvailableSteps(specInfoGatherer.stepsFromRunner)
		newSpecStepMap, conceptInfos := specInfoGatherer.getAllStepsFromSpecs()
		specInfoGatherer.conceptInfos = conceptInfos
		specInfoGatherer.addStepsToAvailableSteps(newSpecStepMap)

		conceptStepsMap := specInfoGatherer.getAllStepsFromConcepts()
		specInfoGatherer.addStepsToAvailableSteps(conceptStepsMap)

		specInfoGatherer.mutex.Unlock()
	}
}

func (specInfoGatherer *specInfoGatherer) getStepsFromRunner(runner *testRunner) []string {
	steps := make([]string, 0)
	if runner == nil {
		runner, connErr := startRunnerAndMakeConnection(getProjectManifest(getCurrentExecutionLogger()), getCurrentExecutionLogger())
		if connErr == nil {
			steps = append(steps, requestForSteps(runner)...)
			apiLog.Debug("Steps got from runner: %v", steps)
			runner.kill(getCurrentExecutionLogger())
		}
		if connErr != nil {
			apiLog.Error("Runner connection failed: %s", connErr)
		}

	} else {
		steps = append(steps, requestForSteps(runner)...)
		apiLog.Debug("Steps got from runner: %v", steps)
	}
	return steps
}
func (specInfoGatherer *specInfoGatherer) findSpecs(specSource string, conceptDictionary *conceptDictionary) []*specification {
	specs, parseResults := findSpecs(specSource, conceptDictionary)
	specInfoGatherer.handleParseFailures(parseResults)
	return specs
}

func (specInfoGatherer *specInfoGatherer) handleParseFailures(parseResults []*parseResult) {
	for _, result := range parseResults {
		if !result.ok {
			apiLog.Error("Spec Parse failure: %s", result.Error())
		}
	}
}

func (specInfoGatherer *specInfoGatherer) findAvailableStepsInSpecs(specs []*specification) map[string][]*step {
	specStepsMap := make(map[string][]*step)
	for _, spec := range specs {
		stepsInSpec := make([]*step, 0)
		stepsInSpec = append(stepsInSpec, spec.contexts...)
		for _, scenario := range spec.scenarios {
			stepsInSpec = append(stepsInSpec, scenario.steps...)
		}
		specStepsMap[spec.fileName] = stepsInSpec
	}
	return specStepsMap
}

func (specInfoGatherer *specInfoGatherer) addStepsToAvailableSteps(newSpecStepsMap map[string][]*step) {
	specInfoGatherer.updateCache(newSpecStepsMap)
	for _, steps := range specInfoGatherer.specStepMapCache {
		for _, step := range steps {
			if step.isConcept {
				continue
			}
			stepValue, err := extractStepValueAndParams(step.lineText, step.hasInlineTable)
			if err == nil {
				if _, ok := specInfoGatherer.availableStepsMap[stepValue.stepValue]; !ok {
					specInfoGatherer.availableStepsMap[stepValue.stepValue] = stepValue
				}
			}
		}
	}

}

func (specInfoGatherer *specInfoGatherer) updateCache(newSpecStepsMap map[string][]*step) {
	for fileName, specsteps := range newSpecStepsMap {
		specInfoGatherer.specStepMapCache[fileName] = specsteps
	}
}

func (specInfoGatherer *specInfoGatherer) addStepValuesToAvailableSteps(stepValues []string) {
	for _, step := range stepValues {
		specInfoGatherer.addToAvailableSteps(step)
	}
}

func (specInfoGatherer *specInfoGatherer) addToAvailableSteps(stepText string) {
	stepValue, err := extractStepValueAndParams(stepText, false)
	if err == nil {
		if _, ok := specInfoGatherer.availableStepsMap[stepValue.stepValue]; !ok {
			specInfoGatherer.availableStepsMap[stepValue.stepValue] = stepValue
		}
	}
}

func (specInfoGatherer *specInfoGatherer) getAvailableSteps() []*stepValue {
	if specInfoGatherer.availableStepsMap == nil {
		specInfoGatherer.makeListOfAvailableSteps(nil)
	}
	specInfoGatherer.mutex.Lock()
	steps := make([]*stepValue, 0)
	for _, stepValue := range specInfoGatherer.availableStepsMap {
		steps = append(steps, stepValue)
	}
	specInfoGatherer.mutex.Unlock()
	return steps
}
