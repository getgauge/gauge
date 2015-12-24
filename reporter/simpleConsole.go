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
)

type simpleConsole struct {
	indentation int
	writer      io.Writer
}

func newSimpleConsole(out io.Writer) *simpleConsole {
	return &simpleConsole{writer: out}
}

func (sc *simpleConsole) SpecStart(heading string) {
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", formatSpec(heading), newline))
}

func (sc *simpleConsole) SpecEnd() {
	fmt.Fprintln(sc.writer)
}

func (sc *simpleConsole) ScenarioStart(heading string) {
	sc.indentation += scenarioIndentation
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(formatScenario(heading), sc.indentation), newline))
}

func (sc *simpleConsole) ScenarioEnd(failed bool) {
	sc.indentation -= scenarioIndentation
}

func (sc *simpleConsole) StepStart(stepText string) {
	sc.indentation += stepIndentation
	if Verbose {
		fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(stepText, sc.indentation), newline))
	}
}

func (sc *simpleConsole) StepEnd(failed bool) {
	sc.indentation -= stepIndentation
}

func (sc *simpleConsole) ConceptStart(conceptHeading string) {
	sc.indentation += stepIndentation
	if Verbose {
		fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(conceptHeading, sc.indentation), newline))
	}
}

func (sc *simpleConsole) ConceptEnd(failed bool) {
	sc.indentation -= stepIndentation
}

func (sc *simpleConsole) DataTable(table string) {
	if Verbose {
		fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(table, sc.indentation), newline))
	}
}

func (sc *simpleConsole) Error(err string, args ...interface{}) {
	errorString := indent(fmt.Sprintf(err, args...), sc.indentation+sysoutIndentation)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", errorString, newline))
}

func (sc *simpleConsole) Write(b []byte) (int, error) {
	if Verbose {
		formattedString := fmt.Sprintf("%s%s", indent(string(b), sc.indentation+sysoutIndentation), newline)
		fmt.Fprint(sc.writer, formattedString)
		return len(formattedString), nil
	}
	return 0, nil
}
