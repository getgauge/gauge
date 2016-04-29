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

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
)

func ParseSpecFiles(specFiles []string, conceptDictionary *gauge.ConceptDictionary) ([]*gauge.Specification, []*ParseResult) {
	parseResultsChan := make(chan *ParseResult, len(specFiles))
	specsChan := make(chan *gauge.Specification, len(specFiles))
	var parseResults []*ParseResult
	var specs []*gauge.Specification

	for _, specFile := range specFiles {
		go parseSpec(specFile, conceptDictionary, specsChan, parseResultsChan)
	}
	for _ = range specFiles {
		parseResults = append(parseResults, <-parseResultsChan)
		spec := <-specsChan
		if spec != nil {
			specs = append(specs, spec)
		}
	}
	return specs, parseResults
}

func parseSpec(specFile string, conceptDictionary *gauge.ConceptDictionary, specChannel chan *gauge.Specification, parseResultChan chan *ParseResult) {
	specFileContent, err := common.ReadFileContents(specFile)
	if err != nil {
		specChannel <- nil
		parseResultChan <- &ParseResult{ParseErrors: []*ParseError{&ParseError{Message: err.Error()}}, Ok: false, FileName: specFile}
		return
	}
	spec, parseResult := new(SpecParser).Parse(specFileContent, conceptDictionary)
	parseResult.FileName = specFile
	if spec != nil {
		spec.FileName = specFile
	}
	specChannel <- spec
	parseResultChan <- parseResult
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
				logger.Warning("%s : %v", result.FileName, warning)
			}
		}
	}
	return failed
}
