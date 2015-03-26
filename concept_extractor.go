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
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge_messages"
	"os"
	"regexp"
	"strings"
)

const (
	SPEC_HEADING_TEMPLATE = "# S\n\n"
	TABLE                 = "table"
)

type extractor struct {
	conceptName    string
	conceptStep    *step
	stepsToExtract []*gauge_messages.Step
	stepsInConcept string
	table          *table
	fileContent    string
	dynamicArgs    []string
	errors         []error
}

func extractConcept(conceptName *gauge_messages.Step, steps []*gauge_messages.Step, conceptFileName string, changeAcrossProject bool, selectedTextInfo *gauge_messages.TextInfo) (bool, error, []string) {
	content := SPEC_HEADING_TEMPLATE
	if isSpecFile(selectedTextInfo.GetFileName()) {
		content, _ = common.ReadFileContents(selectedTextInfo.GetFileName())
	}
	concept, conceptUsageText, err := getExtractedConcept(conceptName, steps, content)
	if err != nil {
		return false, err, []string{}
	}
	writeConceptToFile(concept, conceptUsageText, conceptFileName, selectedTextInfo.GetFileName(), selectedTextInfo)
	return true, nil, []string{}
}

func isSpecFile(fileName string) bool {
	return strings.HasSuffix(fileName, ".spec") || strings.HasSuffix(fileName, ".md")
}

func ReplaceExtractedStepsWithConcept(selectedTextInfo *gauge_messages.TextInfo, conceptText string) string {
	content, _ := common.ReadFileContents(selectedTextInfo.GetFileName())
	return replaceText(content, selectedTextInfo, conceptText)
}

func replaceText(content string, info *gauge_messages.TextInfo, replacement string) string {
	parts := regexp.MustCompile("\r\n|\n").Split(content, -1)
	for i := info.GetStartingLineNo(); i < info.GetEndLineNo(); i++ {
		parts = append(parts[:info.GetStartingLineNo()], parts[info.GetStartingLineNo()+1:]...)
	}
	parts[info.GetStartingLineNo()-1] = replacement
	return strings.Join(parts, "\n")
}

func writeConceptToFile(concept string, conceptUsageText string, conceptFileName string, fileName string, info *gauge_messages.TextInfo) {
	if _, err := os.Stat(conceptFileName); os.IsNotExist(err) {
		os.Create(conceptFileName)
	}
	content, _ := common.ReadFileContents(conceptFileName)
	saveFile(conceptFileName, content+"\n"+concept, true)
	text := ReplaceExtractedStepsWithConcept(info, conceptUsageText)
	saveFile(fileName, text, true)
}

func getExtractedConcept(conceptName *gauge_messages.Step, steps []*gauge_messages.Step, content string) (string, string, error) {
	tokens, _ := new(specParser).generateTokens("* " + conceptName.GetName())
	conceptStep, _ := (&specification{}).createStepUsingLookup(tokens[0], nil)
	specText, err := getContentWithDataTable(content)
	if err != nil {
		return "", "", err
	}
	extractor := &extractor{conceptName: "* " + conceptName.GetName(), stepsInConcept: "", stepsToExtract: steps, conceptStep: conceptStep, table: &table{}, fileContent: specText, errors: make([]error, 0)}
	extractor.extractSteps()
	if len(extractor.errors) != 0 {
		return "", "", err
	}
	conceptStep.replaceArgsWithDynamic(conceptStep.args)
	addArgsFromTable(conceptStep, &extractor.conceptName, extractor.dynamicArgs)
	if extractor.table.isInitialized() {
		extractor.conceptName += "\n" + formatTable(extractor.table)
	}
	return strings.Replace(formatStep(conceptStep), "* ", "# ", 1) + (extractor.stepsInConcept), extractor.conceptName, nil
}

func addArgsFromTable(concept *step, conceptName *string, args []string) {
	for _, arg := range args {
		concept.value += " {}"
		concept.args = append(concept.args, &stepArg{value: arg, argType: dynamic, name: arg})
		*conceptName += fmt.Sprintf(" <%s>", arg)
	}
}

func getContentWithDataTable(content string) (string, error) {
	spec, result := new(specParser).parse(content, &conceptDictionary{})
	if !result.ok {
		return "", errors.New(fmt.Sprintf("Spec Parse failure: %s", result.error))
	}
	newSpec := &specification{heading: &heading{value: "SPECHEADING"}}
	if spec.dataTable.isInitialized() {
		newSpec = &specification{items: []item{&spec.dataTable}, heading: &heading{value: "SPECHEADING"}}
	}
	return formatSpecification(newSpec) + "\n##hello \n* step \n", nil
}

func (self *extractor) extractSteps() {
	for _, step := range self.stepsToExtract {
		tokens, _ := new(specParser).generateTokens("*" + step.GetName())
		stepInConcept, _ := (&specification{}).createStepUsingLookup(tokens[0], nil)
		if step.GetTable() != "" {
			self.handleTable(stepInConcept, step)
		}
		stepInConcept.replaceArgsWithDynamic(self.conceptStep.args)
		self.stepsInConcept += formatItem(stepInConcept)
	}
}

func (self *extractor) handleTable(stepInConcept *step, step *gauge_messages.Step) {
	stepInConcept.value += " {}"
	specText := self.fileContent + step.GetTable()
	spec, result := new(specParser).parse(specText, &conceptDictionary{})
	if !result.ok {
		self.errors = append(self.errors, result.error)
		return
	}
	stepArgs := []*stepArg{spec.scenarios[0].steps[0].args[0]}
	self.addTableAsParam(step, stepArgs)
	stepInConcept.args = append(stepInConcept.args, stepArgs[0])
}

func (self *extractor) addTableAsParam(step *gauge_messages.Step, args []*stepArg) {
	if step.GetParamTableName() != "" {
		self.conceptName = strings.Replace(self.conceptName, fmt.Sprintf("<%s>", step.GetParamTableName()), "", 1)
		self.table = &args[0].table
		args[0] = &stepArg{value: step.GetParamTableName(), argType: dynamic}
	} else {
		self.dynamicArgs = append(self.dynamicArgs, (&args[0].table).getDynamicArgs()...)
	}
}
