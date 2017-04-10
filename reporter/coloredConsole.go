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
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/apoorvam/goterminal"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
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

func (c *coloredConsole) SpecStart(heading string) {
	msg := formatSpec(heading)
	logger.GaugeLog.Info(msg)
	c.displayMessage(msg+newline, ct.Cyan)
	c.writer.Reset()
}

func (c *coloredConsole) SpecEnd(res result.Result) {
	printHookFailureCC(c, res, res.GetPreHook)
	printHookFailureCC(c, res, res.GetPostHook)
	c.displayMessage(newline, ct.None)
	c.writer.Reset()
}

func (c *coloredConsole) ScenarioStart(scenarioHeading string) {
	c.indentation += scenarioIndentation
	msg := formatScenario(scenarioHeading)
	logger.GaugeLog.Info(msg)

	indentedText := indent(msg+"\t", c.indentation)
	c.displayMessage(indentedText, ct.Yellow)
}

func (c *coloredConsole) ScenarioEnd(res result.Result) {
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
	logger.GaugeLog.Debug(stepText)
}

func (c *coloredConsole) StepEnd(step gauge.Step, res result.Result, execInfo gauge_messages.ExecutionInfo) {
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
		logger.GaugeLog.Error(stepText)
		errMsg := prepErrorMessage(stepRes.ProtoStepExecResult().GetExecutionResult().GetErrorMessage())
		logger.GaugeLog.Error(errMsg)
		specInfo := prepSpecInfo(execInfo.GetCurrentSpec().GetFileName(), step.LineNo, step.InConcept())
		logger.GaugeLog.Error(specInfo)
		stacktrace := prepStacktrace(stepRes.ProtoStepExecResult().GetExecutionResult().GetStackTrace())
		logger.GaugeLog.Error(stacktrace)

		failureMsg := formatErrorFragment(stepText, c.indentation) + formatErrorFragment(specInfo, c.indentation) + formatErrorFragment(errMsg, c.indentation) + formatErrorFragment(stacktrace, c.indentation)
		c.sceFailuresBuf.WriteString(failureMsg)
	}
	printHookFailureCC(c, res, res.GetPostHook)
	c.indentation -= stepIndentation
}

func (c *coloredConsole) ConceptStart(conceptHeading string) {
	c.indentation += stepIndentation
	logger.GaugeLog.Debug(conceptHeading)
}

func (c *coloredConsole) ConceptEnd(res result.Result) {
	c.indentation -= stepIndentation
}

func (c *coloredConsole) SuiteEnd(res result.Result) {
	suiteRes := res.(*result.SuiteResult)
	printHookFailureCC(c, res, res.GetPreHook)
	printHookFailureCC(c, res, res.GetPostHook)
	for _, e := range suiteRes.UnhandledErrors {
		logger.GaugeLog.Error(e.Error())
		c.displayMessage(indent(e.Error(), c.indentation+errorIndentation)+newline, ct.Red)
	}
}

func (c *coloredConsole) DataTable(table string) {
	logger.GaugeLog.Debug(table)
	c.displayMessage(table, ct.Yellow)
	c.writer.Reset()
}

func (c *coloredConsole) Errorf(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args...)
	logger.GaugeLog.Error(msg)
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
		logger.GaugeLog.Error(errMsg)
		stacktrace := prepStacktrace(hookFailure()[0].GetStackTrace())
		logger.GaugeLog.Error(stacktrace)
		c.displayMessage(newline+formatErrorFragment(errMsg, c.indentation)+formatErrorFragment(stacktrace, c.indentation), ct.Red)
		return false
	}
	return true
}
