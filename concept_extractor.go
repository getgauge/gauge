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
	SPEC_HEADING_TEMPLATE = "# S\n"
	TABLE                 = "table"
)

type extractor struct {
	conceptName    string
	conceptStep    *step
	stepsToExtract []*gauge_messages.Step
	stepsInConcept string
	table          *table
}

func extractConcept(conceptName *gauge_messages.Step, steps []*gauge_messages.Step, conceptFileName string, changeAcrossProject bool, selectedTextInfo *gauge_messages.TextInfo) (bool, error, []string) {
	concept, conceptUsageText := getExtractedConcept(conceptName, steps)
	specText := ReplaceExtractedStepsWithConcept(selectedTextInfo, conceptUsageText)
	writeConceptToFile(concept, specText, conceptFileName, selectedTextInfo.GetFileName())
	return true, errors.New(""), []string{}
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

func writeConceptToFile(concept string, conceptUsageText string, conceptFileName string, fileName string) {
	os.Create(conceptFileName)
	content, _ := common.ReadFileContents(conceptFileName)
	saveFile(conceptFileName, content+"\n"+concept, true)
	saveFile(fileName, conceptUsageText, true)
}

func getExtractedConcept(conceptName *gauge_messages.Step, steps []*gauge_messages.Step) (string, string) {
	tokens, _ := new(specParser).generateTokens("* " + conceptName.GetName())
	conceptStep, _ := (&specification{}).createStepUsingLookup(tokens[0], nil)
	extractor := &extractor{conceptName: "* " + conceptName.GetName(), stepsInConcept: "", stepsToExtract: steps, conceptStep: conceptStep, table: &table{}}
	extractor.extractSteps()
	conceptStep.replaceArgsWithDynamic(conceptStep.args)
	if extractor.table.isInitialized() {
		extractor.conceptName += "\n" + formatTable(extractor.table)
	}
	return strings.Replace(formatItem(conceptStep), "* ", "# ", 1) + (extractor.stepsInConcept), extractor.conceptName
}

func (self *extractor) extractSteps() {
	for _, step := range self.stepsToExtract {
		tokens, _ := new(specParser).generateTokens("*" + step.GetName())
		stepInConcept, _ := (&specification{}).createStepUsingLookup(tokens[0], nil)
		if step.GetTable() != nil {
			self.handleTable(stepInConcept, step)
		}
		stepInConcept.replaceArgsWithDynamic(self.conceptStep.args)
		self.stepsInConcept += formatItem(stepInConcept)
	}
}

func (self *extractor) handleTable(stepInConcept *step, step *gauge_messages.Step) {
	stepInConcept.value += " {}"
	table := TABLE
	parameterType := gauge_messages.Parameter_Table
	parameter := &gauge_messages.Parameter{Table: step.GetTable(), Name: &table, ParameterType: &parameterType}
	stepArgs := createStepArgsFromProtoArguments([]*gauge_messages.Parameter{parameter})
	self.addTableAsParam(step, stepArgs)
	stepInConcept.args = append(stepInConcept.args, stepArgs[0])
}

func (self *extractor) addTableAsParam(step *gauge_messages.Step, args []*stepArg) {
	if step.GetParamTableName() != "" {
		self.conceptName = strings.Replace(self.conceptName, fmt.Sprintf("\"%s\"", step.GetParamTableName()), "", 1)
		self.table = &args[0].table
		args[0] = &stepArg{value: step.GetParamTableName(), argType: dynamic}
	}
}
