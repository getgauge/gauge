package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestSimpleStepAfterStepValueExtraction(c *C) {
	stepText := "a simple step"
	stepValue, err := extractStepValueAndParams(stepText, false)

	args := stepValue.args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 0)
	c.Assert(stepValue.stepValue, Equals, "a simple step")
	c.Assert(stepValue.parameterizedStepValue, Equals, "a simple step")
}

func (s *MySuite) TestAddingTableParamAfterStepValueExtraction(c *C) {
	stepText := "a simple step"
	stepValue, err := extractStepValueAndParams(stepText, true)

	args := stepValue.args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 1)
	c.Assert(args[0], Equals, "table")
	c.Assert(stepValue.stepValue, Equals, "a simple step {}")
	c.Assert(stepValue.parameterizedStepValue, Equals, "a simple step <table>")
}

func (s *MySuite) TestAddingTableParamAfterStepValueExtractionForStepWithExistingParam(c *C) {
	stepText := "a \"param1\" step with multiple params <param2> <file:specialParam>"
	stepValue, err := extractStepValueAndParams(stepText, true)

	args := stepValue.args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 4)
	c.Assert(args[0], Equals, "param1")
	c.Assert(args[1], Equals, "param2")
	c.Assert(args[2], Equals, "file:specialParam")
	c.Assert(args[3], Equals, "table")
	c.Assert(stepValue.stepValue, Equals, "a {} step with multiple params {} {} {}")
	c.Assert(stepValue.parameterizedStepValue, Equals, "a <param1> step with multiple params <param2> <file:specialParam> <table>")
}

func (s *MySuite) TestAfterStepValueExtractionForStepWithExistingParam(c *C) {
	stepText := "a \"param1\" step with multiple params <param2> <file:specialParam>"
	stepValue, err := extractStepValueAndParams(stepText, false)

	args := stepValue.args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 3)
	c.Assert(args[0], Equals, "param1")
	c.Assert(args[1], Equals, "param2")
	c.Assert(args[2], Equals, "file:specialParam")
	c.Assert(stepValue.stepValue, Equals, "a {} step with multiple params {} {}")
	c.Assert(stepValue.parameterizedStepValue, Equals, "a <param1> step with multiple params <param2> <file:specialParam>")
}
