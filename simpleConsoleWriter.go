package main

import (
	"fmt"
	"github.com/getgauge/gauge/gauge_messages"
)

type simpleConsoleWriter struct{}

func newSimpleConsoleWriter() *simpleConsoleWriter {
	return &simpleConsoleWriter{}
}

func (writer *simpleConsoleWriter) Write(b []byte) (int, error) {
	fmt.Print(string(b))
	return len(b), nil
}

func (writer *simpleConsoleWriter) writeString(value string) {
	writer.Write([]byte(value))
}

func (writer *simpleConsoleWriter) writeError(value string) {
	writer.writeString(value)
}

func (writer *simpleConsoleWriter) writeSpecHeading(heading string) {
	formattedHeading := formatSpecHeading(heading)
	writer.Write([]byte(formattedHeading))
}

func (writer *simpleConsoleWriter) writeItems(items []item) {
	for _, item := range items {
		writer.writeItem(item)
	}
}

func (writer *simpleConsoleWriter) writeSteps(steps []*step) {
	for _, step := range steps {
		writer.writeItem(step)
	}
}

func (writer *simpleConsoleWriter) writeItem(item item) {
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

func (writer *simpleConsoleWriter) writeComment(comment *comment) {
	writer.writeString(formatItem(comment))
}

func (writer *simpleConsoleWriter) writeScenarioHeading(scenarioHeading string) {
	formattedHeading := formatScenarioHeading(scenarioHeading)
	writer.Write([]byte(fmt.Sprintf("\n%s", formattedHeading)))
}

func (writer *simpleConsoleWriter) writeStep(step *step) {
	writer.writeString(formatItem(step))
}

func (writer *simpleConsoleWriter) writeStepStarting(step *step) {
	writer.writeString(fmt.Sprintf("Executing.. => %s", formatItem(step)))
}

//todo: pass protostep instead
func (writer *simpleConsoleWriter) writeStepFinished(step *step, failed bool) {
	var message string
	if failed {
		message = fmt.Sprintf("Step Failed => %s\n", formatItem(step))
	} else {
		message = fmt.Sprintf("Step Passed => %s\n", formatItem(step))
	}
	writer.writeString(message)
}

func (writer *simpleConsoleWriter) writeTable(table *table) {
	writer.writeString(formatTable(table))
}

func (writer *simpleConsoleWriter) writeConceptStarting(protoConcept *gauge_messages.ProtoConcept) {
	writer.writeString(formatConcept(protoConcept))
}

func (writer *simpleConsoleWriter) writeConceptFinished(protoConcept *gauge_messages.ProtoConcept) {
	writer.writeString(formatConcept(protoConcept))
}
