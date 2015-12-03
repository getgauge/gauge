package logger

import (
	"bytes"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/gosuri/uilive"
	"github.com/op/go-logging"
)

type coloredLogger struct {
	writer       *uilive.Writer
	sysoutBuffer *bytes.Buffer
}

func newColoredConsoleWriter() *coloredLogger {
	return &coloredLogger{writer: uilive.New(), sysoutBuffer: new(bytes.Buffer)}
}

func (cLogger *coloredLogger) writeSysoutBuffer(text string) {
	cLogger.sysoutBuffer.WriteString(text)
}

func (cLogger *coloredLogger) Info(msg string, args ...interface{}) {
	Log.Info(msg, args...)
	cLogger.ConsoleWrite(msg, args...)
}

func (cLogger *coloredLogger) Debug(msg string, args ...interface{}) {
	Log.Info(msg, args...)
	if level == logging.DEBUG {
		cLogger.ConsoleWrite(msg, args...)
	}
}

func (cLogger *coloredLogger) ConsoleWrite(msg string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, args...))
}

func (cLogger *coloredLogger) Spec(heading string) {
	msg := formatSpec(heading)
	cLogger.Info(msg)
}

func (cLogger *coloredLogger) ScenarioStart(scenarioHeading string) {
	msg := formatScenario(scenarioHeading)
	Log.Info(msg)
	if level == logging.DEBUG {
		cLogger.ConsoleWrite(indent(msg, scenarioIndentation))
		fmt.Println()
	} else {
		cLogger.writer.Start()
		ct.Foreground(ct.Cyan, true)
		fmt.Fprintln(cLogger.writer, indent(msg, scenarioIndentation))
		ct.ResetColor()
	}
}

func (cLogger *coloredLogger) ScenarioEnd(scenarioHeading string, failed bool) {
	msg := formatScenario(scenarioHeading)
	if level == logging.INFO {
		if failed {
			ct.Foreground(ct.Red, true)
		} else {
			ct.Foreground(ct.Green, true)
		}
		fmt.Fprintln(cLogger.writer, indent(msg, scenarioIndentation))

		cLogger.writer.Stop()
		ct.ResetColor()
		fmt.Println()
	}
}

func (cLogger *coloredLogger) StepStart(stepText string) {
	if level == logging.DEBUG {
		cLogger.writer.Start()
		ct.Foreground(ct.Cyan, true)
		fmt.Fprintln(cLogger.writer, indent(stepText, stepIndentation))
		ct.ResetColor()
	}
}

func (cLogger *coloredLogger) StepEnd(stepText string, failed bool) {
	if level == logging.DEBUG {
		if failed {
			ct.Foreground(ct.Red, true)
		} else {
			ct.Foreground(ct.Green, true)
		}
		fmt.Fprintln(cLogger.writer, indent(stepText, stepIndentation))

		cLogger.writer.Stop()
		ct.ResetColor()
		fmt.Println()
	}

}

func (cLogger *coloredLogger) Table(table string) {
}
