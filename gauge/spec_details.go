package main

import (
	"github.com/getgauge/common"
)

type specInfoGatherer struct {
	availableSpecs    []*specification
	availableStepsMap map[string]*stepValue
}

func (specInfoGatherer *specInfoGatherer) makeListOfAvailableSteps(runner *testRunner) {
	specInfoGatherer.availableStepsMap = make(map[string]*stepValue, 0)
	stepsFromRunner := specInfoGatherer.getStepsFromRunner(runner)
	specInfoGatherer.addStepValuesToAvailableSteps(stepsFromRunner)
	specFiles := findSpecsFilesIn(common.SpecsDirectoryName)
	dictionary, _ := createConceptsDictionary(true)
	specInfoGatherer.availableSpecs = specInfoGatherer.parseSpecFiles(specFiles, dictionary)
	specInfoGatherer.findAvailableStepsInSpecs(specInfoGatherer.availableSpecs)
}

func (specInfoGatherer *specInfoGatherer) getStepsFromRunner(runner *testRunner) []string {
	steps := make([]string, 0)
	if runner == nil {
		runner, connErr := startRunnerAndMakeConnection(getProjectManifest())
		if connErr == nil {
			steps = append(steps, requestForSteps(runner)...)
			runner.kill()
		}
	} else {
		steps = append(steps, requestForSteps(runner)...)
	}
	return steps
}

func (specInfoGatherer *specInfoGatherer) parseSpecFiles(specFiles []string, dictionary *conceptDictionary) []*specification {
	specs := make([]*specification, 0)
	for _, file := range specFiles {
		specContent, err := common.ReadFileContents(file)
		if err != nil {
			continue
		}
		parser := new(specParser)
		specification, result := parser.parse(specContent, dictionary)

		if result.ok {
			specs = append(specs, specification)
		}
	}
	return specs
}

func (specInfoGatherer *specInfoGatherer) findAvailableStepsInSpecs(specs []*specification) {
	for _, spec := range specs {
		specInfoGatherer.addStepsToAvailableSteps(spec.contexts)
		for _, scenario := range spec.scenarios {
			specInfoGatherer.addStepsToAvailableSteps(scenario.steps)
		}
	}
}

func (specInfoGatherer *specInfoGatherer) addStepsToAvailableSteps(steps []*step) {
	for _, step := range steps {
		stepValue, err := extractStepValueAndParams(step.lineText, step.hasInlineTable)
		if err == nil {
			if _, ok := specInfoGatherer.availableStepsMap[stepValue.stepValue]; !ok {
				specInfoGatherer.availableStepsMap[stepValue.stepValue] = stepValue
			}
		}
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
	steps := make([]*stepValue, 0)
	for _, stepValue := range specInfoGatherer.availableStepsMap {
		steps = append(steps, stepValue)
	}
	return steps
}
