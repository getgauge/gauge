/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package conceptExtractor

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"path"

	"github.com/getgauge/common"
	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
)

const (
	SPEC_HEADING_TEMPLATE = "# S\n\n"
	TABLE                 = "table"
)

type extractor struct {
	conceptName    string
	conceptStep    *gauge.Step
	stepsToExtract []*gm.Step
	stepsInConcept string
	table          *gauge.Table
	fileContent    string
	dynamicArgs    []string
	errors         []error
}

// ExtractConcept creates concept form the selected text and writes the concept to the given concept file.
func ExtractConcept(conceptName *gm.Step, steps []*gm.Step, conceptFileName string, info *gm.TextInfo) (bool, []string, error) {
	content := SPEC_HEADING_TEMPLATE
	if util.IsSpec(info.GetFileName()) {
		content, _ = common.ReadFileContents(info.GetFileName())
	}
	concept, cptText, err := getExtractedConcept(conceptName, steps, content, info.GetFileName())
	if err != nil {
		return false, []string{}, err
	}
	err = writeConceptToFile(concept, cptText, conceptFileName, info.GetFileName(), info)
	if err != nil {
		return false, []string{}, err
	}
	return true, []string{conceptFileName, info.GetFileName()}, nil
}

// ReplaceExtractedStepsWithConcept replaces the steps selected for concept extraction with the concept name given.
func ReplaceExtractedStepsWithConcept(selectedTextInfo *gm.TextInfo, conceptText string) (string, int) {
	content, _ := common.ReadFileContents(selectedTextInfo.GetFileName())
	newText := replaceText(content, selectedTextInfo, conceptText)
	if util.GetLineCount(content) > util.GetLineCount(newText) {
		return newText, util.GetLineCount(content)
	}
	return newText, util.GetLineCount(newText)
}

func replaceText(content string, info *gm.TextInfo, replacement string) string {
	parts := regexp.MustCompile("\r\n|\n").Split(content, -1)
	for i := info.GetStartingLineNo(); i < info.GetEndLineNo(); i++ {
		parts = append(parts[:info.GetStartingLineNo()], parts[info.GetStartingLineNo()+1:]...)
	}
	parts[info.GetStartingLineNo()-1] = replacement
	return strings.Join(parts, "\n")
}

func writeConceptToFile(concept string, conceptUsageText string, conceptFileName string, fileName string, info *gm.TextInfo) error {
	if _, err := os.Stat(conceptFileName); os.IsNotExist(err) {
		basepath := path.Dir(conceptFileName)
		if _, err := os.Stat(basepath); os.IsNotExist(err) {
			err = os.MkdirAll(basepath, common.NewDirectoryPermissions)
			return fmt.Errorf("unable to create directory %s: %s", basepath, err.Error())
		}
		_, err = os.Create(conceptFileName)
		if err != nil {
			return fmt.Errorf("unable to create file %s: %s", conceptFileName, err.Error())
		}
	}
	content, err := common.ReadFileContents(conceptFileName)
	if err != nil {
		return fmt.Errorf("unable to read from %s: %s", conceptFileName, err.Error())
	}
	util.SaveFile(conceptFileName, content+"\n"+concept, true)
	text, _ := ReplaceExtractedStepsWithConcept(info, conceptUsageText)
	util.SaveFile(fileName, text, true)
	return nil
}

