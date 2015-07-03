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

package parser

import (
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestThrowsErrorForMultipleSpecHeading(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "Example step", lineNo: 3},
		&Token{kind: specKind, value: "Another Heading", lineNo: 4},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseError.Message, Equals, "Parse error: Multiple spec headings found in same file")
	c.Assert(result.ParseError.LineNo, Equals, 4)
}

func (s *MySuite) TestThrowsErrorForScenarioWithoutSpecHeading(c *C) {
	tokens := []*Token{
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 1},
		&Token{kind: stepKind, value: "Example step", lineNo: 2},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseError.Message, Equals, "Parse error: Scenario should be defined after the spec heading")
	c.Assert(result.ParseError.LineNo, Equals, 1)
}

func (s *MySuite) TestThrowsErrorForDuplicateScenariosWithinTheSameSpec(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "Example step", lineNo: 3},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 4},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseError.Message, Equals, "Parse error: Duplicate scenario definitions are not allowed in the same specification")
	c.Assert(result.ParseError.LineNo, Equals, 4)
}

func (s *MySuite) TestSpecWithHeadingAndSimpleSteps(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "Example step", lineNo: 3},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(len(spec.items), Equals, 1)
	c.Assert(spec.items[0], Equals, spec.scenarios[0])
	scenarioItems := (spec.items[0]).(*Scenario).items
	c.Assert(scenarioItems[0], Equals, spec.scenarios[0].steps[0])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.heading.lineNo, Equals, 1)
	c.Assert(spec.heading.value, Equals, "Spec Heading")

	c.Assert(len(spec.scenarios), Equals, 1)
	c.Assert(spec.scenarios[0].heading.lineNo, Equals, 2)
	c.Assert(spec.scenarios[0].heading.value, Equals, "Scenario Heading")
	c.Assert(len(spec.scenarios[0].steps), Equals, 1)
	c.Assert(spec.scenarios[0].steps[0].value, Equals, "Example step")
}

func (s *MySuite) TestStepsAndComments(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: commentKind, value: "A comment with some text and **bold** characters", lineNo: 2},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
		&Token{kind: commentKind, value: "Another comment", lineNo: 4},
		&Token{kind: stepKind, value: "Example step", lineNo: 5},
		&Token{kind: commentKind, value: "Third comment", lineNo: 6},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))
	c.Assert(len(spec.items), Equals, 2)
	c.Assert(spec.items[0], Equals, spec.comments[0])
	c.Assert(spec.items[1], Equals, spec.scenarios[0])

	scenarioItems := (spec.items[1]).(*Scenario).items
	c.Assert(3, Equals, len(scenarioItems))
	c.Assert(scenarioItems[0], Equals, spec.scenarios[0].comments[0])
	c.Assert(scenarioItems[1], Equals, spec.scenarios[0].steps[0])
	c.Assert(scenarioItems[2], Equals, spec.scenarios[0].comments[1])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.heading.value, Equals, "Spec Heading")

	c.Assert(len(spec.comments), Equals, 1)
	c.Assert(spec.comments[0].lineNo, Equals, 2)
	c.Assert(spec.comments[0].value, Equals, "A comment with some text and **bold** characters")

	c.Assert(len(spec.scenarios), Equals, 1)
	scenario := spec.scenarios[0]

	c.Assert(2, Equals, len(scenario.comments))
	c.Assert(scenario.comments[0].lineNo, Equals, 4)
	c.Assert(scenario.comments[0].value, Equals, "Another comment")

	c.Assert(scenario.comments[1].lineNo, Equals, 6)
	c.Assert(scenario.comments[1].value, Equals, "Third comment")

	c.Assert(scenario.heading.value, Equals, "Scenario Heading")
	c.Assert(len(scenario.steps), Equals, 1)
}

func (s *MySuite) TestStepsWithParam(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: tableHeader, args: []string{"id"}, lineNo: 2},
		&Token{kind: tableRow, args: []string{"1"}, lineNo: 3},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 4},
		&Token{kind: stepKind, value: "enter {static} with {dynamic}", lineNo: 5, args: []string{"user \\n foo", "id"}},
		&Token{kind: stepKind, value: "sample \\{static\\}", lineNo: 6},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))
	c.Assert(result.Ok, Equals, true)
	step := spec.scenarios[0].steps[0]
	c.Assert(step.value, Equals, "enter {} with {}")
	c.Assert(step.lineNo, Equals, 5)
	c.Assert(len(step.args), Equals, 2)
	c.Assert(step.args[0].Value, Equals, "user \\n foo")
	c.Assert(step.args[0].ArgType, Equals, Static)
	c.Assert(step.args[1].Value, Equals, "id")
	c.Assert(step.args[1].ArgType, Equals, Dynamic)
	c.Assert(step.args[1].Name, Equals, "id")

	escapedStep := spec.scenarios[0].steps[1]
	c.Assert(escapedStep.value, Equals, "sample \\{static\\}")
	c.Assert(len(escapedStep.args), Equals, 0)
}

