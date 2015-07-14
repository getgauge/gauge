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

	spec, parseResult := new(specParser).Parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(2))
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.value, Equals, "Third scenario")

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

	spec, parseResult := new(specParser).Parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(3))
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.value, Equals, "Fourth scenario")

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

	spec, parseResult := new(specParser).Parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(0))
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.value, Equals, "First scenario")

}

func (s *MySuite) TestScenarioIndexFilterForSingleScenarioSpec(c *C) {
	specText := SpecBuilder().specHeading("spec heading").
		scenarioHeading("First scenario").
		step("a step").String()

	spec, parseResult := new(specParser).Parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(0))
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.value, Equals, "First scenario")
}

func (s *MySuite) TestScenarioIndexFilterWithWrongScenarioIndex(c *C) {
	specText := SpecBuilder().specHeading("spec heading").
		scenarioHeading("First scenario").
		step("a step").String()

	spec, parseResult := new(specParser).Parse(specText, new(conceptDictionary))
	c.Assert(parseResult.ok, Equals, true)

	spec.filter(newScenarioIndexFilterToRetain(1))
	c.Assert(len(spec.Scenarios), Equals, 0)
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
func (s *MySuite) TestToEvaluateTagExpressionConsistingTrueAndFalseAsTagSDFNames(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "!true"}
	c.Assert(filter.filterTags(nil), Equals, true)
}
func (s *MySuite) TestToEvaluateTagExpressionConsistingSpecialCharacters(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "a && b || c | b & b"}
	c.Assert(filter.filterTags([]string{"a", "b"}), Equals, true)
}

func (s *MySuite) TestToFilterSpecsByTagExpOfTwoTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading1", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
	}
	spec1, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens1 := []*token{
		&token{kind: specKind, value: "Spec Heading2", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 3},
	}
	spec2, result := new(specParser).createSpecification(tokens1, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	c.Assert(specs[0].Tags.values[0], Equals, myTags[0])
	c.Assert(specs[0].Tags.values[1], Equals, myTags[1])
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs), Equals, 1)
}

func (s *MySuite) TestToEvaluateTagExpression(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading1", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 01", lineNo: 2},
		&token{kind: tagKind, args: []string{"tag1"}, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 02", lineNo: 4},
		&token{kind: tagKind, args: []string{"tag3"}, lineNo: 5},
	}
	spec1, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens1 := []*token{
		&token{kind: specKind, value: "Spec Heading2", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
	}
	spec2, result := new(specParser).createSpecification(tokens1, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	specs = filterSpecsByTags(specs, "tag1 & !(tag1 & tag4) & (tag2 | tag3)")
	c.Assert(len(specs), Equals, 1)
	c.Assert(len(specs[0].Scenarios), Equals, 2)
	c.Assert(specs[0].Scenarios[0].Heading.value, Equals, "Scenario Heading 1")
	c.Assert(specs[0].Scenarios[1].Heading.value, Equals, "Scenario Heading 2")
}

func (s *MySuite) TestToFilterSpecsByWrongTagExpression(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading1", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
	}
	spec1, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens1 := []*token{
		&token{kind: specKind, value: "Spec Heading2", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 3},
	}
	spec2, result := new(specParser).createSpecification(tokens1, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	c.Assert(specs[0].Tags.values[0], Equals, myTags[0])
	c.Assert(specs[0].Tags.values[1], Equals, myTags[1])
	specs = filterSpecsByTags(specs, "(tag1 & tag2")
	c.Assert(len(specs), Equals, 0)
}

func (s *MySuite) TestToFilterMultipleScenariosByMultipleTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: tagKind, args: []string{"tag1"}, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: tagKind, args: myTags, lineNo: 5},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 6},
		&token{kind: tagKind, args: myTags, lineNo: 7},
		&token{kind: scenarioKind, value: "Scenario Heading 4", lineNo: 8},
		&token{kind: tagKind, args: []string{"prod", "tag7", "tag1", "tag2"}, lineNo: 9},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec)

	c.Assert(len(specs[0].Scenarios), Equals, 4)
	c.Assert(len(specs[0].Scenarios[0].Tags.values), Equals, 1)
	c.Assert(len(specs[0].Scenarios[1].Tags.values), Equals, 2)
	c.Assert(len(specs[0].Scenarios[2].Tags.values), Equals, 2)
	c.Assert(len(specs[0].Scenarios[3].Tags.values), Equals, 4)

	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(specs[0].Scenarios[0].Heading.value, Equals, "Scenario Heading 2")
	c.Assert(specs[0].Scenarios[1].Heading.value, Equals, "Scenario Heading 3")
	c.Assert(specs[0].Scenarios[2].Heading.value, Equals, "Scenario Heading 4")
}

