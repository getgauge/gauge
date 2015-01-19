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
}

type pluginConsoleWriter struct {
	pluginName string
}

func (writer *pluginConsoleWriter) Write(b []byte) (int, error) {
	message := string(b)
	prefixedMessage := writer.addPrefixToEachLine(message)
	gaugeConsoleWriter := getCurrentConsole()
	_, err := gaugeConsoleWriter.Write([]byte(prefixedMessage))
	return len(message), err
}

func (writer *pluginConsoleWriter) addPrefixToEachLine(text string) string {
	lines := strings.Split(text, "\n")
	prefixedLines := make([]string, 0)
	for i, line := range lines {
		if (i == len(lines)-1) && line == "" {
			prefixedLines = append(prefixedLines, line)
		} else {
			prefixedLines = append(prefixedLines, fmt.Sprintf("[%s Plugin] : %s", writer.pluginName, line))
		}
	}
	return strings.Join(prefixedLines, "\n")
}

func newColoredConsoleWriter() *coloredConsoleWriter {
	return &coloredConsoleWriter{linesAfterLastStep: 0, isInsideStep: false}
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
	message := string(b)
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
	terminal.Stdout.Colorf("%s", formatItem(comment))
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
	conceptText := formatConcept(protoConcept)
	terminal.Stdout.Colorf("@b%s", conceptText)
}

func (writer *coloredConsoleWriter) writeConceptFinished(protoConcept *gauge_messages.ProtoConcept) {
	conceptText := formatConcept(protoConcept)
	terminal.Stdout.Colorf("@g%s", conceptText)
}

func (writer *coloredConsoleWriter) writeStepStarting(step *step) {
	stepText := formatStep(step)
	terminal.Stdout.Colorf("@b%s", stepText)
	writer.isInsideStep = true
	writer.linesAfterLastStep = 0
}

//todo: pass protostep instead
func (writer *coloredConsoleWriter) writeStepFinished(step *step, failed bool) {
	stepText := formatStep(step)
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
	formattedTable := formatTable(table)
	terminal.Stdout.Colorf("@m%s", formattedTable)
}
