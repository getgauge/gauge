package logger

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/gosuri/uilive"
	"github.com/op/go-logging"
	"strings"
)

type coloredLogger struct {
	writer        *uilive.Writer
	currentText   string
	sysoutBuffer  string
	linesToGoDown int
}

func newColoredConsoleWriter() *coloredLogger {
	return &coloredLogger{writer: uilive.New()}
}

func (cLogger *coloredLogger) writeSysoutBuffer(text string) {
	if level == logging.DEBUG {
		text = strings.Replace(text, "\n", "\n\t", -1)
		cLogger.sysoutBuffer += text + "\n"
		cLogger.writeln(cLogger.currentText+cLogger.sysoutBuffer, ct.None, false)
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
	cLogger.writer.Start()

	indentedText := indent(msg, scenarioIndentation)
	if level == logging.INFO {
		cLogger.currentText = indentedText + "\n"
		cLogger.writeln(cLogger.currentText, ct.None, false)
	} else {
		ct.Foreground(ct.Yellow, true)
		ConsoleWrite(indentedText)
		ct.ResetColor()
	}
}

func (cLogger *coloredLogger) ScenarioEnd(failed bool) {
	if level == logging.INFO {
		if failed {
			cLogger.write(cLogger.currentText, ct.Red, true)
		} else {
			cLogger.write(cLogger.currentText, ct.Green, true)
		}
		cLogger.writer.Flush()
	}
	cLogger.writer.Stop()
	cLogger.resetColoredLogger()
}

func (cLogger *coloredLogger) resetColoredLogger() {
	cLogger.writer = uilive.New()
	cLogger.currentText = ""
	cLogger.sysoutBuffer = ""
	cLogger.linesToGoDown = 0
}

func (cLogger *coloredLogger) StepStart(stepText string) {
	Log.Debug(stepText)
	if level == logging.DEBUG {
		for i := 0; i < cLogger.linesToGoDown; i++ {
			fmt.Println()
		}
		cLogger.currentText = indent(stepText, stepIndentation) + "\n"
		cLogger.writeln(cLogger.currentText, ct.None, false)
	}
}

func (cLogger *coloredLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		if failed {
			cLogger.write(cLogger.currentText, ct.Red, false)
		} else {
			cLogger.write(cLogger.currentText, ct.Green, false)
		}
		count := strings.Count(cLogger.currentText, "\n")
		for i := 0; i < count; i++ {
			fmt.Println()
		}
		if len(cLogger.sysoutBuffer) != 0 {
			cLogger.write(cLogger.sysoutBuffer, ct.None, false)
		}

		cLogger.linesToGoDown = strings.Count(cLogger.sysoutBuffer, "\n")
		cLogger.sysoutBuffer = ""
	} else {
		cLogger.sysoutBuffer = ""
		cLogger.linesToGoDown = 0
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
