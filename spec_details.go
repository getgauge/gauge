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
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	"github.com/golang/protobuf/proto"
	fsnotify "gopkg.in/fsnotify.v1"
	"path/filepath"
	"sync"
)

type specInfoGatherer struct {
	availableSpecs    []*specification
	availableStepsMap map[string]*stepValue
	runnerStepValues  []*stepValue
	fileToStepsMap    map[string][]*step
	conceptDictionary *conceptDictionary
	mutex             sync.Mutex
}

func (specInfoGatherer *specInfoGatherer) makeListOfAvailableSteps(runner *testRunner) *testRunner {
	specInfoGatherer.availableStepsMap = make(map[string]*stepValue)
	specInfoGatherer.fileToStepsMap = make(map[string][]*step)
	runner = specInfoGatherer.getStepsFromRunner(runner)

	// Concepts parsed first because we need to create a concept dictionary that spec parsing can use
	specInfoGatherer.findAllStepsFromConcepts()
	specInfoGatherer.findAllStepsFromSpecs()
	specInfoGatherer.updateAllStepsList()

	go specInfoGatherer.watchForFileChanges()
	return runner
}

// Parse all specifications in the project and find all the steps
func (specInfoGatherer *specInfoGatherer) findAllStepsFromSpecs() {
	specFiles := util.FindSpecFilesIn(filepath.Join(config.ProjectRoot, common.SpecsDirectoryName))

	availableSpecs, parseResults := parseSpecFiles(specFiles, specInfoGatherer.getDictionary())
	specInfoGatherer.handleParseFailures(parseResults)

	specInfoGatherer.addStepsForSpecs(availableSpecs)
}

func (specInfoGatherer *specInfoGatherer) createConceptsDictionary() {
	var result *parseResult
	specInfoGatherer.conceptDictionary, result = createConceptsDictionary(true)
	specInfoGatherer.handleParseFailures([]*parseResult{result})
}

func (specInfoGatherer *specInfoGatherer) handleParseFailures(parseResults []*parseResult) {
	for _, result := range parseResults {
		if !result.ok {
			logger.ApiLog.Error("Spec Parse failure: %s", result.Error())
		}
	}
}

// Watch specs and concepts for file changes and update local steps and spec cache used by the gauge api
func (specInfoGatherer *specInfoGatherer) watchForFileChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.ApiLog.Error("Error creating fileWatcher: %s", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				specInfoGatherer.handleEvent(event, watcher)
			case err := <-watcher.Errors:
				logger.ApiLog.Error("Error event while watching specs", err)
			}
		}
	}()

	allDirsToWatch := make([]string, 0)

	specDir := filepath.Join(config.ProjectRoot, common.SpecsDirectoryName)
	allDirsToWatch = append(allDirsToWatch, specDir)
	allDirsToWatch = append(allDirsToWatch, util.FindAllNestedDirs(specDir)...)

	for _, dir := range allDirsToWatch {
		specInfoGatherer.addDirToFileWatcher(watcher, dir)
	}
	<-done
}

func (specInfoGatherer *specInfoGatherer) addDirToFileWatcher(watcher *fsnotify.Watcher, dir string) {
	err := watcher.Add(dir)
	if err != nil {
		logger.ApiLog.Error("Unable to add directory %v to file watcher: %s", dir, err)
	} else {
		logger.ApiLog.Info("Watching directory: %s", dir)
	}
}

func (specInfoGatherer *specInfoGatherer) removeWatcherOn(watcher *fsnotify.Watcher, path string) {
	logger.ApiLog.Error("Removing watcher on : %s", path)
	watcher.Remove(path)
}

func (specInfoGatherer *specInfoGatherer) handleEvent(event fsnotify.Event, watcher *fsnotify.Watcher) {
	filePath, err := filepath.Abs(event.Name)
	if err != nil {
		logger.ApiLog.Error("Failed to get abs file path for %s: %s", event.Name, err)
		return
	}
	switch event.Op {
	case fsnotify.Create:
		specInfoGatherer.fileAdded(watcher, filePath)
	case fsnotify.Write:
		specInfoGatherer.fileModified(watcher, filePath)
	case fsnotify.Rename:
		specInfoGatherer.fileRenamed(watcher, filePath)
	case fsnotify.Remove:
		specInfoGatherer.fileRemoved(watcher, filePath)
	}
}

func (specInfoGatherer *specInfoGatherer) fileAdded(watcher *fsnotify.Watcher, fileName string) {
	if util.IsDir(fileName) {
		specInfoGatherer.addDirToFileWatcher(watcher, fileName)
	}
	specInfoGatherer.fileModified(watcher, fileName)
}