func (s *MySuite) TestStepsWithKeywords(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "sample {static} and {dynamic}", lineNo: 3, args: []string{"name"}},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result, NotNil)
	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}'")
	c.Assert(result.ParseError.LineNo, Equals, 3)
}

func (s *MySuite) TestContextWithKeywords(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: stepKind, value: "sample {static} and {dynamic}", lineNo: 3, args: []string{"name"}},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result, NotNil)
	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}'")
	c.Assert(result.ParseError.LineNo, Equals, 3)
}

func (s *MySuite) TestSpecWithDataTable(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading"},
		&Token{kind: commentKind, value: "Comment before data table"},
		&Token{kind: tableHeader, args: []string{"id", "name"}},
		&Token{kind: tableRow, args: []string{"1", "foo"}},
		&Token{kind: tableRow, args: []string{"2", "bar"}},
		&Token{kind: commentKind, value: "Comment before data table"},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(len(spec.items), Equals, 3)
	c.Assert(spec.items[0], Equals, spec.comments[0])
	c.Assert(spec.items[1], DeepEquals, &spec.dataTable.table)
	c.Assert(spec.items[2], Equals, spec.comments[1])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.dataTable, NotNil)
	c.Assert(len(spec.dataTable.table.Get("id")), Equals, 2)
	c.Assert(len(spec.dataTable.table.Get("name")), Equals, 2)
	c.Assert(spec.dataTable.table.Get("id")[0].value, Equals, "1")
	c.Assert(spec.dataTable.table.Get("id")[0].cellType, Equals, Static)
	c.Assert(spec.dataTable.table.Get("id")[1].value, Equals, "2")
	c.Assert(spec.dataTable.table.Get("id")[1].cellType, Equals, Static)
	c.Assert(spec.dataTable.table.Get("name")[0].value, Equals, "foo")
	c.Assert(spec.dataTable.table.Get("name")[0].cellType, Equals, Static)
	c.Assert(spec.dataTable.table.Get("name")[1].value, Equals, "bar")
	c.Assert(spec.dataTable.table.Get("name")[1].cellType, Equals, Static)
}

func (s *MySuite) TestStepWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "Step with inline table", lineNo: 3},
		&Token{kind: tableHeader, args: []string{"id", "name"}},
		&Token{kind: tableRow, args: []string{"1", "foo"}},
		&Token{kind: tableRow, args: []string{"2", "bar"}},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)
	step := spec.scenarios[0].steps[0]

	c.Assert(step.args[0].ArgType, Equals, TableArg)
	inlineTable := step.args[0].Table
	c.Assert(inlineTable, NotNil)

	c.Assert(step.value, Equals, "Step with inline table {}")
	c.Assert(step.hasInlineTable, Equals, true)
	c.Assert(len(inlineTable.Get("id")), Equals, 2)
	c.Assert(len(inlineTable.Get("name")), Equals, 2)
	c.Assert(inlineTable.Get("id")[0].value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].cellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[1].value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].cellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[0].value, Equals, "foo")
	c.Assert(inlineTable.Get("name")[0].cellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[1].value, Equals, "bar")
	c.Assert(inlineTable.Get("name")[1].cellType, Equals, Static)
}

func (s *MySuite) TestStepWithInlineTableWithDynamicParam(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: tableHeader, args: []string{"type1", "type2"}},
		&Token{kind: tableRow, args: []string{"1", "2"}},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "Step with inline table", lineNo: 3},
		&Token{kind: tableHeader, args: []string{"id", "name"}},
		&Token{kind: tableRow, args: []string{"1", "<type1>"}},
		&Token{kind: tableRow, args: []string{"2", "<type2>"}},
		&Token{kind: tableRow, args: []string{"<2>", "<type3>"}},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)
	step := spec.scenarios[0].steps[0]

	c.Assert(step.args[0].ArgType, Equals, TableArg)
	inlineTable := step.args[0].Table
	c.Assert(inlineTable, NotNil)

	c.Assert(step.value, Equals, "Step with inline table {}")
	c.Assert(len(inlineTable.Get("id")), Equals, 3)
	c.Assert(len(inlineTable.Get("name")), Equals, 3)
	c.Assert(inlineTable.Get("id")[0].value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].cellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[1].value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].cellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[2].value, Equals, "<2>")
	c.Assert(inlineTable.Get("id")[2].cellType, Equals, Static)

	c.Assert(inlineTable.Get("name")[0].value, Equals, "type1")
	c.Assert(inlineTable.Get("name")[0].cellType, Equals, Dynamic)
	c.Assert(inlineTable.Get("name")[1].value, Equals, "type2")
	c.Assert(inlineTable.Get("name")[1].cellType, Equals, Dynamic)
	c.Assert(inlineTable.Get("name")[2].value, Equals, "<type3>")
	c.Assert(inlineTable.Get("name")[2].cellType, Equals, Static)
}

