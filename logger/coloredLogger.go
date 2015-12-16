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

package logger

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/apoorvam/goterminal"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/op/go-logging"
)

const newline = "\n"

type coloredLogger struct {
	writer      *goterminal.Writer
	headingText bytes.Buffer
	buffer      bytes.Buffer
	indentation int
}

func newColoredConsoleWriter() *coloredLogger {
	return &coloredLogger{writer: goterminal.New()}
}

func (cl *coloredLogger) Write(b []byte) (int, error) {
	cl.indentation += sysoutIndentation
	text := strings.Trim(string(b), "\n ")
	text = strings.Replace(text, newline, newline+spaces(cl.indentation), -1)
	if len(text) > 0 {
		msg := spaces(cl.indentation) + text + newline
		cl.buffer.WriteString(msg)
		if level == logging.DEBUG {
			fmt.Fprint(cl.writer, msg)
			cl.writer.Print()
		}
	}
	cl.indentation -= sysoutIndentation
	return len(b), nil
}

func (cl *coloredLogger) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Error(msg)
	fmt.Fprint(cl, msg)
}

func (cl *coloredLogger) Critical(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Critical(msg)
	cl.Write([]byte(msg))
}

func (cl *coloredLogger) Info(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Info(msg)
	fmt.Fprint(cl, msg)
}

func (cl *coloredLogger) Debug(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Debug(msg)
	fmt.Fprint(cl, msg)
}

func (cl *coloredLogger) SpecStart(heading string) {
	msg := formatSpec(heading)
	GaugeLog.Info(msg)
	cl.printViaWriter(msg+newline+newline, ct.Cyan)
	cl.writer.Reset()
}

func (cl *coloredLogger) SpecEnd() {
	cl.printViaWriter("", ct.None)
	cl.writer.Reset()
}

func (cl *coloredLogger) ScenarioStart(scenarioHeading string) {
	cl.indentation = scenarioIndentation
	msg := formatScenario(scenarioHeading)
	GaugeLog.Info(msg)

	indentedText := indent(msg, cl.indentation)
	if level == logging.INFO {
		cl.headingText.WriteString(indentedText + spaces(4))
		cl.printViaWriter(cl.headingText.String(), ct.None)
	} else {
		cl.printViaWriter(indentedText+newline, ct.Yellow)
	}
	cl.writer.Reset()
}

func (cl *coloredLogger) ScenarioEnd(failed bool) {
	cl.printViaWriter("", ct.None)
	if level == logging.INFO {
		cl.printViaWriter(newline, ct.None)
		if failed {
			cl.printViaWriter(cl.buffer.String(), ct.Red)
		}
	}
	cl.writer.Reset()
	cl.indentation -= scenarioIndentation
}

func (cl *coloredLogger) StepStart(stepText string) {
	cl.indentation += stepIndentation
	GaugeLog.Debug(stepText)
	if level == logging.DEBUG {
		cl.headingText.WriteString(indent(stepText, cl.indentation))
		cl.printViaWriter(cl.headingText.String()+newline, ct.None)
	}
}

func (cl *coloredLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		cl.writer.Clear()
		if failed {
			cl.printViaWriter(cl.headingText.String()+"\t ...[FAIL]\n", ct.Red)
		} else {
			cl.printViaWriter(cl.headingText.String()+"\t ...[PASS]\n", ct.Green)
		}
		cl.printViaWriter(cl.buffer.String(), ct.None)
		cl.Reset()
	} else {
		if failed {
			cl.printViaWriter(getFailureSymbol(), ct.Red)
		} else {
			cl.printViaWriter(getSuccessSymbol(), ct.Green)
			cl.Reset()
		}
	}
	cl.writer.Reset()
	cl.indentation -= stepIndentation
}

func (cl *coloredLogger) Reset() {
	cl.headingText.Reset()
	cl.buffer.Reset()
}

func (cl *coloredLogger) ConceptStart(conceptHeading string) {
	cl.indentation += stepIndentation
	GaugeLog.Debug(conceptHeading)
	if level == logging.DEBUG {
		cl.printViaWriter(indent(conceptHeading, cl.indentation)+newline, ct.Magenta)
		cl.writer.Reset()
	}
}

func (cl *coloredLogger) ConceptEnd(failed bool) {
	cl.indentation -= stepIndentation
}

func (cl *coloredLogger) DataTable(table string) {
	GaugeLog.Debug(table)
	if level == logging.DEBUG {
		cl.printViaWriter(table+newline, ct.Yellow)
		cl.writer.Reset()
	}
}

func (cl *coloredLogger) printViaWriter(text string, color ct.Color) {
	ct.Foreground(color, false)
	fmt.Fprint(cl.writer, text)
	cl.writer.Print()
	ct.ResetColor()
}
