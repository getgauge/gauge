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

package reporter

import (
	"fmt"
	"io"

	"github.com/getgauge/gauge/logger"
)

type simpleConsole struct {
	indentation int
	writer      io.Writer
}

func newSimpleConsole(out io.Writer) *simpleConsole {
	return &simpleConsole{writer: out}
}

func (sc *simpleConsole) SpecStart(heading string) {
	formattedHeading := formatSpec(heading)
	logger.GaugeLog.Info(formattedHeading)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", formattedHeading, newline))
}

func (sc *simpleConsole) SpecEnd() {
	fmt.Fprintln(sc.writer)
}

func (sc *simpleConsole) ScenarioStart(heading string) {
	sc.indentation += scenarioIndentation
	formattedHeading := formatScenario(heading)
	logger.GaugeLog.Info(formattedHeading)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(formattedHeading, sc.indentation), newline))
}

func (sc *simpleConsole) ScenarioEnd(failed bool) {
	sc.indentation -= scenarioIndentation
}

func (sc *simpleConsole) StepStart(stepText string) {
	sc.indentation += stepIndentation
	logger.GaugeLog.Debug(stepText)
	if Verbose {
		fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(stepText, sc.indentation), newline))
	}
}

func (sc *simpleConsole) StepEnd(failed bool) {
	sc.indentation -= stepIndentation
}

func (sc *simpleConsole) ConceptStart(conceptHeading string) {
	sc.indentation += stepIndentation
	logger.GaugeLog.Debug(conceptHeading)
	if Verbose {
		fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(conceptHeading, sc.indentation), newline))
	}
}

func (sc *simpleConsole) ConceptEnd(failed bool) {
	sc.indentation -= stepIndentation
}

func (sc *simpleConsole) DataTable(table string) {
	logger.GaugeLog.Debug(table)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", newline, table))
}

func (sc *simpleConsole) Error(err string, args ...interface{}) {
	errorMessage := fmt.Sprintf(err, args...)
	logger.GaugeLog.Error(errorMessage)
	errorString := indent(errorMessage, sc.indentation+sysoutIndentation)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", errorString, newline))
}

func (sc *simpleConsole) Write(b []byte) (int, error) {
	if Verbose {
		formattedString := fmt.Sprintf("%s%s", spaces(sc.indentation+sysoutIndentation), string(b))
		fmt.Fprint(sc.writer, formattedString)
		return len(b), nil
	}
	return 0, nil
}