func (s *MySuite) TestStepWithInlineTableWithUnResolvableDynamicParam(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: tableHeader, args: []string{"type1", "type2"}},
		&Token{kind: tableRow, args: []string{"1", "2"}},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "Step with inline table", lineNo: 3},
		&Token{kind: tableHeader, args: []string{"id", "name"}},
		&Token{kind: tableRow, args: []string{"1", "<invalid>"}},
		&Token{kind: tableRow, args: []string{"2", "<type2>"}},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))
	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.scenarios[0].steps[0].args[0].Table.Get("id")[0].value, Equals, "1")
	c.Assert(spec.scenarios[0].steps[0].args[0].Table.Get("name")[0].value, Equals, "<invalid>")
	c.Assert(result.Warnings[0].message, Equals, "Dynamic param <invalid> could not be resolved, Treating it as static param")
}

func (s *MySuite) TestContextWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading"},
		&Token{kind: stepKind, value: "Context with inline table"},
		&Token{kind: tableHeader, args: []string{"id", "name"}},
		&Token{kind: tableRow, args: []string{"1", "foo"}},
		&Token{kind: tableRow, args: []string{"2", "bar"}},
		&Token{kind: tableRow, args: []string{"3", "not a <dynamic>"}},
		&Token{kind: scenarioKind, value: "Scenario Heading"},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))
	c.Assert(len(spec.items), Equals, 2)
	c.Assert(spec.items[0], DeepEquals, spec.contexts[0])
	c.Assert(spec.items[1], Equals, spec.scenarios[0])

	c.Assert(result.Ok, Equals, true)
	context := spec.contexts[0]

	c.Assert(context.args[0].ArgType, Equals, TableArg)
	inlineTable := context.args[0].Table

	c.Assert(inlineTable, NotNil)
	c.Assert(context.value, Equals, "Context with inline table {}")
	c.Assert(len(inlineTable.Get("id")), Equals, 3)
	c.Assert(len(inlineTable.Get("name")), Equals, 3)
	c.Assert(inlineTable.Get("id")[0].value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].cellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[1].value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].cellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[2].value, Equals, "3")
	c.Assert(inlineTable.Get("id")[2].cellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[0].value, Equals, "foo")
	c.Assert(inlineTable.Get("name")[0].cellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[1].value, Equals, "bar")
	c.Assert(inlineTable.Get("name")[1].cellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[2].value, Equals, "not a <dynamic>")
	c.Assert(inlineTable.Get("name")[2].cellType, Equals, Static)
}

func (s *MySuite) TestErrorWhenDataTableHasOnlyHeader(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading"},
		&Token{kind: tableHeader, args: []string{"id", "name"}, lineNo: 3},
		&Token{kind: scenarioKind, value: "Scenario Heading"},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Data table should have at least 1 data row")
	c.Assert(result.ParseError.LineNo, Equals, 3)
}

func (s *MySuite) TestWarningWhenParsingMultipleDataTable(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading"},
		&Token{kind: commentKind, value: "Comment before data table"},
		&Token{kind: tableHeader, args: []string{"id", "name"}},
		&Token{kind: tableRow, args: []string{"1", "foo"}},
		&Token{kind: tableRow, args: []string{"2", "bar"}},
		&Token{kind: commentKind, value: "Comment before data table"},
		&Token{kind: tableHeader, args: []string{"phone"}, lineNo: 7},
		&Token{kind: tableRow, args: []string{"1"}},
		&Token{kind: tableRow, args: []string{"2"}},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.Warnings), Equals, 1)
	c.Assert(result.Warnings[0].String(), Equals, "line no: 7, Multiple data table present, ignoring table")

}

func (s *MySuite) TestWarningWhenParsingTableOccursWithoutStep(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: tableHeader, args: []string{"id", "name"}, lineNo: 3},
		&Token{kind: tableRow, args: []string{"1", "foo"}, lineNo: 4},
		&Token{kind: tableRow, args: []string{"2", "bar"}, lineNo: 5},
		&Token{kind: stepKind, value: "Step", lineNo: 6},
		&Token{kind: commentKind, value: "comment in between", lineNo: 7},
		&Token{kind: tableHeader, args: []string{"phone"}, lineNo: 8},
		&Token{kind: tableRow, args: []string{"1"}},
		&Token{kind: tableRow, args: []string{"2"}},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))
	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.Warnings), Equals, 2)
	c.Assert(result.Warnings[0].String(), Equals, "line no: 3, Table not associated with a step, ignoring table")
	c.Assert(result.Warnings[1].String(), Equals, "line no: 8, Table not associated with a step, ignoring table")

}

func (s *MySuite) TestAddSpecTags(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: tagKind, args: []string{"tag1", "tag2"}, lineNo: 2},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.tags.values), Equals, 2)
	c.Assert(spec.tags.values[0], Equals, "tag1")
	c.Assert(spec.tags.values[1], Equals, "tag2")
}

func (s *MySuite) TestAddSpecTagsAndScenarioTags(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: tagKind, args: []string{"tag1", "tag2"}, lineNo: 2},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
		&Token{kind: tagKind, args: []string{"tag3", "tag4"}, lineNo: 2},
	}

	spec, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.tags.values), Equals, 2)
	c.Assert(spec.tags.values[0], Equals, "tag1")
	c.Assert(spec.tags.values[1], Equals, "tag2")

	tags := spec.scenarios[0].tags
	c.Assert(len(tags.values), Equals, 2)
	c.Assert(tags.values[0], Equals, "tag3")
	c.Assert(tags.values[1], Equals, "tag4")
}

