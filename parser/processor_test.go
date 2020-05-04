/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestProcessTable(c *C) {
	t := &Token{Kind: gauge.TableRow, Value: "|first second third    |"}
	errors, _ := processTable(new(SpecParser), t)

	c.Assert(len(errors), Equals, 0)
	c.Assert(t.Args[0], Equals, "first second third")
}
