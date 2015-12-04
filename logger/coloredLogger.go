package logger

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/gosuri/uilive"
	"github.com/op/go-logging"
	"time"
)

type coloredLogger struct {
	writer      *uilive.Writer
	currentText string
}

func newColoredConsoleWriter() *coloredLogger {
	return &coloredLogger{writer: uilive.New(), currentText: ""}
}

func (cLogger *coloredLogger) writeSysoutBuffer(text string) {
	cLogger.currentText += text
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

func (cLogger *coloredLogger) SpecStart(heading string) {
	msg := formatSpec(heading)
	Log.Info(msg)
	ct.Foreground(ct.None, true)
	fmt.Println(msg)
	fmt.Println()
	ct.ResetColor()
}

func (coloredLogger *coloredLogger) SpecEnd() {
	fmt.Println()
}

func (cLogger *coloredLogger) ScenarioStart(scenarioHeading string) {
	msg := formatScenario(scenarioHeading)
	Log.Info(msg)

	indentedText := indent(msg, scenarioIndentation)
	if level == logging.INFO {
		cLogger.writer.Start()
		ct.Foreground(ct.Yellow, false)

		fmt.Fprintln(cLogger.writer, indentedText)
		cLogger.currentText = indentedText
		time.Sleep(time.Millisecond * 10)

		ct.ResetColor()
	} else {
		cLogger.writer.Start()
		ct.Foreground(ct.Cyan, true)
		cLogger.ConsoleWrite(indentedText)
		ct.ResetColor()
	}
}

func (cLogger *coloredLogger) ScenarioEnd(failed bool) {
	if level == logging.INFO {
		if failed {
			ct.Foreground(ct.Red, true)
		} else {
			ct.Foreground(ct.Green, true)
		}

		fmt.Fprintln(cLogger.writer, cLogger.currentText)
		time.Sleep(time.Millisecond * 10)

		ct.ResetColor()
	} else {
		cLogger.writer.Stop()
		cLogger.currentText = ""
	}
	if level == logging.INFO {
		cLogger.writer.Flush()
		cLogger.writer.Stop()
	}
	cLogger.writer = uilive.New()
	cLogger.currentText = ""
}

func (cLogger *coloredLogger) StepStart(stepText string) {
	Log.Debug(stepText)
	if level == logging.DEBUG {
		ct.Foreground(ct.Yellow, true)

		cLogger.currentText = indent(stepText, stepIndentation)
		fmt.Fprintln(cLogger.writer, cLogger.currentText)
		time.Sleep(time.Millisecond * 10)

		ct.ResetColor()
	}
}

func (cLogger *coloredLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		if failed {
			ct.Foreground(ct.Red, true)
		} else {
			ct.Foreground(ct.Green, true)
		}

		fmt.Fprintln(cLogger.writer, cLogger.currentText)
		time.Sleep(time.Millisecond * 10)

		ct.ResetColor()
		cLogger.writer.Flush()
		fmt.Println()
		cLogger.currentText = ""
	}
}
