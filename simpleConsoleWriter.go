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
)

type simpleConsoleWriter struct {
	indentation int
}

func newSimpleConsoleWriter() *simpleConsoleWriter {
	return &simpleConsoleWriter{indentation: 0}
}

func (writer *simpleConsoleWriter) Write(b []byte) (int, error) {
	message := indent(string(b), writer.indentation)
	fmt.Print(message)
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
	writer.writeString(formatComment(comment))
}

func (writer *simpleConsoleWriter) writeScenarioHeading(scenarioHeading string) {
	formattedHeading := formatScenarioHeading(scenarioHeading)
	writer.Write([]byte(fmt.Sprintf("\n%s", formattedHeading)))
}

func (writer *simpleConsoleWriter) writeStep(step *step) {
	writer.writeString(formatStep(step))
}

func (writer *simpleConsoleWriter) writeStepStarting(step *step) {
}

//todo: pass protostep instead
func (writer *simpleConsoleWriter) writeStepFinished(step *step, failed bool) {
	var message string
	if failed {
		message = fmt.Sprintf("Step Failed => %s\n", formatStep(step))
	} else {
		message = fmt.Sprintf("Step Passed => %s\n", formatStep(step))
	}
	writer.writeString(message)
}

func (writer *simpleConsoleWriter) writeTable(table *table) {
	writer.writeString(formatTable(table))
}

func (writer *simpleConsoleWriter) writeConceptStarting(protoConcept *gauge_messages.ProtoConcept) {
	writer.writeString(formatConcept(protoConcept))
	writer.indentation += 4
}

func (writer *simpleConsoleWriter) writeConceptFinished(protoConcept *gauge_messages.ProtoConcept) {
	writer.indentation -= 4
}