func (s *MySuite) TestErrorOnAddingDynamicParamterWithoutADataTable(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "Step with a {dynamic}", args: []string{"foo"}, lineNo: 3, lineText: "*Step with a <foo>"},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Dynamic parameter <foo> could not be resolved")
	c.Assert(result.ParseError.LineNo, Equals, 3)

}

func (s *MySuite) TestErrorOnAddingDynamicParamterWithoutDataTableHeaderValue(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: tableHeader, args: []string{"id, name"}, lineNo: 2},
		&Token{kind: tableRow, args: []string{"123, hello"}, lineNo: 3},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 4},
		&Token{kind: stepKind, value: "Step with a {dynamic}", args: []string{"foo"}, lineNo: 5, lineText: "*Step with a <foo>"},
	}

	_, result := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Dynamic parameter <foo> could not be resolved")
	c.Assert(result.ParseError.LineNo, Equals, 5)

}

func (s *MySuite) TestLookupaddArg(c *C) {
	lookup := new(ArgLookup)
	lookup.addArgName("param1")
	lookup.addArgName("param2")

	c.Assert(lookup.paramIndexMap["param1"], Equals, 0)
	c.Assert(lookup.paramIndexMap["param2"], Equals, 1)
	c.Assert(len(lookup.paramValue), Equals, 2)
	c.Assert(lookup.paramValue[0].name, Equals, "param1")
	c.Assert(lookup.paramValue[1].name, Equals, "param2")

}

func (s *MySuite) TestLookupContainsArg(c *C) {
	lookup := new(ArgLookup)
	lookup.addArgName("param1")
	lookup.addArgName("param2")

	c.Assert(lookup.containsArg("param1"), Equals, true)
	c.Assert(lookup.containsArg("param2"), Equals, true)
	c.Assert(lookup.containsArg("param3"), Equals, false)
}

func (s *MySuite) TestaddArgValue(c *C) {
	lookup := new(ArgLookup)
	lookup.addArgName("param1")
	lookup.addArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	lookup.addArgName("param2")
	lookup.addArgValue("param2", &StepArg{Value: "value2", ArgType: Dynamic})

	c.Assert(lookup.getArg("param1").Value, Equals, "value1")
	c.Assert(lookup.getArg("param2").Value, Equals, "value2")
}

func (s *MySuite) TestPanicForInvalidArg(c *C) {
	lookup := new(ArgLookup)

	c.Assert(func() { lookup.addArgValue("param1", &StepArg{Value: "value1", ArgType: Static}) }, Panics, "Accessing an invalid parameter (param1)")
	c.Assert(func() { lookup.getArg("param1") }, Panics, "Accessing an invalid parameter (param1)")
}

func (s *MySuite) TestGetLookupCopy(c *C) {
	originalLookup := new(ArgLookup)
	originalLookup.addArgName("param1")
	originalLookup.addArgValue("param1", &StepArg{Value: "oldValue", ArgType: Dynamic})

	copiedLookup := originalLookup.getCopy()
	copiedLookup.addArgValue("param1", &StepArg{Value: "new value", ArgType: Static})

	c.Assert(copiedLookup.getArg("param1").Value, Equals, "new value")
	c.Assert(originalLookup.getArg("param1").Value, Equals, "oldValue")
}

func (s *MySuite) TestGetLookupFromTableRow(c *C) {
	dataTable := new(Table)
	dataTable.addHeaders([]string{"id", "name"})
	dataTable.addRowValues([]string{"1", "admin"})
	dataTable.addRowValues([]string{"2", "root"})

	emptyLookup := new(ArgLookup).fromDataTableRow(new(Table), 0)
	lookup1 := new(ArgLookup).fromDataTableRow(dataTable, 0)
	lookup2 := new(ArgLookup).fromDataTableRow(dataTable, 1)

	c.Assert(emptyLookup.paramIndexMap, IsNil)

	c.Assert(lookup1.getArg("id").Value, Equals, "1")
	c.Assert(lookup1.getArg("id").ArgType, Equals, Static)
	c.Assert(lookup1.getArg("name").Value, Equals, "admin")
	c.Assert(lookup1.getArg("name").ArgType, Equals, Static)

	c.Assert(lookup2.getArg("id").Value, Equals, "2")
	c.Assert(lookup2.getArg("id").ArgType, Equals, Static)
	c.Assert(lookup2.getArg("name").Value, Equals, "root")
	c.Assert(lookup2.getArg("name").ArgType, Equals, Static)
}

