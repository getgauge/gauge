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
	"fmt"

	"github.com/op/go-logging"
)

const (
	scenarioMsg = 1 << iota
	specMsg     = 1 << iota
	stepMsg     = 1 << iota
)

type message struct {
	value   string
	msgType int32
}

type consoleLogger struct{}

var console consoleLogger

func (c *consoleLogger) Info(msg string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, args...))
}

func (c *consoleLogger) Debug(msg string, args ...interface{}) {
	if level == logging.DEBUG {
		fmt.Println(fmt.Sprintf(msg, args...))
	}
}

func Print(msg string, args ...interface{}) {
	Log.Info(msg, args...)
	console.Debug(msg, args...)
}

func Summary(msg string, args ...interface{}) {
	Log.Info(msg, args...)
	console.Info(msg, args...)
}

func Specification(msg string) {
	formattedMsg := formatMessage(message{value: msg, msgType: specMsg})
	Log.Info(formattedMsg)
	console.Info(formattedMsg)
}

func Scenario(msg string) {
	formattedMsg := formatMessage(message{value: msg, msgType: scenarioMsg})
	Log.Info(formattedMsg)
	console.Info(formattedMsg)
}

func Step(msg string) {
	formattedMsg := formatMessage(message{value: msg, msgType: stepMsg})
	Log.Info(formattedMsg)
	console.Debug(formattedMsg)
}

func formatMessage(msg message) string {
	switch msg.msgType {
	case stepMsg:
		return formatStep(msg.value)
	case scenarioMsg:
		return formatScenario(msg.value)
	case specMsg:
		return formatSpec(msg.value)
	}
	return msg.value
}

func formatScenario(msg string) string {
	return fmt.Sprintf("  ## %s", msg)
}

func formatStep(msg string) string {
	return fmt.Sprintf("    %s", msg)
}

func formatSpec(msg string) string {
	return fmt.Sprintf("# %s", msg)
}

func (g *GaugeLogger) PrintScenarioResult(failed bool) {
	if level == logging.INFO {
		if !failed {
			g.Info("  ...[PASS]")
		} else {
			g.Info("  ...[FAIL]")
		}
	}
}

func (g *GaugeLogger) PrintStepResult(failed bool) {
	if !failed {
		g.Debug("      ...[PASS]")
	} else {
		g.Debug("      ...[FAIL]")
	}
}
