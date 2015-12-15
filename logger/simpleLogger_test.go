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
	"bytes"
	"strings"

	. "gopkg.in/check.v1"
)

func (s *MySuite) TestStepStartAndStepEnd_SimpleLogger(c *C) {
	Initialize(true, "Debug")
	b := &bytes.Buffer{}
	sl := newSimpleConsoleWriter()
	sl.writer.Out = b

	input := "* Say hello to all"
	sl.StepStart(input)
	expectedStepStartOutput := spaces(stepIndentation) + "* Say hello to all\n"
	c.Assert(b.String(), Equals, expectedStepStartOutput)
	b.Reset()

	sl.StepEnd(true)

	if isWindows {
		c.Assert(b.String(), Equals, strings.Repeat(cursorLeftWindows+eraseCharWindows, len(expectedStepStartOutput))+spaces(stepIndentation)+"* Say hello to all\t ...[FAIL]\n")
	} else {
		c.Assert(b.String(), Equals, cursorUpUnix+eraseLineUnix+spaces(stepIndentation)+"* Say hello to all\t ...[FAIL]\n")
	}
}
