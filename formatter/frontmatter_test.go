package formatter

import (
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestFormatSpecificationPreservesYamlFrontMatter(c *C) {
	specText := `---
id: login-success
owner: authentication-team
---

# Login

## Successful login

* Log in with valid credentials
`

	spec, result := new(parser.SpecParser).ParseSpecText(specText, "login.spec")

	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.ParseErrors), Equals, 0)
	c.Assert(spec.Heading.LineNo, Equals, 6)
	c.Assert(FormatSpecification(spec), Equals, specText)
	c.Assert(spec.Items[0].Kind(), Equals, gauge.CommentKind)
}
