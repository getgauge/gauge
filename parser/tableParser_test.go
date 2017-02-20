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
	. "github.com/go-check/check"
)

func (s *MySuite) TestIgnoreCommentLines(c *C) {
	csvContents := "one,two,three\nfoo,bar,baz\n#john,jim,jan"
	table, err := convertCsvToTable(csvContents)

	if err != nil {
		c.Fail()
	}

	c.Assert(table.GetRowCount(), Equals, 1)
	c.Assert(table.Rows()[0][0], Equals, "foo")
	c.Assert(table.Rows()[0][1], Equals, "bar")
	c.Assert(table.Rows()[0][2], Equals, "baz")
}
