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
	"os"
)

// SimpleConsoleOutput represents if coloring should be removed from the Console output
var SimpleConsoleOutput bool

// Verbose represents level of console Reporting. If true its at step level, else at scenario level.
var Verbose bool

const newline = "\n"

// Reporter reports the progress of spec execution. It reports
// 1. Which spec / scenarion / step (if verbose) is currently executing.
// 2. Status (pass/fail) of the spec / scenario / step (if verbose) once its executed.
type Reporter interface {
	SpecStart(string)
	SpecEnd()
	ScenarioStart(string)
	ScenarioEnd(bool)
	StepStart(string)
	StepEnd(bool)
	ConceptStart(string)
	ConceptEnd(bool)
	DataTable(string)

	Error(string, ...interface{})

	io.Writer
}

var currentReporter Reporter

// Current returns the current instance of Reporter, if present. Else, it returns a new Reporter.
func Current() Reporter {
	if currentReporter == nil {
		if SimpleConsoleOutput {
			currentReporter = newSimpleConsole(os.Stdout)
		} else {
			currentReporter = newColoredConsole(os.Stdout)
		}
	}
	return currentReporter
}

type parallelReportWriter struct {
	nRunner int
}

func (p *parallelReportWriter) Write(b []byte) (int, error) {
	return fmt.Printf("[runner: %d] %s", p.nRunner, string(b))
}

// NewParallelConsole returns the instance of parallel console reporter
func NewParallelConsole(n int) Reporter {
	writer := &parallelReportWriter{nRunner: n}
	return newSimpleConsole(writer)
}
