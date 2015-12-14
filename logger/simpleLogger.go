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

func (sl *simpleLogger) Warning(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Warning(msg, args)
	fmt.Fprint(sl, msg)
}

func (sl *simpleLogger) Info(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Info(msg, args)
	fmt.Fprint(sl, msg)
}

func (sl *simpleLogger) Debug(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	GaugeLog.Debug(msg, args)
	fmt.Fprint(sl, msg)
}

func (sl *simpleLogger) SpecStart(heading string) {
	msg := formatSpec(heading)
	GaugeLog.Info(msg)
	fmt.Println(msg + newline)
}

func (simpleLogger *simpleLogger) SpecEnd() {
	fmt.Println()
}

func (sl *simpleLogger) ScenarioStart(scenarioHeading string) {
	sl.indentation = scenarioIndentation
	msg := formatScenario(scenarioHeading)
	GaugeLog.Info(msg)

	indentedText := indent(msg, sl.indentation)
	if level == logging.INFO {
		sl.headingText.WriteString(indentedText + spaces(4))
		fmt.Print(sl.headingText.String())
	} else {
		fmt.Println(indentedText)
	}
}

func (sl *simpleLogger) ScenarioEnd(failed bool) {
	fmt.Println()
	if level == logging.INFO && failed {
		fmt.Print(sl.buffer.String())
	}
	sl.indentation -= scenarioIndentation
}

func (sl *simpleLogger) StepStart(stepText string) {
	sl.indentation += stepIndentation
	GaugeLog.Debug(stepText)
	if level == logging.DEBUG {
		sl.headingText.WriteString(indent(stepText, sl.indentation))
		fmt.Fprintln(sl.writer, sl.headingText.String())
		sl.writer.Print()
	}
}

func (sl *simpleLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		sl.writer.Clear()
		if failed {
			fmt.Fprint(sl.writer, sl.headingText.String()+"\t ...[FAIL]\n"+sl.buffer.String())
			sl.writer.Print()
		} else {
			fmt.Fprint(sl.writer, sl.headingText.String()+"\t ...[PASS]\n"+sl.buffer.String())
			sl.writer.Print()
		}
		sl.writer.Reset()
		sl.Reset()
	} else {
		if failed {
			fmt.Print(getFailureSymbol())
		} else {
			fmt.Print(getSuccessSymbol())
			sl.Reset()
		}
	}
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
		fmt.Println(indent(conceptHeading, sl.indentation))
	}
}

func (sl *simpleLogger) DataTable(table string) {
	GaugeLog.Debug(table)
	if level == logging.DEBUG {
		fmt.Println(table)
	}
}

func (sl *simpleLogger) ConceptEnd(failed bool) {
	sl.indentation -= stepIndentation
}
