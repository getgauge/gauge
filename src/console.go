package main

import (
	"fmt"
	"github.com/wsxiaoys/terminal"
	"strings"
)

type consoleWriter struct {
	linesAfterLastStep int
	isInsideStep       bool
}

func newConsoleWriter() *consoleWriter {
	return &consoleWriter{linesAfterLastStep: 0, isInsideStep: false}
}

var currentConsoleWriter *consoleWriter

func getCurrentConsole() *consoleWriter {
	if currentConsoleWriter == nil {
		currentConsoleWriter = newConsoleWriter()
	}
	return currentConsoleWriter
}

func (writer *consoleWriter) Write(b []byte) (int, error) {
	message := string(b)
	if writer.isInsideStep {
		writer.linesAfterLastStep += strings.Count(message, "\n")
	}
	fmt.Print(message)
	return len(b), nil
}

func (writer *consoleWriter) writeString(value string) {
	writer.Write([]byte(value))
}

func (writer *consoleWriter) writeError(value string) {
	if writer.isInsideStep {
		writer.linesAfterLastStep += strings.Count(value, "\n")
	}
	terminal.Stdout.Colorf("@r%s", value)
}

func (writer *consoleWriter) writeSpecHeading(spec *specification) {
	formattedHeading := formatSpecHeading(spec.heading.value)
	writer.Write([]byte(formattedHeading))
}

func (writer *consoleWriter) writeItems(items []item) {
	for _, item := range items {
		writer.writeItem(item)
	}
}

func (writer *consoleWriter) writeItem(item item) {
	switch item.kind() {
	case commentKind:
		comment := item.(*comment)
		writer.writeComment(comment)
	case stepKind:
		step := item.(*step)
		writer.writeStep(step)
	case tableKind:
		table := item.(*table)
		writer.writeTable(table)
	}
}

func (writer *consoleWriter) writeComment(comment *comment) {
	terminal.Stdout.Colorf("@k%s\n\n", comment.value)
}

func (writer *consoleWriter) writeScenarioHeading(scenarioHeading string) {
	formattedHeading := formatScenarioHeading(scenarioHeading)
	writer.Write([]byte(formattedHeading))
}

func (writer *consoleWriter) writeStep(step *step) {
	terminal.Stdout.Colorf("@b%s", formatStep(step))
	writer.isInsideStep = true
	writer.linesAfterLastStep = 0
}

func (writer *consoleWriter) writeStepFinished(step *step, isPassed bool) {
	stepText := formatStep(step)
	linesInStepText := strings.Count(stepText, "\n")
	if linesInStepText == 0 {
		linesInStepText = 1
	}
	terminal.Stdout.Up(writer.linesAfterLastStep + linesInStepText)
	if isPassed {
		terminal.Stdout.Colorf("@g%s", stepText)
	} else {
		terminal.Stdout.Colorf("@r%s", stepText)
	}
	terminal.Stdout.Down(writer.linesAfterLastStep + linesInStepText)
	writer.isInsideStep = false
}

func (writer *consoleWriter) writeTable(table *table) {
	formattedTable := formatTable(table)
	terminal.Stdout.Colorf("@m%s", formattedTable)
}
