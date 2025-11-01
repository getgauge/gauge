/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package infoGatherer

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
)

// SpecInfoGatherer contains the caches for specs, concepts, and steps
type SpecInfoGatherer struct {
	waitGroup         sync.WaitGroup
	conceptDictionary *gauge.ConceptDictionary
	specsCache        specsCache
	conceptsCache     conceptCache
	stepsCache        stepsCache
	paramsCache       paramsCache
	tagsCache         tagsCache
	SpecDirs          []string
}

type conceptCache struct {
	mutex    sync.RWMutex
	concepts map[string][]*gauge.Concept
}

type stepsCache struct {
	mutex sync.RWMutex
	steps map[string][]*gauge.Step
}

type specsCache struct {
	mutex       sync.RWMutex
	specDetails map[string]*SpecDetail
}

type paramsCache struct {
	mutex         sync.RWMutex
	staticParams  map[string]map[string]gauge.StepArg
	dynamicParams map[string]map[string]gauge.StepArg
}

type tagsCache struct {
	mutex sync.RWMutex
	tags  map[string][]string
}

type SpecDetail struct {
	Spec *gauge.Specification
	Errs []parser.ParseError
}

func (d *SpecDetail) HasSpec() bool {
	return d.Spec != nil && d.Spec.Heading != nil
}

func NewSpecInfoGatherer(conceptDictionary *gauge.ConceptDictionary) *SpecInfoGatherer {
	return &SpecInfoGatherer{conceptDictionary: conceptDictionary, conceptsCache: conceptCache{concepts: make(map[string][]*gauge.Concept)}}
}

// Init initializes all the SpecInfoGatherer caches
func (s *SpecInfoGatherer) Init() {
	// Concepts parsed first because we need to create a concept dictionary that spec parsing can use
	s.initConceptsCache()
	s.initSpecsCache()
	s.initStepsCache()
	s.initParamsCache()
	s.initTagsCache()

	go s.watchForFileChanges()
	s.waitGroup.Wait()
}
func (s *SpecInfoGatherer) initTagsCache() {
	s.tagsCache.mutex.Lock()
	defer s.tagsCache.mutex.Unlock()
	s.specsCache.mutex.Lock()
	s.tagsCache.tags = make(map[string][]string)
	for file, specDetail := range s.specsCache.specDetails {
		s.updateTagsCacheFromSpecs(file, specDetail)
	}
	defer s.specsCache.mutex.Unlock()
}

func (s *SpecInfoGatherer) initParamsCache() {
	s.paramsCache.mutex.Lock()
	defer s.paramsCache.mutex.Unlock()
	s.specsCache.mutex.Lock()
	s.paramsCache.staticParams = make(map[string]map[string]gauge.StepArg)
	s.paramsCache.dynamicParams = make(map[string]map[string]gauge.StepArg)
	for file, specDetail := range s.specsCache.specDetails {
		s.updateParamCacheFromSpecs(file, specDetail)
	}
	s.specsCache.mutex.Unlock()
	s.conceptsCache.mutex.Lock()
	for file, concepts := range s.conceptsCache.concepts {
		s.updateParamsCacheFromConcepts(file, concepts)
	}
	s.conceptsCache.mutex.Unlock()
}

func (s *SpecInfoGatherer) initSpecsCache() {
	details := s.getParsedSpecs(getSpecFiles(s.SpecDirs))

	s.specsCache.mutex.Lock()
	defer s.specsCache.mutex.Unlock()

	s.specsCache.specDetails = make(map[string]*SpecDetail)

	logger.Infof(false, "Initializing specs cache with %d specs", len(details))
	for _, d := range details {
		logger.Debugf(false, "Adding specs from %s", d.Spec.FileName)
		s.addToSpecsCache(d.Spec.FileName, d)
	}
}

func getSpecFiles(specs []string) []string {
	var specFiles []string
	for _, dir := range specs {
		specFiles = append(specFiles, util.FindSpecFilesIn(dir)...)
	}
	return specFiles
}

func (s *SpecInfoGatherer) initConceptsCache() {
	s.conceptsCache.mutex.Lock()
	defer s.conceptsCache.mutex.Unlock()

	parsedConcepts := s.getParsedConcepts()
	s.conceptsCache.concepts = make(map[string][]*gauge.Concept)
	logger.Infof(false, "Initializing concepts cache with %d concepts", len(parsedConcepts))
	for _, concept := range parsedConcepts {
		logger.Debugf(false, "Adding concepts from %s", concept.FileName)
		s.addToConceptsCache(concept.FileName, concept)
	}
}

