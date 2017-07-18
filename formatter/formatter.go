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

package formatter

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
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
			logger.Debugf("Successfully formatted spec: %s", util.RelPathToProjectRoot(spec.FileName))
		}
	}
	if len(filesSkipped) > 0 {
		logger.Errorf("Skipping %d file(s), due to following error(s):", len(filesSkipped))
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
		formattedArg := ""
		if argument.ArgType == gauge.TableArg {
			formattedTable := FormatTable(&argument.Table)
			formattedArg = fmt.Sprintf("\n%s", formattedTable)
		} else if argument.ArgType == gauge.Dynamic {
			formattedArg = fmt.Sprintf("<%s>", parser.GetUnescapedString(argument.Value))
		} else if argument.ArgType == gauge.SpecialString || argument.ArgType == gauge.SpecialTable {
			formattedArg = fmt.Sprintf("<%s>", parser.GetUnescapedString(argument.Name))
		} else {
			formattedArg = fmt.Sprintf("\"%s\"", parser.GetUnescapedString(argument.Value))
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

func FormatConcept(protoConcept *gauge_messages.ProtoConcept) string {
	conceptText := "* "
	for _, fragment := range protoConcept.ConceptStep.GetFragments() {
		if fragment.GetFragmentType() == gauge_messages.Fragment_Text {
			conceptText = conceptText + fragment.GetText()
		} else if fragment.GetFragmentType() == gauge_messages.Fragment_Parameter {
			if fragment.GetParameter().GetParameterType() == (gauge_messages.Parameter_Table | gauge_messages.Parameter_Special_Table) {
				conceptText += "\n" + FormatTable(parser.TableFrom(fragment.GetParameter().GetTable()))
			} else {
				conceptText = conceptText + "\"" + fragment.GetParameter().GetValue() + "\""
			}
		}
	}
	return conceptText + "\n"
}

func FormatHeading(heading, headingChar string) string {
	trimmedHeading := strings.TrimSpace(heading)
	length := len(trimmedHeading)
	return fmt.Sprintf("%s\n%s\n", trimmedHeading, getRepeatedChars(headingChar, length))
}

func FormatTable(table *gauge.Table) string {
	columnToWidthMap := make(map[int]int)
	for i, header := range table.Headers {
		//table.get(header) returns a list of cells in that particular column
		cells := table.Get(header)
		columnToWidthMap[i] = findLongestCellWidth(cells, len(header))
	}

	var tableStringBuffer bytes.Buffer

	tableStringBuffer.WriteString("\n")

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

	return string(tableStringBuffer.Bytes())
}

func addPaddingToCell(cellValue string, width int) string {
	padding := getRepeatedChars(" ", width-len(cellValue))
	return fmt.Sprintf("%s%s", cellValue, padding)
}

func findLongestCellWidth(columnCells []gauge.TableCell, minValue int) int {
	longestLength := minValue
	for _, cellValue := range columnCells {
		cellValueLen := len(cellValue.GetValue())
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
	if tags == nil || len(tags.Values) == 0 {
		return ""
	}
	var b bytes.Buffer
	b.WriteString("tags: ")
	for i, tag := range tags.Values {
		b.WriteString(tag)
		if (i + 1) != len(tags.Values) {
			b.WriteString(", ")
		}
	}
	b.WriteString("\n")
	return string(b.Bytes())
}

func formatExternalDataTable(dataTable *gauge.DataTable) string {
	if dataTable == nil || len(dataTable.Value) == 0 {
		return ""
	}
	var b bytes.Buffer
	b.WriteString(dataTable.Value)
	b.WriteString("\n")
	return string(b.Bytes())
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
	return string(formatter.buffer.Bytes())
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
		return FormatTable(&dataTable.Table)
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
	specFiles := util.GetSpecFiles(filesLocation)
	parseResults := FormatSpecFiles(specFiles...)
	if parser.HandleParseResult(parseResults...) {
		os.Exit(1)
	}
}
