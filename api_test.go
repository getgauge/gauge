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

package main

import (
	. "gopkg.in/check.v1"
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
	c.Assert(args[0], Equals, string(tableArg))
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

func (s *MySuite) TestCreateStepValueFromStep(c *C) {
	step := &step{value: "simple step with {} and {}", args: []*stepArg{staticArg("hello"), dynamicArg("desc")}}
	stepValue := createStepValue(step)

	args := stepValue.args
	c.Assert(len(args), Equals, 2)
	c.Assert(args[0], Equals, "hello")
	c.Assert(args[1], Equals, "desc")
	c.Assert(stepValue.stepValue, Equals, "simple step with {} and {}")
	c.Assert(stepValue.parameterizedStepValue, Equals, "simple step with <hello> and <desc>")
}

func (s *MySuite) TestCreateStepValueFromStepWithSpecialParams(c *C) {
	step := &step{value: "a step with {}, {} and {}", args: []*stepArg{specialTableArg("hello"), specialStringArg("file:user.txt"), tableArgument()}}
	stepValue := createStepValue(step)

	args := stepValue.args
	c.Assert(len(args), Equals, 3)
	c.Assert(args[0], Equals, "hello")
	c.Assert(args[1], Equals, "file:user.txt")
	c.Assert(args[2], Equals, "table")
	c.Assert(stepValue.stepValue, Equals, "a step with {}, {} and {}")
	c.Assert(stepValue.parameterizedStepValue, Equals, "a step with <hello>, <file:user.txt> and <table>")
}

func staticArg(val string) *stepArg {
	return &stepArg{argType: static, value: val}
}

func dynamicArg(val string) *stepArg {
	return &stepArg{argType: dynamic, value: val}
}

func tableArgument() *stepArg {
	return &stepArg{argType: tableArg}
}

func specialTableArg(val string) *stepArg {
	return &stepArg{argType: specialTable, name: val}
}

func specialStringArg(val string) *stepArg {
	return &stepArg{argType: specialString, name: val}
}