func (s *SpecInfoGatherer) initStepsCache() {
	s.stepsCache.mutex.Lock()
	defer s.stepsCache.mutex.Unlock()

	s.stepsCache.steps = make(map[string][]*gauge.Step)
	stepsFromSpecsMap := s.getStepsFromCachedSpecs()
	stepsFromConceptsMap := s.getStepsFromCachedConcepts()

	for filename, steps := range stepsFromConceptsMap {
		s.addToStepsCache(filename, steps)
	}
	for filename, steps := range stepsFromSpecsMap {
		s.addToStepsCache(filename, steps)
	}
	logger.Infof(false, "Initializing steps cache with %d steps", len(stepsFromSpecsMap)+len(stepsFromConceptsMap))
}

func (s *SpecInfoGatherer) updateParamsCacheFromConcepts(file string, concepts []*gauge.Concept) {
	s.paramsCache.staticParams[file] = make(map[string]gauge.StepArg)
	s.paramsCache.dynamicParams[file] = make(map[string]gauge.StepArg)
	for _, concept := range concepts {
		s.addParamsFromSteps([]*gauge.Step{concept.ConceptStep}, file)
		s.addParamsFromSteps(concept.ConceptStep.ConceptSteps, file)
	}
}

func (s *SpecInfoGatherer) updateParamCacheFromSpecs(file string, specDetail *SpecDetail) {
	s.paramsCache.staticParams[file] = make(map[string]gauge.StepArg)
	s.paramsCache.dynamicParams[file] = make(map[string]gauge.StepArg)
	s.addParamsFromSteps(specDetail.Spec.Contexts, file)
	for _, sce := range specDetail.Spec.Scenarios {
		s.addParamsFromSteps(sce.Steps, file)
	}
	s.addParamsFromSteps(specDetail.Spec.TearDownSteps, file)
	if specDetail.Spec.DataTable.IsInitialized() {
		for _, header := range specDetail.Spec.DataTable.Table.Headers {
			s.paramsCache.dynamicParams[file][header] = gauge.StepArg{Value: header, ArgType: gauge.Dynamic}
		}
	}
}

func (s *SpecInfoGatherer) addParamsFromSteps(steps []*gauge.Step, file string) {
	for _, step := range steps {
		for _, arg := range step.Args {
			if arg.ArgType == gauge.Static {
				s.paramsCache.staticParams[file][arg.ArgValue()] = *arg
			} else {
				s.paramsCache.dynamicParams[file][arg.ArgValue()] = *arg
			}
		}
	}
}

func (s *SpecInfoGatherer) updateTagsCacheFromSpecs(file string, specDetail *SpecDetail) {
	if specDetail.Spec.Tags != nil {
		s.tagsCache.tags[file] = specDetail.Spec.Tags.Values()
	}
	for _, sce := range specDetail.Spec.Scenarios {
		if sce.Tags != nil {
			s.tagsCache.tags[file] = append(s.tagsCache.tags[file], sce.Tags.Values()...)
		}
	}
}

func removeDuplicateTags(tags []string) []string {
	encountered := map[string]bool{}
	result := []string{}
	for i := range tags {
		if !encountered[tags[i]] {
			encountered[tags[i]] = true
			result = append(result, tags[i])
		}
	}
	return result
}

func (s *SpecInfoGatherer) addToSpecsCache(key string, value *SpecDetail) {
	if s.specsCache.specDetails == nil {
		return
	}
	s.specsCache.specDetails[key] = value
}

func (s *SpecInfoGatherer) addToConceptsCache(key string, value *gauge.Concept) {
	if s.conceptsCache.concepts == nil {
		return
	}
	if s.conceptsCache.concepts[key] == nil {
		s.conceptsCache.concepts[key] = make([]*gauge.Concept, 0)
	}
	s.conceptsCache.concepts[key] = append(s.conceptsCache.concepts[key], value)
}

func (s *SpecInfoGatherer) deleteFromConceptDictionary(file string) {
	for _, c := range s.conceptsCache.concepts[file] {
		if file == s.conceptDictionary.ConceptsMap[c.ConceptStep.Value].FileName {
			s.conceptDictionary.Remove(c.ConceptStep.Value)
		}
	}
}

