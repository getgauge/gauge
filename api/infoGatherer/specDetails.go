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
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/golang/protobuf/proto"
	fsnotify "gopkg.in/fsnotify.v1"
	"path/filepath"
	"sync"
)

type SpecInfoGatherer struct {
	mutex             sync.Mutex
	conceptDictionary *parser.ConceptDictionary
	specsCache        map[string][]*parser.Specification
	conceptsCache     map[string][]*parser.Concept
	stepsCache        map[string]*parser.StepValue
}

func (s *SpecInfoGatherer) MakeListOfAvailableSteps(runner *runner.TestRunner) {
	s.specsCache = make(map[string][]*parser.Specification, 0)
	s.conceptsCache = make(map[string][]*parser.Concept, 0)
	s.stepsCache = make(map[string]*parser.StepValue, 0)

	// Concepts parsed first because we need to create a concept dictionary that spec parsing can use
	s.initConceptsCache()
	s.initSpecsCache()
	s.initStepsCache(runner)

	go s.watchForFileChanges()
}

func (s *SpecInfoGatherer) initSpecsCache() {
	specFiles := util.FindSpecFilesIn(filepath.Join(config.ProjectRoot, common.SpecsDirectoryName))
	parsedSpecs := s.getParsedSpecs(specFiles)

	logger.ApiLog.Debug("Initializing specs cache with %d specs", len(parsedSpecs))
	for _, spec := range parsedSpecs {
		s.addToSpecsCache(spec.FileName, spec)
	}
}

func (s *SpecInfoGatherer) initConceptsCache() {
	parsedConcepts := s.getParsedConcepts()

	logger.ApiLog.Debug("Initializing concepts cache with %d concepts", len(parsedConcepts))
	for _, concept := range parsedConcepts {
		s.addToConceptsCache(concept.FileName, concept)
	}
}

func (s *SpecInfoGatherer) initStepsCache(runner *runner.TestRunner) {
	stepsFromSpecs := s.getStepsFromCachedSpecs()
	stepsFromConcepts := s.getStepsFromCachedConcepts()
	implementedSteps := s.getImplementedSteps(runner)

	allSteps := append(stepsFromSpecs, stepsFromConcepts...)
	allSteps = append(allSteps, implementedSteps...)

	logger.ApiLog.Debug("Initializing steps cache with %d steps", len(allSteps))
	s.addToStepsCache(allSteps)
}

func (s *SpecInfoGatherer) addToSpecsCache(key string, value *parser.Specification) {
	s.mutex.Lock()
	if s.specsCache[key] == nil {
		s.specsCache[key] = make([]*parser.Specification, 0)
	}
	s.specsCache[key] = append(s.specsCache[key], value)
	s.mutex.Unlock()
}

func (s *SpecInfoGatherer) addToConceptsCache(key string, value *parser.Concept) {
	s.mutex.Lock()
	if s.conceptsCache[key] == nil {
		s.conceptsCache[key] = make([]*parser.Concept, 0)
	}
	s.conceptsCache[key] = append(s.conceptsCache[key], value)
	s.mutex.Unlock()
}

func (s *SpecInfoGatherer) addToStepsCache(allSteps []*parser.StepValue) {
	s.mutex.Lock()
	for _, step := range allSteps {
		if _, ok := s.stepsCache[step.StepValue]; !ok {
			s.stepsCache[step.StepValue] = step
		}
	}
	s.mutex.Unlock()
}

func (s *SpecInfoGatherer) getParsedSpecs(specFiles []string) []*parser.Specification {
	if s.conceptDictionary == nil {
		s.conceptDictionary = parser.NewConceptDictionary()
	}
	parsedSpecs, parseResults := parser.ParseSpecFiles(specFiles, s.conceptDictionary)
	s.handleParseFailures(parseResults)
	return parsedSpecs
}

func (s *SpecInfoGatherer) getParsedConcepts() map[string]*parser.Concept {
	var result *parser.ParseResult
	s.conceptDictionary, result = parser.CreateConceptsDictionary(true)
	s.handleParseFailures([]*parser.ParseResult{result})
	return s.conceptDictionary.ConceptsMap
}

