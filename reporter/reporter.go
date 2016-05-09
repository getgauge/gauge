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

	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
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

	Errorf(string, ...interface{})

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

// ListenExecutionEvents listens to all execution events for reporting on console
func ListenExecutionEvents() {
	ch := make(chan event.ExecutionEvent, 0)
	event.Register(ch, event.SpecStart, event.SpecEnd, event.ScenarioStart, event.ScenarioEnd, event.StepStart, event.StepEnd)

	go func() {
		for {
			e := <-ch
			switch e.Topic {
			case event.SpecStart:
				Current().SpecStart(e.Item.(*gauge.Specification).Heading.Value)
			case event.ScenarioStart:
				Current().ScenarioStart(e.Item.(*gauge.Scenario).Heading.Value)
			case event.StepStart:
				Current().StepStart(formatter.FormatStep(e.Item.(*gauge.Step)))
			case event.StepEnd:
				Current().StepEnd(e.Result.(*result.StepResult).GetFailed())
			case event.ScenarioEnd:
				Current().ScenarioEnd(e.Result.(*result.ScenarioResult).GetFailed())
			case event.SpecEnd:
				Current().SpecEnd()
			}
		}
	}()
}
