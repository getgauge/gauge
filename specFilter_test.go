package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestScenarioIndexFilter(c *C) {
	specText := SpecBuilder().specHeading("spec heading").
	scenarioHeading("First scenario").
	step("a step").
	scenarioHeading("Second scenario").
	step("second step").
	scenarioHeading("Third scenario").
	step("third user").String()

	spec, parseResult := new(specParser).parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilter(2))
	c.Assert(len(spec.scenarios), Equals, 2)
	c.Assert(spec.scenarios[0].heading.value, Equals, "First scenario")
	c.Assert(spec.scenarios[1].heading.value, Equals, "Second scenario")
}




