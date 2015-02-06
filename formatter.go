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
	"bytes"
	"fmt"
	"github.com/getgauge/gauge/gauge_messages"
	"sort"
	"strings"
)

const (
	TABLE_LEFT_SPACING = 5
)

func getRepeatedChars(character string, repeatCount int) string {
	formatted := ""
	for i := 0; i < repeatCount; i++ {
		formatted = fmt.Sprintf("%s%s", formatted, character)
	}
	return formatted
}

func formatSpecHeading(specHeading string) string {
	return formatHeading(specHeading, "=")
}

func formatScenarioHeading(scenarioHeading string) string {
	return fmt.Sprintf("%s", formatHeading(scenarioHeading, "-"))
}

func formatStep(step *step) string {
	text := step.value
	paramCount := strings.Count(text, PARAMETER_PLACEHOLDER)
	for i := 0; i < paramCount; i++ {
		argument := step.args[i]
		formattedArg := ""
		if argument.argType == tableArg {
			formattedTable := formatTable(&argument.table)
			formattedArg = fmt.Sprintf("\n%s", formattedTable)
		} else if argument.argType == dynamic {
			formattedArg = fmt.Sprintf("<%s>", getUnescapedString(argument.value))
		} else if argument.argType == specialString || argument.argType == specialTable {
			formattedArg = fmt.Sprintf("<%s>", getUnescapedString(argument.name))
		} else {
			formattedArg = fmt.Sprintf("\"%s\"", getUnescapedString(argument.value))
		}
		text = strings.Replace(text, PARAMETER_PLACEHOLDER, formattedArg, 1)
	}
	stepText := ""
	if strings.HasSuffix(text, "\n") {
		stepText = fmt.Sprintf("* %s", text)
	} else {
		stepText = fmt.Sprintf("* %s\n", text)
	}
	return stepText
}

func formatConcept(protoConcept *gauge_messages.ProtoConcept) string {
	conceptText := "* "
	for _, fragment := range protoConcept.ConceptStep.GetFragments() {
		if fragment.GetFragmentType() == gauge_messages.Fragment_Text {
			conceptText = conceptText + fragment.GetText()
		} else if fragment.GetFragmentType() == gauge_messages.Fragment_Parameter {
			if fragment.GetParameter().GetParameterType() == (gauge_messages.Parameter_Table | gauge_messages.Parameter_Special_Table) {
				conceptText += "\n" + formatTable(tableFrom(fragment.GetParameter().GetTable()))
			} else {
				conceptText = conceptText + "\"" + fragment.GetParameter().GetValue() + "\""
			}
		}
	}
	return conceptText + "\n"
}

func formatHeading(heading, headingChar string) string {
	trimmedHeading := strings.TrimSpace(heading)
	length := len(trimmedHeading)
	return fmt.Sprintf("%s\n%s\n", trimmedHeading, getRepeatedChars(headingChar, length))
}

func formatTable(table *table) string {
	columnToWidthMap := make(map[int]int)
	for i, header := range table.headers {
		//table.get(header) returns a list of cells in that particular column
		cells := table.get(header)
		columnToWidthMap[i] = findLongestCellWidth(cells, len(header))
	}

	var tableStringBuffer bytes.Buffer
	tableStringBuffer.WriteString(fmt.Sprintf("%s|", getRepeatedChars(" ", TABLE_LEFT_SPACING)))
	for i, header := range table.headers {
		width := columnToWidthMap[i]
		tableStringBuffer.WriteString(fmt.Sprintf("%s|", addPaddingToCell(header, width)))
	}

	tableStringBuffer.WriteString("\n")
	tableStringBuffer.WriteString(fmt.Sprintf("%s|", getRepeatedChars(" ", TABLE_LEFT_SPACING)))
	for i, _ := range table.headers {
		width := columnToWidthMap[i]
		cell := getRepeatedChars("-", width)
		tableStringBuffer.WriteString(fmt.Sprintf("%s|", addPaddingToCell(cell, width)))
	}

	tableStringBuffer.WriteString("\n")
	for _, row := range table.getRows() {
		tableStringBuffer.WriteString(fmt.Sprintf("%s|", getRepeatedChars(" ", TABLE_LEFT_SPACING)))
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

func findLongestCellWidth(columnCells []tableCell, minValue int) int {
	longestLength := minValue
	for _, cellValue := range columnCells {
		cellValueLen := len(cellValue.value)
		if cellValueLen > longestLength {
			longestLength = cellValueLen
		}
	}
	return longestLength
}

func formatComment(comment *comment) string {
	if comment.value == "\n" {
		return comment.value
	}
	return fmt.Sprintf("%s\n", comment.value)
}

func formatTags(tags *tags) string {
	if tags == nil || len(tags.values) == 0 {
		return ""
	}
	var b bytes.Buffer
	b.WriteString("tags: ")
	for i, tag := range tags.values {
		b.WriteString(tag)
		if (i + 1) != len(tags.values) {
			b.WriteString(", ")
		}
	}
	b.WriteString("\n")
	return string(b.Bytes())
}

func formatSpecification(specification *specification) string {
	var formattedSpec bytes.Buffer
	formatter := &formatter{buffer: formattedSpec}
	specification.traverse(formatter)
	return string(formatter.buffer.Bytes())
}

type ByLineNo []*concept

func (s ByLineNo) Len() int {
	return len(s)
}

func (s ByLineNo) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLineNo) Less(i, j int) bool {
	return s[i].conceptStep.lineNo < s[j].conceptStep.lineNo
}

func sortConcepts(conceptDictionary *conceptDictionary, conceptMap map[string]string) []*concept {
	concepts := make([]*concept, 0)
	for _, concept := range conceptDictionary.conceptsMap {
		conceptMap[concept.fileName] = ""
		concepts = append(concepts, concept)
	}
	sort.Sort(ByLineNo(concepts))
	return concepts
}

func formatConceptSteps(conceptMap map[string]string, concept *concept) {
	conceptMap[concept.fileName] += strings.TrimSpace(strings.Replace(formatStep(concept.conceptStep), "*", "#", 1)) + "\n"
	for i := 1; i < len(concept.conceptStep.items); i++ {
		conceptMap[concept.fileName] += formatItem(concept.conceptStep.items[i])
	}
}

func formatConcepts(conceptDictionary *conceptDictionary) map[string]string {
	conceptMap := make(map[string]string)
	for _, concept := range sortConcepts(conceptDictionary, conceptMap) {
		for _, comment := range concept.conceptStep.preComments {
			conceptMap[concept.fileName] += formatComment(comment)
		}
		formatConceptSteps(conceptMap, concept)
	}
	return conceptMap
}

func formatItem(item item) string {
	switch item.kind() {
	case commentKind:
		comment := item.(*comment)
		if comment.value == "\n" {
			return comment.value
		}
		return fmt.Sprintf("%s\n", comment.value)
	case stepKind:
		step := item.(*step)
		return formatStep(step)
	case tableKind:
		table := item.(*table)
		return formatTable(table)
	case tagKind:
		tags := item.(*tags)
		return formatTags(tags)
	}
	return ""
}
