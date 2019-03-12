//// Copyright 2015 ThoughtWorks, Inc.
//
//// This file is part of Gauge.
//
//// Gauge is free software: you can redistribute it and/or modify
//// it under the terms of the GNU General Public License as published by
//// the Free Software Foundation, either version 3 of the License, or
//// (at your option) any later version.
//
//// Gauge is distributed in the hope that it will be useful,
//// but WITHOUT ANY WARRANTY; without even the implied warranty of
//// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//// GNU General Public License for more details.
//
//// You should have received a copy of the GNU General Public License
//// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.
//
package filter

import "fmt"

import (
	"testing"

	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
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

func (s *MySuite) TestToEvaluateTagExpressionConsistingTrueAndFalseAsTagNamesWithNegation(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "!true"}
	c.Assert(filter.filterTags(nil), Equals, true)
}

func (s *MySuite) TestToEvaluateTagExpressionConsistingSpecialCharacters(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "a && b || c | b & b"}
	c.Assert(filter.filterTags([]string{"a", "b"}), Equals, true)
}

func (s *MySuite) TestToEvaluateTagExpressionWhenTagIsSubsetOfTrueOrFalse(c *C) {
	// https://github.com/getgauge/gauge/issues/667
	filter := &ScenarioFilterBasedOnTags{tagExpression: "b || c | b & b && a"}
	c.Assert(filter.filterTags([]string{"a", "b"}), Equals, true)
}

func (s *MySuite) TestParseTagExpression(c *C) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: "b || c | b & b && a"}
	txps, tags := filter.parseTagExpression()

	expectedTxps := []string{"b", "|", "|", "c", "|", "b", "&", "b", "&", "&", "a"}
	expectedTags := []string{"b", "c", "b", "b", "a"}

	for i, t := range txps {
		c.Assert(expectedTxps[i], Equals, t)
	}
	for i, t := range tags {
		c.Assert(expectedTags[i], Equals, t)
	}
}

func (s *MySuite) TestScenarioSpanFilter(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}
	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 7, End: 10},
	}
	scenario4 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Fourth Scenario"},
		Span:    &gauge.Span{Start: 11, End: 15},
	}
	spec := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3, scenario4},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3, scenario4},
	}

	// TODO: assert for another return value as well which consists specs with non filtered items
	specWithFilteredItems, _ := spec.Filter(NewScenarioFilterBasedOnSpan([]int{8}))

	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 1)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario3)
}

func (s *MySuite) TestScenarioSpanFilterLastScenario(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}
	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 7, End: 10},
	}
	scenario4 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Fourth Scenario"},
		Span:    &gauge.Span{Start: 11, End: 15},
	}
	spec := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3, scenario4},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3, scenario4},
	}

	// TODO: assert for another return value as well which consists specs with non filtered items
	specWithFilteredItems, _ := spec.Filter(NewScenarioFilterBasedOnSpan([]int{13}))
	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 1)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario4)

}

func (s *MySuite) TestScenarioSpanFilterFirstScenario(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}
	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 7, End: 10},
	}
	scenario4 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Fourth Scenario"},
		Span:    &gauge.Span{Start: 11, End: 15},
	}
	spec := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3, scenario4},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3, scenario4},
	}

	// TODO: assert for another return value as well which consists specs with non filtered items
	specWithFilteredItems, _ := spec.Filter(NewScenarioFilterBasedOnSpan([]int{2}))

	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 1)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario1)

}

func (s *MySuite) TestScenarioSpanFilterForSingleScenarioSpec(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	spec := &gauge.Specification{
		Items:     []gauge.Item{scenario1},
		Scenarios: []*gauge.Scenario{scenario1},
	}

	// TODO: assert for another return value as well which consists specs with non filtered items
	specWithFilteredItems, _ := spec.Filter(NewScenarioFilterBasedOnSpan([]int{3}))
	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 1)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario1)
}

func (s *MySuite) TestScenarioSpanFilterWithWrongScenarioIndex(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	spec := &gauge.Specification{
		Items:     []gauge.Item{scenario1},
		Scenarios: []*gauge.Scenario{scenario1},
	}

	// TODO: assert for another return value as well which consists specs with non filtered items
	specWithFilteredItems, _ := spec.Filter(NewScenarioFilterBasedOnSpan([]int{5}))
	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 0)
}

