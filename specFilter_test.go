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
		step("third user").
		scenarioHeading("Fourth scenario").
		step("Fourth user").String()

	spec, parseResult := new(specParser).parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(2))
	c.Assert(len(spec.scenarios), Equals, 1)
	c.Assert(spec.scenarios[0].heading.value, Equals, "Third scenario")

}

func (s *MySuite) TestScenarioIndexFilterLastScenario(c *C) {
	specText := SpecBuilder().specHeading("spec heading").
		scenarioHeading("First scenario").
		step("a step").
		scenarioHeading("Second scenario").
		step("second step").
		scenarioHeading("Third scenario").
		step("third user").
		scenarioHeading("Fourth scenario").
		step("Fourth user").String()

	spec, parseResult := new(specParser).parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(3))
	c.Assert(len(spec.scenarios), Equals, 1)
	c.Assert(spec.scenarios[0].heading.value, Equals, "Fourth scenario")

}

func (s *MySuite) TestScenarioIndexFilterFirstScenario(c *C) {
	specText := SpecBuilder().specHeading("spec heading").
		scenarioHeading("First scenario").
		step("a step").
		scenarioHeading("Second scenario").
		step("second step").
		scenarioHeading("Third scenario").
		step("third user").
		scenarioHeading("Fourth scenario").
		step("Fourth user").String()

	spec, parseResult := new(specParser).parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(0))
	c.Assert(len(spec.scenarios), Equals, 1)
	c.Assert(spec.scenarios[0].heading.value, Equals, "First scenario")

}

func (s *MySuite) TestScenarioIndexFilterForSingleScenarioSpec(c *C) {
	specText := SpecBuilder().specHeading("spec heading").
		scenarioHeading("First scenario").
		step("a step").String()

	spec, parseResult := new(specParser).parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(0))
	c.Assert(len(spec.scenarios), Equals, 1)
	c.Assert(spec.scenarios[0].heading.value, Equals, "First scenario")
}

func (s *MySuite) TestScenarioIndexFilterWithWrongScenarioIndex(c *C) {
	specText := SpecBuilder().specHeading("spec heading").
		scenarioHeading("First scenario").
		step("a step").String()

	spec, parseResult := new(specParser).parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(1))
	c.Assert(len(spec.scenarios), Equals, 0)
}