func (s *SpecInfoGatherer) addToStepsCache(fileName string, allSteps []*gauge.Step) {
	if s.stepsCache.steps == nil {
		return
	}
	s.stepsCache.steps[fileName] = allSteps
}

func (s *SpecInfoGatherer) getParsedSpecs(specFiles []string) []*SpecDetail {
	if s.conceptDictionary == nil {
		s.conceptDictionary = gauge.NewConceptDictionary()
	}
	parsedSpecs, parseResults := parser.ParseSpecFiles(specFiles, s.conceptDictionary, gauge.NewBuildErrors())
	specs := make(map[string]*SpecDetail)

	for _, spec := range parsedSpecs {
		specs[spec.FileName] = &SpecDetail{Spec: spec}
	}
	for _, v := range parseResults {
		_, ok := specs[v.FileName]
		if !ok {
			specs[v.FileName] = &SpecDetail{Spec: &gauge.Specification{FileName: v.FileName}, Errs: v.ParseErrors}
		}
	}
	details := make([]*SpecDetail, 0)
	for _, d := range specs {
		details = append(details, d)
	}
	return details
}

func (s *SpecInfoGatherer) getParsedConcepts() map[string]*gauge.Concept {
	var result *parser.ParseResult
	var err error
	s.conceptDictionary, result, err = parser.CreateConceptsDictionary()
	if err != nil {
		logger.Fatalf(true, "Unable to parse concepts : %s", err.Error())
	}
	handleParseFailures([]*parser.ParseResult{result})
	return s.conceptDictionary.ConceptsMap
}

func (s *SpecInfoGatherer) getStepsFromCachedSpecs() map[string][]*gauge.Step {
	s.specsCache.mutex.RLock()
	defer s.specsCache.mutex.RUnlock()

	var stepsFromSpecsMap = make(map[string][]*gauge.Step)
	for _, detail := range s.specsCache.specDetails {
		stepsFromSpecsMap[detail.Spec.FileName] = append(stepsFromSpecsMap[detail.Spec.FileName], getStepsFromSpec(detail.Spec)...)
	}
	return stepsFromSpecsMap
}

func (s *SpecInfoGatherer) getStepsFromCachedConcepts() map[string][]*gauge.Step {
	var stepsFromConceptMap = make(map[string][]*gauge.Step)
	s.conceptsCache.mutex.RLock()
	defer s.conceptsCache.mutex.RUnlock()
	for _, conceptList := range s.conceptsCache.concepts {
		for _, concept := range conceptList {
			stepsFromConceptMap[concept.FileName] = append(stepsFromConceptMap[concept.FileName], getStepsFromConcept(concept)...)
		}
	}
	return stepsFromConceptMap
}

func (s *SpecInfoGatherer) OnSpecFileModify(file string) {
	logger.Debugf(false, "Spec file added / modified: %s", file)

	details := s.getParsedSpecs([]string{file})
	s.specsCache.mutex.Lock()
	s.addToSpecsCache(file, details[0])
	s.specsCache.mutex.Unlock()

	var steps []*gauge.Step
	steps = append(steps, getStepsFromSpec(details[0].Spec)...)
	s.stepsCache.mutex.Lock()
	s.addToStepsCache(file, steps)
	s.stepsCache.mutex.Unlock()

	s.paramsCache.mutex.Lock()
	s.updateParamCacheFromSpecs(file, details[0])
	s.paramsCache.mutex.Unlock()

	s.tagsCache.mutex.Lock()
	s.updateTagsCacheFromSpecs(file, details[0])
	s.tagsCache.mutex.Unlock()
}

func (s *SpecInfoGatherer) OnConceptFileModify(file string) {
	s.conceptsCache.mutex.Lock()
	defer s.conceptsCache.mutex.Unlock()

	logger.Debugf(false, "Concept file added / modified: %s", file)
	s.deleteFromConceptDictionary(file)
	concepts, parseErrors, err := parser.AddConcepts([]string{file}, s.conceptDictionary)
	if err != nil {
		logger.Fatalf(true, "Unable to update concepts : %s", err.Error())
	}
	if len(parseErrors) > 0 {
		res := &parser.ParseResult{}
		res.ParseErrors = append(res.ParseErrors, parseErrors...)
		res.Ok = false
		handleParseFailures([]*parser.ParseResult{res})
	}
	s.conceptsCache.concepts[file] = make([]*gauge.Concept, 0)
	var stepsFromConcept []*gauge.Step
	for _, concept := range concepts {
		c := gauge.Concept{ConceptStep: concept, FileName: file}
		s.addToConceptsCache(file, &c)
		stepsFromConcept = append(stepsFromConcept, getStepsFromConcept(&c)...)
	}
	s.addToStepsCache(file, stepsFromConcept)
	s.paramsCache.mutex.Lock()
	defer s.paramsCache.mutex.Unlock()
	s.updateParamsCacheFromConcepts(file, s.conceptsCache.concepts[file])
}