func (specInfoGatherer *specInfoGatherer) fileModified(watcher *fsnotify.Watcher, fileName string) {
	if util.IsSpec(fileName) {
		specInfoGatherer.addSpec(fileName)
	} else if util.IsConcept(fileName) {
		specInfoGatherer.addConcept(fileName)
	}
}

func (specInfoGatherer *specInfoGatherer) fileRemoved(watcher *fsnotify.Watcher, fileName string) {
	if util.IsSpec(fileName) {
		specInfoGatherer.removeSpec(fileName)
	} else if util.IsConcept(fileName) {
		specInfoGatherer.removeConcept(fileName)
	} else {
		specInfoGatherer.removeWatcherOn(watcher, fileName)
	}
}

func (specInfoGatherer *specInfoGatherer) fileRenamed(watcher *fsnotify.Watcher, fileName string) {
	specInfoGatherer.fileRemoved(watcher, fileName)
}

func (specInfoGatherer *specInfoGatherer) addSpec(fileName string) {
	logger.ApiLog.Info("Spec added/modified: %s", fileName)
	specs, parseResults := parseSpecFiles([]string{fileName}, specInfoGatherer.getDictionary())
	specInfoGatherer.handleParseFailures(parseResults)
	specInfoGatherer.addStepsForSpecs(specs)
	specInfoGatherer.updateAllStepsList()
}

func (specInfoGatherer *specInfoGatherer) addStepsForSpecs(specs []*specification) {
	specInfoGatherer.mutex.Lock()
	specInfoGatherer.addToAvailableSpecs(specs)
	specInfoGatherer.updateCache(specInfoGatherer.findAvailableStepsInSpecs(specs))
	specInfoGatherer.mutex.Unlock()
}

func (specInfoGatherer *specInfoGatherer) addToAvailableSpecs(specs []*specification) {
	if specInfoGatherer.availableSpecs == nil {
		specInfoGatherer.availableSpecs = make([]*specification, 0)
	}
	for _, spec := range specs {
		if _, ok := specInfoGatherer.fileToStepsMap[spec.fileName]; !ok {
			specInfoGatherer.availableSpecs = append(specInfoGatherer.availableSpecs, spec)
		}
	}
}

func (specInfoGatherer *specInfoGatherer) addConcept(fileName string) {
	logger.ApiLog.Info("Concept added/modified: %s", fileName)
	if err := addConcepts(fileName, specInfoGatherer.getDictionary()); err != nil {
		logger.ApiLog.Error("Concept parse failure: %s %s", fileName, err)
		return
	}
	specInfoGatherer.findAllStepsFromConcepts()
	specInfoGatherer.updateAllStepsList()
}

func (specInfoGatherer *specInfoGatherer) removeSpec(fileName string) {
	logger.ApiLog.Info("Spec removed: %s", fileName)
	specInfoGatherer.mutex.Lock()
	delete(specInfoGatherer.fileToStepsMap, fileName)
	specInfoGatherer.updateAllStepsList()
	specInfoGatherer.mutex.Unlock()
}

func (specInfoGatherer *specInfoGatherer) removeConcept(fileName string) {
	logger.ApiLog.Info("Concept removed: %s", fileName)
	delete(specInfoGatherer.fileToStepsMap, fileName)
	specInfoGatherer.createConceptsDictionary()
	specInfoGatherer.findAllStepsFromConcepts()
	specInfoGatherer.updateAllStepsList()
}

// Find all the steps defined in concepts. Look through the created concept dictionary to find concepts
func (specInfoGatherer *specInfoGatherer) findAllStepsFromConcepts() {
	allStepsInConcepts := make(map[string][]*step, 0)
	specInfoGatherer.createConceptsDictionary()
	for _, concept := range specInfoGatherer.getDictionary().conceptsMap {
		stepsInConcept := make([]*step, 0)
		for _, step := range concept.conceptStep.conceptSteps {
			if !step.isConcept {
				stepsInConcept = append(stepsInConcept, step)
			}
		}
		// Concept dictionary contains multiple entries for concepts that belong in the same file. So appending if the list of concepts for that file has been created.
		if _, ok := allStepsInConcepts[concept.fileName]; !ok {
			allStepsInConcepts[concept.fileName] = stepsInConcept
		} else {
			allStepsInConcepts[concept.fileName] = append(allStepsInConcepts[concept.fileName], stepsInConcept...)
		}
	}
	specInfoGatherer.updateCache(allStepsInConcepts)
}

