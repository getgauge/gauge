package main

import (
	. "gopkg.in/check.v1"

)

func(s *MySuite) TestEvaluateTags(c *C) {
	tags := evaluateTags("tag1 & tag2 & tag3")
	c.Assert(tags, Equals, "tag1 , tag2 , tag3")
	
	tags=evaluateTags("tag1 | tag2 | tag3")
	c.Assert(tags, Equals, "tag1 | tag2 | tag3")
}