func (s *SpecInfoGatherer) onSpecFileRemove(file string) {
	logger.Debugf(false, "Spec file removed: %s", file)
	s.specsCache.mutex.Lock()
	defer s.specsCache.mutex.Unlock()
	delete(s.specsCache.specDetails, file)
	s.removeStepsFromCache(file)
}
func (s *SpecInfoGatherer) removeStepsFromCache(fileName string) {
	s.stepsCache.mutex.Lock()
	defer s.stepsCache.mutex.Unlock()
	delete(s.stepsCache.steps, fileName)
}

func (s *SpecInfoGatherer) onConceptFileRemove(file string) {
	logger.Debugf(false, "Concept file removed: %s", file)
	s.conceptsCache.mutex.Lock()
	defer s.conceptsCache.mutex.Unlock()
	s.deleteFromConceptDictionary(file)
	delete(s.conceptsCache.concepts, file)
	s.removeStepsFromCache(file)
}

func (s *SpecInfoGatherer) onFileAdd(watcher *fsnotify.Watcher, file string) {
	if util.IsDir(file) {
		addDirToFileWatcher(watcher, file)
	}
	s.onFileModify(watcher, file)
}

func (s *SpecInfoGatherer) onFileModify(watcher *fsnotify.Watcher, file string) {
	if util.IsSpec(file) {
		s.OnSpecFileModify(file)
	} else if util.IsConcept(file) {
		s.OnConceptFileModify(file)
	}
}

func (s *SpecInfoGatherer) onFileRemove(watcher *fsnotify.Watcher, file string) {
	if util.IsSpec(file) {
		s.onSpecFileRemove(file)
	} else if util.IsConcept(file) {
		s.onConceptFileRemove(file)
	} else {
		removeWatcherOn(watcher, file)
	}
}

func (s *SpecInfoGatherer) onFileRename(watcher *fsnotify.Watcher, file string) {
	s.onFileRemove(watcher, file)
}

func (s *SpecInfoGatherer) handleEvent(event fsnotify.Event, watcher *fsnotify.Watcher) {
	s.waitGroup.Wait()

	file, err := filepath.Abs(event.Name)
	if err != nil {
		logger.Errorf(false, "Failed to get abs file path for %s: %s", event.Name, err)
		return
	}
	if util.IsSpec(file) || util.IsConcept(file) || util.IsDir(file) {
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

func (s *SpecInfoGatherer) watchForFileChanges() {
	s.waitGroup.Add(1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Errorf(false, "Error creating fileWatcher: %s", err)
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				s.handleEvent(event, watcher)
			case err := <-watcher.Errors:
				logger.Errorf(false, "Error event while watching specs %s", err)
			}
		}
	}()

	var allDirsToWatch []string
	var specDir string

	for _, dir := range s.SpecDirs {
		specDir = filepath.Join(config.ProjectRoot, dir)
		allDirsToWatch = append(allDirsToWatch, specDir)
		allDirsToWatch = append(allDirsToWatch, util.FindAllNestedDirs(specDir)...)
	}

	for _, dir := range allDirsToWatch {
		addDirToFileWatcher(watcher, dir)
	}
	s.waitGroup.Done()
	<-done
}

// GetAvailableSpecs returns the list of all the specs in the gauge project
func (s *SpecInfoGatherer) GetAvailableSpecDetails(specs []string) []*SpecDetail {
	if len(specs) < 1 {
		specs = util.GetSpecDirs()
	}
	specFiles := getSpecFiles(specs)
	s.specsCache.mutex.RLock()
	defer s.specsCache.mutex.RUnlock()
	var details []*SpecDetail
	for _, f := range specFiles {
		if d, ok := s.specsCache.specDetails[f]; ok {
			details = append(details, d)
		}
	}
	return details
}

func (s *SpecInfoGatherer) GetSpecDirs() []string {
	return s.SpecDirs
}

