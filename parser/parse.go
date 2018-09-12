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

/*Package parser parses all the specs in the list of directories given and also de-duplicates all specs passed through `specDirs` before parsing specs.
  Gets all the specs files in the given directory and generates token for each spec file.
  While parsing a concept file, concepts are inlined i.e. concept in the spec file is replaced with steps that concept has in the concept file.
  While creating a specification file parser applies the converter functions.
  Parsing a spec file gives a specification with parseresult. ParseResult contains ParseErrors, CriticalErrors, Warnings and FileName

  Errors can be generated, While
	- Generating tokens
	- Applying converters
	- After Applying converters

  If a parse error is found in a spec, only that spec is ignored and others will continue execution.
  This doesn't invoke the language runner.
  Eg : Multiple spec headings found in same file.
       Scenario should be defined after the spec heading.

  Critical error :
  	Circular reference of concepts - Doesn't parse specs becz it goes in recursion and crashes
*/
package parser

import (
	"runtime/debug"
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

// ParseSpecFiles gets all the spec files and parse each spec file.
// Generates specifications and parse results.
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

// ParseSpecs parses specs in the give directory and gives specification and pass/fail status, used in validation.
func ParseSpecs(args []string, conceptsDictionary *gauge.ConceptDictionary, buildErrors *gauge.BuildErrors) ([]*gauge.Specification, bool) {
	specs, failed := parseSpecsInDirs(conceptsDictionary, args, buildErrors)
	specsToExecute := order.Sort(filter.FilterSpecs(specs))
	return specsToExecute, failed
}

// ParseConcepts creates concept dictionary and concept parse result.
func ParseConcepts() (*gauge.ConceptDictionary, *ParseResult, error) {
	conceptsDictionary, conceptParseResult, err := CreateConceptsDictionary()
	if err != nil {
		return nil, nil, err
	}
	HandleParseResult(conceptParseResult)
	return conceptsDictionary, conceptParseResult, nil
}

func recoverPanic() {
	if r := recover(); r != nil {
		logger.Fatalf(true, "%v\n%s", r, string(debug.Stack()))
	}
}

func parseSpec(specFile string, conceptDictionary *gauge.ConceptDictionary, specChannel chan *gauge.Specification, parseResultChan chan *ParseResult) {
	defer recoverPanic()
	specFileContent, err := common.ReadFileContents(specFile)
	if err != nil {
		specChannel <- nil
		parseResultChan <- &ParseResult{ParseErrors: []ParseError{ParseError{FileName: specFile, Message: err.Error()}}, Ok: false}
		return
	}
	spec, parseResult, err := new(SpecParser).Parse(specFileContent, conceptDictionary, specFile)
	if err != nil {
		logger.Fatalf(true, err.Error())
	}
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
			files := util.GetSpecFiles([]string{specName})
			if len(files) < 1 {
				continue
			}
			specificationFile, created := addSpecFile(&specFiles, files[0])
			if created || len(specificationFile.indices) > 0 {
				specificationFile.indices = append(specificationFile.indices, index)
			}
			givenSpecs = append(givenSpecs, files[0])
		} else {
			files := util.GetSpecFiles([]string{specSource})
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

// ExtractStepValueAndParams parses a stepText string into a StepValue struct
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

	return &gauge.StepValue{Args: args, StepValue: extractedStepValue, ParameterizedStepValue: parameterizedStepValue}, nil

}

// CreateStepValue converts a Step to StepValue
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

// HandleParseResult collates list of parse result and determines if gauge has to break flow.
func HandleParseResult(results ...*ParseResult) bool {
	var failed = false
	for _, result := range results {
		if !result.Ok {
			for _, err := range result.Errors() {
				logger.Errorf(true, err)
			}
			failed = true
		}
		if result.Warnings != nil {
			for _, warning := range result.Warnings {
				logger.Warningf(true, "[ParseWarning] %s", warning)
			}
		}
	}
	return failed
}
