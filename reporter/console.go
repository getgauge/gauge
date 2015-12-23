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
	"github.com/getgauge/gauge/logger"
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

// Current returns an instance of Reporter.
// It returns the current instance of Reporter, if present. Else, it returns a new Reporter.
func Current() Reporter {
	if currentReporter == nil {
		currentReporter = newConsole(!SimpleConsoleOutput)
	}
	return currentReporter
}

type console struct {
	writer               *goterminal.Writer
	headingBuffer        bytes.Buffer
	pluginMessagesBuffer bytes.Buffer
	indentation          int
	isColored            bool
}

func newConsole(isColored bool) *console {
	return &console{writer: goterminal.New(), isColored: isColored}
}

// NewParallelConsole returns the instance of parallel console reporter
func NewParallelConsole(n int) Reporter {
	parallelLogger := Current()
	//	parallelLogger := &GaugeLogger{logging.MustGetLogger("gauge")}
	// 	stdOutLogger := logging.NewLogBackend(os.Stdout, "", 0)
	//	stdOutFormatter := logging.NewBackendFormatter(stdOutLogger, logging.MustStringFormatter("[runner:"+strconv.Itoa(n)+"] %{message}"))
	//	stdOutLoggerLeveled := logging.AddModuleLevel(stdOutFormatter)
	//	stdOutLoggerLeveled.SetLevel(level, "")
	//  parallelLogger.SetBackend(stdOutLoggerLeveled)

	return parallelLogger
}

func (c *console) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	logger.GaugeLog.Error(msg)
	fmt.Fprint(c, msg)
}

func (c *console) SpecStart(heading string) {
	msg := formatSpec(heading)
	logger.GaugeLog.Info(msg)
	c.displayMessage(msg+newline+newline, ct.Cyan)
	c.writer.Reset()
}

func (c *console) SpecEnd() {
	c.displayMessage(newline, ct.None)
	c.writer.Reset()
}

func (c *console) ScenarioStart(scenarioHeading string) {
	c.indentation = scenarioIndentation
	msg := formatScenario(scenarioHeading)
	logger.GaugeLog.Info(msg)

	indentedText := indent(msg, c.indentation)
	if !Verbose {
		c.headingBuffer.WriteString(indentedText + spaces(4))
		c.displayMessage(c.headingBuffer.String(), ct.None)
	} else {
		c.displayMessage(indentedText+newline, ct.Yellow)
	}
	c.writer.Reset()
}

func (c *console) ScenarioEnd(failed bool) {
	if !Verbose {
		c.displayMessage(newline, ct.None)
		if failed {
			c.displayMessage(c.pluginMessagesBuffer.String(), ct.Red)
		}
	}
	c.writer.Reset()
	c.indentation -= scenarioIndentation
}

func (c *console) StepStart(stepText string) {
	c.indentation += stepIndentation
	logger.GaugeLog.Debug(stepText)
	if Verbose {
		c.headingBuffer.WriteString(indent(stepText, c.indentation))
		c.displayMessage(c.headingBuffer.String()+newline, ct.None)
	}
}

func (c *console) StepEnd(failed bool) {
	if Verbose {
		c.writer.Clear()
		if failed {
			c.displayMessage(c.headingBuffer.String()+"\t ...[FAIL]\n", ct.Red)
		} else {
			c.displayMessage(c.headingBuffer.String()+"\t ...[PASS]\n", ct.Green)
		}
		c.displayMessage(c.pluginMessagesBuffer.String(), ct.None)
		c.resetBuffers()
	} else {
		if failed {
			c.displayMessage(getFailureSymbol(), ct.Red)
		} else {
			c.displayMessage(getSuccessSymbol(), ct.Green)
			c.resetBuffers()
		}
	}
	c.writer.Reset()
	c.indentation -= stepIndentation
}

func (c *console) ConceptStart(conceptHeading string) {
	c.indentation += stepIndentation
	logger.GaugeLog.Debug(conceptHeading)
	if Verbose {
		c.displayMessage(indent(conceptHeading, c.indentation)+newline, ct.Magenta)
		c.writer.Reset()
	}
}

func (c *console) ConceptEnd(failed bool) {
	c.indentation -= stepIndentation
}

func (c *console) DataTable(table string) {
	logger.GaugeLog.Debug(table)
	if Verbose {
		c.displayMessage(table+newline, ct.Yellow)
		c.writer.Reset()
	}
}

// Write writes the bytes to console via goterminal's writer.
// This is called when any sysouts are to be printed on console.
func (c *console) Write(b []byte) (int, error) {
	c.indentation += sysoutIndentation
	text := c.formatText(string(b))
	if len(text) > 0 {
		c.pluginMessagesBuffer.WriteString(text)
		if Verbose {
			c.displayMessage(text, ct.None)
		}
	}
	c.indentation -= sysoutIndentation
	return len(b), nil
}

func (c *console) formatText(text string) string {
	formattedText := strings.Trim(text, "\n ")
	formattedText = strings.Replace(formattedText, newline, newline+spaces(c.indentation), -1)
	if len(formattedText) > 0 {
		formattedText = spaces(c.indentation) + formattedText + newline
	}
	return formattedText
}

func (c *console) displayMessage(msg string, color ct.Color) {
	if c.isColored {
		ct.Foreground(color, false)
		defer ct.ResetColor()
	}
	fmt.Fprint(c.writer, msg)
	c.writer.Print()
}

func (c *console) resetBuffers() {
	c.headingBuffer.Reset()
	c.pluginMessagesBuffer.Reset()
}