func (s *MySuite) TestCreateStepFromSimpleConcept(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "concept step", lineNo: 3},
	}

	conceptDictionary := new(ConceptDictionary)
	firstStep := &Step{value: "step 1"}
	secondStep := &Step{value: "step 2"}
	conceptStep := &Step{value: "concept step", isConcept: true, conceptSteps: []*Step{firstStep, secondStep}}
	conceptDictionary.add([]*Step{conceptStep}, "file.cpt")
	spec, result := new(SpecParser).createSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.scenarios[0].steps), Equals, 1)
	specConceptStep := spec.scenarios[0].steps[0]
	c.Assert(specConceptStep.isConcept, Equals, true)
	c.Assert(specConceptStep.conceptSteps[0], Equals, firstStep)
	c.Assert(specConceptStep.conceptSteps[1], Equals, secondStep)
}

func (s *MySuite) TestCreateStepFromConceptWithParameters(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "create user {static}", args: []string{"foo"}, lineNo: 3},
		&Token{kind: stepKind, value: "create user {static}", args: []string{"bar"}, lineNo: 4},
	}

	concepts, _ := new(ConceptParser).parse("#create user <username> \n * enter user <username> \n *select \"finish\"")
	conceptsDictionary := new(ConceptDictionary)
	conceptsDictionary.add(concepts, "file.cpt")

	spec, result := new(SpecParser).createSpecification(tokens, conceptsDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.scenarios[0].steps), Equals, 2)

	firstConceptStep := spec.scenarios[0].steps[0]
	c.Assert(firstConceptStep.isConcept, Equals, true)
	c.Assert(firstConceptStep.conceptSteps[0].value, Equals, "enter user {}")
	c.Assert(firstConceptStep.conceptSteps[0].args[0].Value, Equals, "username")
	c.Assert(firstConceptStep.conceptSteps[1].value, Equals, "select {}")
	c.Assert(firstConceptStep.conceptSteps[1].args[0].Value, Equals, "finish")
	c.Assert(len(firstConceptStep.lookup.paramValue), Equals, 1)
	c.Assert(firstConceptStep.getArg("username").Value, Equals, "foo")

	secondConceptStep := spec.scenarios[0].steps[1]
	c.Assert(secondConceptStep.isConcept, Equals, true)
	c.Assert(secondConceptStep.conceptSteps[0].value, Equals, "enter user {}")
	c.Assert(secondConceptStep.conceptSteps[0].args[0].Value, Equals, "username")
	c.Assert(secondConceptStep.conceptSteps[1].value, Equals, "select {}")
	c.Assert(secondConceptStep.conceptSteps[1].args[0].Value, Equals, "finish")
	c.Assert(len(secondConceptStep.lookup.paramValue), Equals, 1)
	c.Assert(secondConceptStep.getArg("username").Value, Equals, "bar")

}

func (s *MySuite) TestCreateStepFromConceptWithDynamicParameters(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: tableHeader, args: []string{"id", "description"}, lineNo: 2},
		&Token{kind: tableRow, args: []string{"123", "Admin fellow"}, lineNo: 3},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 4},
		&Token{kind: stepKind, value: "create user {dynamic} and {dynamic}", args: []string{"id", "description"}, lineNo: 5},
		&Token{kind: stepKind, value: "create user {static} and {static}", args: []string{"456", "Regular fellow"}, lineNo: 6},
	}

	concepts, _ := new(ConceptParser).parse("#create user <user-id> and <user-description> \n * enter user <user-id> and <user-description> \n *select \"finish\"")
	conceptsDictionary := new(ConceptDictionary)
	conceptsDictionary.add(concepts, "file.cpt")
	spec, result := new(SpecParser).createSpecification(tokens, conceptsDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.items), Equals, 2)
	c.Assert(spec.items[0], DeepEquals, &spec.dataTable.table)
	c.Assert(spec.items[1], Equals, spec.scenarios[0])

	scenarioItems := (spec.items[1]).(*Scenario).items
	c.Assert(scenarioItems[0], Equals, spec.scenarios[0].steps[0])
	c.Assert(scenarioItems[1], DeepEquals, spec.scenarios[0].steps[1])

	c.Assert(len(spec.scenarios[0].steps), Equals, 2)

	firstConcept := spec.scenarios[0].steps[0]
	c.Assert(firstConcept.isConcept, Equals, true)
	c.Assert(firstConcept.conceptSteps[0].value, Equals, "enter user {} and {}")
	c.Assert(firstConcept.conceptSteps[0].args[0].ArgType, Equals, Dynamic)
	c.Assert(firstConcept.conceptSteps[0].args[0].Value, Equals, "user-id")
	c.Assert(firstConcept.conceptSteps[0].args[0].Name, Equals, "user-id")
	c.Assert(firstConcept.conceptSteps[0].args[1].ArgType, Equals, Dynamic)
	c.Assert(firstConcept.conceptSteps[0].args[1].Value, Equals, "user-description")
	c.Assert(firstConcept.conceptSteps[0].args[1].Name, Equals, "user-description")
	c.Assert(firstConcept.conceptSteps[1].value, Equals, "select {}")
	c.Assert(firstConcept.conceptSteps[1].args[0].Value, Equals, "finish")
	c.Assert(firstConcept.conceptSteps[1].args[0].ArgType, Equals, Static)

	c.Assert(len(firstConcept.lookup.paramValue), Equals, 2)
	arg1 := firstConcept.lookup.getArg("user-id")
	c.Assert(arg1.Value, Equals, "id")
	c.Assert(arg1.ArgType, Equals, Dynamic)

	arg2 := firstConcept.lookup.getArg("user-description")
	c.Assert(arg2.Value, Equals, "description")
	c.Assert(arg2.ArgType, Equals, Dynamic)

	secondConcept := spec.scenarios[0].steps[1]
	c.Assert(secondConcept.isConcept, Equals, true)
	c.Assert(secondConcept.conceptSteps[0].value, Equals, "enter user {} and {}")
	c.Assert(secondConcept.conceptSteps[0].args[0].ArgType, Equals, Dynamic)
	c.Assert(secondConcept.conceptSteps[0].args[0].Value, Equals, "user-id")
	c.Assert(secondConcept.conceptSteps[0].args[0].Name, Equals, "user-id")
	c.Assert(secondConcept.conceptSteps[0].args[1].ArgType, Equals, Dynamic)
	c.Assert(secondConcept.conceptSteps[0].args[1].Value, Equals, "user-description")
	c.Assert(secondConcept.conceptSteps[0].args[1].Name, Equals, "user-description")
	c.Assert(secondConcept.conceptSteps[1].value, Equals, "select {}")
	c.Assert(secondConcept.conceptSteps[1].args[0].Value, Equals, "finish")
	c.Assert(secondConcept.conceptSteps[1].args[0].ArgType, Equals, Static)

	c.Assert(len(secondConcept.lookup.paramValue), Equals, 2)
	arg1 = secondConcept.lookup.getArg("user-id")
	arg2 = secondConcept.lookup.getArg("user-description")
	c.Assert(arg1.Value, Equals, "456")
	c.Assert(arg1.ArgType, Equals, Static)
	c.Assert(arg2.Value, Equals, "Regular fellow")
	c.Assert(arg2.ArgType, Equals, Static)

}

