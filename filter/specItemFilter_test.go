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

package filter

import (
	"fmt"
	"github.com/getgauge/gauge/parser"
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

type specBuilder struct {
	lines []string
}

func SpecBuilder() *specBuilder {
	return &specBuilder{lines: make([]string, 0)}
}

func (specBuilder *specBuilder) addPrefix(prefix string, line string) string {
	return fmt.Sprintf("%s%s\n", prefix, line)
}

func (specBuilder *specBuilder) String() string {
	var result string
	for _, line := range specBuilder.lines {
		result = fmt.Sprintf("%s%s", result, line)
	}
	return result
}

func (specBuilder *specBuilder) specHeading(heading string) *specBuilder {
	line := specBuilder.addPrefix("#", heading)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) scenarioHeading(heading string) *specBuilder {
	line := specBuilder.addPrefix("##", heading)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) step(stepText string) *specBuilder {
	line := specBuilder.addPrefix("* ", stepText)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) tags(tags ...string) *specBuilder {
	tagText := ""
	for i, tag := range tags {
		tagText = fmt.Sprintf("%s%s", tagText, tag)
		if i != len(tags)-1 {
			tagText = fmt.Sprintf("%s,", tagText)
		}
	}
	line := specBuilder.addPrefix("tags: ", tagText)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) tableHeader(cells ...string) *specBuilder {
	return specBuilder.tableRow(cells...)
}
func (specBuilder *specBuilder) tableRow(cells ...string) *specBuilder {
	rowInMarkdown := "|"
	for _, cell := range cells {
		rowInMarkdown = fmt.Sprintf("%s%s|", rowInMarkdown, cell)
	}
	specBuilder.lines = append(specBuilder.lines, fmt.Sprintf("%s\n", rowInMarkdown))
	return specBuilder
}

func (specBuilder *specBuilder) text(comment string) *specBuilder {
	specBuilder.lines = append(specBuilder.lines, fmt.Sprintf("%s\n", comment))
	return specBuilder
}

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

	spec, parseResult := new(parser.SpecParser).Parse(specText, new(parser.ConceptDictionary))
	c.Assert(parseResult.Ok, Equals, true)

	spec.Filter(newScenarioIndexFilterToRetain(2))
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.Value, Equals, "Third scenario")

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

	spec, parseResult := new(parser.SpecParser).Parse(specText, new(parser.ConceptDictionary))
	c.Assert(parseResult.Ok, Equals, true)

	spec.Filter(newScenarioIndexFilterToRetain(3))
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.Value, Equals, "Fourth scenario")

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

	spec, parseResult := new(parser.SpecParser).Parse(specText, new(parser.ConceptDictionary))
	c.Assert(parseResult.Ok, Equals, true)

	spec.Filter(newScenarioIndexFilterToRetain(0))
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.Value, Equals, "First scenario")

}

func (s *MySuite) TestScenarioIndexFilterForSingleScenarioSpec(c *C) {
	specText := SpecBuilder().specHeading("spec heading").
		scenarioHeading("First scenario").
		step("a step").String()

	spec, parseResult := new(parser.SpecParser).Parse(specText, new(parser.ConceptDictionary))
	c.Assert(parseResult.Ok, Equals, true)

	spec.Filter(newScenarioIndexFilterToRetain(0))
	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.Value, Equals, "First scenario")
}

func (s *MySuite) TestScenarioIndexFilterWithWrongScenarioIndex(c *C) {
	specText := SpecBuilder().specHeading("spec heading").
		scenarioHeading("First scenario").
		step("a step").String()

	spec, parseResult := new(parser.SpecParser).Parse(specText, new(parser.ConceptDictionary))
	c.Assert(parseResult.Ok, Equals, true)

	spec.Filter(newScenarioIndexFilterToRetain(1))
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
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading1", LineNo: 1},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 3},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 4},
	}
	spec1, result := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	tokens1 := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading2", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 3},
	}
	spec2, result := new(parser.SpecParser).CreateSpecification(tokens1, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	var specs []*parser.Specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	c.Assert(specs[0].Tags.Values[0], Equals, myTags[0])
	c.Assert(specs[0].Tags.Values[1], Equals, myTags[1])
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs), Equals, 1)
}

func (s *MySuite) TestToEvaluateTagExpression(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading1", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 01", LineNo: 2},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag1"}, LineNo: 3},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 02", LineNo: 4},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag3"}, LineNo: 5},
	}
	spec1, result := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	tokens1 := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading2", LineNo: 1},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 3},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 4},
	}
	spec2, result := new(parser.SpecParser).CreateSpecification(tokens1, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	var specs []*parser.Specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	specs = filterSpecsByTags(specs, "tag1 & !(tag1 & tag4) & (tag2 | tag3)")
	c.Assert(len(specs), Equals, 1)
	c.Assert(len(specs[0].Scenarios), Equals, 2)
	c.Assert(specs[0].Scenarios[0].Heading.Value, Equals, "Scenario Heading 1")
	c.Assert(specs[0].Scenarios[1].Heading.Value, Equals, "Scenario Heading 2")
}

