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

import "github.com/op/go-logging"

func (g *GaugeLogger) Specification(msg string) {
	g.Info("# %s", msg)
}

func (g *GaugeLogger) Scenario(msg string) {
	g.Info("  ## %s", msg)
}

func (g *GaugeLogger) Step(msg string) {
	g.Debug("    %s", msg)
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
