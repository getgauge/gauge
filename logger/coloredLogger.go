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
	writer       *uilive.Writer
	currentText  bytes.Buffer
	sysoutBuffer bytes.Buffer
}

func newColoredConsoleWriter() *coloredLogger {
	return &coloredLogger{writer: uilive.New()}
}

func (cLogger *coloredLogger) writeSysoutBuffer(text string) {
	if level == logging.DEBUG {
		text = strings.Replace(text, "\n", "\n\t", -1)
		cLogger.sysoutBuffer.WriteString(text + "\n")
		cLogger.writeln(cLogger.currentText.String()+cLogger.sysoutBuffer.String(), ct.None, false)
	}
}

func (cLogger *coloredLogger) SpecStart(heading string) {
	msg := formatSpec(heading)
	Log.Info(msg)
	ct.Foreground(ct.Cyan, true)
	ConsoleWrite(msg)
	fmt.Println()
	ct.ResetColor()
}

func (coloredLogger *coloredLogger) SpecEnd() {
}

func (cLogger *coloredLogger) ScenarioStart(scenarioHeading string) {
	msg := formatScenario(scenarioHeading)
	Log.Info(msg)

	indentedText := indent(msg, scenarioIndentation)
	if level == logging.INFO {
		cLogger.writer.Start()
		cLogger.currentText.WriteString(indentedText + "\n")
		cLogger.writeln(cLogger.currentText.String(), ct.None, false)
	} else {
		ct.Foreground(ct.Yellow, true)
		ConsoleWrite(indentedText)
		ct.ResetColor()
	}
}

func (cLogger *coloredLogger) ScenarioEnd(failed bool) {
	if level == logging.INFO {
		if failed {
			cLogger.write(cLogger.currentText.String(), ct.Red, true)
		} else {
			cLogger.write(cLogger.currentText.String(), ct.Green, true)
		}
		cLogger.writer.Flush()
		cLogger.writer.Stop()
	}
	cLogger.resetColoredLogger()
}

func (cLogger *coloredLogger) resetColoredLogger() {
	cLogger.writer = uilive.New()
	cLogger.currentText.Reset()
	cLogger.sysoutBuffer.Reset()
}

func (cLogger *coloredLogger) StepStart(stepText string) {
	Log.Debug(stepText)
	if level == logging.DEBUG {
		cLogger.writer.Start()
		cLogger.currentText.WriteString(indent(stepText, stepIndentation) + "\n")
		cLogger.writeln(cLogger.currentText.String(), ct.None, false)
	}
}

func (cLogger *coloredLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		if failed {
			cLogger.write(cLogger.currentText.String()+cLogger.sysoutBuffer.String(), ct.Red, false)
		} else {
			cLogger.write(cLogger.currentText.String()+cLogger.sysoutBuffer.String(), ct.Green, false)
		}
		cLogger.writer.Stop()
		cLogger.resetColoredLogger()
	} else {
		cLogger.sysoutBuffer.Reset()
	}
}

func (cLogger *coloredLogger) writeln(text string, color ct.Color, isBright bool) {
	cLogger.write(text+"\n", color, isBright)
}

func (cLogger *coloredLogger) write(text string, color ct.Color, isBright bool) {
	ct.Foreground(color, isBright)
	fmt.Fprint(cLogger.writer, text)
	cLogger.writer.Flush()
	ct.ResetColor()
}
