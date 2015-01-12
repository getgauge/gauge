package main

import . "gopkg.in/check.v1"

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

func (s *MySuite) TestToEvaluateTagExpressionWithTwoTags(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "tag1 & tag3"}
	c.Assert(filter.filterTags([]string{"tag1", "tag2"}), Equals, false)
}

func (s *MySuite) TestToEvaluateTagExpressionWithComplexTagExpression(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "tag1 & ((tag3 | tag2) & (tag5 | tag4 | tag3) & tag7) | tag6"}
	c.Assert(filter.filterTags([]string{"tag1", "tag2", "tag7", "tag4"}), Equals, true)
}

func (s *MySuite) TestToEvaluateTagExpressionWithFailingTagExpression(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "tag1 & ((tag3 | tag2) & (tag5 | tag4 | tag3) & tag7) & tag6"}
	c.Assert(filter.filterTags([]string{"tag1", "tag2", "tag7", "tag4"}), Equals, false)
}

func (s *MySuite) TestToEvaluateTagExpressionWithWrongTagExpression(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "tag1 & ((((tag3 | tag2) & (tag5 | tag4 | tag3) & tag7) & tag6"}
	c.Assert(filter.filterTags([]string{"tag1", "tag2", "tag7", "tag4"}), Equals, false)
}

func (s *MySuite) TestToEvaluateTagExpressionConsistingOfSpaces(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "tag 1 & tag3"}
	c.Assert(filter.filterTags([]string{"tag 1", "tag3"}), Equals, true)
}

func (s *MySuite) TestToEvaluateTagExpressionConsistingLogicalNotOperator(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "!tag 1 & tag3"}
	c.Assert(filter.filterTags([]string{"tag2", "tag3"}), Equals, true)
}

func (s *MySuite) TestToEvaluateTagExpressionConsistingManyLogicalNotOperator(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "!(!(tag 1 | !(tag6 | !(tag5))) & tag2)"}
	value := filter.filterTags([]string{"tag2", "tag4"})
	c.Assert(value, Equals, true)
}

func (s *MySuite) TestToEvaluateTagExpressionConsistingParallelLogicalNotOperator(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "!(tag1) & ! (tag3 & ! (tag3))"}
	value := filter.filterTags([]string{"tag2", "tag4"})
	c.Assert(value, Equals, true)
}

func (s *MySuite) TestToEvaluateTagExpressionConsistingComma(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "tag 1 , tag3"}
	c.Assert(filter.filterTags([]string{"tag2", "tag3"}), Equals, false)
}

func (s *MySuite) TestToEvaluateTagExpressionConsistingCommaGivesTrue(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "tag 1 , tag3"}
	c.Assert(filter.filterTags([]string{"tag1", "tag3"}), Equals, true)
}
func (s *MySuite) TestToEvaluateTagExpressionConsistingTrueAndFalseAsTagNames(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "true , false"}
	c.Assert(filter.filterTags([]string{"true", "false"}), Equals, true)
}