func (specInfoGatherer *specInfoGatherer) createConceptInfos() []*gauge_messages.ConceptInfo {
	conceptInfos := make([]*gauge_messages.ConceptInfo, 0)
	for _, concept := range specInfoGatherer.getDictionary().conceptsMap {
		stepValue := createStepValue(concept.conceptStep)
		conceptInfos = append(conceptInfos, &gauge_messages.ConceptInfo{StepValue: convertToProtoStepValue(&stepValue), Filepath: proto.String(concept.fileName), LineNumber: proto.Int(concept.conceptStep.lineNo)})
	}
	return conceptInfos
}

// Gets all steps list from the runner
func (specInfoGatherer *specInfoGatherer) getStepsFromRunner(runner *testRunner) *testRunner {
	steps := make([]string, 0)
	if runner == nil {
		var connErr error
		runner, connErr = startRunnerAndMakeConnection(getProjectManifest(getCurrentExecutionLogger()), getCurrentExecutionLogger())
		if connErr == nil {
			steps = append(steps, requestForSteps(runner)...)
			logger.ApiLog.Debug("Steps got from runner: %v", steps)
		}
		if connErr != nil {
			logger.ApiLog.Error("Runner connection failed: %s", connErr)
		}

	} else {
		steps = append(steps, requestForSteps(runner)...)
		logger.ApiLog.Debug("Steps got from runner: %v", steps)
	}
	specInfoGatherer.runnerStepValues = specInfoGatherer.convertToStepValues(steps)
	return runner
}

func (specInfoGatherer *specInfoGatherer) convertToStepValues(steps []string) []*stepValue {
	stepValues := make([]*stepValue, 0)
	for _, step := range steps {
		stepValue, err := extractStepValueAndParams(step, false)
		if err != nil {
			logger.ApiLog.Error("Failed to extract stepvalue for step - %s : %s", step, err)
			continue
		}
		stepValues = append(stepValues, stepValue)
	}
	return stepValues
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
	specInfoGatherer.updateAllStepsList()
}

func (specInfoGatherer *specInfoGatherer) updateAllStepsList() {
	specInfoGatherer.availableStepsMap = make(map[string]*stepValue, 0)
	for _, steps := range specInfoGatherer.fileToStepsMap {
		for _, step := range steps {
			if step.isConcept {
				continue
			}
			stepValue := createStepValue(step)
			if _, ok := specInfoGatherer.availableStepsMap[stepValue.stepValue]; !ok {
				specInfoGatherer.availableStepsMap[stepValue.stepValue] = &stepValue
			}
		}
	}
	specInfoGatherer.addStepValuesToAvailableSteps(specInfoGatherer.runnerStepValues)
}

func (specInfoGatherer *specInfoGatherer) updateCache(newSpecStepsMap map[string][]*step) {
	if specInfoGatherer.fileToStepsMap == nil {
		specInfoGatherer.fileToStepsMap = make(map[string][]*step, 0)
	}
	for fileName, specSteps := range newSpecStepsMap {
		specInfoGatherer.fileToStepsMap[fileName] = specSteps
	}
}

func (specInfoGatherer *specInfoGatherer) addStepValuesToAvailableSteps(stepValues []*stepValue) {
	for _, stepValue := range stepValues {
		specInfoGatherer.addToAvailableSteps(stepValue)
	}
}

func (specInfoGatherer *specInfoGatherer) addToAvailableSteps(stepValue *stepValue) {
	if _, ok := specInfoGatherer.availableStepsMap[stepValue.stepValue]; !ok {
		specInfoGatherer.availableStepsMap[stepValue.stepValue] = stepValue
	}
}

func (specInfoGatherer *specInfoGatherer) getAvailableSteps() []*stepValue {
	steps := make([]*stepValue, 0)
	if specInfoGatherer.availableStepsMap == nil {
		return steps
	}
	specInfoGatherer.mutex.Lock()
	for _, stepValue := range specInfoGatherer.availableStepsMap {
		steps = append(steps, stepValue)
	}
	specInfoGatherer.mutex.Unlock()
	return steps
}

func (specInfoGatherer *specInfoGatherer) getDictionary() *conceptDictionary {
	if specInfoGatherer.conceptDictionary == nil {
		specInfoGatherer.conceptDictionary = newConceptDictionary()
	}
	return specInfoGatherer.conceptDictionary
}

func (specInfoGatherer *specInfoGatherer) getConceptInfos() []*gauge_messages.ConceptInfo {
	return specInfoGatherer.createConceptInfos()
}
