package parser

import (
	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestYamlFrontMatterIsPreservedAsComments(c *C) {
	specText := `---
id: login-success
owner: authentication-team
tags: smoke
# not a spec
* not a step
---
# Login
## Successful login
* Log in with valid credentials
`

	parser := new(SpecParser)
	tokens, errs := parser.GenerateTokens(specText, "login.spec")

	c.Assert(len(errs), Equals, 0)
	c.Assert(len(tokens), Equals, 10)
	for i := 0; i < 7; i++ {
		c.Assert(tokens[i].Kind, Equals, gauge.CommentKind)
		c.Assert(tokens[i].LineNo, Equals, i+1)
	}
	c.Assert(tokens[7].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[7].LineNo, Equals, 8)
	c.Assert(tokens[8].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[8].LineNo, Equals, 9)
	c.Assert(tokens[9].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[9].LineNo, Equals, 10)
}

func (s *MySuite) TestYamlFrontMatterParsesWithoutChangingSourceLineNumbers(c *C) {
	specText := `---
id: login-success
owner: authentication-team
---

# Login

## Successful login

* Log in with valid credentials
`

	spec, result, err := new(SpecParser).Parse(specText, gauge.NewConceptDictionary(), "login.spec")

	c.Assert(err, IsNil)
	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.ParseErrors), Equals, 0)
	c.Assert(spec.Heading.LineNo, Equals, 6)
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.LineNo, Equals, 8)
	c.Assert(len(spec.Scenarios[0].Steps), Equals, 1)
	c.Assert(spec.Scenarios[0].Steps[0].LineNo, Equals, 10)
}

func (s *MySuite) TestScenarioUnderlineStillWorksOutsideFrontMatter(c *C) {
	specText := `# Login
Successful login
----------------
* Log in with valid credentials
`

	spec, result, err := new(SpecParser).Parse(specText, gauge.NewConceptDictionary(), "login.spec")

	c.Assert(err, IsNil)
	c.Assert(result.Ok, Equals, true)
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.Value, Equals, "Successful login")
	c.Assert(spec.Scenarios[0].Heading.LineNo, Equals, 2)
	c.Assert(spec.Scenarios[0].Heading.SpanEnd, Equals, 3)
}
