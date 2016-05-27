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

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
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

func (sc *simpleConsole) SpecEnd(res result.Result) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	printHookFailureSC(sc, res, res.GetPreHook)
	printHookFailureSC(sc, res, res.GetPostHook)
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

func (sc *simpleConsole) ScenarioEnd(res result.Result) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	printHookFailureSC(sc, res, res.GetPreHook)
	printHookFailureSC(sc, res, res.GetPostHook)
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

func (sc *simpleConsole) StepEnd(step gauge.Step, res result.Result, execInfo gauge_messages.ExecutionInfo) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	printHookFailureSC(sc, res, res.GetPreHook)
	stepRes := res.(*result.StepResult)
	if stepRes.GetStepFailed() {
		stepText := prepStepMsg(step.LineText)
		logger.GaugeLog.Error(stepText)
		specInfo := prepSpecInfo(execInfo.GetCurrentSpec().GetFileName(), step.LineNo)
		logger.GaugeLog.Error(specInfo)
		errMsg := prepErrorMessage(stepRes.ProtoStepExecResult().GetExecutionResult().GetErrorMessage())
		logger.GaugeLog.Error(errMsg)
		stacktrace := prepStacktrace(stepRes.ProtoStepExecResult().GetExecutionResult().GetStackTrace())
		logger.GaugeLog.Error(stacktrace)

		msg := formatErrorFragment(stepText, sc.indentation) + formatErrorFragment(specInfo, sc.indentation) + formatErrorFragment(errMsg, sc.indentation) + formatErrorFragment(stacktrace, sc.indentation)
		fmt.Fprint(sc.writer, msg)
	}
	printHookFailureSC(sc, res, res.GetPostHook)
	sc.indentation -= stepIndentation
}

func (sc *simpleConsole) ConceptStart(conceptHeading string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.indentation += stepIndentation
	logger.GaugeLog.Debug(conceptHeading)
	if Verbose {
		fmt.Fprint(sc.writer, fmt.Sprintf("%s%s", indent(strings.TrimSpace(conceptHeading), sc.indentation), newline))
	}
}

func (sc *simpleConsole) ConceptEnd(res result.Result) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.indentation -= stepIndentation
}

func (sc *simpleConsole) SuiteEnd(res result.Result) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	printHookFailureSC(sc, res, res.GetPreHook)
	printHookFailureSC(sc, res, res.GetPostHook)
	suiteRes := res.(*result.SuiteResult)
	for _, e := range suiteRes.UnhandledErrors {
		logger.GaugeLog.Error(e.Error())
		fmt.Fprint(sc.writer, indent(e.Error(), sc.indentation+errorIndentation)+newline)
	}
}

func (sc *simpleConsole) DataTable(table string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	logger.GaugeLog.Debug(table)
	fmt.Fprint(sc.writer, fmt.Sprintf("%s", table))
}

func (sc *simpleConsole) Errorf(err string, args ...interface{}) {
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
	fmt.Fprint(sc.writer, string(b))
	return len(b), nil
}

func printHookFailureSC(sc *simpleConsole, res result.Result, hookFailure func() **(gauge_messages.ProtoHookFailure)) {
	if hookFailure() != nil && *hookFailure() != nil {
		errMsg := prepErrorMessage((*hookFailure()).GetErrorMessage())
		logger.GaugeLog.Error(errMsg)
		stacktrace := prepStacktrace((*hookFailure()).GetStackTrace())
		logger.GaugeLog.Error(stacktrace)
		fmt.Fprint(sc.writer, formatErrorFragment(errMsg, sc.indentation), formatErrorFragment(stacktrace, sc.indentation))
	}
}
