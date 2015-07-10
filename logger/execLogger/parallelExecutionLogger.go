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

package execLogger

import (
	"fmt"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
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
	logger.Log.Critical(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) Info(formatString string, args ...interface{}) {
	logger.Log.Info(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) Warning(formatString string, args ...interface{}) {
	logger.Log.Warning(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) Debug(formatString string, args ...interface{}) {
	logger.Log.Debug(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) Error(formatString string, args ...interface{}) {
	logger.Log.Error(addPrefixToEachLine(fmt.Sprintf("[%s] : ", writer.name), formatString), args)
}

func (writer *parallelExecutionLogger) SpecHeading(heading string) {
	formattedHeading := fmt.Sprintf("Executing specification => %s \n", heading)
	writer.Write([]byte(formattedHeading))
}

func (writer *parallelExecutionLogger) Steps(steps []*parser.Step) {
	for _, step := range steps {
		writer.Step(step)
	}
}

func (writer *parallelExecutionLogger) Comment(comment *parser.Comment) {
	writer.Text(formatter.FormatComment(comment))
}

func (writer *parallelExecutionLogger) ScenarioHeading(scenarioHeading string) {
	formattedHeading := fmt.Sprintf("Executing scenario => %s \n", scenarioHeading)
	writer.Write([]byte(fmt.Sprintf("\n%s", formattedHeading)))
}

func (writer *parallelExecutionLogger) Step(step *parser.Step) {
}

func (writer *parallelExecutionLogger) StepStarting(step *parser.Step) {
}

//todo: pass protostep instead
func (writer *parallelExecutionLogger) StepFinished(step *parser.Step, failed bool) {
	StepFinished(step, failed, writer)
}

func StepFinished(step *parser.Step, failed bool, writer ExecutionLogger) {
	var message string
	if failed {
		message = fmt.Sprintf("Step Failed => %s\n", formatter.FormatStep(step))
	} else {
		message = fmt.Sprintf("Step Passed => %s\n", formatter.FormatStep(step))
	}
	writer.Text(message)
}

func (writer *parallelExecutionLogger) Table(table *parser.Table) {
	writer.Text(formatter.FormatTable(table))
}

func (writer *parallelExecutionLogger) ConceptStarting(protoConcept *gauge_messages.ProtoConcept) {
	writer.Text(formatter.FormatConcept(protoConcept))
	writer.indentation += 4
}

func (writer *parallelExecutionLogger) ConceptFinished(protoConcept *gauge_messages.ProtoConcept) {
	writer.indentation -= 4
}