func (s *MySuite) TestToFilterSpecsByWrongTagExpression(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading1", LineNo: 1},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 3},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 4},
	}
	spec1, result := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	tokens1 := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading2", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 3},
	}
	spec2, result := new(parser.SpecParser).CreateSpecification(tokens1, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	var specs []*parser.Specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	c.Assert(specs[0].Tags.Values[0], Equals, myTags[0])
	c.Assert(specs[0].Tags.Values[1], Equals, myTags[1])
	specs = filterSpecsByTags(specs, "(tag1 & tag2")
	c.Assert(len(specs), Equals, 0)
}

func (s *MySuite) TestToFilterMultipleScenariosByMultipleTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag1"}, LineNo: 3},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 4},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 5},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 3", LineNo: 6},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 7},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 4", LineNo: 8},
		&parser.Token{Kind: parser.TagKind, Args: []string{"prod", "tag7", "tag1", "tag2"}, LineNo: 9},
	}
	spec, result := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	var specs []*parser.Specification
	specs = append(specs, spec)

	c.Assert(len(specs[0].Scenarios), Equals, 4)
	c.Assert(len(specs[0].Scenarios[0].Tags.Values), Equals, 1)
	c.Assert(len(specs[0].Scenarios[1].Tags.Values), Equals, 2)
	c.Assert(len(specs[0].Scenarios[2].Tags.Values), Equals, 2)
	c.Assert(len(specs[0].Scenarios[3].Tags.Values), Equals, 4)

	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(specs[0].Scenarios[0].Heading.Value, Equals, "Scenario Heading 2")
	c.Assert(specs[0].Scenarios[1].Heading.Value, Equals, "Scenario Heading 3")
	c.Assert(specs[0].Scenarios[2].Heading.Value, Equals, "Scenario Heading 4")
}

func (s *MySuite) TestToFilterScenariosByTagsAtSpecLevel(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 3},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 4},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 3", LineNo: 5},
	}
	spec, result := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	var specs []*parser.Specification
	specs = append(specs, spec)

	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(len(specs[0].Tags.Values), Equals, 2)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(specs[0].Scenarios[0].Heading.Value, Equals, "Scenario Heading 1")
	c.Assert(specs[0].Scenarios[1].Heading.Value, Equals, "Scenario Heading 2")
	c.Assert(specs[0].Scenarios[2].Heading.Value, Equals, "Scenario Heading 3")
}

func (s *MySuite) TestToFilterSpecsByTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading1", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 1},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 3},
	}
	spec1, result := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	tokens1 := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading2", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 2},
	}
	spec2, result := new(parser.SpecParser).CreateSpecification(tokens1, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	tokens2 := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading3", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 1},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 3},
	}
	spec3, result := new(parser.SpecParser).CreateSpecification(tokens2, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec1.Scenarios), Equals, 2)
	c.Assert(len(spec1.Scenarios[0].Tags.Values), Equals, 2)
	c.Assert(len(spec2.Scenarios), Equals, 2)

	var specs []*parser.Specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)
	specs = append(specs, spec3)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs), Equals, 2)
	c.Assert(len(specs[0].Scenarios), Equals, 1)
	c.Assert(len(specs[1].Scenarios), Equals, 1)
	c.Assert(specs[0].Heading.Value, Equals, "Spec Heading1")
	c.Assert(specs[1].Heading.Value, Equals, "Spec Heading3")
}

func (s *MySuite) TestToFilterScenariosByTag(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 4},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 3},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 3", LineNo: 5},
	}
	spec, result := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Scenarios), Equals, 3)
	c.Assert(len(spec.Scenarios[1].Tags.Values), Equals, 2)

	var specs []*parser.Specification
	specs = append(specs, spec)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 1)
	c.Assert(specs[0].Scenarios[0].Heading.Value, Equals, "Scenario Heading 2")
}

func (s *MySuite) TestToFilterMultipleScenariosByTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag1"}, LineNo: 3},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 4},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 5},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 3", LineNo: 6},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 7},
	}
	spec, result := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	var specs []*parser.Specification
	specs = append(specs, spec)
	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(len(specs[0].Scenarios[0].Tags.Values), Equals, 1)
	c.Assert(len(specs[0].Scenarios[1].Tags.Values), Equals, 2)
	specs = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 2)
	c.Assert(specs[0].Scenarios[0].Heading.Value, Equals, "Scenario Heading 2")
	c.Assert(specs[0].Scenarios[1].Heading.Value, Equals, "Scenario Heading 3")
}

func (s *MySuite) TestToFilterScenariosByUnavailableTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 1", LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 2", LineNo: 4},
		&parser.Token{Kind: parser.TagKind, Args: myTags, LineNo: 3},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading 3", LineNo: 5},
	}
	spec, result := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Scenarios), Equals, 3)
	c.Assert(len(spec.Scenarios[1].Tags.Values), Equals, 2)

	var specs []*parser.Specification
	specs = append(specs, spec)
	specs = filterSpecsByTags(specs, "tag3")
	c.Assert(len(specs), Equals, 0)
}
