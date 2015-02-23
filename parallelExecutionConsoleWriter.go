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
	"strconv"
)

type parallelExecutionConsoleWriter struct {
	name        string
	indentation int
}

func newParallelExecutionConsoleWriter(id int) *parallelExecutionConsoleWriter {
	return &parallelExecutionConsoleWriter{indentation: 0, name: "Worker:" + strconv.Itoa(id)}
}

func (writer *parallelExecutionConsoleWriter) Write(b []byte) (int, error) {
	message := indent(string(b), writer.indentation)
	fmt.Print(addPrefixToEachLine(message, fmt.Sprintf("[%s] : ", writer.name)))
	return len(b), nil
}

func (writer *parallelExecutionConsoleWriter) writeString(value string) {
	writer.Write([]byte(value))
}

func (writer *parallelExecutionConsoleWriter) writeError(value string) {
	writer.writeString(value)
}

func (writer *parallelExecutionConsoleWriter) writeSpecHeading(heading string) {
	formattedHeading := fmt.Sprintf("Executing specification => %s \n", heading)
	writer.Write([]byte(formattedHeading))
}

func (writer *parallelExecutionConsoleWriter) writeItems(items []item) {
	for _, item := range items {
		writer.writeItem(item)
	}
}

func (writer *parallelExecutionConsoleWriter) writeSteps(steps []*step) {
	for _, step := range steps {
		writer.writeItem(step)
	}
}

func (writer *parallelExecutionConsoleWriter) writeItem(item item) {
	writeItem(item,writer)
}

func writeItem(item item, writer consoleWriter){
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

func (writer *parallelExecutionConsoleWriter) writeComment(comment *comment) {
	writer.writeString(formatComment(comment))
}

func (writer *parallelExecutionConsoleWriter) writeScenarioHeading(scenarioHeading string) {
	formattedHeading := fmt.Sprintf("Executing scenario => %s \n", scenarioHeading)
	writer.Write([]byte(fmt.Sprintf("\n%s", formattedHeading)))
}

func (writer *parallelExecutionConsoleWriter) writeStep(step *step) {
}

func (writer *parallelExecutionConsoleWriter) writeStepStarting(step *step) {
}

//todo: pass protostep instead
func (writer *parallelExecutionConsoleWriter) writeStepFinished(step *step, failed bool) {
	writeStepFinished(step,failed,writer)
}

func writeStepFinished(step *step, failed bool, writer consoleWriter) {
	var message string
	if failed {
		message = fmt.Sprintf("Step Failed => %s\n", formatStep(step))
	} else {
		message = fmt.Sprintf("Step Passed => %s\n", formatStep(step))
	}
	writer.writeString(message)
}

func (writer *parallelExecutionConsoleWriter) writeTable(table *table) {
	writer.writeString(formatTable(table))
}

func (writer *parallelExecutionConsoleWriter) writeConceptStarting(protoConcept *gauge_messages.ProtoConcept) {
	writer.writeString(formatConcept(protoConcept))
	writer.indentation += 4
}

func (writer *parallelExecutionConsoleWriter) writeConceptFinished(protoConcept *gauge_messages.ProtoConcept) {
	writer.indentation -= 4
}
