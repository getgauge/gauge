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

package infoGatherer

import (
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/logger/execLogger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/golang/protobuf/proto"
	fsnotify "gopkg.in/fsnotify.v1"
	"path/filepath"
	"sync"
)

type SpecInfoGatherer struct {
	AvailableSpecs    []*parser.Specification
	availableStepsMap map[string]*parser.StepValue
	runnerStepValues  []*parser.StepValue
	fileToStepsMap    map[string][]*parser.Step
	conceptDictionary *parser.ConceptDictionary
	mutex             sync.Mutex
}

func (specInfoGatherer *SpecInfoGatherer) MakeListOfAvailableSteps(killChannel chan bool) (*runner.TestRunner, error) {
	specInfoGatherer.availableStepsMap = make(map[string]*parser.StepValue)
	specInfoGatherer.fileToStepsMap = make(map[string][]*parser.Step)
	runner, err := specInfoGatherer.getStepsFromRunner(killChannel)
	// Concepts parsed first because we need to create a concept dictionary that spec parsing can use
	specInfoGatherer.findAllStepsFromConcepts()
	specInfoGatherer.findAllStepsFromSpecs()
	specInfoGatherer.updateAllStepsList()
	go specInfoGatherer.watchForFileChanges()
	return runner, err
}

// Parse all specifications in the project and find all the steps
func (specInfoGatherer *SpecInfoGatherer) findAllStepsFromSpecs() {
	specFiles := util.FindSpecFilesIn(filepath.Join(config.ProjectRoot, common.SpecsDirectoryName))

	availableSpecs, parseResults := parser.ParseSpecFiles(specFiles, specInfoGatherer.getDictionary())
	specInfoGatherer.handleParseFailures(parseResults)

	specInfoGatherer.addStepsForSpecs(availableSpecs)
}

func (specInfoGatherer *SpecInfoGatherer) createConceptsDictionary() {
	var result *parser.ParseResult
	specInfoGatherer.conceptDictionary, result = parser.CreateConceptsDictionary(true)
	specInfoGatherer.handleParseFailures([]*parser.ParseResult{result})
}

func (specInfoGatherer *SpecInfoGatherer) handleParseFailures(parseResults []*parser.ParseResult) {
	for _, result := range parseResults {
		if !result.Ok {
			logger.ApiLog.Error("Spec Parse failure: %s", result.Error())
		}
	}
}