func getExtractedConcept(conceptName *gm.Step, steps []*gm.Step, content string, cptFileName string) (string, string, error) {
	tokens, _ := new(parser.SpecParser).GenerateTokens("* "+conceptName.GetName(), cptFileName)
	conceptStep, _ := parser.CreateStepUsingLookup(tokens[0], nil, cptFileName)
	cptDict, _, err := parser.ParseConcepts()
	if err != nil {
		return "", "", err
	}
	if isDuplicateConcept(conceptStep, cptDict) {
		return "", "", fmt.Errorf("Concept `%s` already present", conceptName.GetName())
	}
	specText, err := getContentWithDataTable(content, cptFileName)
	if err != nil {
		return "", "", err
	}
	extractor := &extractor{conceptName: "* " + conceptName.GetName(), stepsInConcept: "", stepsToExtract: steps, conceptStep: conceptStep, table: &gauge.Table{}, fileContent: specText, errors: make([]error, 0)}
	err = extractor.extractSteps(cptFileName)
	if err != nil {
		return "", "", err
	}
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

func addArgsFromTable(concept *gauge.Step, conceptName *string, args []string) {
	for _, arg := range args {
		concept.Value += " {}"
		concept.Args = append(concept.Args, &gauge.StepArg{Value: arg, ArgType: gauge.Dynamic, Name: arg})
		*conceptName += fmt.Sprintf(" <%s>", arg)
	}
}

func getContentWithDataTable(content, cptFileName string) (string, error) {
	spec, result, err := new(parser.SpecParser).Parse(content, &gauge.ConceptDictionary{}, cptFileName)
	if err != nil {
		return "", err
	}
	if !result.Ok {
		return "", fmt.Errorf("Spec Parse failure: %s", result.ParseErrors)
	}
	newSpec := &gauge.Specification{Heading: &gauge.Heading{Value: "SPECHEADING"}}
	if spec.DataTable.IsInitialized() {
		newSpec = &gauge.Specification{Items: []gauge.Item{&spec.DataTable}, Heading: &gauge.Heading{Value: "SPECHEADING"}}
	}
	return formatter.FormatSpecification(newSpec) + "\n##hello \n* step \n", nil
}

func isDuplicateConcept(concept *gauge.Step, cptDict *gauge.ConceptDictionary) bool {
	for _, cpt := range cptDict.ConceptsMap {
		if strings.TrimSpace(cpt.ConceptStep.Value) == strings.TrimSpace(concept.Value) {
			return true
		}
	}
	return false
}

func (e *extractor) extractSteps(cptFileName string) error {
	for _, step := range e.stepsToExtract {
		tokens, _ := new(parser.SpecParser).GenerateTokens("*"+step.GetName(), cptFileName)
		stepInConcept, _ := parser.CreateStepUsingLookup(tokens[0], nil, cptFileName)
		if step.GetTable() != "" {
			if err := e.handleTable(stepInConcept, step, cptFileName); err != nil {
				return err
			}
		}
		stepInConcept.ReplaceArgsWithDynamic(e.conceptStep.Args)
		e.stepsInConcept += formatter.FormatStep(stepInConcept)
	}
	return nil
}

func (e *extractor) handleTable(stepInConcept *gauge.Step, step *gm.Step, cptFileName string) error {
	stepInConcept.Value += " {}"
	specText := e.fileContent + step.GetTable()
	spec, result, err := new(parser.SpecParser).Parse(specText, &gauge.ConceptDictionary{}, cptFileName)
	if err != nil {
		return err
	}

	if !result.Ok {
		for _, err := range result.ParseErrors {
			e.errors = append(e.errors, err)
		}
		return nil
	}
	stepArgs := []*gauge.StepArg{spec.Scenarios[0].Steps[0].Args[0]}
	e.addTableAsParam(step, stepArgs)
	stepInConcept.Args = append(stepInConcept.Args, stepArgs[0])
	return nil
}

func (e *extractor) addTableAsParam(step *gm.Step, args []*gauge.StepArg) {
	if step.GetParamTableName() != "" {
		e.conceptName = strings.Replace(e.conceptName, fmt.Sprintf("<%s>", step.GetParamTableName()), "", 1)
		e.table = &args[0].Table
		args[0] = &gauge.StepArg{Name: step.GetParamTableName(), ArgType: gauge.Dynamic}
	} else {
		e.dynamicArgs = append(e.dynamicArgs, (&args[0].Table).GetDynamicArgs()...)
	}
}
