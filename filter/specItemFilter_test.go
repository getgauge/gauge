/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package filter

import (
	"os"
	"testing"

	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func before() {
	_ = os.Setenv("allow_case_sensitive_tags", "false")
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

	specWithFilteredItems, specWithOtherItems := spec.Filter(NewScenarioFilterBasedOnSpan([]int{8}))

	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 1)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario3)

	c.Assert(len(specWithOtherItems.Scenarios), Equals, 3)
	c.Assert(specWithOtherItems.Scenarios[0], Equals, scenario1)
	c.Assert(specWithOtherItems.Scenarios[1], Equals, scenario2)
	c.Assert(specWithOtherItems.Scenarios[2], Equals, scenario4)
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

	specWithFilteredItems, specWithOtherItems := spec.Filter(NewScenarioFilterBasedOnSpan([]int{13}))
	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 1)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario4)

	c.Assert(len(specWithOtherItems.Scenarios), Equals, 3)
	c.Assert(specWithOtherItems.Scenarios[0], Equals, scenario1)
	c.Assert(specWithOtherItems.Scenarios[1], Equals, scenario2)
	c.Assert(specWithOtherItems.Scenarios[2], Equals, scenario3)

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

	specWithFilteredItems, specWithOtherItems := spec.Filter(NewScenarioFilterBasedOnSpan([]int{2}))

	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 1)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario1)

	c.Assert(len(specWithOtherItems.Scenarios), Equals, 3)
	c.Assert(specWithOtherItems.Scenarios[0], Equals, scenario2)
	c.Assert(specWithOtherItems.Scenarios[1], Equals, scenario3)
	c.Assert(specWithOtherItems.Scenarios[2], Equals, scenario4)

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

	specWithFilteredItems, specWithOtherItems := spec.Filter(NewScenarioFilterBasedOnSpan([]int{3}))
	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 1)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario1)

	c.Assert(len(specWithOtherItems.Scenarios), Equals, 0)
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

	specWithFilteredItems, specWithOtherItems := spec.Filter(NewScenarioFilterBasedOnSpan([]int{5}))
	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 0)

	c.Assert(len(specWithOtherItems.Scenarios), Equals, 1)
	c.Assert(specWithOtherItems.Scenarios[0], Equals, scenario1)
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

	specWithFilteredItems, specWithOtherItems := spec.Filter(NewScenarioFilterBasedOnSpan([]int{3, 13}))

	c.Assert(len(specWithFilteredItems.Scenarios), Equals, 2)
	c.Assert(specWithFilteredItems.Scenarios[0], Equals, scenario1)
	c.Assert(specWithFilteredItems.Scenarios[1], Equals, scenario4)

	c.Assert(len(specWithOtherItems.Scenarios), Equals, 2)
	c.Assert(specWithOtherItems.Scenarios[0], Equals, scenario2)
	c.Assert(specWithOtherItems.Scenarios[1], Equals, scenario3)

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
	specs = append(specs, spec1, spec2)

	c.Assert(specs[0].Tags.Values()[0], Equals, myTags[0])
	c.Assert(specs[0].Tags.Values()[1], Equals, myTags[1])

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(filteredSpecs), Equals, 1)

	c.Assert(len(otherSpecs), Equals, 1)
}

func (s *MySuite) TestToEvaluateTagExpression(c *C) {
	myTags := []string{"tag1", "tag2"}

	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
		Tags:    &gauge.Tags{RawValues: [][]string{{myTags[0]}}},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{{"tag3"}}},
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
	specs = append(specs, spec1, spec2)

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag1 & !(tag1 & tag4) & (tag2 | tag3)")
	c.Assert(len(filteredSpecs), Equals, 1)
	c.Assert(len(filteredSpecs[0].Scenarios), Equals, 2)
	c.Assert(filteredSpecs[0].Scenarios[0], Equals, scenario3)
	c.Assert(filteredSpecs[0].Scenarios[1], Equals, scenario4)

	c.Assert(len(otherSpecs), Equals, 1)
	c.Assert(len(otherSpecs[0].Scenarios), Equals, 2)
	c.Assert(otherSpecs[0].Scenarios[0], Equals, scenario1)
	c.Assert(otherSpecs[0].Scenarios[1], Equals, scenario2)
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
	specs = append(specs, spec1, spec2)

	c.Assert(specs[0].Tags.Values()[0], Equals, myTags[0])
	c.Assert(specs[0].Tags.Values()[1], Equals, myTags[1])

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "(tag1 & tag2")
	c.Assert(len(filteredSpecs), Equals, 0)

	c.Assert(len(otherSpecs), Equals, 2)

}

