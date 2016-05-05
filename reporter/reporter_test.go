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
	. "gopkg.in/check.v1"
    "github.com/getgauge/gauge/execution/event"
)

func (s *MySuite) TestSubscribeSpecEnd(c *C) {
    dw, sc := setupSimpleConsole()
    currentReporter = sc
    SimpleConsoleOutput = true
    event.InitRegistry()

    ListenExecutionEvents()

    event.Notify(event.NewExecutionEvent(event.SpecEnd, nil, nil))
    c.Assert(dw.output, Equals, "\n")
}
