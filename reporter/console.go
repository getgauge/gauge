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
var Verbose bool

const newline = "\n"

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
	Critical(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})

	io.Writer
}

var currentReporter Reporter

func Current() Reporter {
	if currentReporter == nil {
		currentReporter = newConsole(!SimpleConsoleOutput)
	}
	return currentReporter
}

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

type console struct {
	writer      *goterminal.Writer
	headingText bytes.Buffer
	buffer      bytes.Buffer
	indentation int
	isColored   bool
}

func newConsole(isColored bool) *console {
	return &console{writer: goterminal.New(), isColored: isColored}
}

func (c *console) Write(b []byte) (int, error) {
	c.indentation += sysoutIndentation
	text := strings.Trim(string(b), "\n ")
	text = strings.Replace(text, newline, newline+spaces(c.indentation), -1)
	if len(text) > 0 {
		msg := spaces(c.indentation) + text + newline
		c.buffer.WriteString(msg)
		if Verbose {
			fmt.Fprint(c.writer, msg)
			c.writer.Print()
		}
	}
	c.indentation -= sysoutIndentation
	return len(b), nil
}

func (c *console) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	logger.GaugeLog.Error(msg)
	fmt.Fprint(c, msg)
}

func (c *console) Critical(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	logger.GaugeLog.Critical(msg)
	c.Write([]byte(msg))
}

func (c *console) Info(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	logger.GaugeLog.Info(msg)
	fmt.Fprint(c, msg)
}

func (c *console) Debug(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	logger.GaugeLog.Debug(msg)
	fmt.Fprint(c, msg)
}

func (c *console) SpecStart(heading string) {
	msg := formatSpec(heading)
	logger.GaugeLog.Info(msg)
	c.printViaWriter(msg+newline+newline, ct.Cyan)
	c.writer.Reset()
}

func (c *console) SpecEnd() {
	c.printViaWriter("", ct.None)
	c.writer.Reset()
}

func (c *console) ScenarioStart(scenarioHeading string) {
	c.indentation = scenarioIndentation
	msg := formatScenario(scenarioHeading)
	logger.GaugeLog.Info(msg)

	indentedText := indent(msg, c.indentation)
	if !Verbose {
		c.headingText.WriteString(indentedText + spaces(4))
		c.printViaWriter(c.headingText.String(), ct.None)
	} else {
		c.printViaWriter(indentedText+newline, ct.Yellow)
	}
	c.writer.Reset()
}

func (c *console) ScenarioEnd(failed bool) {
	c.printViaWriter("", ct.None)
	if !Verbose {
		c.printViaWriter(newline, ct.None)
		if failed {
			c.printViaWriter(c.buffer.String(), ct.Red)
		}
	}
	c.writer.Reset()
	c.indentation -= scenarioIndentation
}

func (c *console) StepStart(stepText string) {
	c.indentation += stepIndentation
	logger.GaugeLog.Debug(stepText)
	if Verbose {
		c.headingText.WriteString(indent(stepText, c.indentation))
		c.printViaWriter(c.headingText.String()+newline, ct.None)
	}
}

func (c *console) StepEnd(failed bool) {
	if Verbose {
		c.writer.Clear()
		if failed {
			c.printViaWriter(c.headingText.String()+"\t ...[FAIL]\n", ct.Red)
		} else {
			c.printViaWriter(c.headingText.String()+"\t ...[PASS]\n", ct.Green)
		}
		c.printViaWriter(c.buffer.String(), ct.None)
		c.Reset()
	} else {
		if failed {
			c.printViaWriter(getFailureSymbol(), ct.Red)
		} else {
			c.printViaWriter(getSuccessSymbol(), ct.Green)
			c.Reset()
		}
	}
	c.writer.Reset()
	c.indentation -= stepIndentation
}

func (c *console) Reset() {
	c.headingText.Reset()
	c.buffer.Reset()
}

func (c *console) ConceptStart(conceptHeading string) {
	c.indentation += stepIndentation
	logger.GaugeLog.Debug(conceptHeading)
	if Verbose {
		c.printViaWriter(indent(conceptHeading, c.indentation)+newline, ct.Magenta)
		c.writer.Reset()
	}
}

func (c *console) ConceptEnd(failed bool) {
	c.indentation -= stepIndentation
}

func (c *console) DataTable(table string) {
	logger.GaugeLog.Debug(table)
	if Verbose {
		c.printViaWriter(table+newline, ct.Yellow)
		c.writer.Reset()
	}
}

func (c *console) printViaWriter(text string, color ct.Color) {
	if c.isColored {
		ct.Foreground(color, false)
		defer ct.ResetColor()
	}
	fmt.Fprint(c.writer, text)
	c.writer.Print()
}
