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

package parser

import (
	"github.com/getgauge/gauge/gauge"
	. "github.com/go-check/check"
)

func (s *MySuite) TestProcessTable(c *C) {
	t := &Token{Kind: gauge.TableRow, Value: "|first second third    |"}
	errors, _ := processTable(new(SpecParser), t)

	c.Assert(len(errors), Equals, 0)
	c.Assert(t.Args[0], Equals, "first second third")
}