func (s *MySuite) TestCreateStepFromConceptWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 4},
		&Token{kind: stepKind, value: "create users", lineNo: 3},
		&Token{kind: tableHeader, args: []string{"id", "description"}, lineNo: 4},
		&Token{kind: tableRow, args: []string{"123", "Admin"}, lineNo: 5},
		&Token{kind: tableRow, args: []string{"456", "normal fellow"}, lineNo: 6},
	}

	concepts, _ := new(ConceptParser).parse("#create users <table> \n * enter details from <table> \n *select \"finish\"")
	conceptsDictionary := new(ConceptDictionary)
	conceptsDictionary.add(concepts, "file.cpt")
	spec, result := new(SpecParser).createSpecification(tokens, conceptsDictionary)
	c.Assert(result.Ok, Equals, true)

	steps := spec.scenarios[0].steps
	c.Assert(len(steps), Equals, 1)
	c.Assert(steps[0].isConcept, Equals, true)
	c.Assert(steps[0].value, Equals, "create users {}")
	c.Assert(len(steps[0].args), Equals, 1)
	c.Assert(steps[0].args[0].ArgType, Equals, TableArg)
	c.Assert(len(steps[0].conceptSteps), Equals, 2)
}

func (s *MySuite) TestCreateStepFromConceptWithInlineTableHavingDynamicParam(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: tableHeader, args: []string{"id", "description"}, lineNo: 2},
		&Token{kind: tableRow, args: []string{"123", "Admin"}, lineNo: 3},
		&Token{kind: tableRow, args: []string{"456", "normal fellow"}, lineNo: 4},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 5},
		&Token{kind: stepKind, value: "create users", lineNo: 6},
		&Token{kind: tableHeader, args: []string{"user-id", "description", "name"}, lineNo: 7},
		&Token{kind: tableRow, args: []string{"<id>", "<description>", "root"}, lineNo: 8},
		&Token{kind: stepKind, value: "create users", lineNo: 9},
		&Token{kind: tableHeader, args: []string{"user-id", "description", "name"}, lineNo: 10},
		&Token{kind: tableRow, args: []string{"1", "normal", "wheel"}, lineNo: 11},
	}

	concepts, _ := new(ConceptParser).parse("#create users <id> \n * enter details from <id> \n *select \"finish\"")
	conceptsDictionary := new(ConceptDictionary)
	conceptsDictionary.add(concepts, "file.cpt")
	spec, result := new(SpecParser).createSpecification(tokens, conceptsDictionary)
	c.Assert(result.Ok, Equals, true)

	steps := spec.scenarios[0].steps
	c.Assert(len(steps), Equals, 2)
	c.Assert(steps[0].isConcept, Equals, true)
	c.Assert(steps[1].isConcept, Equals, true)
	c.Assert(steps[0].value, Equals, "create users {}")
	c.Assert(len(steps[0].args), Equals, 1)
	c.Assert(steps[0].args[0].ArgType, Equals, TableArg)
	table := steps[0].args[0].Table
	c.Assert(table.Get("user-id")[0].value, Equals, "id")
	c.Assert(table.Get("user-id")[0].cellType, Equals, Dynamic)
	c.Assert(table.Get("description")[0].value, Equals, "description")
	c.Assert(table.Get("description")[0].cellType, Equals, Dynamic)
	c.Assert(table.Get("name")[0].value, Equals, "root")
	c.Assert(table.Get("name")[0].cellType, Equals, Static)
	c.Assert(len(steps[0].conceptSteps), Equals, 2)
}