func (s *MySuite) TestScenarioSpanFilterWithMultipleLineNumbers(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}
	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 7, End: 10},
	}
	scenario4 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Fourth Scenario"},
		Span:    &gauge.Span{Start: 11, End: 15},
	}
	spec := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3, scenario4},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3, scenario4},
	}

	// TODO: assert for another return value as well which consists specs with non filtered items
	specWithFilteredItems, _ := spec.Filter(NewScenarioFilterBasedOnSpan([]int{3, 13}))

	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 2)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario1)
	c.Assert(specWithFilteredItems.Scenarios[1], Equals, scenario4)

}

func (s *MySuite) TestToFilterSpecsByTagExpOfTwoTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}
	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
		Tags:      &gauge.Tags{RawValues: [][]string{myTags}},
	}

	spec2 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	c.Assert(specs[0].Tags.Values()[0], Equals, myTags[0])
	c.Assert(specs[0].Tags.Values()[1], Equals, myTags[1])
	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs), Equals, 1)
}

func (s *MySuite) TestToEvaluateTagExpression(c *C) {
	myTags := []string{"tag1", "tag2"}

	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
		Tags:    &gauge.Tags{RawValues: [][]string{[]string{myTags[0]}}},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{[]string{"tag3"}}},
	}

	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario4 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Fourth Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}

	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
	}

	spec2 := &gauge.Specification{
		Items:     []gauge.Item{scenario3, scenario4},
		Scenarios: []*gauge.Scenario{scenario3, scenario4},
		Tags:      &gauge.Tags{RawValues: [][]string{myTags}},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)
	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "tag1 & !(tag1 & tag4) & (tag2 | tag3)")
	c.Assert(len(specs), Equals, 1)
	c.Assert(len(specs[0].Scenarios), Equals, 2)
	c.Assert(specs[0].Scenarios[0], Equals, scenario3)
	c.Assert(specs[0].Scenarios[1], Equals, scenario4)
}

func (s *MySuite) TestToFilterSpecsByWrongTagExpression(c *C) {
	myTags := []string{"tag1", "tag2"}
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}
	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
		Tags:      &gauge.Tags{RawValues: [][]string{myTags}},
	}

	spec2 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)

	c.Assert(specs[0].Tags.Values()[0], Equals, myTags[0])
	c.Assert(specs[0].Tags.Values()[1], Equals, myTags[1])
	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "(tag1 & tag2")
	c.Assert(len(specs), Equals, 0)
}

func (s *MySuite) TestToFilterMultipleScenariosByMultipleTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
		Tags:    &gauge.Tags{RawValues: [][]string{[]string{"tag1"}}},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{myTags}},
	}

	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
		Tags:    &gauge.Tags{RawValues: [][]string{myTags}},
	}
	scenario4 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Fourth Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{[]string{"prod", "tag7", "tag1", "tag2"}}},
	}
	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3, scenario4},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3, scenario4},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)

	c.Assert(len(specs[0].Scenarios), Equals, 4)
	c.Assert(len(specs[0].Scenarios[0].Tags.Values()), Equals, 1)
	c.Assert(len(specs[0].Scenarios[1].Tags.Values()), Equals, 2)
	c.Assert(len(specs[0].Scenarios[2].Tags.Values()), Equals, 2)
	c.Assert(len(specs[0].Scenarios[3].Tags.Values()), Equals, 4)

	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(specs[0].Scenarios[0], Equals, scenario2)
	c.Assert(specs[0].Scenarios[1], Equals, scenario3)
	c.Assert(specs[0].Scenarios[2], Equals, scenario4)
}

func (s *MySuite) TestToFilterScenariosByTagsAtSpecLevel(c *C) {
	myTags := []string{"tag1", "tag2"}

	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}

	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}

	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3},
		Tags:      &gauge.Tags{RawValues: [][]string{myTags}},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)

	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(len(specs[0].Tags.Values()), Equals, 2)
	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 3)
	c.Assert(specs[0].Scenarios[0], Equals, scenario1)
	c.Assert(specs[0].Scenarios[1], Equals, scenario2)
	c.Assert(specs[0].Scenarios[2], Equals, scenario3)
}

func (s *MySuite) TestToFilterScenariosByTagExpWithDuplicateTagNames(c *C) {
	myTags := []string{"tag1", "tag12"}
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
		Tags:    &gauge.Tags{RawValues: [][]string{myTags}},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{[]string{"tag1"}}},
	}

	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{[]string{"tag12"}}},
	}

	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)
	c.Assert(len(specs), Equals, 1)

	c.Assert(len(specs[0].Scenarios), Equals, 3)
	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "tag1 & tag12")
	c.Assert(len(specs[0].Scenarios), Equals, 1)
	c.Assert(specs[0].Scenarios[0], Equals, scenario1)
}

