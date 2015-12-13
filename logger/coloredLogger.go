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
}

func newColoredConsoleWriter() *coloredLogger {
	return &coloredLogger{writer: goterminal.New()}
}

func (cl *coloredLogger) Write(b []byte) (int, error) {
	text := strings.Trim(string(b), "\n ")
	text = strings.Replace(text, newline, newline+spaces(sysoutIndentation), -1)
	if len(text) > 0 {
		msg := spaces(sysoutIndentation) + text + newline
		cl.buffer.WriteString(msg)
		if level == logging.DEBUG {
			fmt.Fprint(cl.writer, msg)
			cl.writer.Print()
		}
	}
	return len(b), nil
}

func (cl *coloredLogger) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Error(msg)
	fmt.Fprint(cl, msg)
}

func (cl *coloredLogger) Critical(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Critical(msg)
	cl.Write([]byte(msg))
}

func (cl *coloredLogger) Warning(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Warning(msg)
	fmt.Fprint(cl, msg)
}

func (cl *coloredLogger) Info(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Info(msg)
	fmt.Fprint(cl, msg)
}

func (cl *coloredLogger) Debug(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Debug(msg)
	fmt.Fprint(cl, msg)
}

func (cl *coloredLogger) SpecStart(heading string) {
	msg := formatSpec(heading)
	Log.Info(msg)
	cl.writeToConsole(msg+newline+newline, ct.Cyan, true)
}

func (coloredLogger *coloredLogger) SpecEnd() {
	fmt.Println()
}

func (cl *coloredLogger) ScenarioStart(scenarioHeading string) {
	msg := formatScenario(scenarioHeading)
	Log.Info(msg)

	indentedText := indent(msg, scenarioIndentation)
	if level == logging.INFO {
		cl.headingText.WriteString(indentedText + spaces(4))
		cl.writeToConsole(cl.headingText.String(), ct.None, false)
	} else {
		ct.Foreground(ct.Yellow, false)
		ConsoleWrite(indentedText)
		ct.ResetColor()
	}
}

func (cl *coloredLogger) ScenarioEnd(failed bool) {
	fmt.Println()
	if level == logging.INFO && failed {
		cl.writeToConsole(cl.buffer.String(), ct.Red, false)
	}
}

func (cl *coloredLogger) StepStart(stepText string) {
	Log.Debug(stepText)
	if level == logging.DEBUG {
		cl.headingText.WriteString(indent(stepText, stepIndentation))
		cl.print(cl.headingText.String()+newline, ct.None, false)
	}
}

func (cl *coloredLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		cl.writer.Clear()
		if failed {
			cl.print(cl.headingText.String()+"\t ...[FAIL]\n", ct.Red, false)
			cl.print(cl.buffer.String(), ct.Red, false)
		} else {
			cl.print(cl.headingText.String()+"\t ...[PASS]\n", ct.Green, false)
			cl.print(cl.buffer.String(), ct.None, false)
		}
		cl.writer.Reset()
		cl.Reset()
	} else {
		if failed {
			cl.writeToConsole(getFailureSymbol(), ct.Red, false)
		} else {
			cl.writeToConsole(getSuccessSymbol(), ct.Green, false)
			cl.Reset()
		}
	}
}

func (cl *coloredLogger) Reset() {
	cl.buffer.Reset()
	cl.headingText.Reset()
}

func (cl *coloredLogger) ConceptStart(conceptHeading string) {
	Log.Debug(conceptHeading)
	if level == logging.DEBUG {
		cl.writeToConsole(indent(conceptHeading, stepIndentation)+newline, ct.Magenta, false)
	}
}

func (cl *coloredLogger) print(text string, color ct.Color, isBright bool) {
	ct.Foreground(color, isBright)
	fmt.Fprint(cl.writer, text)
	cl.writer.Print()
	ct.ResetColor()
}

func (cl *coloredLogger) writeToConsole(text string, color ct.Color, isBright bool) {
	ct.Foreground(color, isBright)
	fmt.Print(text)
	ct.ResetColor()
}
