/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package reporter

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
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

func (sc *simpleConsole) SuiteStart() {
}

func (sc *simpleConsole) SpecStart(spec *gauge.Specification, res result.Result) {
	if res.(*result.SpecResult).Skipped {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	formattedHeading := formatSpec(spec.Heading.Value)
	logger.Info(false, formattedHeading)
	fmt.Fprintf(sc.writer, "%s%s", formattedHeading, newline)
}

func (sc *simpleConsole) SpecEnd(spec *gauge.Specification, res result.Result) {
	if res.(*result.SpecResult).Skipped {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	printHookFailureSC(sc, res, res.GetPreHook)
	printHookFailureSC(sc, res, res.GetPostHook)
	fmt.Fprintln(sc.writer)
}

func (sc *simpleConsole) ScenarioStart(scenario *gauge.Scenario, i *gauge_messages.ExecutionInfo, res result.Result) {
	if res.(*result.ScenarioResult).ProtoScenario.ExecutionStatus == gauge_messages.ExecutionStatus_SKIPPED {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.indentation += scenarioIndentation
	formattedHeading := formatScenario(scenario.Heading.Value)
	logger.Info(false, formattedHeading)
	fmt.Fprintf(sc.writer, "%s%s", indent(formattedHeading, sc.indentation), newline)
}

func (sc *simpleConsole) ScenarioEnd(scenario *gauge.Scenario, res result.Result, i *gauge_messages.ExecutionInfo) {
	if res.(*result.ScenarioResult).ProtoScenario.ExecutionStatus == gauge_messages.ExecutionStatus_SKIPPED {
		return
	}
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
	logger.Debug(false, stepText)
	if Verbose {
		fmt.Fprintf(sc.writer, "%s%s", indent(strings.TrimSpace(stepText), sc.indentation), newline)
	}
}

func (sc *simpleConsole) StepEnd(step gauge.Step, res result.Result, execInfo *gauge_messages.ExecutionInfo) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	printHookFailureSC(sc, res, res.GetPreHook)
	stepRes := res.(*result.StepResult)
	if stepRes.GetStepFailed() {
		stepText := prepStepMsg(step.LineText)
		logger.Error(false, stepText)

		specInfo := prepSpecInfo(execInfo.GetCurrentSpec().GetFileName(), step.LineNo, step.InConcept())
		logger.Error(false, specInfo)

		errMsg := prepErrorMessage(stepRes.ProtoStepExecResult().GetExecutionResult().GetErrorMessage())
		logger.Error(false, errMsg)
		stacktrace := prepStacktrace(stepRes.ProtoStepExecResult().GetExecutionResult().GetStackTrace())
		logger.Error(false, stacktrace)

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
	logger.Debug(false, conceptHeading)
	if Verbose {
		fmt.Fprintf(sc.writer, "%s%s", indent(strings.TrimSpace(conceptHeading), sc.indentation), newline)
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
		logger.Error(false, e.Error())
		fmt.Fprint(sc.writer, indent(e.Error(), sc.indentation+errorIndentation)+newline)
	}
}

func (sc *simpleConsole) DataTable(table string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	logger.Debug(false, table)
	fmt.Fprint(sc.writer, table)
}

func (sc *simpleConsole) Errorf(err string, args ...interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	errorMessage := fmt.Sprintf(err, args...)
	logger.Error(false, errorMessage)
	errorString := indent(errorMessage, sc.indentation+errorIndentation)
	fmt.Fprintf(sc.writer, "%s%s", errorString, newline)
}

func (sc *simpleConsole) Write(b []byte) (int, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	fmt.Fprint(sc.writer, string(b))
	return len(b), nil
}

func printHookFailureSC(sc *simpleConsole, res result.Result, hookFailure func() []*gauge_messages.ProtoHookFailure) {
	if len(hookFailure()) > 0 {
		errMsg := prepErrorMessage(hookFailure()[0].GetErrorMessage())
		logger.Error(false, errMsg)
		stacktrace := prepStacktrace(hookFailure()[0].GetStackTrace())
		logger.Error(false, stacktrace)
		fmt.Fprint(sc.writer, formatErrorFragment(errMsg, sc.indentation), formatErrorFragment(stacktrace, sc.indentation))
	}
}
