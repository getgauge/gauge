/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
 package gauge

import . "gopkg.in/check.v1"

func (s *MySuite) TestUsesArgsInStep(c *C) {
	stepArg := &StepArg{
		Value:   "foo",
		ArgType: Dynamic,
	}
	step1 := &Step{
		Value:    "Some Step",
		LineText: "abc <foo>",
		Args:     []*StepArg{stepArg},
	}
	step2 := &Step{
		Value:    "Some Step",
		LineText: "abc",
		Args:     []*StepArg{},
	}
	scenario := &Scenario{
		Steps: []*Step{step1, step2},
	}

	c.Assert(scenario.UsesArgsInSteps("foo"), Equals, true)
}

func (s *MySuite) TestDoesNotUseDynamicArgsInStep(c *C) {
	stepArg := &StepArg{
		Value:   "foo",
		ArgType: Static,
	}

	step1 := &Step{
		Value:    "Some Step",
		LineText: "abc <foo>",
		Args:     []*StepArg{stepArg},
	}
	step2 := &Step{
		Value:    "Some Step",
		LineText: "abc",
		Args:     []*StepArg{},
	}
	scenario := &Scenario{
		Steps: []*Step{step1, step2},
	}

	c.Assert(scenario.UsesArgsInSteps("foo"), Equals, false)
}

func (s *MySuite) TestDoesNotUseArgsInStep(c *C) {
	stepArg := &StepArg{
		Value:   "abc",
		ArgType: Dynamic,
	}
	step1 := &Step{
		Value:    "Some Step",
		LineText: "abc <foo>",
		Args:     []*StepArg{stepArg},
	}
	step2 := &Step{
		Value:    "Some Step",
		LineText: "abc",
		Args:     []*StepArg{},
	}
	scenario := &Scenario{
		Steps: []*Step{step1, step2},
	}

	c.Assert(scenario.UsesArgsInSteps("foo"), Equals, false)
}
