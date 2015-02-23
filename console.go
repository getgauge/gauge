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
	"fmt"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/wsxiaoys/terminal"
	"strings"
)

type consoleWriter interface {
	Write([]byte) (int, error)
	writeString(string)
	writeError(string)
	writeSpecHeading(string)
	writeScenarioHeading(string)
	writeComment(*comment)
	writeStep(*step)
	writeStepStarting(*step)
	writeStepFinished(*step, bool)
	writeTable(*table)
	writeConceptStarting(*gauge_messages.ProtoConcept)
	writeConceptFinished(*gauge_messages.ProtoConcept)
}

var currentConsoleWriter consoleWriter

type coloredConsoleWriter struct {
	linesAfterLastStep int
	isInsideStep       bool
	indentation        int
}

type pluginConsoleWriter struct {
	pluginName string
}

func (writer *pluginConsoleWriter) Write(b []byte) (int, error) {
	message := string(b)
	prefixedMessage := addPrefixToEachLine(message, fmt.Sprintf("[%s Plugin] : ", writer.pluginName))
	gaugeConsoleWriter := getCurrentConsole()
	_, err := gaugeConsoleWriter.Write([]byte(prefixedMessage))
	return len(message), err
}

func addPrefixToEachLine(text string, template string) string {
	lines := strings.Split(text, "\n")
	prefixedLines := make([]string, 0)
	for i, line := range lines {
		if (i == len(lines)-1) && line == "" {
			prefixedLines = append(prefixedLines, line)
		} else {
			prefixedLines = append(prefixedLines, template+line)
		}
	}
	return strings.Join(prefixedLines, "\n")
}

func newColoredConsoleWriter() *coloredConsoleWriter {
	return &coloredConsoleWriter{linesAfterLastStep: 0, isInsideStep: false, indentation: 0}
}

func getCurrentConsole() consoleWriter {
	if currentConsoleWriter == nil {
		if *simpleConsoleOutput {
			currentConsoleWriter = newSimpleConsoleWriter()
		} else {
			currentConsoleWriter = newColoredConsoleWriter()
		}
	}
	return currentConsoleWriter
}

func (writer *coloredConsoleWriter) Write(b []byte) (int, error) {
	message := indent(string(b), writer.indentation)
	if writer.isInsideStep {
		writer.linesAfterLastStep += strings.Count(message, "\n")
	}
	fmt.Print(message)
	return len(b), nil
}

func (writer *coloredConsoleWriter) writeString(value string) {
	writer.Write([]byte(value))
}

func (writer *coloredConsoleWriter) writeError(value string) {
	if writer.isInsideStep {
		writer.linesAfterLastStep += strings.Count(value, "\n")
	}
	terminal.Stdout.Colorf("@r%s", value)
}

func (writer *coloredConsoleWriter) writeSpecHeading(heading string) {
	formattedHeading := formatSpecHeading(heading)
	writer.Write([]byte(formattedHeading))
}

func (writer *coloredConsoleWriter) writeComment(comment *comment) {
	writer.Write([]byte(formatComment(comment)))
}

func (writer *coloredConsoleWriter) writeScenarioHeading(scenarioHeading string) {
	formattedHeading := formatScenarioHeading(scenarioHeading)
	writer.Write([]byte(fmt.Sprintf("\n%s", formattedHeading)))
}

func (writer *coloredConsoleWriter) writeContextStep(step *step) {
	writer.writeStep(step)
}

func (writer *coloredConsoleWriter) writeStep(step *step) {
	stepText := formatStep(step)
	terminal.Stdout.Colorf("@b%s", stepText)
	writer.isInsideStep = true
	writer.linesAfterLastStep = 0
}

func (writer *coloredConsoleWriter) writeConceptStarting(protoConcept *gauge_messages.ProtoConcept) {
	conceptText := indent(formatConcept(protoConcept), writer.indentation)
	terminal.Stdout.Colorf("@b%s", conceptText)
	writer.indentation += 4
}

func (writer *coloredConsoleWriter) writeConceptFinished(protoConcept *gauge_messages.ProtoConcept) {
	writer.indentation -= 4
}

func (writer *coloredConsoleWriter) writeStepStarting(step *step) {
	stepText := formatStep(step)
	terminal.Stdout.Colorf("@b%s", stepText)
	writer.isInsideStep = true
	writer.linesAfterLastStep = 0
}

//todo: pass protostep instead
func (writer *coloredConsoleWriter) writeStepFinished(step *step, failed bool) {
	stepText := indent(formatStep(step), writer.indentation)
	linesInStepText := strings.Count(stepText, "\n")
	if linesInStepText == 0 {
		linesInStepText = 1
	}
	linesToMoveUp := writer.linesAfterLastStep + linesInStepText
	terminal.Stdout.Up(linesToMoveUp)
	if failed {
		terminal.Stdout.Colorf("@r%s", stepText)
	} else {
		terminal.Stdout.Colorf("@g%s", stepText)
	}
	terminal.Stdout.Down(linesToMoveUp)
	writer.isInsideStep = false
}

func (writer *coloredConsoleWriter) writeTable(table *table) {
	formattedTable := indent(formatTable(table), writer.indentation)
	terminal.Stdout.Colorf("@m%s", formattedTable)
}

func indent(message string, indentation int) string {
	if indentation == 0 {
		return message
	}
	lines := strings.Split(message, "\n")
	prefixedLines := make([]string, 0)
	spaces := getEmptySpacedString(indentation)
	for i, line := range lines {
		if (i == len(lines)-1) && line == "" {
			prefixedLines = append(prefixedLines, line)
		} else {
			prefixedLines = append(prefixedLines, spaces+line)
		}
	}
	return strings.Join(prefixedLines, "\n")
}

func getEmptySpacedString(numOfSpaces int) string {
	text := ""
	for i := 0; i < numOfSpaces; i++ {
		text += " "
	}
	return text
}
