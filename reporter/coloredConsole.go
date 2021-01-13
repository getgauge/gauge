/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package reporter

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/apoorvam/goterminal"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
)

type coloredConsole struct {
	writer         *goterminal.Writer
	indentation    int
	sceFailuresBuf bytes.Buffer
}

func newColoredConsole(out io.Writer) *coloredConsole {
	return &coloredConsole{writer: goterminal.New(out)}
}

func (c *coloredConsole) SuiteStart() {
}

func (c *coloredConsole) SpecStart(spec *gauge.Specification, res result.Result) {
	if res.(*result.SpecResult).Skipped {
		return
	}
	msg := formatSpec(spec.Heading.Value)
	logger.Info(false, msg)
	c.displayMessage(msg+newline, ct.Cyan)
	c.writer.Reset()
}

func (c *coloredConsole) SpecEnd(spec *gauge.Specification, res result.Result) {
	if res.(*result.SpecResult).Skipped {
		return
	}
	printHookFailureCC(c, res, res.GetPreHook)
	printHookFailureCC(c, res, res.GetPostHook)
	c.displayMessage(newline, ct.None)
	c.writer.Reset()
}

func (c *coloredConsole) ScenarioStart(scenario *gauge.Scenario, i *gauge_messages.ExecutionInfo, res result.Result) {
	if res.(*result.ScenarioResult).ProtoScenario.ExecutionStatus == gauge_messages.ExecutionStatus_SKIPPED {
		return
	}
	c.indentation += scenarioIndentation
	msg := formatScenario(scenario.Heading.Value)
	logger.Info(false, msg)

	indentedText := indent(msg+"\t", c.indentation)
	c.displayMessage(indentedText, ct.Yellow)
}

func (c *coloredConsole) ScenarioEnd(scenario *gauge.Scenario, res result.Result, i *gauge_messages.ExecutionInfo) {
	if res.(*result.ScenarioResult).ProtoScenario.ExecutionStatus == gauge_messages.ExecutionStatus_SKIPPED {
		return
	}
	if printHookFailureCC(c, res, res.GetPreHook) {
		if c.sceFailuresBuf.Len() != 0 {
			c.displayMessage(newline+strings.Trim(c.sceFailuresBuf.String(), newline)+newline, ct.Red)
		} else {
			c.displayMessage(newline, ct.None)
		}
	}

	printHookFailureCC(c, res, res.GetPostHook)
	c.indentation -= scenarioIndentation
	c.writer.Reset()
	c.sceFailuresBuf.Reset()
}

func (c *coloredConsole) StepStart(stepText string) {
	c.indentation += stepIndentation
	logger.Debug(false, stepText)
}

func (c *coloredConsole) StepEnd(step gauge.Step, res result.Result, execInfo *gauge_messages.ExecutionInfo) {
	stepRes := res.(*result.StepResult)
	if !(hookFailed(res.GetPreHook) || hookFailed(res.GetPostHook)) {
		if stepRes.GetStepFailed() {
			c.displayMessage(getFailureSymbol(), ct.Red)
		} else {
			c.displayMessage(getSuccessSymbol(), ct.Green)
		}
	}
	if printHookFailureCC(c, res, res.GetPreHook) && stepRes.GetStepFailed() {
		stepText := strings.TrimLeft(prepStepMsg(step.LineText), newline)
		logger.Error(false, stepText)
		errMsg := prepErrorMessage(stepRes.ProtoStepExecResult().GetExecutionResult().GetErrorMessage())
		logger.Error(false, errMsg)
		specInfo := prepSpecInfo(execInfo.GetCurrentSpec().GetFileName(), step.LineNo, step.InConcept())
		logger.Error(false, specInfo)
		stacktrace := prepStacktrace(stepRes.ProtoStepExecResult().GetExecutionResult().GetStackTrace())
		logger.Error(false, stacktrace)

		failureMsg := formatErrorFragment(stepText, c.indentation) + formatErrorFragment(specInfo, c.indentation) + formatErrorFragment(errMsg, c.indentation) + formatErrorFragment(stacktrace, c.indentation)
		_, err := c.sceFailuresBuf.WriteString(failureMsg)
		if err != nil {
			logger.Errorf(true, "Error writing to scenario failure buffer: %s", err.Error())
		}
	}
	printHookFailureCC(c, res, res.GetPostHook)
	c.indentation -= stepIndentation
}

func (c *coloredConsole) ConceptStart(conceptHeading string) {
	c.indentation += stepIndentation
	logger.Debug(false, conceptHeading)
}

func (c *coloredConsole) ConceptEnd(res result.Result) {
	c.indentation -= stepIndentation
}

func (c *coloredConsole) SuiteEnd(res result.Result) {
	suiteRes := res.(*result.SuiteResult)
	printHookFailureCC(c, res, res.GetPreHook)
	printHookFailureCC(c, res, res.GetPostHook)
	for _, e := range suiteRes.UnhandledErrors {
		logger.Error(false, e.Error())
		c.displayMessage(indent(e.Error(), c.indentation+errorIndentation)+newline, ct.Red)
	}
}

func (c *coloredConsole) DataTable(table string) {
	logger.Debug(false, table)
	c.displayMessage(table, ct.Yellow)
	c.writer.Reset()
}

func (c *coloredConsole) Errorf(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args...)
	logger.Error(false, msg)
	msg = indent(msg, c.indentation+errorIndentation) + newline
	c.displayMessage(msg, ct.Red)
}

// Write writes the bytes to console via goterminal's writer.
// This is called when any sysouts are to be printed on console.
func (c *coloredConsole) Write(b []byte) (int, error) {
	text := string(b)
	c.displayMessage(text, ct.None)
	return len(b), nil
}

func (c *coloredConsole) displayMessage(msg string, color ct.Color) {
	ct.Foreground(color, false)
	defer ct.ResetColor()
	fmt.Fprint(c.writer, msg)
	c.writer.Print()
}

func printHookFailureCC(c *coloredConsole, res result.Result, hookFailure func() []*gauge_messages.ProtoHookFailure) bool {
	if len(hookFailure()) > 0 {
		errMsg := prepErrorMessage(hookFailure()[0].GetErrorMessage())
		logger.Error(false, errMsg)
		stacktrace := prepStacktrace(hookFailure()[0].GetStackTrace())
		logger.Error(false, stacktrace)
		c.displayMessage(newline+formatErrorFragment(errMsg, c.indentation)+formatErrorFragment(stacktrace, c.indentation), ct.Red)
		return false
	}
	return true
}
