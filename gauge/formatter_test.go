package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestFormatSpecification(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: "Example step", lineNo: 3},
		&token{kind: stepKind, value: "Step with inline table", lineNo: 3},
		&token{kind: tableHeader, args: []string{"id", "name"}},
		&token{kind: tableRow, args: []string{"1", "foo"}},
		&token{kind: tableRow, args: []string{"2", "bar"}},
	}

	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))

	formatted := formatSpecification(spec)

	c.Assert(formatted, Equals,
		`Spec Heading
============
Scenario Heading
----------------
* Example step
* Step with inline table 
     |id|name|
     |--|----|
     |1 |foo |
     |2 |bar |

`)
}
