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
	"github.com/op/go-logging"
)

type simpleLogger struct {
	writer      *goterminal.Writer
	headingText bytes.Buffer
	buffer      bytes.Buffer
	indentation int
}

func newSimpleConsoleWriter() *simpleLogger {
	return &simpleLogger{writer: goterminal.New()}
}

func (sl *simpleLogger) Write(b []byte) (int, error) {
	sl.indentation += sysoutIndentation
	text := strings.Trim(string(b), "\n ")
	text = strings.Replace(text, newline, newline+spaces(sl.indentation), -1)
	if len(text) > 0 {
		msg := spaces(sl.indentation) + text + newline
		sl.buffer.WriteString(msg)
		if level == logging.DEBUG {
			fmt.Fprint(sl.writer, msg)
			sl.writer.Print()
		}
	}
	sl.indentation -= sysoutIndentation
	return len(b), nil
}

func (sl *simpleLogger) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Error(msg)
	fmt.Fprint(sl, msg)
}

func (sl *simpleLogger) Critical(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Critical(msg)
	fmt.Fprint(sl, msg)
}

func (sl *simpleLogger) Debug(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Debug(msg)
	fmt.Fprint(sl, msg)
}

func (sl *simpleLogger) Info(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Info(msg)
	fmt.Fprint(sl, msg)
}

func (sl *simpleLogger) SpecStart(heading string) {
	msg := formatSpec(heading)
	GaugeLog.Info(msg)
	sl.printViaWriter(msg + newline)
	sl.writer.Reset()
}

func (sl *simpleLogger) SpecEnd() {
	sl.printViaWriter("")
	sl.writer.Reset()
}

func (sl *simpleLogger) ScenarioStart(scenarioHeading string) {
	sl.indentation = scenarioIndentation
	msg := formatScenario(scenarioHeading)
	GaugeLog.Info(msg)

	indentedText := indent(msg, sl.indentation)
	if level == logging.INFO {
		sl.headingText.WriteString(indentedText + spaces(4))
		sl.printViaWriter(sl.headingText.String())
	} else {
		sl.printViaWriter(indentedText + newline)
	}
	sl.writer.Reset()
}

func (sl *simpleLogger) ScenarioEnd(failed bool) {
	sl.printViaWriter("")
	if level == logging.INFO {
		sl.printViaWriter(newline)
		if failed {
			sl.printViaWriter(sl.buffer.String())
		}
	}
	sl.writer.Reset()
	sl.indentation -= scenarioIndentation
}

func (sl *simpleLogger) StepStart(stepText string) {
	sl.indentation += stepIndentation
	GaugeLog.Debug(stepText)
	if level == logging.DEBUG {
		sl.headingText.WriteString(indent(stepText, sl.indentation))
		sl.printViaWriter(sl.headingText.String() + newline)
	}
}

func (sl *simpleLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		sl.writer.Clear()
		if failed {
			sl.printViaWriter(sl.headingText.String() + "\t ...[FAIL]\n")
		} else {
			sl.printViaWriter(sl.headingText.String() + "\t ...[PASS]\n")
		}
		sl.printViaWriter(sl.buffer.String())
		sl.Reset()
	} else {
		if failed {
			sl.printViaWriter(getFailureSymbol())
		} else {
			sl.printViaWriter(getSuccessSymbol())
			sl.Reset()
		}
	}
	sl.writer.Reset()
	sl.indentation -= stepIndentation
}

func (sl *simpleLogger) Reset() {
	sl.headingText.Reset()
	sl.buffer.Reset()
}

func (sl *simpleLogger) ConceptStart(conceptHeading string) {
	sl.indentation += stepIndentation
	GaugeLog.Debug(conceptHeading)
	if level == logging.DEBUG {
		sl.printViaWriter(indent(conceptHeading, sl.indentation) + newline)
		sl.writer.Reset()
	}
}

func (sl *simpleLogger) ConceptEnd(failed bool) {
	sl.indentation -= stepIndentation
}

func (sl *simpleLogger) DataTable(table string) {
	GaugeLog.Debug(table)
	if level == logging.DEBUG {
		sl.printViaWriter(table + newline)
		sl.writer.Reset()
	}
}

func (sl *simpleLogger) printViaWriter(text string) {
	fmt.Fprint(sl.writer, text)
	sl.writer.Print()
}
