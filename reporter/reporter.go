/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package reporter

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"sync"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
)

// IsParallel represents console reporting format based on simple/parallel execution
var IsParallel bool

// NumberOfExecutionStreams indicates the total number of parallel execution streams
var NumberOfExecutionStreams int

// SimpleConsoleOutput represents if coloring should be removed from the Console output
var SimpleConsoleOutput bool

// Verbose represents level of console Reporting. If true its at step level, else at scenario level.
var Verbose bool

// MachineReadable represents if output should be in JSON format.
var MachineReadable bool

const newline = "\n"

// Reporter reports the progress of spec execution. It reports
// 1. Which spec / scenarion / step (if verbose) is currently executing.
// 2. Status (pass/fail) of the spec / scenario / step (if verbose) once its executed.
type Reporter interface {
	SuiteStart()
	SpecStart(*gauge.Specification, result.Result)
	SpecEnd(*gauge.Specification, result.Result)
	ScenarioStart(*gauge.Scenario, *gauge_messages.ExecutionInfo, result.Result)
	ScenarioEnd(*gauge.Scenario, result.Result, *gauge_messages.ExecutionInfo)
	StepStart(string)
	StepEnd(gauge.Step, result.Result, *gauge_messages.ExecutionInfo)
	ConceptStart(string)
	ConceptEnd(result.Result)
	DataTable(string)
	SuiteEnd(result.Result)

	Errorf(string, ...interface{})

	io.Writer
}

var currentReporter Reporter

func reporter(e event.ExecutionEvent) Reporter {
	if IsParallel {
		return ParallelReporter(e.Stream)
	}
	return Current()
}

// Current returns the current instance of Reporter, if present. Else, it returns a new Reporter.
func Current() Reporter {
	if currentReporter == nil {
		if MachineReadable {
			currentReporter = newJSONConsole(os.Stdout, IsParallel, 0)
		} else if SimpleConsoleOutput {
			currentReporter = newSimpleConsole(os.Stdout)
		} else if Verbose {
			currentReporter = newVerboseColoredConsole(os.Stdout)
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

// ParallelReporter returns the instance of parallel console reporter
func ParallelReporter(n int) Reporter {
	if r, ok := parallelReporters[n]; ok {
		return r
	}
	return Current()
}

var parallelReporters map[int]Reporter

func initParallelReporters() {
	parallelReporters = make(map[int]Reporter, NumberOfExecutionStreams)
	for i := 1; i <= NumberOfExecutionStreams; i++ {
		if MachineReadable {
			parallelReporters[i] = newJSONConsole(os.Stdout, true, i)
		} else {
			writer := &parallelReportWriter{nRunner: i}
			parallelReporters[i] = newSimpleConsole(writer)
		}
	}
}

// ListenExecutionEvents listens to all execution events for reporting on console
func ListenExecutionEvents(wg *sync.WaitGroup) {
	ch := make(chan event.ExecutionEvent)
	initParallelReporters()
	event.Register(ch, event.SuiteStart, event.SpecStart, event.SpecEnd, event.ScenarioStart, event.ScenarioEnd, event.StepStart, event.StepEnd, event.ConceptStart, event.ConceptEnd, event.SuiteEnd)
	var r Reporter
	wg.Add(1)

	go func() {
		defer recoverPanic()
		for {
			e := <-ch
			r = reporter(e)
			switch e.Topic {
			case event.SuiteStart:
				r.SuiteStart()
			case event.SpecStart:
				r.SpecStart(e.Item.(*gauge.Specification), e.Result)
			case event.ScenarioStart:
				skipped := e.Result.(*result.ScenarioResult).ProtoScenario.GetExecutionStatus() == gauge_messages.ExecutionStatus_SKIPPED
				sce := e.Item.(*gauge.Scenario)
				// if it is datatable driven execution
				if !skipped {
					if sce.SpecDataTableRow.GetRowCount() != 0 {
						r.DataTable(formatter.FormatTable(&sce.SpecDataTableRow))
					}
					if sce.ScenarioDataTableRow.GetRowCount() != 0 {
						r.DataTable(formatter.FormatTable(&sce.ScenarioDataTableRow))
					}
				}
				r.ScenarioStart(sce, e.ExecutionInfo, e.Result)
			case event.ConceptStart:
				r.ConceptStart(formatter.FormatStep(e.Item.(*gauge.Step)))
			case event.StepStart:
				r.StepStart(formatter.FormatStepWithResolvedArgs(e.Item.(*gauge.Step)))
			case event.StepEnd:
				r.StepEnd(e.Item.(gauge.Step), e.Result, e.ExecutionInfo)
			case event.ConceptEnd:
				r.ConceptEnd(e.Result)
			case event.ScenarioEnd:
				r.ScenarioEnd(e.Item.(*gauge.Scenario), e.Result, e.ExecutionInfo)
			case event.SpecEnd:
				r.SpecEnd(e.Item.(*gauge.Specification), e.Result)
			case event.SuiteEnd:
				r.SuiteEnd(e.Result)
				wg.Done()
			}
		}
	}()
}

func recoverPanic() {
	if r := recover(); r != nil {
		logger.Fatalf(true, "%v\n%s", r, string(debug.Stack()))
	}
}