func (s *MySuite) TestPopulateFragmentsForSimpleStep(c *C) {
	step := &Step{value: "This is a simple step"}

	step.populateFragments()

	c.Assert(len(step.fragments), Equals, 1)
	fragment := step.fragments[0]
	c.Assert(fragment.GetText(), Equals, "This is a simple step")
	c.Assert(fragment.GetFragmentType(), Equals, gauge_messages.Fragment_Text)
}

func (s *MySuite) TestGetArgForStep(c *C) {
	lookup := new(ArgLookup)
	lookup.addArgName("param1")
	lookup.addArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	step := &Step{lookup: *lookup}

	c.Assert(step.getArg("param1").Value, Equals, "value1")
}

func (s *MySuite) TestGetArgForConceptStep(c *C) {
	lookup := new(ArgLookup)
	lookup.addArgName("param1")
	lookup.addArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	concept := &Step{lookup: *lookup, isConcept: true}
	stepLookup := new(ArgLookup)
	stepLookup.addArgName("param1")
	stepLookup.addArgValue("param1", &StepArg{Value: "param1", ArgType: Dynamic})
	step := &Step{parent: concept, lookup: *stepLookup}

	c.Assert(step.getArg("param1").Value, Equals, "value1")
}

func (s *MySuite) TestPopulateFragmentsForStepWithParameters(c *C) {
	arg1 := &StepArg{Value: "first", ArgType: Static}
	arg2 := &StepArg{Value: "second", ArgType: Dynamic, Name: "second"}
	argTable := new(Table)
	headers := []string{"header1", "header2"}
	row1 := []string{"row1", "row2"}
	argTable.addHeaders(headers)
	argTable.addRowValues(row1)
	arg3 := &StepArg{ArgType: SpecialString, Value: "text from file", Name: "file:foo.txt"}
	arg4 := &StepArg{Table: *argTable, ArgType: TableArg}
	stepArgs := []*StepArg{arg1, arg2, arg3, arg4}
	step := &Step{value: "{} step with {} and {}, {}", args: stepArgs}

	step.populateFragments()

	c.Assert(len(step.fragments), Equals, 7)
	fragment1 := step.fragments[0]
	c.Assert(fragment1.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment1.GetParameter().GetValue(), Equals, "first")
	c.Assert(fragment1.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Static)

	fragment2 := step.fragments[1]
	c.Assert(fragment2.GetText(), Equals, " step with ")
	c.Assert(fragment2.GetFragmentType(), Equals, gauge_messages.Fragment_Text)

	fragment3 := step.fragments[2]
	c.Assert(fragment3.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment3.GetParameter().GetValue(), Equals, "second")
	c.Assert(fragment3.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)

	fragment4 := step.fragments[3]
	c.Assert(fragment4.GetText(), Equals, " and ")
	c.Assert(fragment4.GetFragmentType(), Equals, gauge_messages.Fragment_Text)

	fragment5 := step.fragments[4]
	c.Assert(fragment5.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment5.GetParameter().GetValue(), Equals, "text from file")
	c.Assert(fragment5.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Special_String)
	c.Assert(fragment5.GetParameter().GetName(), Equals, "file:foo.txt")

	fragment6 := step.fragments[5]
	c.Assert(fragment6.GetText(), Equals, ", ")
	c.Assert(fragment6.GetFragmentType(), Equals, gauge_messages.Fragment_Text)

	fragment7 := step.fragments[6]
	c.Assert(fragment7.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment7.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Table)
	protoTable := fragment7.GetParameter().GetTable()
	c.Assert(protoTable.GetHeaders().GetCells(), DeepEquals, headers)
	c.Assert(len(protoTable.GetRows()), Equals, 1)
	c.Assert(protoTable.GetRows()[0].GetCells(), DeepEquals, row1)
}

func (s *MySuite) TestUpdatePropertiesFromAnotherStep(c *C) {
	argsInStep := []*StepArg{&StepArg{Name: "arg1", Value: "arg value", ArgType: Dynamic}}
	fragments := []*gauge_messages.Fragment{&gauge_messages.Fragment{Text: proto.String("foo")}}
	originalStep := &Step{lineNo: 12,
		value:          "foo {}",
		lineText:       "foo <bar>",
		args:           argsInStep,
		isConcept:      false,
		fragments:      fragments,
		hasInlineTable: false}

	destinationStep := new(Step)
	destinationStep.copyFrom(originalStep)

	c.Assert(destinationStep, DeepEquals, originalStep)
}