func (s *MySuite) TestToFilterMultipleScenariosByMultipleTags(c *C) {
	myTags := []string{"tag1", "tag2"}
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
		Tags:    &gauge.Tags{RawValues: [][]string{{"tag1"}}},
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
		Tags:    &gauge.Tags{RawValues: [][]string{{"prod", "tag7", "tag1", "tag2"}}},
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

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(filteredSpecs[0].Scenarios), Equals, 3)
	c.Assert(filteredSpecs[0].Scenarios[0], Equals, scenario2)
	c.Assert(filteredSpecs[0].Scenarios[1], Equals, scenario3)
	c.Assert(filteredSpecs[0].Scenarios[2], Equals, scenario4)

	c.Assert(len(otherSpecs[0].Scenarios), Equals, 1)
	c.Assert(otherSpecs[0].Scenarios[0], Equals, scenario1)

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

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(filteredSpecs[0].Scenarios), Equals, 3)
	c.Assert(filteredSpecs[0].Scenarios[0], Equals, scenario1)
	c.Assert(filteredSpecs[0].Scenarios[1], Equals, scenario2)
	c.Assert(filteredSpecs[0].Scenarios[2], Equals, scenario3)

	c.Assert(len(otherSpecs), Equals, 0)

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
		Tags:    &gauge.Tags{RawValues: [][]string{{"tag1"}}},
	}

	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{{"tag12"}}},
	}

	spec1 := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3},
	}

	var specs []*gauge.Specification
	specs = append(specs, spec1)
	c.Assert(len(specs), Equals, 1)

	c.Assert(len(specs[0].Scenarios), Equals, 3)

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag1 & tag12")
	c.Assert(len(filteredSpecs[0].Scenarios), Equals, 1)
	c.Assert(filteredSpecs[0].Scenarios[0], Equals, scenario1)

	c.Assert(len(otherSpecs), Equals, 1)
	c.Assert(otherSpecs[0].Scenarios[0], Equals, scenario2)
	c.Assert(otherSpecs[0].Scenarios[1], Equals, scenario3)
}

func (s *MySuite) TestFilterTags(c *C) {
	specTags := []string{"abcd", "foo", "bar", "foo bar"}
	tagFilter := NewScenarioFilterBasedOnTags(specTags, "abcd & foo bar")
	evaluateTrue := tagFilter.filterTags(specTags)
	c.Assert(evaluateTrue, Equals, true)
}

func (s *MySuite) TestFilterMixedCaseTags(c *C) {
	before()
	specTags := []string{"abcd", "foo", "BAR", "foo bar"}
	tagFilter := NewScenarioFilterBasedOnTags(specTags, "abcd & FOO bar")
	evaluateTrue := tagFilter.filterTags(specTags)
	c.Assert(evaluateTrue, Equals, true)
}

func (s *MySuite) TestSanitizeTags(c *C) {
	specTags := []string{"abcd", "foo", "bar", "foo bar"}
	tagFilter := NewScenarioFilterBasedOnTags(specTags, "abcd & foo bar | true")
	evaluateTrue := tagFilter.filterTags(specTags)
	c.Assert(evaluateTrue, Equals, true)
}

func (s *MySuite) TestSanitizeMixedCaseTags(c *C) {
	before()
	specTags := []string{"abcd", "foo", "bar", "foo bar"}
	tagFilter := NewScenarioFilterBasedOnTags(specTags, "abcd & foo Bar | true")
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
	specs = append(specs, spec1, spec2, spec3)

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(filteredSpecs), Equals, 2)
	c.Assert(len(filteredSpecs[0].Scenarios), Equals, 1)
	c.Assert(len(filteredSpecs[1].Scenarios), Equals, 1)
	c.Assert(filteredSpecs[0].Scenarios[0], Equals, scenario1)
	c.Assert(filteredSpecs[1].Scenarios[0], Equals, scenario1)

	c.Assert(len(otherSpecs), Equals, 3)
	c.Assert(len(otherSpecs[0].Scenarios), Equals, 1)
	c.Assert(len(otherSpecs[1].Scenarios), Equals, 2)
	c.Assert(len(otherSpecs[2].Scenarios), Equals, 1)
	c.Assert(otherSpecs[0].Scenarios[0], Equals, scenario2)
	c.Assert(otherSpecs[1].Scenarios[0], Equals, scenario2)
	c.Assert(otherSpecs[1].Scenarios[1], Equals, scenario3)
	c.Assert(otherSpecs[2].Scenarios[0], Equals, scenario3)

}

