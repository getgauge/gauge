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
)

type simpleLogger struct {
	indentation int
}

func newSimpleConsoleWriter() *simpleLogger {
	return &simpleLogger{indentation: 0}
}

func (writer *simpleLogger) Write(b []byte) (int, error) {
	message := indent(string(b), writer.indentation)
	fmt.Print(message)
	return len(b), nil
}

func (writer *simpleLogger) Text(value string) {
	writer.Write([]byte(value))
}

func (writer *simpleLogger) PrintError(value string) {
	writer.Text(value)
}

func (writer *simpleLogger) Critical(formatString string, args ...interface{}) {
	logger.Log.Critical(formatString, args...)
}

func (writer *simpleLogger) Info(formatString string, args ...interface{}) {
	logger.Log.Info(formatString, args...)
}

func (writer *simpleLogger) Warning(formatString string, args ...interface{}) {
	logger.Log.Warning(formatString, args...)
}

func (writer *simpleLogger) Debug(formatString string, args ...interface{}) {
	logger.Log.Debug(formatString, args...)
}

func (writer *simpleLogger) Error(formatString string, args ...interface{}) {
	logger.Log.Error(formatString, args...)
}

func (writer *simpleLogger) SpecHeading(heading string) {
	formattedHeading := formatSpecHeading(heading)
	writer.Write([]byte(formattedHeading))
}

func (writer *simpleLogger) writeItems(items []item) {
	for _, item := range items {
		writer.writeItem(item)
	}
}

func (writer *simpleLogger) Steps(steps []*step) {
	for _, step := range steps {
		writer.writeItem(step)
	}
}

func (writer *simpleLogger) writeItem(item item) {
	writeItem(item, writer)
}

func (writer *simpleLogger) Comment(comment *comment) {
	writer.Text(formatComment(comment))
}

func (writer *simpleLogger) ScenarioHeading(scenarioHeading string) {
	formattedHeading := formatScenarioHeading(scenarioHeading)
	writer.Write([]byte(fmt.Sprintf("\n%s", formattedHeading)))
}

func (writer *simpleLogger) Step(step *step) {
	writer.Text(formatStep(step))
}

func (writer *simpleLogger) StepStarting(step *step) {
}

//todo: pass protostep instead
func (writer *simpleLogger) StepFinished(step *step, failed bool) {
	StepFinished(step, failed, writer)
}

func (writer *simpleLogger) Table(table *table) {
	writer.Text(formatTable(table))
}

func (writer *simpleLogger) ConceptStarting(protoConcept *gauge_messages.ProtoConcept) {
	writer.Text(formatConcept(protoConcept))
	writer.indentation += 4
}

func (writer *simpleLogger) ConceptFinished(protoConcept *gauge_messages.ProtoConcept) {
	writer.indentation -= 4
}
