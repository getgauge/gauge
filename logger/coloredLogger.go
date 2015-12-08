package logger

import (
	"bytes"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/gosuri/uilive"
	"github.com/op/go-logging"
	"strings"
)

const (
	success     = "✔ "
	failure     = "✘ "
	successChar = "P"
	failureChar = "F"
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

func (cLogger *coloredLogger) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Error(msg, args)
	cLogger.sysoutBuffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		cLogger.writeln(cLogger.currentText.String()+cLogger.sysoutBuffer.String(), ct.Red, false)
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
		cLogger.currentText.WriteString(indentedText + spaces(4))
		cLogger.writeToConsole(cLogger.currentText.String(), ct.None, false)
	} else {
		ct.Foreground(ct.Yellow, true)
		ConsoleWrite(indentedText)
		ct.ResetColor()
	}
}

func (cLogger *coloredLogger) ScenarioEnd(failed bool) {
	if level == logging.INFO {
		fmt.Println()
		cLogger.writeToConsole(cLogger.sysoutBuffer.String(), ct.Red, false)
		fmt.Println()
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
		if failed {
			cLogger.writeToConsole(getFailureSymbol(), ct.Red, false)
		} else {
			cLogger.writeToConsole(getSuccessSymbol(), ct.Green, false)
		}
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

func (cLogger *coloredLogger) writeToConsole(text string, color ct.Color, isBright bool) {
	ct.Foreground(color, isBright)
	fmt.Print(text)
	ct.ResetColor()
}

func getFailureSymbol() string {
	if isWindows {
		return spaces(1) + failureChar
	}
	return spaces(1) + failure
}

func getSuccessSymbol() string {
	if isWindows {
		return spaces(1) + successChar
	}
	return spaces(1) + success
}
