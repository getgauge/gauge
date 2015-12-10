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

	"github.com/apoorvam/uilive"
	"github.com/op/go-logging"
)

type simpleLogger struct {
	writer      *uilive.Writer
	headingText bytes.Buffer
	buffer      bytes.Buffer
}

func newSimpleConsoleWriter() *simpleLogger {
	return &simpleLogger{writer: uilive.New()}
}

func (sl *simpleLogger) Write(b []byte) (int, error) {
	if level == logging.DEBUG {
		text := strings.Trim(string(b), "\n ")
		text = strings.Replace(text, "\n", "\n\t", -1)
		if len(text) > 0 {
			sl.buffer.WriteString(fmt.Sprintf("\t%s\n", text))
			sl.write(sl.headingText.String() + sl.buffer.String())
		}
	}
	return len(b), nil
}

func (sl *simpleLogger) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Error(msg, args)
	sl.buffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		sl.write(sl.headingText.String() + sl.buffer.String())
	}
}

func (sl *simpleLogger) Critical(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Critical(msg, args)
	sl.buffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		sl.write(sl.headingText.String() + sl.buffer.String())
	}
}

func (sl *simpleLogger) Warning(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Warning(msg, args)
	sl.buffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		sl.write(sl.headingText.String() + sl.buffer.String())
	}
}

func (sl *simpleLogger) Info(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Info(msg, args)
	sl.buffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		sl.write(sl.headingText.String() + sl.buffer.String())
	}
}

func (sl *simpleLogger) Debug(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Debug(msg, args)
	sl.buffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		sl.write(sl.headingText.String() + sl.buffer.String())
	}
}

func (sl *simpleLogger) SpecStart(heading string) {
	msg := formatSpec(heading)
	Log.Info(msg)
	fmt.Println()
	ConsoleWrite(msg)
	fmt.Println()
}

func (simpleLogger *simpleLogger) SpecEnd() {
}

func (sl *simpleLogger) ScenarioStart(scenarioHeading string) {
	msg := formatScenario(scenarioHeading)
	Log.Info(msg)

	indentedText := indent(msg, scenarioIndentation)
	if level == logging.INFO {
		sl.headingText.WriteString(indentedText + spaces(4))
		fmt.Print(sl.headingText.String())
	} else {
		ConsoleWrite(indentedText)
	}
}

func (sl *simpleLogger) ScenarioEnd(failed bool) {
	if level == logging.INFO {
		fmt.Println()
		fmt.Print(sl.buffer.String())
	}
	sl.resetsimpleLogger()
}

func (sl *simpleLogger) resetsimpleLogger() {
	sl.writer = uilive.New()
	sl.headingText.Reset()
	sl.buffer.Reset()
}

func (sl *simpleLogger) StepStart(stepText string) {
	Log.Debug(stepText)
	if level == logging.DEBUG {
		sl.writer.Start()
		sl.headingText.WriteString(indent(stepText, stepIndentation) + "\n")
		sl.write(sl.headingText.String())
	}
}

func (sl *simpleLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		heading := strings.Trim(sl.headingText.String(), "\n")
		if failed {
			sl.write(heading + "\t ...[FAIL]\n" + sl.buffer.String())
		} else {
			sl.write(heading + "\t ...[PASS]\n" + sl.buffer.String())
		}
		sl.writer.Stop()
		sl.resetsimpleLogger()
	} else {
		if failed {
			fmt.Print(getFailureSymbol())
		} else {
			fmt.Print(getSuccessSymbol())
		}
	}
}

func (sl *simpleLogger) writeln(text string) {
	sl.write(text + "\n")
}

func (sl *simpleLogger) write(text string) {
	fmt.Fprint(sl.writer, text)
	sl.writer.Flush()
}
