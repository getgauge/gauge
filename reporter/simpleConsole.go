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
	"strings"
	"sync"

	"github.com/getgauge/gauge/logger"
)

type simpleConsole struct {
	mu          *sync.Mutex
	indentation int
	writer      io.Writer
}

func newSimpleConsole(out io.Writer) *simpleConsole {
	return &simpleConsole{mu: &sync.Mutex{}, writer: out}
}

func (sc *simpleConsole) SpecStart(heading string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	formattedHeading := formatSpec(heading)
	logger.GaugeLog.Info(formattedHeading)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", formattedHeading, newline))
}

func (sc *simpleConsole) SpecEnd() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	fmt.Fprintln(sc.writer)
}

func (sc *simpleConsole) ScenarioStart(heading string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.indentation += scenarioIndentation
	formattedHeading := formatScenario(heading)
	logger.GaugeLog.Info(formattedHeading)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(formattedHeading, sc.indentation), newline))
}

func (sc *simpleConsole) ScenarioEnd(failed bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.indentation -= scenarioIndentation
}

func (sc *simpleConsole) StepStart(stepText string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.indentation += stepIndentation
	logger.GaugeLog.Debug(stepText)
	if Verbose {
		fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(strings.TrimSpace(stepText), sc.indentation), newline))
	}
}

func (sc *simpleConsole) StepEnd(failed bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.indentation -= stepIndentation
}

func (sc *simpleConsole) ConceptStart(conceptHeading string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.indentation += stepIndentation
	logger.GaugeLog.Debug(conceptHeading)
	if Verbose {
		fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(conceptHeading, sc.indentation), newline))
	}
}

func (sc *simpleConsole) ConceptEnd(failed bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.indentation -= stepIndentation
}

func (sc *simpleConsole) DataTable(table string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	logger.GaugeLog.Debug(table)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", newline, table))
}

func (sc *simpleConsole) Error(err string, args ...interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	errorMessage := fmt.Sprintf(err, args...)
	logger.GaugeLog.Error(errorMessage)
	errorString := indent(errorMessage, sc.indentation+errorIndentation)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", errorString, newline))
}

func (sc *simpleConsole) Write(b []byte) (int, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	formattedString := string(b)
	fmt.Fprint(sc.writer, formattedString)
	return len(b), nil
}
