/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	. "gopkg.in/check.v1"
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