func (s *MySuite) TestFilterTags(c *C) {
	specTags := []string{"abcd", "foo", "bar", "foo bar"}
	tagFilter := newScenarioFilterBasedOnTags(specTags, "abcd & foo bar")
	evaluateTrue := tagFilter.filterTags(specTags)
	c.Assert(evaluateTrue, Equals, true)
}

func (s *MySuite) TestSanitizeTags(c *C) {
	specTags := []string{"abcd", "foo", "bar", "foo bar"}
	tagFilter := newScenarioFilterBasedOnTags(specTags, "abcd & foo bar | true")
	evaluateTrue := tagFilter.filterTags(specTags)
	c.Assert(evaluateTrue, Equals, true)
}

func (s *MySuite) TestToFilterSpecsByTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
		Tags:    &gauge.Tags{RawValues: [][]string{myTags}},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}
	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}

	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
	}
	spec2 := &gauge.Specification{
		Items:     []gauge.Item{scenario2, scenario3},
		Scenarios: []*gauge.Scenario{scenario2, scenario3},
	}

	spec3 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario3},
		Scenarios: []*gauge.Scenario{scenario1, scenario3},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)
	specs = append(specs, spec2)
	specs = append(specs, spec3)
	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs), Equals, 2)
	c.Assert(len(specs[0].Scenarios), Equals, 1)
	c.Assert(len(specs[1].Scenarios), Equals, 1)
	c.Assert(specs[0].Scenarios[0], Equals, scenario1)
	c.Assert(specs[1].Scenarios[0], Equals, scenario1)
}

func (s *MySuite) TestToFilterScenariosByTag(c *C) {
	myTags := []string{"tag1", "tag2"}

	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{myTags}},
	}

	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}

	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)
	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(specs[0].Scenarios), Equals, 1)
	c.Assert(specs[0].Scenarios[0], Equals, scenario2)
}

func (s *MySuite) TestToFilterMultipleScenariosByTags(c *C) {
	myTags := []string{"tag1", "tag2"}

	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
		Tags:    &gauge.Tags{RawValues: [][]string{[]string{"tag1"}}},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{myTags}},
	}

	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{myTags}},
	}

	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)

	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "tag1 & tag2")

	c.Assert(len(specs[0].Scenarios), Equals, 2)
	c.Assert(specs[0].Scenarios[0], Equals, scenario2)
	c.Assert(specs[0].Scenarios[1], Equals, scenario3)
}

func (s *MySuite) TestToFilterScenariosByUnavailableTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{myTags}},
	}

	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
	}

	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)

	// TODO: assert for another return value as well which consists specs with non filtered items
	specs, _ = filterSpecsByTags(specs, "tag3")

	c.Assert(len(specs), Equals, 0)
}

func (s *MySuite) TestFilterScenariosByName(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
	}
	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
	}
	var scenarios = []string{"First Scenario"}

	var specs []*gauge.Specification
	specs = append(specs, spec1)

	c.Assert(len(specs[0].Scenarios), Equals, 2)
	specs = filterSpecsByScenarioName(specs, scenarios)
	c.Assert(len(specs[0].Scenarios), Equals, 1)
}

func (s *MySuite) TestFilterScenarioWhichDoesNotExists(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
	}
	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
	}
	var scenarios = []string{"Third Scenario"}

	var specs []*gauge.Specification
	specs = append(specs, spec1)

	c.Assert(len(specs[0].Scenarios), Equals, 2)
	specs = filterSpecsByScenarioName(specs, scenarios)
	c.Assert(len(specs), Equals, 0)
}

func (s *MySuite) TestFilterMultipleScenariosByName(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
	}
	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
	}
	var scenarios = []string{"First Scenario", "Second Scenario"}

	var specs []*gauge.Specification
	specs = append(specs, spec1)

	c.Assert(len(specs[0].Scenarios), Equals, 2)
	specs = filterSpecsByScenarioName(specs, scenarios)
	c.Assert(len(specs[0].Scenarios), Equals, 2)
}

func (s *MySuite) TestFilterInvalidScenarios(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
	}
	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2},
		Scenarios: []*gauge.Scenario{scenario1, scenario2},
	}
	var scenarios = []string{"First Scenario", "Third Scenario"}

	var specs []*gauge.Specification
	specs = append(specs, spec1)

	c.Assert(len(specs[0].Scenarios), Equals, 2)
	filteredScenarios := filterValidScenarios(specs, scenarios)
	c.Assert(len(filteredScenarios), Equals, 1)
	c.Assert(filteredScenarios[0], Equals, "First Scenario")
}