func (s *SpecInfoGatherer) getParsedStepValues(steps []string) []*parser.StepValue {
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

func (s *SpecInfoGatherer) getStepsFromCachedSpecs() []*parser.StepValue {
	stepValues := make([]*parser.StepValue, 0)
	s.mutex.Lock()
	for _, specList := range s.specsCache {
		for _, spec := range specList {
			stepValues = append(stepValues, s.getStepsFromSpec(spec)...)
		}
	}
	s.mutex.Unlock()
	return stepValues
}

func (s *SpecInfoGatherer) getStepsFromCachedConcepts() []*parser.StepValue {
	stepValues := make([]*parser.StepValue, 0)
	s.mutex.Lock()
	for _, conceptList := range s.conceptsCache {
		for _, concept := range conceptList {
			stepValues = append(stepValues, s.getStepsFromConcept(concept)...)
		}
	}
	s.mutex.Unlock()
	return stepValues
}

func (s *SpecInfoGatherer) getStepsFromSpec(spec *parser.Specification) []*parser.StepValue {
	stepValues := make([]*parser.StepValue, 0)
	for _, scenario := range spec.Scenarios {
		for _, step := range scenario.Steps {
			if !step.IsConcept {
				stepValue := parser.CreateStepValue(step)
				stepValues = append(stepValues, &stepValue)
			}
		}
	}
	return stepValues
}

func (s *SpecInfoGatherer) getStepsFromConcept(concept *parser.Concept) []*parser.StepValue {
	stepValues := make([]*parser.StepValue, 0)
	for _, step := range concept.ConceptStep.ConceptSteps {
		if !step.IsConcept {
			stepValue := parser.CreateStepValue(step)
			stepValues = append(stepValues, &stepValue)
		}
	}
	return stepValues
}

func (s *SpecInfoGatherer) getImplementedSteps(runner *runner.TestRunner) []*parser.StepValue {
	stepValues := make([]*parser.StepValue, 0)
	message, err := conn.GetResponseForMessageWithTimeout(createGetStepNamesRequest(), runner.Connection, config.RunnerRequestTimeout())
	if err != nil {
		logger.ApiLog.Error("Error response from runner on getStepNamesRequest: %s", err)
		return stepValues
	}

	allSteps := message.GetStepNamesResponse().GetSteps()
	return s.getParsedStepValues(allSteps)
}

func (s *SpecInfoGatherer) onSpecFileModify(file string) {
	logger.ApiLog.Info("Spec file added / modified: %s", file)
	parsedSpec := s.getParsedSpecs([]string{file})[0]
	s.addToSpecsCache(file, parsedSpec)

	stepsFromSpec := s.getStepsFromSpec(parsedSpec)
	s.addToStepsCache(stepsFromSpec)
}

func (s *SpecInfoGatherer) onConceptFileModify(file string) {
	logger.ApiLog.Info("Concept file added / modified: %s", file)
	conceptParser := new(parser.ConceptParser)
	concepts, parseResults := conceptParser.ParseFile(file)
	if parseResults != nil && parseResults.Error != nil {
		logger.ApiLog.Error("Error parsing concepts: ", parseResults.Error)
		return
	}

	for _, concept := range concepts {
		c := parser.Concept{concept, file}
		s.addToConceptsCache(file, &c)
		stepsFromConcept := s.getStepsFromConcept(&c)
		s.addToStepsCache(stepsFromConcept)
	}
}

func (s *SpecInfoGatherer) onSpecFileRemove(file string) {
	logger.ApiLog.Info("Spec file removed: %s", file)
	s.mutex.Lock()
	delete(s.specsCache, file)
	s.mutex.Unlock()
}

func (s *SpecInfoGatherer) onConceptFileRemove(file string) {
	logger.ApiLog.Info("Concept file removed: %s", file)
	s.mutex.Lock()
	delete(s.conceptsCache, file)
	s.mutex.Unlock()
}

func (s *SpecInfoGatherer) createConceptsDictionary() {
	var result *parser.ParseResult
	s.conceptDictionary, result = parser.CreateConceptsDictionary(true)
	s.handleParseFailures([]*parser.ParseResult{result})
}

func (s *SpecInfoGatherer) handleParseFailures(parseResults []*parser.ParseResult) {
	for _, result := range parseResults {
		if !result.Ok {
			logger.ApiLog.Error("Spec Parse failure: %s", result.Error())
		}
	}
}

func (s *SpecInfoGatherer) watchForFileChanges() {
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
				s.handleEvent(event, watcher)
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
		s.addDirToFileWatcher(watcher, dir)
	}
	<-done
}