func (s *MySuite) TestToFilterScenariosByTagsAtSpecLevel(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 5},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec)

	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(len(specs[0].Tags.values), Equals, 2)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(specs[0].Scenarios[0].Heading.value, Equals, "Scenario Heading 1")
	c.Assert(specs[0].Scenarios[1].Heading.value, Equals, "Scenario Heading 2")
	c.Assert(specs[0].Scenarios[2].Heading.value, Equals, "Scenario Heading 3")
}

func (s *MySuite) TestToFilterSpecsByTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading1", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 3},
	}
	spec1, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens1 := []*token{
		&token{kind: specKind, value: "Spec Heading2", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 2},
	}
	spec2, result := new(specParser).createSpecification(tokens1, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	tokens2 := []*token{
		&token{kind: specKind, value: "Spec Heading3", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 1},
		&token{kind: tagKind, args: myTags, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 3},
	}
	spec3, result := new(specParser).createSpecification(tokens2, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec1.Scenarios), Equals, 2)
	c.Assert(len(spec1.Scenarios[0].Tags.values), Equals, 2)
	c.Assert(len(spec2.Scenarios), Equals, 2)

	var specs []*specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)
	specs = append(specs, spec3)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs), Equals, 2)
	c.Assert(len(specs[0].Scenarios), Equals, 1)
	c.Assert(len(specs[1].Scenarios), Equals, 1)
	c.Assert(specs[0].Heading.value, Equals, "Spec Heading1")
	c.Assert(specs[1].Heading.value, Equals, "Spec Heading3")
}

func (s *MySuite) TestToFilterScenariosByTag(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: tagKind, args: myTags, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 5},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec.Scenarios), Equals, 3)
	c.Assert(len(spec.Scenarios[1].Tags.values), Equals, 2)

	var specs []*specification
	specs = append(specs, spec)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 1)
	c.Assert(specs[0].Scenarios[0].Heading.value, Equals, "Scenario Heading 2")
}

func (s *MySuite) TestToFilterMultipleScenariosByTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: tagKind, args: []string{"tag1"}, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: tagKind, args: myTags, lineNo: 5},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 6},
		&token{kind: tagKind, args: myTags, lineNo: 7},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	var specs []*specification
	specs = append(specs, spec)
	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(len(specs[0].Scenarios[0].Tags.values), Equals, 1)
	c.Assert(len(specs[0].Scenarios[1].Tags.values), Equals, 2)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 2)
	c.Assert(specs[0].Scenarios[0].Heading.value, Equals, "Scenario Heading 2")
	c.Assert(specs[0].Scenarios[1].Heading.value, Equals, "Scenario Heading 3")
}

func (s *MySuite) TestToFilterScenariosByUnavailableTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading 1", lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading 2", lineNo: 4},
		&token{kind: tagKind, args: myTags, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading 3", lineNo: 5},
	}
	spec, result := new(specParser).createSpecification(tokens, new(conceptDictionary))
	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec.Scenarios), Equals, 3)
	c.Assert(len(spec.Scenarios[1].Tags.values), Equals, 2)

	var specs []*specification
	specs = append(specs, spec)
	specs = filterSpecsByTags(specs, "tag3")
	c.Assert(len(specs), Equals, 0)
}
