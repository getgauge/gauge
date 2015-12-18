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

type consoleWriter struct {
	writer      *goterminal.Writer
	headingText bytes.Buffer
	buffer      bytes.Buffer
	indentation int
	isColored   bool
}

func newConsoleWriter(isColored bool) *consoleWriter {
	return &consoleWriter{writer: goterminal.New(), isColored: isColored}
}

func (cw *consoleWriter) Write(b []byte) (int, error) {
	cw.indentation += sysoutIndentation
	text := strings.Trim(string(b), "\n ")
	text = strings.Replace(text, newline, newline+spaces(cw.indentation), -1)
	if len(text) > 0 {
		msg := spaces(cw.indentation) + text + newline
		cw.buffer.WriteString(msg)
		if level == logging.DEBUG {
			fmt.Fprint(cw.writer, msg)
			cw.writer.Print()
		}
	}
	cw.indentation -= sysoutIndentation
	return len(b), nil
}

func (cw *consoleWriter) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Error(msg)
	fmt.Fprint(cw, msg)
}

func (cw *consoleWriter) Critical(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Critical(msg)
	cw.Write([]byte(msg))
}

func (cw *consoleWriter) Info(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Info(msg)
	fmt.Fprint(cw, msg)
}

func (cw *consoleWriter) Debug(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Debug(msg)
	fmt.Fprint(cw, msg)
}

func (cw *consoleWriter) SpecStart(heading string) {
	msg := formatSpec(heading)
	GaugeLog.Info(msg)
	cw.printViaWriter(msg+newline+newline, ct.Cyan)
	cw.writer.Reset()
}

func (cw *consoleWriter) SpecEnd() {
	cw.printViaWriter("", ct.None)
	cw.writer.Reset()
}

func (cw *consoleWriter) ScenarioStart(scenarioHeading string) {
	cw.indentation = scenarioIndentation
	msg := formatScenario(scenarioHeading)
	GaugeLog.Info(msg)

	indentedText := indent(msg, cw.indentation)
	if level == logging.INFO {
		cw.headingText.WriteString(indentedText + spaces(4))
		cw.printViaWriter(cw.headingText.String(), ct.None)
	} else {
		cw.printViaWriter(indentedText+newline, ct.Yellow)
	}
	cw.writer.Reset()
}

func (cw *consoleWriter) ScenarioEnd(failed bool) {
	cw.printViaWriter("", ct.None)
	if level == logging.INFO {
		cw.printViaWriter(newline, ct.None)
		if failed {
			cw.printViaWriter(cw.buffer.String(), ct.Red)
		}
	}
	cw.writer.Reset()
	cw.indentation -= scenarioIndentation
}

func (cw *consoleWriter) StepStart(stepText string) {
	cw.indentation += stepIndentation
	GaugeLog.Debug(stepText)
	if level == logging.DEBUG {
		cw.headingText.WriteString(indent(stepText, cw.indentation))
		cw.printViaWriter(cw.headingText.String()+newline, ct.None)
	}
}

func (cw *consoleWriter) StepEnd(failed bool) {
	if level == logging.DEBUG {
		cw.writer.Clear()
		if failed {
			cw.printViaWriter(cw.headingText.String()+"\t ...[FAIL]\n", ct.Red)
		} else {
			cw.printViaWriter(cw.headingText.String()+"\t ...[PASS]\n", ct.Green)
		}
		cw.printViaWriter(cw.buffer.String(), ct.None)
		cw.Reset()
	} else {
		if failed {
			cw.printViaWriter(getFailureSymbol(), ct.Red)
		} else {
			cw.printViaWriter(getSuccessSymbol(), ct.Green)
			cw.Reset()
		}
	}
	cw.writer.Reset()
	cw.indentation -= stepIndentation
}

func (cw *consoleWriter) Reset() {
	cw.headingText.Reset()
	cw.buffer.Reset()
}

func (cw *consoleWriter) ConceptStart(conceptHeading string) {
	cw.indentation += stepIndentation
	GaugeLog.Debug(conceptHeading)
	if level == logging.DEBUG {
		cw.printViaWriter(indent(conceptHeading, cw.indentation)+newline, ct.Magenta)
		cw.writer.Reset()
	}
}

func (cw *consoleWriter) ConceptEnd(failed bool) {
	cw.indentation -= stepIndentation
}

func (cw *consoleWriter) DataTable(table string) {
	GaugeLog.Debug(table)
	if level == logging.DEBUG {
		cw.printViaWriter(table+newline, ct.Yellow)
		cw.writer.Reset()
	}
}

func (cw *consoleWriter) printViaWriter(text string, color ct.Color) {
	if cw.isColored {
		ct.Foreground(color, false)
		defer ct.ResetColor()
	}
	fmt.Fprint(cw.writer, text)
	cw.writer.Print()
}
