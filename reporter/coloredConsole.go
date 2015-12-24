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

package reporter

import (
	"bytes"
	"fmt"
	"io"

	"github.com/apoorvam/goterminal"
	ct "github.com/daviddengcn/go-colortext"
	"github.com/getgauge/gauge/logger"
)

type coloredConsole struct {
	writer               *goterminal.Writer
	headingBuffer        bytes.Buffer
	pluginMessagesBuffer bytes.Buffer
	indentation          int
}

func newColoredConsole(out io.Writer) *coloredConsole {
	return &coloredConsole{writer: goterminal.New(out)}
}

func (c *coloredConsole) SpecStart(heading string) {
	msg := formatSpec(heading)
	logger.GaugeLog.Info(msg)
	c.displayMessage(msg+newline, ct.Cyan)
	c.writer.Reset()
}

func (c *coloredConsole) SpecEnd() {
	c.displayMessage(newline, ct.None)
	c.writer.Reset()
}

func (c *coloredConsole) ScenarioStart(scenarioHeading string) {
	c.indentation = scenarioIndentation
	msg := formatScenario(scenarioHeading)
	logger.GaugeLog.Info(msg)

	indentedText := indent(msg, c.indentation)
	if !Verbose {
		c.headingBuffer.WriteString(indentedText + spaces(4))
		c.displayMessage(c.headingBuffer.String(), ct.None)
	} else {
		c.displayMessage(indentedText+newline, ct.Yellow)
	}
	c.writer.Reset()
}

func (c *coloredConsole) ScenarioEnd(failed bool) {
	if !Verbose {
		c.displayMessage(newline, ct.None)
		if failed {
			c.displayMessage(c.pluginMessagesBuffer.String(), ct.Red)
		}
	}
	c.writer.Reset()
	c.indentation -= scenarioIndentation
}

func (c *coloredConsole) StepStart(stepText string) {
	c.indentation += stepIndentation
	logger.GaugeLog.Debug(stepText)
	if Verbose {
		c.headingBuffer.WriteString(indent(stepText, c.indentation))
		c.displayMessage(c.headingBuffer.String()+newline, ct.None)
	}
}

func (c *coloredConsole) StepEnd(failed bool) {
	if Verbose {
		c.writer.Clear()
		if failed {
			c.displayMessage(c.headingBuffer.String()+"\t ...[FAIL]\n", ct.Red)
		} else {
			c.displayMessage(c.headingBuffer.String()+"\t ...[PASS]\n", ct.Green)
		}
		c.displayMessage(c.pluginMessagesBuffer.String(), ct.None)
		c.resetBuffers()
	} else {
		if failed {
			c.displayMessage(getFailureSymbol(), ct.Red)
		} else {
			c.displayMessage(getSuccessSymbol(), ct.Green)
			c.resetBuffers()
		}
	}
	c.writer.Reset()
	c.indentation -= stepIndentation
}

func (c *coloredConsole) ConceptStart(conceptHeading string) {
	c.indentation += stepIndentation
	logger.GaugeLog.Debug(conceptHeading)
	if Verbose {
		c.displayMessage(indent(conceptHeading, c.indentation)+newline, ct.Magenta)
		c.writer.Reset()
	}
}

func (c *coloredConsole) ConceptEnd(failed bool) {
	c.indentation -= stepIndentation
}

func (c *coloredConsole) DataTable(table string) {
	logger.GaugeLog.Debug(table)
	if Verbose {
		c.displayMessage(table+newline, ct.Yellow)
		c.writer.Reset()
	}
}

func (c *coloredConsole) Error(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args)
	logger.GaugeLog.Error(msg)
	fmt.Fprint(c, msg)
}

// Write writes the bytes to console via goterminal's writer.
// This is called when any sysouts are to be printed on console.
func (c *coloredConsole) Write(b []byte) (int, error) {
	c.indentation += sysoutIndentation
	text := indent(string(b), c.indentation)
	if len(text) > 0 {
		text += newline
		c.pluginMessagesBuffer.WriteString(text)
		if Verbose {
			c.displayMessage(text, ct.None)
		}
	}
	c.indentation -= sysoutIndentation
	return len(b), nil
}

func (c *coloredConsole) displayMessage(msg string, color ct.Color) {
	ct.Foreground(color, false)
	defer ct.ResetColor()
	fmt.Fprint(c.writer, msg)
	c.writer.Print()
}

func (c *coloredConsole) resetBuffers() {
	c.headingBuffer.Reset()
	c.pluginMessagesBuffer.Reset()
}
