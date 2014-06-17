package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestExtractingStepValueAndParamForSimpleStep(c *C) {
	stepText := "My step foo"
	stepValue, params := extractStepValueAndParams(stepText)
	c.Assert(stepValue, Equals, "My step foo")
	c.Assert(len(params), Equals, 0)
}

func (s *MySuite) TestExtractingStepValueAndParamsInQuotes(c *C) {
	stepText := "My step foo with \"param1\""
	stepValue, params := extractStepValueAndParams(stepText)
	c.Assert(stepValue, Equals, "My step foo with {}")
	c.Assert(len(params), Equals, 1)
	c.Assert(params[0], Equals, "param1")
}

func (s *MySuite) TestExtractingStepValueAndParamsInAngularBrackets(c *C) {
	stepText := "My step foo with <param1> and <param2>"
	stepValue, params := extractStepValueAndParams(stepText)
	c.Assert(stepValue, Equals, "My step foo with {} and {}")
	c.Assert(len(params), Equals, 2)
	c.Assert(params[0], Equals, "param1")
	c.Assert(params[1], Equals, "param2")
}

func (s *MySuite) TestExtractingStepValueAndParamsWithEscapedQuotes(c *C) {
	stepText := "My step foo with <\"> and <\">"
	stepValue, params := extractStepValueAndParams(stepText)
	c.Assert(stepValue, Equals, "My step foo with {} and {}")
	c.Assert(len(params), Equals, 2)
	c.Assert(params[0], Equals, "\"")
	c.Assert(params[1], Equals, "\"")
}

func (s *MySuite) TestExtractingStepValueAndQuotedParamsWithEscapedQuotes(c *C) {
	stepText := "step and \"param with \\\"\" and <another>"
	stepValue, params := extractStepValueAndParams(stepText)
	c.Assert(stepValue, Equals, "step and {} and {}")
	c.Assert(len(params), Equals, 2)
	c.Assert(params[0], Equals, "param with \"")
	c.Assert(params[1], Equals, "another")
}

func (s *MySuite) TestExtractingStepValueForStepWithUnicodeCharacters(c *C) {
	stepText := "ÀÃÙĝ <汉语漢語> and \"पेरामीटर\""
	stepValue, params := extractStepValueAndParams(stepText)
	c.Assert(stepValue, Equals, "ÀÃÙĝ {} and {}")
	c.Assert(len(params), Equals, 2)
	c.Assert(params[0], Equals, "汉语漢語")
	c.Assert(params[1], Equals, "पेरामीटर")
}
