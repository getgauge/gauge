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
	"github.com/getgauge/gauge/logger"
	"github.com/wsxiaoys/terminal"
	"strings"
)

type executionLogger interface {
	Write([]byte) (int, error)
	Text(string)
	PrintError(string)
	SpecHeading(string)
	ScenarioHeading(string)
	Comment(*comment)
	Step(*step)
	StepStarting(*step)
	StepFinished(*step, bool)
	Table(*table)
	Critical(string, ...interface{})
	Warning(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
	Error(string, ...interface{})
	ConceptStarting(*gauge_messages.ProtoConcept)
	ConceptFinished(*gauge_messages.ProtoConcept)
}

var currentLogger executionLogger

type coloredLogger struct {
	linesAfterLastStep int
	isInsideStep       bool
	indentation        int
}

type pluginLogger struct {
	pluginName string
}

func (writer *pluginLogger) Write(b []byte) (int, error) {
	message := string(b)
	prefixedMessage := addPrefixToEachLine(message, fmt.Sprintf("[%s Plugin] : ", writer.pluginName))
	gaugeConsoleWriter := getCurrentExecutionLogger()
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

func newColoredConsoleWriter() *coloredLogger {
	return &coloredLogger{linesAfterLastStep: 0, isInsideStep: false, indentation: 0}
}

func getCurrentExecutionLogger() executionLogger {
	if currentLogger == nil {
		if *simpleConsoleOutput {
			currentLogger = newSimpleConsoleWriter()
		} else {
			currentLogger = newColoredConsoleWriter()
		}
	}
	return currentLogger
}

func (writer *coloredLogger) Write(b []byte) (int, error) {
	message := indent(string(b), writer.indentation)
	if writer.isInsideStep {
		writer.linesAfterLastStep += strings.Count(message, "\n")
	}
	fmt.Print(message)
	return len(b), nil
}

func (writer *coloredLogger) Text(value string) {
	writer.Write([]byte(value))
}

func (writer *coloredLogger) PrintError(value string) {
	if writer.isInsideStep {
		writer.linesAfterLastStep += strings.Count(value, "\n")
	}
	terminal.Stdout.Colorf("@r%s", value)
}

func (writer *coloredLogger) Critical(formatString string, args ...interface{}) {
	logger.Log.Critical(formatString, args)
}

func (writer *coloredLogger) Info(formatString string, args ...interface{}) {
	logger.Log.Info(formatString, args)
}

func (writer *coloredLogger) Warning(formatString string, args ...interface{}) {
	logger.Log.Warning(formatString, args)
}

func (writer *coloredLogger) Debug(formatString string, args ...interface{}) {
	logger.Log.Debug(formatString, args)
}

func (writer *coloredLogger) Error(formatString string, args ...interface{}) {
	logger.Log.Error(formatString, args)
}

func (writer *coloredLogger) SpecHeading(heading string) {
	formattedHeading := formatSpecHeading(heading)
	writer.Write([]byte(formattedHeading))
}

func (writer *coloredLogger) Comment(comment *comment) {
	writer.Write([]byte(formatComment(comment)))
}

func (writer *coloredLogger) ScenarioHeading(scenarioHeading string) {
	formattedHeading := formatScenarioHeading(scenarioHeading)
	writer.Write([]byte(fmt.Sprintf("\n%s", formattedHeading)))
}

func (writer *coloredLogger) writeContextStep(step *step) {
	writer.Step(step)
}

func (writer *coloredLogger) Step(step *step) {
	stepText := formatStep(step)
	terminal.Stdout.Colorf("@b%s", stepText)
	writer.isInsideStep = true
	writer.linesAfterLastStep = 0
}

func (writer *coloredLogger) ConceptStarting(protoConcept *gauge_messages.ProtoConcept) {
	conceptText := indent(formatConcept(protoConcept), writer.indentation)
	terminal.Stdout.Colorf("@b%s", conceptText)
	writer.indentation += 4
}

func (writer *coloredLogger) ConceptFinished(protoConcept *gauge_messages.ProtoConcept) {
	writer.indentation -= 4
}

func (writer *coloredLogger) StepStarting(step *step) {
	stepText := formatStep(step)
	terminal.Stdout.Colorf("@b%s", stepText)
	writer.isInsideStep = true
	writer.linesAfterLastStep = 0
}

//todo: pass protostep instead
func (writer *coloredLogger) StepFinished(step *step, failed bool) {
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

func (writer *coloredLogger) Table(table *table) {
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
