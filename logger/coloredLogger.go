package logger

import (
	"bytes"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/gosuri/uilive"
	"github.com/op/go-logging"
	"strings"
)

type coloredLogger struct {
	writer *uilive.Writer

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
	cLogger.writer.Start()
	if level == logging.INFO {
		cLogger.start(indent(msg, scenarioIndentation))
	} else {

		fmt.Println(indent(msg, scenarioIndentation))
		fmt.Println()
	}
}

func (cLogger *coloredLogger) ScenarioEnd(scenarioHeading string, failed bool) {
	msg := formatScenario(scenarioHeading)
	if level == logging.INFO {
		cLogger.end(indent(msg, scenarioIndentation), failed)
	}
	cLogger.writer.Stop()
}

func (cLogger *coloredLogger) StepStart(stepText string) {
	Log.Debug(stepText)
	if level == logging.DEBUG {
		cLogger.start(indent(stepText, stepIndentation))
	}
}

func (cLogger *coloredLogger) StepEnd(stepText string, failed bool) {
	if level == logging.DEBUG {
		cLogger.end(indent(stepText, stepIndentation), failed)
	}
}

func (cLogger *coloredLogger) Table(table string) {
}

func (cLogger *coloredLogger) start(text string) {
	ct.Foreground(ct.Cyan, true)
	fragments := strings.Split(text, "\n")
	for i := 0; i < len(fragments); i++ {
		fmt.Fprintln(cLogger.writer, fragments[i])
	}
	cLogger.writer.Flush()
	ct.ResetColor()
}

func (cLogger *coloredLogger) end(text string, failed bool) {
	if failed {
		ct.Foreground(ct.Red, true)
	} else {
		ct.Foreground(ct.Green, true)
	}
	fragments := strings.Split(text, "\n")
	for i := 0; i < len(fragments); i++ {
		fmt.Fprintln(cLogger.writer, fragments[i])
	}
	cLogger.writer.Flush()
	ct.ResetColor()
	fmt.Println()
}
