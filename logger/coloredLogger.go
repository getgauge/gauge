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
	headingText  bytes.Buffer
	sysoutBuffer bytes.Buffer
}

func newColoredConsoleWriter() *coloredLogger {
	return &coloredLogger{writer: uilive.New()}
}

func (cl *coloredLogger) Write(b []byte) (int, error) {
	if level == logging.DEBUG {
		text := strings.Trim(string(b), "\n ")
		text = strings.Replace(text, "\n", "\n\t", -1)
		if len(text) > 0 {
			cl.sysoutBuffer.WriteString(fmt.Sprintf("\t%s\n", text))
			cl.write(cl.headingText.String()+cl.sysoutBuffer.String(), ct.None, false)
		}
	}
	return len(b), nil
}

func (cl *coloredLogger) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Error(msg, args)
	cl.sysoutBuffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		cl.write(cl.headingText.String()+cl.sysoutBuffer.String(), ct.Red, false)
	}
}

func (cl *coloredLogger) Critical(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Critical(msg, args)
	cl.sysoutBuffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		cl.write(cl.headingText.String()+cl.sysoutBuffer.String(), ct.Red, false)
	}
}

func (cl *coloredLogger) Warning(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Warning(msg, args)
	cl.sysoutBuffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		cl.write(cl.headingText.String()+cl.sysoutBuffer.String(), ct.Yellow, false)
	}
}

func (cl *coloredLogger) Info(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Info(msg, args)
	cl.sysoutBuffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		cl.write(cl.headingText.String()+cl.sysoutBuffer.String(), ct.None, false)
	}
}

func (cl *coloredLogger) Debug(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	Log.Debug(msg, args)
	cl.sysoutBuffer.WriteString(msg + "\n")
	if level == logging.DEBUG {
		cl.write(cl.headingText.String()+cl.sysoutBuffer.String(), ct.None, false)
	}
}

func (cl *coloredLogger) SpecStart(heading string) {
	msg := formatSpec(heading)
	Log.Info(msg)
	ct.Foreground(ct.Cyan, true)
	ConsoleWrite(msg)
	fmt.Println()
	ct.ResetColor()
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
		cl.writeToConsole(cl.sysoutBuffer.String(), ct.Red, false)
	}
	cl.resetColoredLogger()
}

func (cl *coloredLogger) resetColoredLogger() {
	cl.writer = uilive.New()
	cl.headingText.Reset()
	cl.sysoutBuffer.Reset()
}

func (cl *coloredLogger) StepStart(stepText string) {
	Log.Debug(stepText)
	if level == logging.DEBUG {
		cl.writer.Start()
		cl.headingText.WriteString(indent(stepText, stepIndentation) + "\n")
		cl.write(cl.headingText.String(), ct.None, false)
	}
}

func (cl *coloredLogger) StepEnd(failed bool) {
	if level == logging.DEBUG {
		if failed {
			cl.write(cl.headingText.String()+cl.sysoutBuffer.String(), ct.Red, false)
		} else {
			cl.write(cl.headingText.String()+cl.sysoutBuffer.String(), ct.Green, false)
		}
		cl.writer.Stop()
		cl.resetColoredLogger()
	} else {
		if failed {
			cl.writeToConsole(getFailureSymbol(), ct.Red, false)
		} else {
			cl.writeToConsole(getSuccessSymbol(), ct.Green, false)
		}
	}
}

func (cl *coloredLogger) writeln(text string, color ct.Color, isBright bool) {
	cl.write(text+"\n", color, isBright)
}

func (cl *coloredLogger) write(text string, color ct.Color, isBright bool) {
	ct.Foreground(color, isBright)
	fmt.Fprint(cl.writer, text)
	cl.writer.Flush()
	ct.ResetColor()
}

func (cl *coloredLogger) writeToConsole(text string, color ct.Color, isBright bool) {
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
