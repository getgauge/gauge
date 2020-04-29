/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package reporter

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestIndent(c *C) {
	c.Assert(indent("foo bar", 2), Equals, "  foo bar")
	c.Assert(indent("\nfoo bar", 2), Equals, "  \n  foo bar")
}