// Steps returns the list of all the steps in the gauge project. Duplicate steps are filtered
func (s *SpecInfoGatherer) Steps(filterConcepts bool) []*gauge.Step {
	s.stepsCache.mutex.RLock()
	defer s.stepsCache.mutex.RUnlock()
	filteredSteps := make(map[string]*gauge.Step)
	for _, steps := range s.stepsCache.steps {
		for _, s := range steps {
			if !filterConcepts || !s.IsConcept {
				filteredSteps[s.Value] = s
			}
		}
	}
	var steps []*gauge.Step
	for _, sv := range filteredSteps {
		steps = append(steps, sv)
	}
	return steps
}

// Steps returns the list of all the steps in the gauge project including duplicate steps
func (s *SpecInfoGatherer) AllSteps(filterConcepts bool) []*gauge.Step {
	s.stepsCache.mutex.RLock()
	defer s.stepsCache.mutex.RUnlock()
	var allSteps []*gauge.Step
	for _, steps := range s.stepsCache.steps {
		if filterConcepts {
			for _, s := range steps {
				if !s.IsConcept {
					allSteps = append(allSteps, s)
				}
			}
		} else {
			allSteps = append(allSteps, steps...)
		}
	}
	return allSteps
}

// Steps returns the list of all the steps in the gauge project
func (s *SpecInfoGatherer) Params(filePath string, argType gauge.ArgType) []gauge.StepArg {
	s.paramsCache.mutex.RLock()
	defer s.paramsCache.mutex.RUnlock()
	var params []gauge.StepArg
	if argType == gauge.Static {
		for _, param := range s.paramsCache.staticParams[filePath] {
			params = append(params, param)
		}
	} else {
		for _, param := range s.paramsCache.dynamicParams[filePath] {
			params = append(params, param)
		}
	}
	return params
}

// Concepts returns an array containing information about all the concepts present in the Gauge project
func (s *SpecInfoGatherer) Concepts() []*gauge_messages.ConceptInfo {
	var conceptInfos []*gauge_messages.ConceptInfo
	s.conceptsCache.mutex.RLock()
	defer s.conceptsCache.mutex.RUnlock()
	for _, conceptList := range s.conceptsCache.concepts {
		for _, concept := range conceptList {
			stepValue := parser.CreateStepValue(concept.ConceptStep)
			conceptInfos = append(conceptInfos, &gauge_messages.ConceptInfo{StepValue: gauge.ConvertToProtoStepValue(&stepValue), Filepath: concept.FileName, LineNumber: int32(concept.ConceptStep.LineNo)})
		}
	}
	return conceptInfos
}

func (s *SpecInfoGatherer) Tags() []string {
	s.tagsCache.mutex.RLock()
	defer s.tagsCache.mutex.RUnlock()
	var allTags []string
	for _, tags := range s.tagsCache.tags {
		allTags = append(allTags, tags...)
	}
	return removeDuplicateTags(allTags)
}

// SearchConceptDictionary searches for a concept in concept dictionary
func (s *SpecInfoGatherer) SearchConceptDictionary(stepValue string) *gauge.Concept {
	return s.conceptDictionary.Search(stepValue)
}

func getStepsFromSpec(spec *gauge.Specification) []*gauge.Step {
	steps := spec.Contexts
	for _, scenario := range spec.Scenarios {
		steps = append(steps, scenario.Steps...)
	}
	steps = append(steps, spec.TearDownSteps...)
	return steps
}

func getStepsFromConcept(concept *gauge.Concept) []*gauge.Step {
	return concept.ConceptStep.ConceptSteps
}

func handleParseFailures(parseResults []*parser.ParseResult) {
	for _, result := range parseResults {
		if !result.Ok {
			logger.Errorf(false, "Parse failure: %s", result.Errors())
		}
	}
}

func addDirToFileWatcher(watcher *fsnotify.Watcher, dir string) {
	err := watcher.Add(dir)
	if err != nil {
		logger.Errorf(false, "Unable to add directory %v to file watcher: %s", dir, err.Error())
	} else {
		logger.Debugf(false, "Watching directory: %s", dir)
		files, _ := os.ReadDir(dir)
		logger.Debugf(false, "Found %d files", len(files))
	}
}

func removeWatcherOn(watcher *fsnotify.Watcher, path string) {
	logger.Debugf(false, "Removing watcher on : %s", path)
	err := watcher.Remove(path)
	if err != nil {
		logger.Errorf(false, "Unable to remove watcher on: %s. %s", path, err.Error())
	}
}
