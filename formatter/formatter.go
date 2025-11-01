/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package formatter

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
)

const (
	tableLeftSpacing = 3
)

func FormatSpecFiles(specFiles ...string) []*parser.ParseResult {
	specs, results := parser.ParseSpecFiles(specFiles, &gauge.ConceptDictionary{}, gauge.NewBuildErrors())
	resultsMap := getParseResult(results)
	filesSkipped := make([]string, 0)
	for _, spec := range specs {
		result := resultsMap[spec.FileName]
		if !result.Ok {
			filesSkipped = append(filesSkipped, spec.FileName)
			continue
		}
		if err := formatAndSave(spec); err != nil {
			result.ParseErrors = []parser.ParseError{parser.ParseError{Message: err.Error()}}
		} else {
			logger.Debugf(true, "Successfully formatted spec: %s", util.RelPathToProjectRoot(spec.FileName))
		}
	}
	if len(filesSkipped) > 0 {
		logger.Errorf(true, "Skipping %d file(s), due to following error(s):", len(filesSkipped))
	}
	return results
}

func getParseResult(results []*parser.ParseResult) map[string]*parser.ParseResult {
	resultsMap := make(map[string]*parser.ParseResult)
	for _, result := range results {
		resultsMap[result.FileName] = result
	}
	return resultsMap
}

func FormatStep(step *gauge.Step) string {
	text := step.Value
	paramCount := strings.Count(text, gauge.ParameterPlaceholder)
	for i := 0; i < paramCount; i++ {
		argument := step.Args[i]
		var formattedArg string
		stripBeforeArg := ""
		switch argument.ArgType {
		case gauge.TableArg:
			formattedArg = fmt.Sprintf("\n%s", FormatTable(&argument.Table))
			stripBeforeArg = " "
		case gauge.Dynamic, gauge.SpecialString, gauge.SpecialTable:
			formattedArg = fmt.Sprintf("<%s>", parser.GetUnescapedString(argument.Name))
		case gauge.MultilineString:
			formattedArg = fmt.Sprintf("\n\"\"\"\n%s\n\"\"\"\n", argument.Value)
			stripBeforeArg = " "
		default:
			formattedArg = fmt.Sprintf("\"%s\"", parser.GetUnescapedString(argument.Value))
		}
		text = strings.Replace(text, stripBeforeArg+gauge.ParameterPlaceholder, formattedArg, 1)
	}
	stepText := ""
	if strings.HasSuffix(text, "\n") {
		stepText = fmt.Sprintf("* %s", text)
	} else {
		stepText = fmt.Sprintf("* %s%s\n", text, step.Suffix)
	}
	return stepText
}

func FormatStepWithResolvedArgs(step *gauge.Step) string {
	text := step.Value
	paramCount := strings.Count(text, gauge.ParameterPlaceholder)
	sf := make([]*gauge_messages.Fragment, 0)
	for _, f := range step.GetFragments() {
		if f.FragmentType == gauge_messages.Fragment_Parameter {
			sf = append(sf, f)
		}
	}
	for i := 0; i < paramCount; i++ {
		a := step.Args[i]
		var formattedArg string
		if a.ArgType == gauge.TableArg && sf[i].Parameter.ParameterType == gauge_messages.Parameter_Table {
			formattedArg = fmt.Sprintf("\n%s", FormatTable(&a.Table))
		} else {
			formattedArg = fmt.Sprintf("\"%s\"", sf[i].GetParameter().Value)
		}
		text = strings.Replace(text, gauge.ParameterPlaceholder, formattedArg, 1)
	}
	stepText := ""
	if strings.HasSuffix(text, "\n") {
		stepText = fmt.Sprintf("* %s", text)
	} else {
		stepText = fmt.Sprintf("* %s%s\n", text, step.Suffix)
	}
	return stepText
}

func FormatHeading(heading, headingChar string) string {
	trimmedHeading := strings.TrimSpace(heading)
	return fmt.Sprintf("%s %s\n", headingChar, trimmedHeading)
}

func FormatTable(table *gauge.Table) string {
	columnToWidthMap := make(map[int]int)
	for i, header := range table.Headers {
		//table.get(header) returns a list of cells in that particular column
		cells, _ := table.Get(header)
		columnToWidthMap[i] = findLongestCellWidth(cells, len(header))
	}

	var tableStringBuffer bytes.Buffer

	if !config.CurrentGaugeSettings().Format.SkipEmptyLineInsertions {
		tableStringBuffer.WriteString("\n")
	}

	tableStringBuffer.WriteString(fmt.Sprintf("%s|", getRepeatedChars(" ", tableLeftSpacing)))
	for i, header := range table.Headers {
		width := columnToWidthMap[i]
		tableStringBuffer.WriteString(fmt.Sprintf("%s|", addPaddingToCell(header, width)))
	}

	tableStringBuffer.WriteString("\n")
	tableStringBuffer.WriteString(fmt.Sprintf("%s|", getRepeatedChars(" ", tableLeftSpacing)))
	for i := range table.Headers {
		width := columnToWidthMap[i]
		cell := getRepeatedChars("-", width)
		tableStringBuffer.WriteString(fmt.Sprintf("%s|", addPaddingToCell(cell, width)))
	}

	tableStringBuffer.WriteString("\n")
	for _, row := range table.Rows() {
		tableStringBuffer.WriteString(fmt.Sprintf("%s|", getRepeatedChars(" ", tableLeftSpacing)))
		for i, cell := range row {
			width := columnToWidthMap[i]
			tableStringBuffer.WriteString(fmt.Sprintf("%s|", addPaddingToCell(cell, width)))
		}
		tableStringBuffer.WriteString("\n")
	}

	return tableStringBuffer.String()
}

