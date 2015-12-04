package logger

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/gosuri/uilive"
	"github.com/op/go-logging"
	"time"
)

type coloredLogger struct {
	writer *uilive.Writer
	myText string
}

func newColoredConsoleWriter() *coloredLogger {
	return &coloredLogger{writer: uilive.New(), myText: ""}
}

func (cLogger *coloredLogger) writeSysoutBuffer(text string) {
	cLogger.myText += text
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
	cLogger.writer.Start()

	indentedText := indent(msg, scenarioIndentation)
	if level == logging.INFO {
		ct.Foreground(ct.Cyan, false)

		fmt.Fprintln(cLogger.writer, indentedText)
		cLogger.myText = indentedText
		time.Sleep(time.Millisecond * 100)
		cLogger.writer.Flush()

		ct.ResetColor()
	} else {
		ct.Foreground(ct.Cyan, false)

		cLogger.myText += indentedText
		fmt.Fprintln(cLogger.writer, cLogger.myText)
		time.Sleep(time.Millisecond * 100)
		cLogger.writer.Flush()

		ct.ResetColor()
		cLogger.myText += "\n"
	}
}

func (cLogger *coloredLogger) ScenarioEnd(failed bool) {
	if level == logging.INFO {
		if failed {
			ct.Foreground(ct.Red, true)
		} else {
			ct.Foreground(ct.Green, true)
		}

		fmt.Fprintln(cLogger.writer, cLogger.myText)
		time.Sleep(time.Millisecond * 100)
		cLogger.writer.Flush()

		cLogger.myText += ""
		ct.ResetColor()
	} else {
		cLogger.myText = ""
	}
	fmt.Println()
	cLogger.writer.Stop()
}

func (cLogger *coloredLogger) StepStart(stepText string) {
	Log.Debug(stepText)
	if level == logging.DEBUG {
		ct.Foreground(ct.Cyan, true)

		cLogger.myText += indent(stepText, stepIndentation)
		fmt.Fprintln(cLogger.writer, cLogger.myText)
		time.Sleep(time.Millisecond * 100)
		cLogger.writer.Flush()

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

		fmt.Fprintln(cLogger.writer, cLogger.myText)
		time.Sleep(time.Millisecond * 100)
		cLogger.writer.Flush()

		cLogger.myText += "\n"
		ct.ResetColor()
		//		fmt.Println()
		//		cLogger.myText = ""
	}
}