func (s *MySuite) TestToFilterSpecsByMixedCaseTags(c *C) {
	before()
	myTags := []string{"tag1", "TAG2"}
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
	specs = append(specs, spec1, spec2, spec3)

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag1 & Tag2")
	c.Assert(len(filteredSpecs), Equals, 2)
	c.Assert(len(filteredSpecs[0].Scenarios), Equals, 1)
	c.Assert(len(filteredSpecs[1].Scenarios), Equals, 1)
	c.Assert(filteredSpecs[0].Scenarios[0], Equals, scenario1)
	c.Assert(filteredSpecs[1].Scenarios[0], Equals, scenario1)

	c.Assert(len(otherSpecs), Equals, 3)
	c.Assert(len(otherSpecs[0].Scenarios), Equals, 1)
	c.Assert(len(otherSpecs[1].Scenarios), Equals, 2)
	c.Assert(len(otherSpecs[2].Scenarios), Equals, 1)
	c.Assert(otherSpecs[0].Scenarios[0], Equals, scenario2)
	c.Assert(otherSpecs[1].Scenarios[0], Equals, scenario2)
	c.Assert(otherSpecs[1].Scenarios[1], Equals, scenario3)
	c.Assert(otherSpecs[2].Scenarios[0], Equals, scenario3)

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

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag1 & tag2")
	c.Assert(len(filteredSpecs[0].Scenarios), Equals, 1)
	c.Assert(filteredSpecs[0].Scenarios[0], Equals, scenario2)

	c.Assert(len(otherSpecs), Equals, 1)
	c.Assert(len(otherSpecs[0].Scenarios), Equals, 2)
	c.Assert(otherSpecs[0].Scenarios[0], Equals, scenario1)
	c.Assert(otherSpecs[0].Scenarios[1], Equals, scenario3)
}

func (s *MySuite) TestToFilterScenariosByMixedCaseTag(c *C) {
	before()
	myTags := []string{"Tag-1", "tag2"}

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

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag-1 & tag2")
	c.Assert(len(filteredSpecs[0].Scenarios), Equals, 1)
	c.Assert(filteredSpecs[0].Scenarios[0], Equals, scenario2)

	c.Assert(len(otherSpecs), Equals, 1)
	c.Assert(len(otherSpecs[0].Scenarios), Equals, 2)
	c.Assert(otherSpecs[0].Scenarios[0], Equals, scenario1)
	c.Assert(otherSpecs[0].Scenarios[1], Equals, scenario3)
}

func (s *MySuite) TestToFilterMultipleScenariosByTags(c *C) {
	myTags := []string{"tag1", "tag2"}

	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
		Tags:    &gauge.Tags{RawValues: [][]string{{"tag1"}}},
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

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag1 & tag2")

	c.Assert(len(filteredSpecs[0].Scenarios), Equals, 2)
	c.Assert(filteredSpecs[0].Scenarios[0], Equals, scenario2)
	c.Assert(filteredSpecs[0].Scenarios[1], Equals, scenario3)

	c.Assert(len(otherSpecs), Equals, 1)
	c.Assert(len(otherSpecs[0].Scenarios), Equals, 1)
	c.Assert(otherSpecs[0].Scenarios[0], Equals, scenario1)
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

	filteredSpecs, otherSpecs := filterSpecsByTags(specs, "tag3")

	c.Assert(len(filteredSpecs), Equals, 0)

	c.Assert(len(otherSpecs), Equals, 1)
	c.Assert(len(otherSpecs[0].Scenarios), Equals, 3)
	c.Assert(otherSpecs[0].Scenarios[0], Equals, scenario1)
	c.Assert(otherSpecs[0].Scenarios[1], Equals, scenario2)
	c.Assert(otherSpecs[0].Scenarios[2], Equals, scenario3)
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
func (s *MySuite) TestScenarioTagFilterShouldNotRemoveNonScenarioKindItems(c *C) {
	scenario1 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "First Scenario"},
		Span:    &gauge.Span{Start: 1, End: 3},
	}
	scenario2 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Second Scenario"},
		Span:    &gauge.Span{Start: 4, End: 6},
		Tags:    &gauge.Tags{RawValues: [][]string{{"tag2"}}},
	}
	scenario3 := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Third Scenario"},
		Span:    &gauge.Span{Start: 7, End: 10},
		Tags:    &gauge.Tags{RawValues: [][]string{{"tag1"}}},
	}
	spec := &gauge.Specification{
		Items:     []gauge.Item{scenario1, scenario2, scenario3, &gauge.Table{}, &gauge.Comment{Value: "Comment", LineNo: 1}, &gauge.Step{}},
		Scenarios: []*gauge.Scenario{scenario1, scenario2, scenario3},
	}

	specWithFilteredItems, specWithOtherItems := filterSpecsByTags([]*gauge.Specification{spec}, "tag1 | tag2")

	c.Assert(len(specWithFilteredItems), Equals, 1)
	c.Assert(len(specWithFilteredItems[0].Items), Equals, 5)

	c.Assert(len(specWithOtherItems), Equals, 1)
	c.Assert(len(specWithOtherItems[0].Items), Equals, 4)
}