func addPaddingToCell(cellValue string, width int) string {
	cellRunes := []rune(cellValue)
	padding := getRepeatedChars(" ", width-len(cellRunes))
	return fmt.Sprintf("%s%s", string(cellRunes), padding)
}

func findLongestCellWidth(columnCells []gauge.TableCell, minValue int) int {
	longestLength := minValue
	for _, cellValue := range columnCells {
		cellValueLen := len([]rune(cellValue.GetValue()))
		if cellValueLen > longestLength {
			longestLength = cellValueLen
		}
	}
	return longestLength
}

func FormatComment(comment *gauge.Comment) string {
	if comment.Value == "\n" {
		return comment.Value
	}
	return fmt.Sprintf("%s\n", comment.Value)
}

func FormatTags(tags *gauge.Tags) string {
	if tags == nil || len(tags.RawValues) == 0 {
		return ""
	}
	var b bytes.Buffer
	b.WriteString("tags: ")
	for i, tag := range tags.RawValues {
		for j, tagString := range tag {
			b.WriteString(tagString)
			if (i != len(tags.RawValues)-1) || (j != len(tag)-1) {
				b.WriteString(", ")
			}
		}
		b.WriteString("\n")
		if i != len(tags.RawValues)-1 {
			b.WriteString("      ")
		}
	}
	return b.String()
}

func formatExternalDataTable(dataTable *gauge.DataTable) string {
	if dataTable == nil || len(dataTable.Value) == 0 {
		return ""
	}
	var b bytes.Buffer
	b.WriteString(dataTable.Value)
	b.WriteString("\n")
	return b.String()
}

func formatAndSave(spec *gauge.Specification) error {
	formatted := FormatSpecification(spec)
	if err := common.SaveFile(spec.FileName, formatted, true); err != nil {
		return err
	}
	return nil
}

func FormatSpecification(specification *gauge.Specification) string {
	var formattedSpec bytes.Buffer
	queue := &gauge.ItemQueue{Items: specification.AllItems()}
	formatter := &formatter{buffer: formattedSpec, itemQueue: queue}
	specification.Traverse(formatter, queue)
	return formatter.buffer.String()
}

func sortConcepts(conceptDictionary *gauge.ConceptDictionary, conceptMap map[string]string) []*gauge.Concept {
	var concepts []*gauge.Concept
	for _, concept := range conceptDictionary.ConceptsMap {
		conceptMap[concept.FileName] = ""
		concepts = append(concepts, concept)
	}
	sort.Sort(gauge.ByLineNo(concepts))
	return concepts
}

func formatConceptSteps(conceptMap map[string]string, concept *gauge.Concept) {
	conceptMap[concept.FileName] += strings.TrimSpace(strings.Replace(FormatStep(concept.ConceptStep), "*", "#", 1)) + "\n"
	for i := 1; i < len(concept.ConceptStep.Items); i++ {
		conceptMap[concept.FileName] += formatItem(concept.ConceptStep.Items[i])
	}
}

func FormatConcepts(conceptDictionary *gauge.ConceptDictionary) map[string]string {
	conceptMap := make(map[string]string)
	for _, concept := range sortConcepts(conceptDictionary, conceptMap) {
		for _, comment := range concept.ConceptStep.PreComments {
			conceptMap[concept.FileName] += FormatComment(comment)
		}
		formatConceptSteps(conceptMap, concept)
	}
	return conceptMap
}

func formatItem(item gauge.Item) string {
	switch item.Kind() {
	case gauge.CommentKind:
		comment := item.(*gauge.Comment)
		if comment.Value == "\n" {
			return comment.Value
		}
		return fmt.Sprintf("%s\n", comment.Value)
	case gauge.StepKind:
		step := item.(*gauge.Step)
		return FormatStep(step)
	case gauge.DataTableKind:
		dataTable := item.(*gauge.DataTable)
		return FormatTable(dataTable.Table)
	case gauge.TagKind:
		tags := item.(*gauge.Tags)
		return FormatTags(tags)
	}
	return ""
}

func getRepeatedChars(character string, repeatCount int) string {
	formatted := ""
	for i := 0; i < repeatCount; i++ {
		formatted = fmt.Sprintf("%s%s", formatted, character)
	}
	return formatted
}

func FormatSpecFilesIn(filesLocation string) {
	specFiles := util.GetSpecFiles([]string{filesLocation})
	parseResults := FormatSpecFiles(specFiles...)
	if parser.HandleParseResult(parseResults...) {
		os.Exit(1)
	}
}

func FormatConceptFilesIn(filesLocation string) {
	conceptFiles := util.FindConceptFiles([]string{filesLocation})
	conceptsDictionary := gauge.NewConceptDictionary()
	if _, errs, e := parser.AddConcepts(conceptFiles, conceptsDictionary); len(errs) > 0 {
		for _, err := range errs {
			logger.Errorf(false, "Concept parse failure: %s %s", conceptFiles[0], err)
		}
		if e != nil {
			logger.Errorf(false, "Concept failure: %s %s", conceptFiles[0], e)
			os.Exit(1)
		}
	}
	conceptMap := FormatConcepts(conceptsDictionary)
	for file, formatted := range conceptMap {
		e := common.SaveFile(file, formatted, true)
		if e != nil {
			logger.Errorf(false, "Concept file save failure: %s", e)
			os.Exit(1)
		}
	}
}
