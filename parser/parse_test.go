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
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestSimpleStepAfterStepValueExtraction(c *C) {
	stepText := "a simple step"
	stepValue, err := ExtractStepValueAndParams(stepText, false)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 0)
	c.Assert(stepValue.StepValue, Equals, "a simple step")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a simple step")
}

func (s *MySuite) TestStepWithColonAfterStepValueExtraction(c *C) {
	stepText := "a : simple step \"hello\""
	stepValue, err := ExtractStepValueAndParams(stepText, false)
	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 1)
	c.Assert(stepValue.StepValue, Equals, "a : simple step {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a : simple step <hello>")
}

func (s *MySuite) TestSimpleStepAfterStepValueExtractionForStepWithAParam(c *C) {
	stepText := "Comment <a>"
	stepValue, err := ExtractStepValueAndParams(stepText, false)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 1)
	c.Assert(stepValue.StepValue, Equals, "Comment {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "Comment <a>")
}

func (s *MySuite) TestAddingTableParamAfterStepValueExtraction(c *C) {
	stepText := "a simple step"
	stepValue, err := ExtractStepValueAndParams(stepText, true)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 1)
	c.Assert(args[0], Equals, string(gauge.TableArg))
	c.Assert(stepValue.StepValue, Equals, "a simple step {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a simple step <table>")
}

func (s *MySuite) TestAddingTableParamAfterStepValueExtractionForStepWithExistingParam(c *C) {
	stepText := "a \"param1\" step with multiple params <param2> <file:specialParam>"
	stepValue, err := ExtractStepValueAndParams(stepText, true)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 4)
	c.Assert(args[0], Equals, "param1")
	c.Assert(args[1], Equals, "param2")
	c.Assert(args[2], Equals, "file:specialParam")
	c.Assert(args[3], Equals, "table")
	c.Assert(stepValue.StepValue, Equals, "a {} step with multiple params {} {} {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a <param1> step with multiple params <param2> <file:specialParam> <table>")
}

func (s *MySuite) TestAfterStepValueExtractionForStepWithExistingParam(c *C) {
	stepText := "a \"param1\" step with multiple params <param2> <file:specialParam>"
	stepValue, err := ExtractStepValueAndParams(stepText, false)

	args := stepValue.Args
	c.Assert(err, Equals, nil)
	c.Assert(len(args), Equals, 3)
	c.Assert(args[0], Equals, "param1")
	c.Assert(args[1], Equals, "param2")
	c.Assert(args[2], Equals, "file:specialParam")
	c.Assert(stepValue.StepValue, Equals, "a {} step with multiple params {} {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a <param1> step with multiple params <param2> <file:specialParam>")
}

func (s *MySuite) TestCreateStepValueFromStep(c *C) {
	step := &gauge.Step{Value: "simple step with {} and {}", Args: []*gauge.StepArg{staticArg("hello"), dynamicArg("desc")}}
	stepValue := CreateStepValue(step)

	args := stepValue.Args
	c.Assert(len(args), Equals, 2)
	c.Assert(args[0], Equals, "hello")
	c.Assert(args[1], Equals, "desc")
	c.Assert(stepValue.StepValue, Equals, "simple step with {} and {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "simple step with <hello> and <desc>")
}

func (s *MySuite) TestCreateStepValueFromStepWithSpecialParams(c *C) {
	step := &gauge.Step{Value: "a step with {}, {} and {}", Args: []*gauge.StepArg{specialTableArg("hello"), specialStringArg("file:user.txt"), tableArgument()}}
	stepValue := CreateStepValue(step)

	args := stepValue.Args
	c.Assert(len(args), Equals, 3)
	c.Assert(args[0], Equals, "hello")
	c.Assert(args[1], Equals, "file:user.txt")
	c.Assert(args[2], Equals, "table")
	c.Assert(stepValue.StepValue, Equals, "a step with {}, {} and {}")
	c.Assert(stepValue.ParameterizedStepValue, Equals, "a step with <hello>, <file:user.txt> and <table>")
}

func staticArg(val string) *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.Static, Value: val}
}

func dynamicArg(val string) *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.Dynamic, Value: val}
}

func tableArgument() *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.TableArg}
}

func specialTableArg(val string) *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.SpecialTable, Name: val}
}

func specialStringArg(val string) *gauge.StepArg {
	return &gauge.StepArg{ArgType: gauge.SpecialString, Name: val}
}