func (s *MySuite) TestUpdatePropertiesFromAnotherConcept(c *C) {
	argsInStep := []*StepArg{&StepArg{Name: "arg1", Value: "arg value", ArgType: Dynamic}}
	argLookup := new(ArgLookup)
	argLookup.addArgName("name")
	argLookup.addArgName("id")
	fragments := []*gauge_messages.Fragment{&gauge_messages.Fragment{Text: proto.String("foo")}}
	conceptSteps := []*Step{&Step{value: "step 1"}}
	originalConcept := &Step{
		lineNo:         12,
		value:          "foo {}",
		lineText:       "foo <bar>",
		args:           argsInStep,
		isConcept:      true,
		lookup:         *argLookup,
		fragments:      fragments,
		conceptSteps:   conceptSteps,
		hasInlineTable: false}

	destinationConcept := new(Step)
	destinationConcept.copyFrom(originalConcept)

	c.Assert(destinationConcept, DeepEquals, originalConcept)
}

func (s *MySuite) TestCreateConceptStep(c *C) {
	conceptText := SpecBuilder().
		specHeading("concept with <foo>").
		step("nested concept with <foo>").
		specHeading("nested concept with <baz>").
		step("nested concept step wiht <baz>").String()
	concepts, _ := new(ConceptParser).parse(conceptText)

	dictionary := new(ConceptDictionary)
	dictionary.add(concepts, "file.cpt")

	argsInStep := []*StepArg{&StepArg{Name: "arg1", Value: "value", ArgType: Static}}
	originalStep := &Step{
		lineNo:         12,
		value:          "concept with {}",
		lineText:       "concept with \"value\"",
		args:           argsInStep,
		isConcept:      true,
		hasInlineTable: false}
	new(Specification).createConceptStep(dictionary.search("concept with {}").ConceptStep, originalStep)
	c.Assert(originalStep.isConcept, Equals, true)
	c.Assert(len(originalStep.conceptSteps), Equals, 1)
	c.Assert(originalStep.args[0].Value, Equals, "value")

	c.Assert(originalStep.lookup.getArg("foo").Value, Equals, "value")

	nestedConcept := originalStep.conceptSteps[0]
	c.Assert(nestedConcept.isConcept, Equals, true)
	c.Assert(len(nestedConcept.conceptSteps), Equals, 1)

	c.Assert(nestedConcept.args[0].ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.args[0].Value, Equals, "foo")

	c.Assert(nestedConcept.conceptSteps[0].args[0].ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.conceptSteps[0].args[0].Value, Equals, "baz")

	c.Assert(nestedConcept.lookup.getArg("baz").ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.lookup.getArg("baz").Value, Equals, "foo")
}

func (s *MySuite) TestRenameStep(c *C) {
	argsInStep := []*StepArg{&StepArg{Name: "arg1", Value: "value", ArgType: Static}, &StepArg{Name: "arg2", Value: "value1", ArgType: Static}}
	originalStep := &Step{
		lineNo:         12,
		value:          "step with {}",
		args:           argsInStep,
		isConcept:      false,
		hasInlineTable: false}

	argsInStep = []*StepArg{&StepArg{Name: "arg2", Value: "value1", ArgType: Static}, &StepArg{Name: "arg1", Value: "value", ArgType: Static}}
	newStep := &Step{
		lineNo:         12,
		value:          "step from {} {}",
		args:           argsInStep,
		isConcept:      false,
		hasInlineTable: false}
	orderMap := make(map[int]int)
	orderMap[0] = 1
	orderMap[1] = 0
	isConcept := false
	isRefactored := originalStep.rename(*originalStep, *newStep, false, orderMap, &isConcept)

	c.Assert(isRefactored, Equals, true)
	c.Assert(originalStep.value, Equals, "step from {} {}")
	c.Assert(originalStep.args[0].Name, Equals, "arg2")
	c.Assert(originalStep.args[1].Name, Equals, "arg1")
}

func (s *MySuite) TestGetLineTextForStep(c *C) {
	step := &Step{lineText: "foo"}

	c.Assert(step.getLineText(), Equals, "foo")
}

func (s *MySuite) TestGetLineTextForStepWithTable(c *C) {
	step := &Step{
		lineText:       "foo",
		hasInlineTable: true}

	c.Assert(step.getLineText(), Equals, "foo <table>")
}

func (s *MySuite) TestCreateInValidSpecialArgInStep(c *C) {
	tokens := []*Token{
		&Token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&Token{kind: tableHeader, args: []string{"unknown:foo", "description"}, lineNo: 2},
		&Token{kind: tableRow, args: []string{"123", "Admin"}, lineNo: 3},
		&Token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&Token{kind: stepKind, value: "Example {special} step", lineNo: 3, args: []string{"unknown:foo"}},
	}
	spec, parseResults := new(SpecParser).createSpecification(tokens, new(ConceptDictionary))
	c.Assert(spec.scenarios[0].steps[0].args[0].ArgType, Equals, Dynamic)
	c.Assert(len(parseResults.Warnings), Equals, 1)
	c.Assert(parseResults.Warnings[0].message, Equals, "Could not resolve special param type <unknown:foo>. Treating it as dynamic param.")
}
