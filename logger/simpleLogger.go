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
}

func newSimpleConsoleWriter() *simpleLogger {
	return &simpleLogger{writer: goterminal.New()}
}

func (sl *simpleLogger) Write(b []byte) (int, error) {
	text := strings.Trim(string(b), "\n ")
	text = strings.Replace(text, newline, newline+spaces(sysoutIndentation), -1)
	if len(text) > 0 {
		msg := spaces(sysoutIndentation) + text + newline
		sl.buffer.WriteString(msg)
		if level == logging.DEBUG {
			fmt.Fprint(sl.writer, msg)
			sl.writer.Print()
		}
	}
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
	msg := formatScenario(scenarioHeading)
	GaugeLog.Info(msg)

	indentedText := indent(msg, scenarioIndentation)
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
}

func (sl *simpleLogger) StepStart(stepText string) {
	GaugeLog.Debug(stepText)
	if level == logging.DEBUG {
		sl.headingText.WriteString(indent(stepText, stepIndentation))
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
}

func (sl *simpleLogger) Reset() {
	sl.headingText.Reset()
	sl.buffer.Reset()
}

func (sl *simpleLogger) ConceptStart(conceptHeading string) {
	GaugeLog.Debug(conceptHeading)
	if level == logging.DEBUG {
		fmt.Println(indent(conceptHeading, stepIndentation))
	}
}
