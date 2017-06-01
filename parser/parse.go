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

func addSpecsToMap(specs []*gauge.Specification, specsMap map[string]*gauge.Specification) {
	for _, spec := range specs {
		if _, ok := specsMap[spec.FileName]; ok {
			specsMap[spec.FileName].Scenarios = append(specsMap[spec.FileName].Scenarios, spec.Scenarios...)
			for _, sce := range spec.Scenarios {
				specsMap[spec.FileName].Items = append(specsMap[spec.FileName].Items, sce)
			}
			continue
		}
		specsMap[spec.FileName] = spec
	}
}

// parseSpecsInDirs parses all the specs in list of dirs given.
// It also merges the scenarios belonging to same spec which are passed as different arguments in `specDirs`
func parseSpecsInDirs(conceptDictionary *gauge.ConceptDictionary, specDirs []string, buildErrors *gauge.BuildErrors) ([]*gauge.Specification, bool) {
	specsMap := make(map[string]*gauge.Specification)
	passed := true
	var givenSpecs []*gauge.Specification
	for _, specSource := range specDirs {
		var specs []*gauge.Specification
		var specParseResults []*ParseResult
		if isIndexedSpec(specSource) {
			specs, specParseResults = getSpecWithScenarioIndex(specSource, conceptDictionary, buildErrors)
		} else {
			specs, specParseResults = ParseSpecFiles(util.GetSpecFiles(specSource), conceptDictionary, buildErrors)
		}
		passed = !HandleParseResult(specParseResults...) && passed
		givenSpecs = append(givenSpecs, specs...)
		addSpecsToMap(specs, specsMap)
	}
	var allSpecs []*gauge.Specification
	for _, spec := range givenSpecs {
		if _, ok := specsMap[spec.FileName]; ok {
			allSpecs = append(allSpecs, specsMap[spec.FileName])
			delete(specsMap, spec.FileName)
		}
	}
	return allSpecs, !passed
}

func getSpecWithScenarioIndex(specSource string, conceptDictionary *gauge.ConceptDictionary, buildErrors *gauge.BuildErrors) ([]*gauge.Specification, []*ParseResult) {
	specName, indexToFilter := getIndexedSpecName(specSource)
	parsedSpecs, parseResult := ParseSpecFiles(util.GetSpecFiles(specName), conceptDictionary, buildErrors)
	return filter.FilterSpecsItems(parsedSpecs, filter.NewScenarioFilterBasedOnSpan(indexToFilter)), parseResult
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
		switch arg.ArgType {
		case gauge.Static, gauge.Dynamic:
			args = append(args, arg.Value)
		case gauge.TableArg:
			args = append(args, "table")
		case gauge.SpecialString, gauge.SpecialTable:
			args = append(args, arg.Name)
		}
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
				logger.Warning("[ParseWarning] %s", warning)
			}
		}
	}
	return failed
}
