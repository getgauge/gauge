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

type verboseColoredConsole struct {
	writer               *goterminal.Writer
	headingBuffer        bytes.Buffer
	pluginMessagesBuffer bytes.Buffer
	errorMessagesBuffer  bytes.Buffer
	indentation          int
}

func newVerboseColoredConsole(out io.Writer) *verboseColoredConsole {
	return &verboseColoredConsole{writer: goterminal.New(out)}
}

func (c *verboseColoredConsole) SpecStart(heading string) {
	msg := formatSpec(heading)
	logger.GaugeLog.Info(msg)
	c.displayMessage(msg+newline, ct.Cyan)
	c.writer.Reset()
}

func (c *verboseColoredConsole) SpecEnd(res result.Result) {
	printHookFailureVCC(c, res, res.GetPreHook)
	printHookFailureVCC(c, res, res.GetPostHook)
	c.displayMessage(newline, ct.None)
	c.writer.Reset()
}

func (c *verboseColoredConsole) ScenarioStart(scenarioHeading string) {
	c.indentation += scenarioIndentation
	msg := formatScenario(scenarioHeading)
	logger.GaugeLog.Info(msg)

	indentedText := indent(msg+"\t", c.indentation)
	c.displayMessage(indentedText+newline, ct.Yellow)
	c.writer.Reset()
}

func (c *verboseColoredConsole) ScenarioEnd(res result.Result) {
	printHookFailureVCC(c, res, res.GetPreHook)
	printHookFailureVCC(c, res, res.GetPostHook)

	c.writer.Reset()
	c.indentation -= scenarioIndentation
}

func (c *verboseColoredConsole) StepStart(stepText string) {
	c.resetBuffers()
	c.writer.Reset()

	c.indentation += stepIndentation
	logger.GaugeLog.Debug(stepText)
	c.headingBuffer.WriteString(indent(strings.TrimSpace(stepText), c.indentation))
	c.displayMessage(c.headingBuffer.String()+newline, ct.None)
}

func (c *verboseColoredConsole) StepEnd(step gauge.Step, res result.Result, execInfo gauge_messages.ExecutionInfo) {
	stepRes := res.(*result.StepResult)
	c.writer.Clear()
	if !(hookFailed(res.GetPreHook) || hookFailed(res.GetPostHook)) {
		if stepRes.GetStepFailed() {
			c.displayMessage(c.headingBuffer.String()+"\t ...[FAIL]\n", ct.Red)
		} else {
			c.displayMessage(c.headingBuffer.String()+"\t ...[PASS]\n", ct.Green)
		}
	} else {
		c.displayMessage(c.headingBuffer.String()+newline, ct.None)
	}
	printHookFailureVCC(c, res, res.GetPreHook)
	c.displayMessage(c.pluginMessagesBuffer.String(), ct.None)
	c.displayMessage(c.errorMessagesBuffer.String(), ct.Red)
	if stepRes.GetStepFailed() {
		stepText := prepStepMsg(step.LineText)
		logger.GaugeLog.Error(stepText)
		errMsg := prepErrorMessage(stepRes.ProtoStepExecResult().GetExecutionResult().GetErrorMessage())
		logger.GaugeLog.Error(errMsg)
		specInfo := prepSpecInfo(execInfo.GetCurrentSpec().GetFileName(), step.LineNo, step.InConcept())
		logger.GaugeLog.Error(specInfo)
		stacktrace := prepStacktrace(stepRes.ProtoStepExecResult().GetExecutionResult().GetStackTrace())
		logger.GaugeLog.Error(stacktrace)

		msg := formatErrorFragment(stepText, c.indentation) + formatErrorFragment(specInfo, c.indentation) + formatErrorFragment(errMsg, c.indentation) + formatErrorFragment(stacktrace, c.indentation)

		c.displayMessage(msg, ct.Red)
	}
	printHookFailureVCC(c, res, res.GetPostHook)
	c.indentation -= stepIndentation
	c.writer.Reset()
	c.resetBuffers()
}

func (c *verboseColoredConsole) ConceptStart(conceptHeading string) {
	c.indentation += stepIndentation
	logger.GaugeLog.Debug(conceptHeading)
	c.displayMessage(indent(strings.TrimSpace(conceptHeading), c.indentation)+newline, ct.Magenta)
	c.writer.Reset()
}

func (c *verboseColoredConsole) ConceptEnd(res result.Result) {
	c.indentation -= stepIndentation
}

func (c *verboseColoredConsole) SuiteEnd(res result.Result) {
	suiteRes := res.(*result.SuiteResult)
	printHookFailureVCC(c, res, res.GetPreHook)
	printHookFailureVCC(c, res, res.GetPostHook)
	for _, e := range suiteRes.UnhandledErrors {
		logger.GaugeLog.Error(e.Error())
		c.displayMessage(indent(e.Error(), c.indentation+errorIndentation)+newline, ct.Red)
	}
}

func (c *verboseColoredConsole) DataTable(table string) {
	logger.GaugeLog.Debug(table)
	c.displayMessage(table, ct.Yellow)
	c.writer.Reset()
}

func (c *verboseColoredConsole) Errorf(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args...)
	logger.GaugeLog.Error(msg)
	msg = indent(msg, c.indentation+errorIndentation) + newline
	c.displayMessage(msg, ct.Red)
	c.errorMessagesBuffer.WriteString(msg)
}

// Write writes the bytes to console via goterminal's writer.
// This is called when any sysouts are to be printed on console.
func (c *verboseColoredConsole) Write(b []byte) (int, error) {
	text := string(b)
	c.pluginMessagesBuffer.WriteString(text)
	c.displayMessage(text, ct.None)
	return len(b), nil
}

func (c *verboseColoredConsole) displayMessage(msg string, color ct.Color) {
	ct.Foreground(color, false)
	defer ct.ResetColor()
	fmt.Fprint(c.writer, msg)
	c.writer.Print()
}

func (c *verboseColoredConsole) resetBuffers() {
	c.headingBuffer.Reset()
	c.pluginMessagesBuffer.Reset()
	c.errorMessagesBuffer.Reset()
}

func printHookFailureVCC(c *verboseColoredConsole, res result.Result, hookFailure func() []*gauge_messages.ProtoHookFailure) bool {
	if hookFailed(hookFailure) {
		errMsg := prepErrorMessage(hookFailure()[0].GetErrorMessage())
		logger.GaugeLog.Error(errMsg)
		stacktrace := prepStacktrace(hookFailure()[0].GetStackTrace())
		logger.GaugeLog.Error(stacktrace)
		c.displayMessage(formatErrorFragment(errMsg, c.indentation)+formatErrorFragment(stacktrace, c.indentation), ct.Red)
		return false
	}
	return true
}

func hookFailed(hookFailure func() []*gauge_messages.ProtoHookFailure) bool {
	return len(hookFailure()) > 0
}
