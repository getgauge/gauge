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

package conceptExtractor

import (
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
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
	conceptStep    *parser.Step
	stepsToExtract []*gauge_messages.Step
	stepsInConcept string
	table          *parser.Table
	fileContent    string
	dynamicArgs    []string
	errors         []error
}

func ExtractConcept(conceptName *gauge_messages.Step, steps []*gauge_messages.Step, conceptFileName string, changeAcrossProject bool, selectedTextInfo *gauge_messages.TextInfo) (bool, error, []string) {
	content := SPEC_HEADING_TEMPLATE
	if isSpecFile(selectedTextInfo.GetFileName()) {
		content, _ = common.ReadFileContents(selectedTextInfo.GetFileName())
	}
	concept, conceptUsageText, err := getExtractedConcept(conceptName, steps, content)
	if err != nil {
		return false, err, []string{}
	}
	writeConceptToFile(concept, conceptUsageText, conceptFileName, selectedTextInfo.GetFileName(), selectedTextInfo)
	return true, errors.New(""), []string{conceptFileName, selectedTextInfo.GetFileName()}
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
	util.SaveFile(conceptFileName, content+"\n"+concept, true)
	text := ReplaceExtractedStepsWithConcept(info, conceptUsageText)
	util.SaveFile(fileName, text, true)
}

func getExtractedConcept(conceptName *gauge_messages.Step, steps []*gauge_messages.Step, content string) (string, string, error) {
	tokens, _ := new(parser.SpecParser).GenerateTokens("* " + conceptName.GetName())
	conceptStep, _ := (&parser.Specification{}).CreateStepUsingLookup(tokens[0], nil)
	specText, err := getContentWithDataTable(content)
	if err != nil {
		return "", "", err
	}
	extractor := &extractor{conceptName: "* " + conceptName.GetName(), stepsInConcept: "", stepsToExtract: steps, conceptStep: conceptStep, table: &parser.Table{}, fileContent: specText, errors: make([]error, 0)}
	extractor.extractSteps()
	if len(extractor.errors) != 0 {
		return "", "", err
	}
	conceptStep.ReplaceArgsWithDynamic(conceptStep.Args)
	addArgsFromTable(conceptStep, &extractor.conceptName, extractor.dynamicArgs)
	if extractor.table.IsInitialized() {
		extractor.conceptName += "\n" + formatter.FormatTable(extractor.table)
	}
	return strings.Replace(formatter.FormatStep(conceptStep), "* ", "# ", 1) + (extractor.stepsInConcept), extractor.conceptName, nil
}

func addArgsFromTable(concept *parser.Step, conceptName *string, args []string) {
	for _, arg := range args {
		concept.Value += " {}"
		concept.Args = append(concept.Args, &parser.StepArg{Value: arg, ArgType: parser.Dynamic, Name: arg})
		*conceptName += fmt.Sprintf(" <%s>", arg)
	}
}

func getContentWithDataTable(content string) (string, error) {
	spec, result := new(parser.SpecParser).Parse(content, &parser.ConceptDictionary{})
	if !result.Ok {
		return "", errors.New(fmt.Sprintf("Spec Parse failure: %s", result.ParseError))
	}
	newSpec := &parser.Specification{Heading: &parser.Heading{Value: "SPECHEADING"}}
	if spec.DataTable.IsInitialized() {
		newSpec = &parser.Specification{Items: []parser.Item{&spec.DataTable}, Heading: &parser.Heading{Value: "SPECHEADING"}}
	}
	return formatter.FormatSpecification(newSpec) + "\n##hello \n* step \n", nil
}

func (self *extractor) extractSteps() {
	for _, step := range self.stepsToExtract {
		tokens, _ := new(parser.SpecParser).GenerateTokens("*" + step.GetName())
		stepInConcept, _ := (&parser.Specification{}).CreateStepUsingLookup(tokens[0], nil)
		if step.GetTable() != "" {
			self.handleTable(stepInConcept, step)
		}
		stepInConcept.ReplaceArgsWithDynamic(self.conceptStep.Args)
		self.stepsInConcept += formatter.FormatStep(stepInConcept)
	}
}

func (self *extractor) handleTable(stepInConcept *parser.Step, step *gauge_messages.Step) {
	stepInConcept.Value += " {}"
	specText := self.fileContent + step.GetTable()
	spec, result := new(parser.SpecParser).Parse(specText, &parser.ConceptDictionary{})
	if !result.Ok {
		self.errors = append(self.errors, result.ParseError)
		return
	}
	stepArgs := []*parser.StepArg{spec.Scenarios[0].Steps[0].Args[0]}
	self.addTableAsParam(step, stepArgs)
	stepInConcept.Args = append(stepInConcept.Args, stepArgs[0])
}

func (self *extractor) addTableAsParam(step *gauge_messages.Step, args []*parser.StepArg) {
	if step.GetParamTableName() != "" {
		self.conceptName = strings.Replace(self.conceptName, fmt.Sprintf("<%s>", step.GetParamTableName()), "", 1)
		self.table = &args[0].Table
		args[0] = &parser.StepArg{Value: step.GetParamTableName(), ArgType: parser.Dynamic}
	} else {
		self.dynamicArgs = append(self.dynamicArgs, (&args[0].Table).GetDynamicArgs()...)
	}
}