// Watch specs and concepts for file changes and update local steps and spec cache used by the gauge api
func (specInfoGatherer *SpecInfoGatherer) watchForFileChanges() {
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

func (specInfoGatherer *SpecInfoGatherer) addDirToFileWatcher(watcher *fsnotify.Watcher, dir string) {
	err := watcher.Add(dir)
	if err != nil {
		logger.ApiLog.Error("Unable to add directory %v to file watcher: %s", dir, err)
	} else {
		logger.ApiLog.Info("Watching directory: %s", dir)
	}
}

func (specInfoGatherer *SpecInfoGatherer) removeWatcherOn(watcher *fsnotify.Watcher, path string) {
	logger.ApiLog.Error("Removing watcher on : %s", path)
	watcher.Remove(path)
}

func (specInfoGatherer *SpecInfoGatherer) handleEvent(event fsnotify.Event, watcher *fsnotify.Watcher) {
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

func (specInfoGatherer *SpecInfoGatherer) fileAdded(watcher *fsnotify.Watcher, fileName string) {
	if util.IsDir(fileName) {
		specInfoGatherer.addDirToFileWatcher(watcher, fileName)
	}
	specInfoGatherer.fileModified(watcher, fileName)
}

func (specInfoGatherer *SpecInfoGatherer) fileModified(watcher *fsnotify.Watcher, fileName string) {
	if util.IsSpec(fileName) {
		specInfoGatherer.addSpec(fileName)
	} else if util.IsConcept(fileName) {
		specInfoGatherer.addConcept(fileName)
	}
}

func (specInfoGatherer *SpecInfoGatherer) fileRemoved(watcher *fsnotify.Watcher, fileName string) {
	if util.IsSpec(fileName) {
		specInfoGatherer.removeSpec(fileName)
	} else if util.IsConcept(fileName) {
		specInfoGatherer.removeConcept(fileName)
	} else {
		specInfoGatherer.removeWatcherOn(watcher, fileName)
	}
}

func (specInfoGatherer *SpecInfoGatherer) fileRenamed(watcher *fsnotify.Watcher, fileName string) {
	specInfoGatherer.fileRemoved(watcher, fileName)
}

func (specInfoGatherer *SpecInfoGatherer) addSpec(fileName string) {
	logger.ApiLog.Info("Spec added/modified: %s", fileName)
	specs, parseResults := parser.ParseSpecFiles([]string{fileName}, specInfoGatherer.getDictionary())
	specInfoGatherer.handleParseFailures(parseResults)
	specInfoGatherer.addStepsForSpecs(specs)
	specInfoGatherer.updateAllStepsList()
}

func (specInfoGatherer *SpecInfoGatherer) addStepsForSpecs(specs []*parser.Specification) {
	specInfoGatherer.mutex.Lock()
	specInfoGatherer.addToAvailableSpecs(specs)
	specInfoGatherer.updateCache(specInfoGatherer.findAvailableStepsInSpecs(specs))
	specInfoGatherer.mutex.Unlock()
}

func (specInfoGatherer *SpecInfoGatherer) addToAvailableSpecs(specs []*parser.Specification) {
	if specInfoGatherer.AvailableSpecs == nil {
		specInfoGatherer.AvailableSpecs = make([]*parser.Specification, 0)
	}
	for _, spec := range specs {
		if _, ok := specInfoGatherer.fileToStepsMap[spec.FileName]; !ok {
			specInfoGatherer.AvailableSpecs = append(specInfoGatherer.AvailableSpecs, spec)
		}
	}
}

func (specInfoGatherer *SpecInfoGatherer) addConcept(fileName string) {
	logger.ApiLog.Info("Concept added/modified: %s", fileName)
	if err := parser.AddConcepts(fileName, specInfoGatherer.getDictionary()); err != nil {
		logger.ApiLog.Error("Concept parse failure: %s %s", fileName, err)
		return
	}
	specInfoGatherer.findAllStepsFromConcepts()
	specInfoGatherer.updateAllStepsList()
}

func (specInfoGatherer *SpecInfoGatherer) removeSpec(fileName string) {
	logger.ApiLog.Info("Spec removed: %s", fileName)
	specInfoGatherer.mutex.Lock()
	delete(specInfoGatherer.fileToStepsMap, fileName)
	specInfoGatherer.updateAllStepsList()
	specInfoGatherer.mutex.Unlock()
}

func (specInfoGatherer *SpecInfoGatherer) removeConcept(fileName string) {
	logger.ApiLog.Info("Concept removed: %s", fileName)
	delete(specInfoGatherer.fileToStepsMap, fileName)
	specInfoGatherer.createConceptsDictionary()
	specInfoGatherer.findAllStepsFromConcepts()
	specInfoGatherer.updateAllStepsList()
}

// Find all the steps defined in concepts. Look through the created concept dictionary to find concepts
func (specInfoGatherer *SpecInfoGatherer) findAllStepsFromConcepts() {
	allStepsInConcepts := make(map[string][]*parser.Step, 0)
	specInfoGatherer.createConceptsDictionary()
	for _, concept := range specInfoGatherer.getDictionary().ConceptsMap {
		stepsInConcept := make([]*parser.Step, 0)
		for _, step := range concept.ConceptStep.ConceptSteps {
			if !step.IsConcept {
				stepsInConcept = append(stepsInConcept, step)
			}
		}
		// Concept dictionary contains multiple entries for concepts that belong in the same file. So appending if the list of concepts for that file has been created.
		if _, ok := allStepsInConcepts[concept.FileName]; !ok {
			allStepsInConcepts[concept.FileName] = stepsInConcept
		} else {
			allStepsInConcepts[concept.FileName] = append(allStepsInConcepts[concept.FileName], stepsInConcept...)
		}
	}
	specInfoGatherer.updateCache(allStepsInConcepts)
}

func (specInfoGatherer *SpecInfoGatherer) createConceptInfos() []*gauge_messages.ConceptInfo {
	conceptInfos := make([]*gauge_messages.ConceptInfo, 0)
	for _, concept := range specInfoGatherer.getDictionary().ConceptsMap {
		stepValue := parser.CreateStepValue(concept.ConceptStep)
		conceptInfos = append(conceptInfos, &gauge_messages.ConceptInfo{StepValue: parser.ConvertToProtoStepValue(&stepValue), Filepath: proto.String(concept.FileName), LineNumber: proto.Int(concept.ConceptStep.LineNo)})
	}
	return conceptInfos
}

// Gets all steps list from the runner
func (specInfoGatherer *SpecInfoGatherer) getStepsFromRunner(killChannel chan bool) (*runner.TestRunner, error) {
	steps := make([]string, 0)
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		execLogger.CriticalError(err)
	}
	testRunner, connErr := runner.StartRunnerAndMakeConnection(manifest, execLogger.Current(), killChannel)
	if connErr == nil {
		steps = append(steps, requestForSteps(testRunner)...)
		logger.ApiLog.Debug("Steps got from runner: %v", steps)
	} else {
		logger.ApiLog.Error("Runner connection failed: %s", connErr)
	}
	specInfoGatherer.runnerStepValues = specInfoGatherer.convertToStepValues(steps)
	return testRunner, connErr
}

func requestForSteps(runner *runner.TestRunner) []string {
	message, err := conn.GetResponseForMessageWithTimeout(createGetStepNamesRequest(), runner.Connection, config.RunnerRequestTimeout())
	if err == nil {
		allStepsResponse := message.GetStepNamesResponse()
		return allStepsResponse.GetSteps()
	}
	logger.ApiLog.Error("Error response from runner on getStepNamesRequest: %s", err)
	return make([]string, 0)
}

func createGetStepNamesRequest() *gauge_messages.Message {
	return &gauge_messages.Message{MessageType: gauge_messages.Message_StepNamesRequest.Enum(), StepNamesRequest: &gauge_messages.StepNamesRequest{}}
}

func (specInfoGatherer *SpecInfoGatherer) convertToStepValues(steps []string) []*parser.StepValue {
	stepValues := make([]*parser.StepValue, 0)
	for _, step := range steps {
		stepValue, err := parser.ExtractStepValueAndParams(step, false)
		if err != nil {
			logger.ApiLog.Error("Failed to extract stepvalue for step - %s : %s", step, err)
			continue
		}
		stepValues = append(stepValues, stepValue)
	}
	return stepValues
}

func (specInfoGatherer *SpecInfoGatherer) findAvailableStepsInSpecs(specs []*parser.Specification) map[string][]*parser.Step {
	specStepsMap := make(map[string][]*parser.Step)
	for _, spec := range specs {
		stepsInSpec := make([]*parser.Step, 0)
		stepsInSpec = append(stepsInSpec, spec.Contexts...)
		for _, scenario := range spec.Scenarios {
			stepsInSpec = append(stepsInSpec, scenario.Steps...)
		}
		specStepsMap[spec.FileName] = stepsInSpec
	}
	return specStepsMap
}

func (specInfoGatherer *SpecInfoGatherer) addStepsToAvailableSteps(newSpecStepsMap map[string][]*parser.Step) {
	specInfoGatherer.updateCache(newSpecStepsMap)
	specInfoGatherer.updateAllStepsList()
}

func (specInfoGatherer *SpecInfoGatherer) updateAllStepsList() {
	specInfoGatherer.availableStepsMap = make(map[string]*parser.StepValue, 0)
	specInfoGatherer.addStepValuesToAvailableSteps(specInfoGatherer.runnerStepValues)
	for _, steps := range specInfoGatherer.fileToStepsMap {
		for _, step := range steps {
			if step.IsConcept {
				continue
			}
			stepValue := parser.CreateStepValue(step)
			if _, ok := specInfoGatherer.availableStepsMap[stepValue.StepValue]; !ok {
				specInfoGatherer.availableStepsMap[stepValue.StepValue] = &stepValue
			}
		}
	}
}

func (specInfoGatherer *SpecInfoGatherer) updateCache(newSpecStepsMap map[string][]*parser.Step) {
	if specInfoGatherer.fileToStepsMap == nil {
		specInfoGatherer.fileToStepsMap = make(map[string][]*parser.Step, 0)
	}
	for fileName, specSteps := range newSpecStepsMap {
		specInfoGatherer.fileToStepsMap[fileName] = specSteps
	}
}

func (specInfoGatherer *SpecInfoGatherer) addStepValuesToAvailableSteps(stepValues []*parser.StepValue) {
	for _, stepValue := range stepValues {
		specInfoGatherer.addToAvailableSteps(stepValue)
	}
}

func (specInfoGatherer *SpecInfoGatherer) addToAvailableSteps(stepValue *parser.StepValue) {
	if _, ok := specInfoGatherer.availableStepsMap[stepValue.StepValue]; !ok {
		specInfoGatherer.availableStepsMap[stepValue.StepValue] = stepValue
	}
}

func (specInfoGatherer *SpecInfoGatherer) GetAvailableSteps() []*parser.StepValue {
	steps := make([]*parser.StepValue, 0)
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

func (specInfoGatherer *SpecInfoGatherer) getDictionary() *parser.ConceptDictionary {
	if specInfoGatherer.conceptDictionary == nil {
		specInfoGatherer.conceptDictionary = parser.NewConceptDictionary()
	}
	return specInfoGatherer.conceptDictionary
}

func (specInfoGatherer *SpecInfoGatherer) GetConceptInfos() []*gauge_messages.ConceptInfo {
	return specInfoGatherer.createConceptInfos()
}
