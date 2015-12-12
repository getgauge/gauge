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
	if level == logging.DEBUG {
		text := strings.Trim(string(b), "\n ")
		text = strings.Replace(text, newline, "\n\t", -1)
		if len(text) > 0 {
			msg := fmt.Sprintf("\t%s\n", text)
			cl.buffer.WriteString(msg)
			cl.print(msg, ct.None, false)
		}
	}
	return len(b), nil
}

func (cl *coloredLogger) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Error(msg, args)
	cl.buffer.WriteString(msg + newline)
	if level == logging.DEBUG {
		cl.print(msg+newline, ct.Red, false)
	}
}

func (cl *coloredLogger) Critical(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Critical(msg, args)
	cl.buffer.WriteString(msg + newline)
	if level == logging.DEBUG {
		cl.print(msg+newline, ct.Red, false)
	}
}

func (cl *coloredLogger) Warning(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Warning(msg, args)
	cl.buffer.WriteString(msg + newline)
	if level == logging.DEBUG {
		cl.print(msg+newline, ct.Yellow, false)
	}
}

func (cl *coloredLogger) Info(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Info(msg, args)
	cl.buffer.WriteString(msg + newline)
	if level == logging.DEBUG {
		cl.print(msg+newline, ct.None, false)
	}
}

func (cl *coloredLogger) Debug(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Debug(msg, args)
	cl.buffer.WriteString(msg + newline)
	if level == logging.DEBUG {
		cl.print(msg+newline, ct.None, false)
	}
}

func (cl *coloredLogger) SpecStart(heading string) {
	msg := formatSpec(heading)
	Log.Info(msg)
	cl.writeToConsole(newline+msg+newline+newline, ct.Cyan, true)
}

func (coloredLogger *coloredLogger) SpecEnd() {
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
	if level == logging.INFO {
		fmt.Println()
		cl.writeToConsole(cl.buffer.String(), ct.Red, false)
	}
}

func (cl *coloredLogger) StepStart(stepText string) {
	Log.Debug(stepText)
	if level == logging.DEBUG {
		cl.headingText.WriteString(indent(stepText, stepIndentation) + newline)
		cl.print(cl.headingText.String(), ct.None, false)
	}
}

func (cl *coloredLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		cl.writer.Clear()
		heading := strings.Trim(cl.headingText.String(), newline)
		if failed {
			cl.print(heading+"\t ...[FAIL]"+newline, ct.Red, false)
			cl.print(cl.buffer.String(), ct.Red, false)
		} else {
			cl.print(heading+"\t ...[PASS]"+newline, ct.Green, false)
			cl.print(cl.buffer.String(), ct.None, false)
		}
		cl.Reset()
	} else {
		if failed {
			cl.writeToConsole(getFailureSymbol(), ct.Red, false)
		} else {
			cl.writeToConsole(getSuccessSymbol(), ct.Green, false)
		}
	}
}

func (cl *coloredLogger) Reset() {
	cl.writer.Reset()
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
