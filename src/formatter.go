package main

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	HEADING_UNDERLINE_LENGTH = 20
	TABLE_LEFT_SPACING       = 5
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
			formattedArg = fmt.Sprintf("<%s>", argument.value)
		} else {
			formattedArg = fmt.Sprintf("\"%s\"", argument.value)
		}
		text = strings.Replace(text, PARAMETER_PLACEHOLDER, formattedArg, 1)
	}
	stepText := fmt.Sprintf("* %s\n", text)
	return stepText
}

func formatHeading(heading, headingChar string) string {
	length := len(heading)
	if length > HEADING_UNDERLINE_LENGTH {
		length = HEADING_UNDERLINE_LENGTH
	}

	return fmt.Sprintf("%s\n%s\n", heading, getRepeatedChars(headingChar, length))
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
	case scenarioKind:
		scenario := item.(*scenario)
		var b bytes.Buffer
		b.WriteString(formatScenarioHeading(scenario.heading.value))
		b.WriteString(formatItems(scenario.items))
		return string(b.Bytes())
	case tagKind:
		tags := item.(*tags)
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
	return ""
}

func formatItems(items []item) string {
	var result bytes.Buffer
	for _, item := range items {
		result.WriteString(formatItem(item))
	}
	return string(result.Bytes())
}

func formatSpecification(specification *specification) string {
	var formattedText bytes.Buffer
	formattedText.WriteString(formatSpecHeading(specification.heading.value))
	formattedText.WriteString(formatItems(specification.items))
	return string(formattedText.Bytes())
}
