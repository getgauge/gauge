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
	writer               *goterminal.Writer
	headingBuffer        bytes.Buffer
	pluginMessagesBuffer bytes.Buffer
	errorMessagesBuffer  bytes.Buffer
	indentation          int
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
	c.displayMessage(newline, ct.None)
	c.writer.Reset()
	printHookFailureCC(c, res, res.GetPreHook)
	printHookFailureCC(c, res, res.GetPostHook)
}

func (c *coloredConsole) ScenarioStart(scenarioHeading string) {
	c.indentation += scenarioIndentation
	msg := formatScenario(scenarioHeading)
	logger.GaugeLog.Info(msg)

	indentedText := indent(msg+"\t", c.indentation)
	c.displayMessage(indentedText, ct.Yellow)
	if Verbose {
		c.displayMessage(newline, ct.None)
	}
	c.writer.Reset()
}

func (c *coloredConsole) ScenarioEnd(res result.Result) {
	if !Verbose {
		c.displayMessage(newline, ct.None)
	}
	c.writer.Reset()
	printHookFailureCC(c, res, res.GetPreHook)
	printHookFailureCC(c, res, res.GetPostHook)
	c.indentation -= scenarioIndentation
}

func (c *coloredConsole) StepStart(stepText string) {
	c.resetBuffers()
	c.writer.Reset()

	c.indentation += stepIndentation
	logger.GaugeLog.Debug(stepText)
	if Verbose {
		c.headingBuffer.WriteString(indent(strings.TrimSpace(stepText), c.indentation))
		c.displayMessage(c.headingBuffer.String()+newline, ct.None)
	}
}

func (c *coloredConsole) StepEnd(step gauge.Step, res result.Result) {
	stepRes := res.(*result.StepResult)
	if Verbose {
		c.writer.Clear()
		if stepRes.GetStepFailed() {
			c.displayMessage(c.headingBuffer.String()+"\t ...[FAIL]\n", ct.Red)
		} else {
			c.displayMessage(c.headingBuffer.String()+"\t ...[PASS]\n", ct.Green)
		}
		c.displayMessage(c.pluginMessagesBuffer.String(), ct.None)
		c.displayMessage(c.errorMessagesBuffer.String(), ct.Red)
	} else {
		if stepRes.GetStepFailed() {
			c.displayMessage(getFailureSymbol()+newline, ct.Red)
		} else {
			c.displayMessage(getSuccessSymbol(), ct.Green)
		}
	}
	if stepRes.GetStepFailed() {
		stepText := prepStepMsg(step.LineText)
		logger.GaugeLog.Error(stepText)
		errMsg := prepErrorMessage(stepRes.ProtoStepExecResult().GetExecutionResult().GetErrorMessage())
		logger.GaugeLog.Error(errMsg)
		stacktrace := prepStacktrace(stepRes.ProtoStepExecResult().GetExecutionResult().GetStackTrace())
		logger.GaugeLog.Error(stacktrace)

		msg := formatStepText(stepText, c.indentation) + formatErrorMessage(errMsg, c.indentation) + formatStacktrace(stacktrace, c.indentation)
		c.displayMessage(msg, ct.Red)
	}
	printHookFailureCC(c, res, res.GetPreHook)
	printHookFailureCC(c, res, res.GetPostHook)
	c.writer.Reset()
	c.resetBuffers()
	c.indentation -= stepIndentation
}

func (c *coloredConsole) ConceptStart(conceptHeading string) {
	c.indentation += stepIndentation
	logger.GaugeLog.Debug(conceptHeading)
	if Verbose {
		c.displayMessage(indent(strings.TrimSpace(conceptHeading), c.indentation)+newline, ct.Magenta)
		c.writer.Reset()
	}
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
	c.errorMessagesBuffer.WriteString(msg)
}

// Write writes the bytes to console via goterminal's writer.
// This is called when any sysouts are to be printed on console.
func (c *coloredConsole) Write(b []byte) (int, error) {
	text := string(b)
	c.pluginMessagesBuffer.WriteString(text)
	c.displayMessage(text, ct.None)
	return len(b), nil
}

func (c *coloredConsole) displayMessage(msg string, color ct.Color) {
	ct.Foreground(color, false)
	defer ct.ResetColor()
	fmt.Fprint(c.writer, msg)
	c.writer.Print()
}

func (c *coloredConsole) resetBuffers() {
	c.headingBuffer.Reset()
	c.pluginMessagesBuffer.Reset()
	c.errorMessagesBuffer.Reset()
}

func printHookFailureCC(c *coloredConsole, res result.Result, hookFailure func() **(gauge_messages.ProtoHookFailure)) {
	if hookFailure() != nil && *hookFailure() != nil {
		errMsg := prepErrorMessage((*hookFailure()).GetErrorMessage())
		logger.GaugeLog.Error(errMsg)
		stacktrace := prepStacktrace((*hookFailure()).GetStackTrace())
		logger.GaugeLog.Error(stacktrace)
		c.displayMessage(formatErrorMessage(errMsg, c.indentation)+formatStacktrace(stacktrace, c.indentation), ct.Red)
	}
}
