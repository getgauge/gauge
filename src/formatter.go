package main

import (
	"bytes"
	"fmt"
)

const (
	HEADING_UNDERLINE_LENGTH = 20
	TABLE_LEFT_SPACING       = 10
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
	return fmt.Sprintf("\n%s", formatHeading(scenarioHeading, "-"))
}

func formatStepText(stepText string) string {
	return fmt.Sprintf("* %s\n", stepText)
}

func formatHeading(heading, headingChar string) string {
	length := len(heading)
	if length > HEADING_UNDERLINE_LENGTH {
		length = HEADING_UNDERLINE_LENGTH
	}

	return fmt.Sprintf("%s\n%s\n\n", heading, getRepeatedChars(headingChar, length))
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
