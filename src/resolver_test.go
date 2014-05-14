package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestParsingSimpleSpecialType(c *C) {
	resolver := newSpecialTypeResolver()
	resolver.predefinedResolvers["file"] = func(value string) *stepArg {
		return &stepArg{value:"dummy", argType:static}
	}

	stepArg := resolver.resolve("file:foo")
	c.Assert(stepArg.value, Equals, "dummy")
	c.Assert(stepArg.argType, Equals, static)
}

func (s *MySuite) TestParsingUnknownSpecialType(c *C) {
	resolver := newSpecialTypeResolver()

	stepArg := resolver.resolve("unknown:foo")
	c.Assert(stepArg.value, Equals, "unknown:foo")
	c.Assert(stepArg.argType, Equals, static)
}
