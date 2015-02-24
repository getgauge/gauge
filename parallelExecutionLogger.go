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
	"strings"
)

const WORKER = "Worker:"

type parallelExecutionLogger struct {
	name        string
	indentation int
}

func newParallelExecutionConsoleWriter(id int) *parallelExecutionLogger {
	return &parallelExecutionLogger{indentation: 0, name: WORKER + strconv.Itoa(id)}
}

func (writer *parallelExecutionLogger) Write(b []byte) (int, error) {
	message := indent(string(b), writer.indentation)
	value := addPrefixToEachLine(message, fmt.Sprintf("[%s] : ", writer.name))
	if strings.TrimSpace(message) == "" {
		value = message
	}
	fmt.Print(value)
	return len(b), nil
}

func (writer *parallelExecutionLogger) Text(value string) {
	writer.Write([]byte(value))
}

func (writer *parallelExecutionLogger) PrintError(value string) {
	writer.Text(value)
}

func (writer *parallelExecutionLogger) Critical(formatString string, args ...interface{}) {
	log.Critical(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) Info(formatString string, args ...interface{}) {
	log.Info(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) Warning(formatString string, args ...interface{}) {
	log.Warning(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) Debug(formatString string, args ...interface{}) {
	log.Debug(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) Error(formatString string, args ...interface{}) {
	log.Error(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) SpecHeading(heading string) {
	formattedHeading := fmt.Sprintf("Executing specification => %s \n", heading)
	writer.Write([]byte(formattedHeading))
}

func (writer *parallelExecutionLogger) writeItems(items []item) {
	for _, item := range items {
		writer.writeItem(item)
	}
}

func (writer *parallelExecutionLogger) Steps(steps []*step) {
	for _, step := range steps {
		writer.writeItem(step)
	}
}

func (writer *parallelExecutionLogger) writeItem(item item) {
	writeItem(item, writer)
}

func writeItem(item item, writer executionLogger) {
	switch item.kind() {
	case commentKind:
		comment := item.(*comment)
		writer.Comment(comment)
	case stepKind:
		step := item.(*step)
		writer.Step(step)
	case tableKind:
		table := item.(*table)
		writer.Table(table)
	}
}

func (writer *parallelExecutionLogger) Comment(comment *comment) {
	writer.Text(formatComment(comment))
}

func (writer *parallelExecutionLogger) ScenarioHeading(scenarioHeading string) {
	formattedHeading := fmt.Sprintf("Executing scenario => %s \n", scenarioHeading)
	writer.Write([]byte(fmt.Sprintf("\n%s", formattedHeading)))
}

func (writer *parallelExecutionLogger) Step(step *step) {
}

func (writer *parallelExecutionLogger) StepStarting(step *step) {
}

//todo: pass protostep instead
func (writer *parallelExecutionLogger) StepFinished(step *step, failed bool) {
	StepFinished(step, failed, writer)
}

func StepFinished(step *step, failed bool, writer executionLogger) {
	var message string
	if failed {
		message = fmt.Sprintf("Step Failed => %s\n", formatStep(step))
	} else {
		message = fmt.Sprintf("Step Passed => %s\n", formatStep(step))
	}
	writer.Text(message)
}

func (writer *parallelExecutionLogger) Table(table *table) {
	writer.Text(formatTable(table))
}

func (writer *parallelExecutionLogger) ConceptStarting(protoConcept *gauge_messages.ProtoConcept) {
	writer.Text(formatConcept(protoConcept))
	writer.indentation += 4
}

func (writer *parallelExecutionLogger) ConceptFinished(protoConcept *gauge_messages.ProtoConcept) {
	writer.indentation -= 4
}
