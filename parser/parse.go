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

package parser

import (
	"strings"

	"regexp"
	"strconv"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/order"
	"github.com/getgauge/gauge/util"
)

// TODO: Use single channel instead of one for spec and another for result, so that mapping is consistent
func ParseSpecFiles(specFiles []string, conceptDictionary *gauge.ConceptDictionary, buildErrors *gauge.BuildErrors) ([]*gauge.Specification, []*ParseResult) {
	parseResultsChan := make(chan *ParseResult, len(specFiles))
	specsChan := make(chan *gauge.Specification, len(specFiles))
	var parseResults []*ParseResult
	var specs []*gauge.Specification

	for _, specFile := range specFiles {
		go parseSpec(specFile, conceptDictionary, specsChan, parseResultsChan)
	}
	for range specFiles {
		parseRes := <-parseResultsChan
		spec := <-specsChan
		if spec != nil {
			specs = append(specs, spec)
			var parseErrs []error
			for _, e := range parseRes.CriticalErrors {
				parseErrs = append(parseErrs, e)
			}
			for _, e := range parseRes.ParseErrors {
				parseErrs = append(parseErrs, e)
			}
			if len(parseErrs) != 0 {
				buildErrors.SpecErrs[spec] = parseErrs
			}
		}
		parseResults = append(parseResults, parseRes)
	}
	return specs, parseResults
}

func ParseSpecs(args []string, conceptsDictionary *gauge.ConceptDictionary, buildErrors *gauge.BuildErrors) ([]*gauge.Specification, bool) {
	specs, failed := parseSpecsInDirs(conceptsDictionary, args, buildErrors)
	specsToExecute := order.Sort(filter.FilterSpecs(specs))
	return specsToExecute, failed
}

func ParseConcepts() (*gauge.ConceptDictionary, *ParseResult) {
	conceptsDictionary, conceptParseResult := CreateConceptsDictionary()
	HandleParseResult(conceptParseResult)
	return conceptsDictionary, conceptParseResult
}

func parseSpec(specFile string, conceptDictionary *gauge.ConceptDictionary, specChannel chan *gauge.Specification, parseResultChan chan *ParseResult) {
	specFileContent, err := common.ReadFileContents(specFile)
	if err != nil {
		specChannel <- nil
		parseResultChan <- &ParseResult{ParseErrors: []ParseError{ParseError{FileName: specFile, Message: err.Error()}}, Ok: false}
		return
	}
	spec, parseResult := new(SpecParser).Parse(specFileContent, conceptDictionary, specFile)
	specChannel <- spec
	parseResultChan <- parseResult
}

type specFile struct {
	filePath string
	indices  []int
}

// parseSpecsInDirs parses all the specs in list of dirs given.
// It also de-duplicates all specs passed through `specDirs` before parsing specs.
func parseSpecsInDirs(conceptDictionary *gauge.ConceptDictionary, specDirs []string, buildErrors *gauge.BuildErrors) ([]*gauge.Specification, bool) {
	passed := true
	givenSpecs, specFiles := getAllSpecFiles(specDirs)
	var specs []*gauge.Specification
	var specParseResults []*ParseResult
	allSpecs := make([]*gauge.Specification, len(specFiles))
	specs, specParseResults = ParseSpecFiles(givenSpecs, conceptDictionary, buildErrors)
	passed = !HandleParseResult(specParseResults...) && passed
	for _, spec := range specs {
		i, _ := getIndexFor(specFiles, spec.FileName)
		specFile := specFiles[i]
		if len(specFile.indices) > 0 {
			spec.Filter(filter.NewScenarioFilterBasedOnSpan(specFile.indices))
		}
		allSpecs[i] = spec
	}
	return allSpecs, !passed
}

func getAllSpecFiles(specDirs []string) (givenSpecs []string, specFiles []*specFile) {
	for _, specSource := range specDirs {
		if isIndexedSpec(specSource) {
			var specName string
			specName, index := getIndexedSpecName(specSource)
			files := util.GetSpecFiles(specName)
			if len(files) < 1 {
				continue
			}
			specificationFile, created := addSpecFile(&specFiles, files[0])
			if created || len(specificationFile.indices) > 0 {
				specificationFile.indices = append(specificationFile.indices, index)
			}
			givenSpecs = append(givenSpecs, files[0])
		} else {
			files := util.GetSpecFiles(specSource)
			for _, file := range files {
				specificationFile, _ := addSpecFile(&specFiles, file)
				specificationFile.indices = specificationFile.indices[0:0]
			}
			givenSpecs = append(givenSpecs, files...)
		}
	}
	return
}

func addSpecFile(specFiles *[]*specFile, file string) (*specFile, bool) {
	i, exists := getIndexFor(*specFiles, file)
	if !exists {
		specificationFile := &specFile{filePath: file}
		*specFiles = append(*specFiles, specificationFile)
		return specificationFile, true
	}
	return (*specFiles)[i], false
}

func getIndexFor(files []*specFile, file string) (int, bool) {
	for index, f := range files {
		if f.filePath == file {
			return index, true
		}
	}
	return -1, false
}

func isIndexedSpec(specSource string) bool {
	return getIndex(specSource) != 0
}

func getIndexedSpecName(indexedSpec string) (string, int) {
	index := getIndex(indexedSpec)
	specName := indexedSpec[:index]
	scenarioNum := indexedSpec[index+1:]
	scenarioNumber, _ := strconv.Atoi(scenarioNum)
	return specName, scenarioNumber
}

func getIndex(specSource string) int {
	re, _ := regexp.Compile(":[0-9]+$")
	index := re.FindStringSubmatchIndex(specSource)
	if index != nil {
		return index[0]
	}
	return 0
}

func ExtractStepValueAndParams(stepText string, hasInlineTable bool) (*gauge.StepValue, error) {
	stepValueWithPlaceHolders, args, err := processStepText(stepText)
	if err != nil {
		return nil, err
	}

	extractedStepValue, _ := extractStepValueAndParameterTypes(stepValueWithPlaceHolders)
	if hasInlineTable {
		extractedStepValue += " " + gauge.ParameterPlaceholder
		args = append(args, string(gauge.TableArg))
	}
	parameterizedStepValue := getParameterizeStepValue(extractedStepValue, args)

	return &gauge.StepValue{args, extractedStepValue, parameterizedStepValue}, nil

}

func CreateStepValue(step *gauge.Step) gauge.StepValue {
	stepValue := gauge.StepValue{StepValue: step.Value}
	args := make([]string, 0)
	for _, arg := range step.Args {
		args = append(args, arg.ArgValue())
	}
	stepValue.Args = args
	stepValue.ParameterizedStepValue = getParameterizeStepValue(stepValue.StepValue, args)
	return stepValue
}

func getParameterizeStepValue(stepValue string, params []string) string {
	for _, param := range params {
		stepValue = strings.Replace(stepValue, gauge.ParameterPlaceholder, "<"+param+">", 1)
	}
	return stepValue
}

func HandleParseResult(results ...*ParseResult) bool {
	var failed = false
	for _, result := range results {
		if !result.Ok {
			for _, err := range result.Errors() {
				logger.Errorf(err)
			}
			failed = true
		}
		if result.Warnings != nil {
			for _, warning := range result.Warnings {
				logger.Warningf("[ParseWarning] %s", warning)
			}
		}
	}
	return failed
}
