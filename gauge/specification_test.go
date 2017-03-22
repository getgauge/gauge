package gauge

import . "gopkg.in/check.v1"

func (s *MySuite) TestUsesArgsInContext(c *C) {
	spec := &Specification{
		Contexts: []*Step{
			&Step{Value: "some ",
				LineText:  "sfd <foo>",
				IsConcept: false,
				Args: []*StepArg{
					&StepArg{
						Value:   "foo",
						ArgType: Dynamic,
					},
				},
			},
		},
		TearDownSteps: []*Step{},
	}

	c.Assert(spec.UsesArgsInContextTeardown("foo"), Equals, true)
}

func (s *MySuite) TestDoesNotUseDynamicArgsInContext(c *C) {
	spec := &Specification{
		Contexts: []*Step{
			&Step{Value: "some ",
				LineText:  "sfd <foo>",
				IsConcept: false,
				Args: []*StepArg{
					&StepArg{
						Value:   "foo",
						ArgType: Static,
					},
				},
			},
		},
		TearDownSteps: []*Step{},
	}

	c.Assert(spec.UsesArgsInContextTeardown("foo"), Equals, false)
}

func (s *MySuite) TestDoesNotUseArgsInTeardown(c *C) {
	spec := &Specification{
		Contexts: []*Step{},
		TearDownSteps: []*Step{
			&Step{Value: "some ",
				LineText:  "sfd <foo>",
				IsConcept: false,
				Args: []*StepArg{
					&StepArg{
						Value:   "foo",
						ArgType: Static,
					},
				},
			},
		},
	}

	c.Assert(spec.UsesArgsInContextTeardown("foo"), Equals, false)
}