func (s *SpecInfoGatherer) addDirToFileWatcher(watcher *fsnotify.Watcher, dir string) {
	err := watcher.Add(dir)
	if err != nil {
		logger.ApiLog.Error("Unable to add directory %v to file watcher: %s", dir, err)
	} else {
		logger.ApiLog.Info("Watching directory: %s", dir)
	}
}

func (s *SpecInfoGatherer) removeWatcherOn(watcher *fsnotify.Watcher, path string) {
	logger.ApiLog.Info("Removing watcher on : %s", path)
	watcher.Remove(path)
}

func (s *SpecInfoGatherer) handleEvent(event fsnotify.Event, watcher *fsnotify.Watcher) {
	file, err := filepath.Abs(event.Name)
	if err != nil {
		logger.ApiLog.Error("Failed to get abs file path for %s: %s", event.Name, err)
		return
	}
	if filepath.Ext(file) == ".spec" || filepath.Ext(file) == ".cpt" || filepath.Ext(file) == ".md" {
		switch event.Op {
		case fsnotify.Create:
			s.onFileAdd(watcher, file)
		case fsnotify.Write:
			s.onFileModify(watcher, file)
		case fsnotify.Rename:
			s.onFileRename(watcher, file)
		case fsnotify.Remove:
			s.onFileRemove(watcher, file)
		}
	}
}

func (s *SpecInfoGatherer) onFileAdd(watcher *fsnotify.Watcher, file string) {
	if util.IsDir(file) {
		s.addDirToFileWatcher(watcher, file)
	}
	s.onFileModify(watcher, file)
}

func (s *SpecInfoGatherer) onFileModify(watcher *fsnotify.Watcher, file string) {
	if util.IsSpec(file) {
		s.onSpecFileModify(file)
	} else if util.IsConcept(file) {
		s.onConceptFileModify(file)
	}
}

func (s *SpecInfoGatherer) onFileRemove(watcher *fsnotify.Watcher, file string) {
	if util.IsSpec(file) {
		s.onSpecFileRemove(file)
	} else if util.IsConcept(file) {
		s.onConceptFileRemove(file)
	} else {
		s.removeWatcherOn(watcher, file)
	}
}

func (s *SpecInfoGatherer) onFileRename(watcher *fsnotify.Watcher, file string) {
	s.onFileRemove(watcher, file)
}

func (s *SpecInfoGatherer) GetAvailableSpecs() []*parser.Specification {
	allSpecs := make([]*parser.Specification, 0)
	s.mutex.Lock()
	for _, specs := range s.specsCache {
		allSpecs = append(allSpecs, specs...)
	}
	s.mutex.Unlock()
	return allSpecs
}

func (s *SpecInfoGatherer) GetAvailableSteps() []*parser.StepValue {
	steps := make([]*parser.StepValue, 0)
	s.mutex.Lock()
	for _, stepValue := range s.stepsCache {
		steps = append(steps, stepValue)
	}
	s.mutex.Unlock()
	return steps
}

func (s *SpecInfoGatherer) GetConceptInfos() []*gauge_messages.ConceptInfo {
	conceptInfos := make([]*gauge_messages.ConceptInfo, 0)
	s.mutex.Lock()
	for _, conceptList := range s.conceptsCache {
		for _, concept := range conceptList {
			stepValue := parser.CreateStepValue(concept.ConceptStep)
			conceptInfos = append(conceptInfos, &gauge_messages.ConceptInfo{StepValue: parser.ConvertToProtoStepValue(&stepValue), Filepath: proto.String(concept.FileName), LineNumber: proto.Int(concept.ConceptStep.LineNo)})
		}
	}
	s.mutex.Unlock()
	return conceptInfos
}

func createGetStepNamesRequest() *gauge_messages.Message {
	return &gauge_messages.Message{MessageType: gauge_messages.Message_StepNamesRequest.Enum(), StepNamesRequest: &gauge_messages.StepNamesRequest{}}
}
