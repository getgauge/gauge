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
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "Example step", LineNo: 3},
		&Token{Kind: SpecKind, Value: "Another Heading", LineNo: 4},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseError.Message, Equals, "Parse error: Multiple spec headings found in same file")
	c.Assert(result.ParseError.LineNo, Equals, 4)
}

func (s *MySuite) TestThrowsErrorForScenarioWithoutSpecHeading(c *C) {
	tokens := []*Token{
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 1},
		&Token{Kind: StepKind, Value: "Example step", LineNo: 2},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseError.Message, Equals, "Parse error: Scenario should be defined after the spec heading")
	c.Assert(result.ParseError.LineNo, Equals, 1)
}

func (s *MySuite) TestThrowsErrorForDuplicateScenariosWithinTheSameSpec(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "Example step", LineNo: 3},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 4},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseError.Message, Equals, "Parse error: Duplicate scenario definitions are not allowed in the same specification")
	c.Assert(result.ParseError.LineNo, Equals, 4)
}

func (s *MySuite) TestSpecWithHeadingAndSimpleSteps(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "Example step", LineNo: 3},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(len(spec.Items), Equals, 1)
	c.Assert(spec.Items[0], Equals, spec.Scenarios[0])
	scenarioItems := (spec.Items[0]).(*Scenario).Items
	c.Assert(scenarioItems[0], Equals, spec.Scenarios[0].Steps[0])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.Heading.LineNo, Equals, 1)
	c.Assert(spec.Heading.Value, Equals, "Spec Heading")

	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.LineNo, Equals, 2)
	c.Assert(spec.Scenarios[0].Heading.Value, Equals, "Scenario Heading")
	c.Assert(len(spec.Scenarios[0].Steps), Equals, 1)
	c.Assert(spec.Scenarios[0].Steps[0].Value, Equals, "Example step")
}

func (s *MySuite) TestStepsAndComments(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: CommentKind, Value: "A comment with some text and **bold** characters", LineNo: 2},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&Token{Kind: CommentKind, Value: "Another comment", LineNo: 4},
		&Token{Kind: StepKind, Value: "Example step", LineNo: 5},
		&Token{Kind: CommentKind, Value: "Third comment", LineNo: 6},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))
	c.Assert(len(spec.Items), Equals, 2)
	c.Assert(spec.Items[0], Equals, spec.Comments[0])
	c.Assert(spec.Items[1], Equals, spec.Scenarios[0])

	scenarioItems := (spec.Items[1]).(*Scenario).Items
	c.Assert(3, Equals, len(scenarioItems))
	c.Assert(scenarioItems[0], Equals, spec.Scenarios[0].Comments[0])
	c.Assert(scenarioItems[1], Equals, spec.Scenarios[0].Steps[0])
	c.Assert(scenarioItems[2], Equals, spec.Scenarios[0].Comments[1])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.Heading.Value, Equals, "Spec Heading")

	c.Assert(len(spec.Comments), Equals, 1)
	c.Assert(spec.Comments[0].LineNo, Equals, 2)
	c.Assert(spec.Comments[0].Value, Equals, "A comment with some text and **bold** characters")

	c.Assert(len(spec.Scenarios), Equals, 1)
	scenario := spec.Scenarios[0]

	c.Assert(2, Equals, len(scenario.Comments))
	c.Assert(scenario.Comments[0].LineNo, Equals, 4)
	c.Assert(scenario.Comments[0].Value, Equals, "Another comment")

	c.Assert(scenario.Comments[1].LineNo, Equals, 6)
	c.Assert(scenario.Comments[1].Value, Equals, "Third comment")

	c.Assert(scenario.Heading.Value, Equals, "Scenario Heading")
	c.Assert(len(scenario.Steps), Equals, 1)
}

func (s *MySuite) TestStepsWithParam(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: TableHeader, Args: []string{"id"}, LineNo: 2},
		&Token{Kind: TableRow, Args: []string{"1"}, LineNo: 3},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: StepKind, Value: "enter {static} with {dynamic}", LineNo: 5, Args: []string{"user \\n foo", "id"}},
		&Token{Kind: StepKind, Value: "sample \\{static\\}", LineNo: 6},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))
	c.Assert(result.Ok, Equals, true)
	step := spec.Scenarios[0].Steps[0]
	c.Assert(step.Value, Equals, "enter {} with {}")
	c.Assert(step.LineNo, Equals, 5)
	c.Assert(len(step.Args), Equals, 2)
	c.Assert(step.Args[0].Value, Equals, "user \\n foo")
	c.Assert(step.Args[0].ArgType, Equals, Static)
	c.Assert(step.Args[1].Value, Equals, "id")
	c.Assert(step.Args[1].ArgType, Equals, Dynamic)
	c.Assert(step.Args[1].Name, Equals, "id")

	escapedStep := spec.Scenarios[0].Steps[1]
	c.Assert(escapedStep.Value, Equals, "sample \\{static\\}")
	c.Assert(len(escapedStep.Args), Equals, 0)
}

func (s *MySuite) TestStepsWithKeywords(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "sample {static} and {dynamic}", LineNo: 3, Args: []string{"name"}},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result, NotNil)
	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}'")
	c.Assert(result.ParseError.LineNo, Equals, 3)
}

func (s *MySuite) TestContextWithKeywords(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: StepKind, Value: "sample {static} and {dynamic}", LineNo: 3, Args: []string{"name"}},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result, NotNil)
	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}'")
	c.Assert(result.ParseError.LineNo, Equals, 3)
}

func (s *MySuite) TestSpecWithDataTable(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading"},
		&Token{Kind: CommentKind, Value: "Comment before data table"},
		&Token{Kind: TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: TableRow, Args: []string{"2", "bar"}},
		&Token{Kind: CommentKind, Value: "Comment before data table"},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(len(spec.Items), Equals, 3)
	c.Assert(spec.Items[0], Equals, spec.Comments[0])
	c.Assert(spec.Items[1], DeepEquals, &spec.DataTable.Table)
	c.Assert(spec.Items[2], Equals, spec.Comments[1])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.DataTable, NotNil)
	c.Assert(len(spec.DataTable.Table.Get("id")), Equals, 2)
	c.Assert(len(spec.DataTable.Table.Get("name")), Equals, 2)
	c.Assert(spec.DataTable.Table.Get("id")[0].Value, Equals, "1")
	c.Assert(spec.DataTable.Table.Get("id")[0].CellType, Equals, Static)
	c.Assert(spec.DataTable.Table.Get("id")[1].Value, Equals, "2")
	c.Assert(spec.DataTable.Table.Get("id")[1].CellType, Equals, Static)
	c.Assert(spec.DataTable.Table.Get("name")[0].Value, Equals, "foo")
	c.Assert(spec.DataTable.Table.Get("name")[0].CellType, Equals, Static)
	c.Assert(spec.DataTable.Table.Get("name")[1].Value, Equals, "bar")
	c.Assert(spec.DataTable.Table.Get("name")[1].CellType, Equals, Static)
}

func (s *MySuite) TestStepWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "Step with inline table", LineNo: 3},
		&Token{Kind: TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: TableRow, Args: []string{"2", "bar"}},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)
	step := spec.Scenarios[0].Steps[0]

	c.Assert(step.Args[0].ArgType, Equals, TableArg)
	inlineTable := step.Args[0].Table
	c.Assert(inlineTable, NotNil)

	c.Assert(step.Value, Equals, "Step with inline table {}")
	c.Assert(step.HasInlineTable, Equals, true)
	c.Assert(len(inlineTable.Get("id")), Equals, 2)
	c.Assert(len(inlineTable.Get("name")), Equals, 2)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[0].Value, Equals, "foo")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "bar")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, Static)
}

func (s *MySuite) TestStepWithInlineTableWithDynamicParam(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: TableHeader, Args: []string{"type1", "type2"}},
		&Token{Kind: TableRow, Args: []string{"1", "2"}},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "Step with inline table", LineNo: 3},
		&Token{Kind: TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: TableRow, Args: []string{"1", "<type1>"}},
		&Token{Kind: TableRow, Args: []string{"2", "<type2>"}},
		&Token{Kind: TableRow, Args: []string{"<2>", "<type3>"}},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)
	step := spec.Scenarios[0].Steps[0]

	c.Assert(step.Args[0].ArgType, Equals, TableArg)
	inlineTable := step.Args[0].Table
	c.Assert(inlineTable, NotNil)

	c.Assert(step.Value, Equals, "Step with inline table {}")
	c.Assert(len(inlineTable.Get("id")), Equals, 3)
	c.Assert(len(inlineTable.Get("name")), Equals, 3)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[2].Value, Equals, "<2>")
	c.Assert(inlineTable.Get("id")[2].CellType, Equals, Static)

	c.Assert(inlineTable.Get("name")[0].Value, Equals, "type1")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, Dynamic)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "type2")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, Dynamic)
	c.Assert(inlineTable.Get("name")[2].Value, Equals, "<type3>")
	c.Assert(inlineTable.Get("name")[2].CellType, Equals, Static)
}

func (s *MySuite) TestStepWithInlineTableWithUnResolvableDynamicParam(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: TableHeader, Args: []string{"type1", "type2"}},
		&Token{Kind: TableRow, Args: []string{"1", "2"}},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "Step with inline table", LineNo: 3},
		&Token{Kind: TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: TableRow, Args: []string{"1", "<invalid>"}},
		&Token{Kind: TableRow, Args: []string{"2", "<type2>"}},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))
	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.Scenarios[0].Steps[0].Args[0].Table.Get("id")[0].Value, Equals, "1")
	c.Assert(spec.Scenarios[0].Steps[0].Args[0].Table.Get("name")[0].Value, Equals, "<invalid>")
	c.Assert(result.Warnings[0].Message, Equals, "Dynamic param <invalid> could not be resolved, Treating it as static param")
}

func (s *MySuite) TestContextWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading"},
		&Token{Kind: StepKind, Value: "Context with inline table"},
		&Token{Kind: TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: TableRow, Args: []string{"2", "bar"}},
		&Token{Kind: TableRow, Args: []string{"3", "not a <dynamic>"}},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading"},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))
	c.Assert(len(spec.Items), Equals, 2)
	c.Assert(spec.Items[0], DeepEquals, spec.Contexts[0])
	c.Assert(spec.Items[1], Equals, spec.Scenarios[0])

	c.Assert(result.Ok, Equals, true)
	context := spec.Contexts[0]

	c.Assert(context.Args[0].ArgType, Equals, TableArg)
	inlineTable := context.Args[0].Table

	c.Assert(inlineTable, NotNil)
	c.Assert(context.Value, Equals, "Context with inline table {}")
	c.Assert(len(inlineTable.Get("id")), Equals, 3)
	c.Assert(len(inlineTable.Get("name")), Equals, 3)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, Static)
	c.Assert(inlineTable.Get("id")[2].Value, Equals, "3")
	c.Assert(inlineTable.Get("id")[2].CellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[0].Value, Equals, "foo")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "bar")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, Static)
	c.Assert(inlineTable.Get("name")[2].Value, Equals, "not a <dynamic>")
	c.Assert(inlineTable.Get("name")[2].CellType, Equals, Static)
}

func (s *MySuite) TestErrorWhenDataTableHasOnlyHeader(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading"},
		&Token{Kind: TableHeader, Args: []string{"id", "name"}, LineNo: 3},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading"},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Data table should have at least 1 data row")
	c.Assert(result.ParseError.LineNo, Equals, 3)
}

func (s *MySuite) TestWarningWhenParsingMultipleDataTable(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading"},
		&Token{Kind: CommentKind, Value: "Comment before data table"},
		&Token{Kind: TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: TableRow, Args: []string{"2", "bar"}},
		&Token{Kind: CommentKind, Value: "Comment before data table"},
		&Token{Kind: TableHeader, Args: []string{"phone"}, LineNo: 7},
		&Token{Kind: TableRow, Args: []string{"1"}},
		&Token{Kind: TableRow, Args: []string{"2"}},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.Warnings), Equals, 1)
	c.Assert(result.Warnings[0].String(), Equals, "line no: 7, Multiple data table present, ignoring table")

}

func (s *MySuite) TestWarningWhenParsingTableOccursWithoutStep(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: TableHeader, Args: []string{"id", "name"}, LineNo: 3},
		&Token{Kind: TableRow, Args: []string{"1", "foo"}, LineNo: 4},
		&Token{Kind: TableRow, Args: []string{"2", "bar"}, LineNo: 5},
		&Token{Kind: StepKind, Value: "Step", LineNo: 6},
		&Token{Kind: CommentKind, Value: "comment in between", LineNo: 7},
		&Token{Kind: TableHeader, Args: []string{"phone"}, LineNo: 8},
		&Token{Kind: TableRow, Args: []string{"1"}},
		&Token{Kind: TableRow, Args: []string{"2"}},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))
	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.Warnings), Equals, 2)
	c.Assert(result.Warnings[0].String(), Equals, "line no: 3, Table not associated with a step, ignoring table")
	c.Assert(result.Warnings[1].String(), Equals, "line no: 8, Table not associated with a step, ignoring table")

}

func (s *MySuite) TestAddSpecTags(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 3},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Tags.Values), Equals, 2)
	c.Assert(spec.Tags.Values[0], Equals, "tag1")
	c.Assert(spec.Tags.Values[1], Equals, "tag2")
}

func (s *MySuite) TestAddSpecTagsAndScenarioTags(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&Token{Kind: TagKind, Args: []string{"tag3", "tag4"}, LineNo: 2},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Tags.Values), Equals, 2)
	c.Assert(spec.Tags.Values[0], Equals, "tag1")
	c.Assert(spec.Tags.Values[1], Equals, "tag2")

	tags := spec.Scenarios[0].Tags
	c.Assert(len(tags.Values), Equals, 2)
	c.Assert(tags.Values[0], Equals, "tag3")
	c.Assert(tags.Values[1], Equals, "tag4")
}

func (s *MySuite) TestErrorOnAddingDynamicParamterWithoutADataTable(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "Step with a {dynamic}", Args: []string{"foo"}, LineNo: 3, LineText: "*Step with a <foo>"},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Dynamic parameter <foo> could not be resolved")
	c.Assert(result.ParseError.LineNo, Equals, 3)

}

func (s *MySuite) TestErrorOnAddingDynamicParamterWithoutDataTableHeaderValue(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: TableHeader, Args: []string{"id, name"}, LineNo: 2},
		&Token{Kind: TableRow, Args: []string{"123, hello"}, LineNo: 3},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: StepKind, Value: "Step with a {dynamic}", Args: []string{"foo"}, LineNo: 5, LineText: "*Step with a <foo>"},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))

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

	emptyLookup := new(ArgLookup).FromDataTableRow(new(Table), 0)
	lookup1 := new(ArgLookup).FromDataTableRow(dataTable, 0)
	lookup2 := new(ArgLookup).FromDataTableRow(dataTable, 1)

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
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "concept step", LineNo: 3},
	}

	conceptDictionary := new(ConceptDictionary)
	firstStep := &Step{Value: "step 1"}
	secondStep := &Step{Value: "step 2"}
	conceptStep := &Step{Value: "concept step", IsConcept: true, ConceptSteps: []*Step{firstStep, secondStep}}
	conceptDictionary.Add([]*Step{conceptStep}, "file.cpt")
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Scenarios[0].Steps), Equals, 1)
	specConceptStep := spec.Scenarios[0].Steps[0]
	c.Assert(specConceptStep.IsConcept, Equals, true)
	c.Assert(specConceptStep.ConceptSteps[0], Equals, firstStep)
	c.Assert(specConceptStep.ConceptSteps[1], Equals, secondStep)
}

func (s *MySuite) TestCreateStepFromConceptWithParameters(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "create user {static}", Args: []string{"foo"}, LineNo: 3},
		&Token{Kind: StepKind, Value: "create user {static}", Args: []string{"bar"}, LineNo: 4},
	}

	concepts, _ := new(ConceptParser).Parse("#create user <username> \n * enter user <username> \n *select \"finish\"")
	conceptsDictionary := new(ConceptDictionary)
	conceptsDictionary.Add(concepts, "file.cpt")

	spec, result := new(SpecParser).CreateSpecification(tokens, conceptsDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Scenarios[0].Steps), Equals, 2)

	firstConceptStep := spec.Scenarios[0].Steps[0]
	c.Assert(firstConceptStep.IsConcept, Equals, true)
	c.Assert(firstConceptStep.ConceptSteps[0].Value, Equals, "enter user {}")
	c.Assert(firstConceptStep.ConceptSteps[0].Args[0].Value, Equals, "username")
	c.Assert(firstConceptStep.ConceptSteps[1].Value, Equals, "select {}")
	c.Assert(firstConceptStep.ConceptSteps[1].Args[0].Value, Equals, "finish")
	c.Assert(len(firstConceptStep.Lookup.paramValue), Equals, 1)
	c.Assert(firstConceptStep.getArg("username").Value, Equals, "foo")

	secondConceptStep := spec.Scenarios[0].Steps[1]
	c.Assert(secondConceptStep.IsConcept, Equals, true)
	c.Assert(secondConceptStep.ConceptSteps[0].Value, Equals, "enter user {}")
	c.Assert(secondConceptStep.ConceptSteps[0].Args[0].Value, Equals, "username")
	c.Assert(secondConceptStep.ConceptSteps[1].Value, Equals, "select {}")
	c.Assert(secondConceptStep.ConceptSteps[1].Args[0].Value, Equals, "finish")
	c.Assert(len(secondConceptStep.Lookup.paramValue), Equals, 1)
	c.Assert(secondConceptStep.getArg("username").Value, Equals, "bar")

}

func (s *MySuite) TestCreateStepFromConceptWithDynamicParameters(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: TableHeader, Args: []string{"id", "description"}, LineNo: 2},
		&Token{Kind: TableRow, Args: []string{"123", "Admin fellow"}, LineNo: 3},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: StepKind, Value: "create user {dynamic} and {dynamic}", Args: []string{"id", "description"}, LineNo: 5},
		&Token{Kind: StepKind, Value: "create user {static} and {static}", Args: []string{"456", "Regular fellow"}, LineNo: 6},
	}

	concepts, _ := new(ConceptParser).Parse("#create user <user-id> and <user-description> \n * enter user <user-id> and <user-description> \n *select \"finish\"")
	conceptsDictionary := new(ConceptDictionary)
	conceptsDictionary.Add(concepts, "file.cpt")
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptsDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Items), Equals, 2)
	c.Assert(spec.Items[0], DeepEquals, &spec.DataTable.Table)
	c.Assert(spec.Items[1], Equals, spec.Scenarios[0])

	scenarioItems := (spec.Items[1]).(*Scenario).Items
	c.Assert(scenarioItems[0], Equals, spec.Scenarios[0].Steps[0])
	c.Assert(scenarioItems[1], DeepEquals, spec.Scenarios[0].Steps[1])

	c.Assert(len(spec.Scenarios[0].Steps), Equals, 2)

	firstConcept := spec.Scenarios[0].Steps[0]
	c.Assert(firstConcept.IsConcept, Equals, true)
	c.Assert(firstConcept.ConceptSteps[0].Value, Equals, "enter user {} and {}")
	c.Assert(firstConcept.ConceptSteps[0].Args[0].ArgType, Equals, Dynamic)
	c.Assert(firstConcept.ConceptSteps[0].Args[0].Value, Equals, "user-id")
	c.Assert(firstConcept.ConceptSteps[0].Args[0].Name, Equals, "user-id")
	c.Assert(firstConcept.ConceptSteps[0].Args[1].ArgType, Equals, Dynamic)
	c.Assert(firstConcept.ConceptSteps[0].Args[1].Value, Equals, "user-description")
	c.Assert(firstConcept.ConceptSteps[0].Args[1].Name, Equals, "user-description")
	c.Assert(firstConcept.ConceptSteps[1].Value, Equals, "select {}")
	c.Assert(firstConcept.ConceptSteps[1].Args[0].Value, Equals, "finish")
	c.Assert(firstConcept.ConceptSteps[1].Args[0].ArgType, Equals, Static)

	c.Assert(len(firstConcept.Lookup.paramValue), Equals, 2)
	arg1 := firstConcept.Lookup.getArg("user-id")
	c.Assert(arg1.Value, Equals, "id")
	c.Assert(arg1.ArgType, Equals, Dynamic)

	arg2 := firstConcept.Lookup.getArg("user-description")
	c.Assert(arg2.Value, Equals, "description")
	c.Assert(arg2.ArgType, Equals, Dynamic)

	secondConcept := spec.Scenarios[0].Steps[1]
	c.Assert(secondConcept.IsConcept, Equals, true)
	c.Assert(secondConcept.ConceptSteps[0].Value, Equals, "enter user {} and {}")
	c.Assert(secondConcept.ConceptSteps[0].Args[0].ArgType, Equals, Dynamic)
	c.Assert(secondConcept.ConceptSteps[0].Args[0].Value, Equals, "user-id")
	c.Assert(secondConcept.ConceptSteps[0].Args[0].Name, Equals, "user-id")
	c.Assert(secondConcept.ConceptSteps[0].Args[1].ArgType, Equals, Dynamic)
	c.Assert(secondConcept.ConceptSteps[0].Args[1].Value, Equals, "user-description")
	c.Assert(secondConcept.ConceptSteps[0].Args[1].Name, Equals, "user-description")
	c.Assert(secondConcept.ConceptSteps[1].Value, Equals, "select {}")
	c.Assert(secondConcept.ConceptSteps[1].Args[0].Value, Equals, "finish")
	c.Assert(secondConcept.ConceptSteps[1].Args[0].ArgType, Equals, Static)

	c.Assert(len(secondConcept.Lookup.paramValue), Equals, 2)
	arg1 = secondConcept.Lookup.getArg("user-id")
	arg2 = secondConcept.Lookup.getArg("user-description")
	c.Assert(arg1.Value, Equals, "456")
	c.Assert(arg1.ArgType, Equals, Static)
	c.Assert(arg2.Value, Equals, "Regular fellow")
	c.Assert(arg2.ArgType, Equals, Static)

}

func (s *MySuite) TestCreateStepFromConceptWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: StepKind, Value: "create users", LineNo: 3},
		&Token{Kind: TableHeader, Args: []string{"id", "description"}, LineNo: 4},
		&Token{Kind: TableRow, Args: []string{"123", "Admin"}, LineNo: 5},
		&Token{Kind: TableRow, Args: []string{"456", "normal fellow"}, LineNo: 6},
	}

	concepts, _ := new(ConceptParser).Parse("#create users <table> \n * enter details from <table> \n *select \"finish\"")
	conceptsDictionary := new(ConceptDictionary)
	conceptsDictionary.Add(concepts, "file.cpt")
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptsDictionary)
	c.Assert(result.Ok, Equals, true)

	steps := spec.Scenarios[0].Steps
	c.Assert(len(steps), Equals, 1)
	c.Assert(steps[0].IsConcept, Equals, true)
	c.Assert(steps[0].Value, Equals, "create users {}")
	c.Assert(len(steps[0].Args), Equals, 1)
	c.Assert(steps[0].Args[0].ArgType, Equals, TableArg)
	c.Assert(len(steps[0].ConceptSteps), Equals, 2)
}

func (s *MySuite) TestCreateStepFromConceptWithInlineTableHavingDynamicParam(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: TableHeader, Args: []string{"id", "description"}, LineNo: 2},
		&Token{Kind: TableRow, Args: []string{"123", "Admin"}, LineNo: 3},
		&Token{Kind: TableRow, Args: []string{"456", "normal fellow"}, LineNo: 4},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 5},
		&Token{Kind: StepKind, Value: "create users", LineNo: 6},
		&Token{Kind: TableHeader, Args: []string{"user-id", "description", "name"}, LineNo: 7},
		&Token{Kind: TableRow, Args: []string{"<id>", "<description>", "root"}, LineNo: 8},
		&Token{Kind: StepKind, Value: "create users", LineNo: 9},
		&Token{Kind: TableHeader, Args: []string{"user-id", "description", "name"}, LineNo: 10},
		&Token{Kind: TableRow, Args: []string{"1", "normal", "wheel"}, LineNo: 11},
	}

	concepts, _ := new(ConceptParser).Parse("#create users <id> \n * enter details from <id> \n *select \"finish\"")
	conceptsDictionary := new(ConceptDictionary)
	conceptsDictionary.Add(concepts, "file.cpt")
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptsDictionary)
	c.Assert(result.Ok, Equals, true)

	steps := spec.Scenarios[0].Steps
	c.Assert(len(steps), Equals, 2)
	c.Assert(steps[0].IsConcept, Equals, true)
	c.Assert(steps[1].IsConcept, Equals, true)
	c.Assert(steps[0].Value, Equals, "create users {}")
	c.Assert(len(steps[0].Args), Equals, 1)
	c.Assert(steps[0].Args[0].ArgType, Equals, TableArg)
	table := steps[0].Args[0].Table
	c.Assert(table.Get("user-id")[0].Value, Equals, "id")
	c.Assert(table.Get("user-id")[0].CellType, Equals, Dynamic)
	c.Assert(table.Get("description")[0].Value, Equals, "description")
	c.Assert(table.Get("description")[0].CellType, Equals, Dynamic)
	c.Assert(table.Get("name")[0].Value, Equals, "root")
	c.Assert(table.Get("name")[0].CellType, Equals, Static)
	c.Assert(len(steps[0].ConceptSteps), Equals, 2)
}

func (s *MySuite) TestPopulateFragmentsForSimpleStep(c *C) {
	step := &Step{Value: "This is a simple step"}

	step.PopulateFragments()

	c.Assert(len(step.Fragments), Equals, 1)
	fragment := step.Fragments[0]
	c.Assert(fragment.GetText(), Equals, "This is a simple step")
	c.Assert(fragment.GetFragmentType(), Equals, gauge_messages.Fragment_Text)
}

func (s *MySuite) TestGetArgForStep(c *C) {
	lookup := new(ArgLookup)
	lookup.addArgName("param1")
	lookup.addArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	step := &Step{Lookup: *lookup}

	c.Assert(step.getArg("param1").Value, Equals, "value1")
}

func (s *MySuite) TestGetArgForConceptStep(c *C) {
	lookup := new(ArgLookup)
	lookup.addArgName("param1")
	lookup.addArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	concept := &Step{Lookup: *lookup, IsConcept: true}
	stepLookup := new(ArgLookup)
	stepLookup.addArgName("param1")
	stepLookup.addArgValue("param1", &StepArg{Value: "param1", ArgType: Dynamic})
	step := &Step{Parent: concept, Lookup: *stepLookup}

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
	step := &Step{Value: "{} step with {} and {}, {}", Args: stepArgs}

	step.PopulateFragments()

	c.Assert(len(step.Fragments), Equals, 7)
	fragment1 := step.Fragments[0]
	c.Assert(fragment1.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment1.GetParameter().GetValue(), Equals, "first")
	c.Assert(fragment1.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Static)

	fragment2 := step.Fragments[1]
	c.Assert(fragment2.GetText(), Equals, " step with ")
	c.Assert(fragment2.GetFragmentType(), Equals, gauge_messages.Fragment_Text)

	fragment3 := step.Fragments[2]
	c.Assert(fragment3.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment3.GetParameter().GetValue(), Equals, "second")
	c.Assert(fragment3.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)

	fragment4 := step.Fragments[3]
	c.Assert(fragment4.GetText(), Equals, " and ")
	c.Assert(fragment4.GetFragmentType(), Equals, gauge_messages.Fragment_Text)

	fragment5 := step.Fragments[4]
	c.Assert(fragment5.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment5.GetParameter().GetValue(), Equals, "text from file")
	c.Assert(fragment5.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Special_String)
	c.Assert(fragment5.GetParameter().GetName(), Equals, "file:foo.txt")

	fragment6 := step.Fragments[5]
	c.Assert(fragment6.GetText(), Equals, ", ")
	c.Assert(fragment6.GetFragmentType(), Equals, gauge_messages.Fragment_Text)

	fragment7 := step.Fragments[6]
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
	originalStep := &Step{LineNo: 12,
		Value:          "foo {}",
		LineText:       "foo <bar>",
		Args:           argsInStep,
		IsConcept:      false,
		Fragments:      fragments,
		HasInlineTable: false}

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
	conceptSteps := []*Step{&Step{Value: "step 1"}}
	originalConcept := &Step{
		LineNo:         12,
		Value:          "foo {}",
		LineText:       "foo <bar>",
		Args:           argsInStep,
		IsConcept:      true,
		Lookup:         *argLookup,
		Fragments:      fragments,
		ConceptSteps:   conceptSteps,
		HasInlineTable: false}

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
	concepts, _ := new(ConceptParser).Parse(conceptText)

	dictionary := new(ConceptDictionary)
	dictionary.Add(concepts, "file.cpt")

	argsInStep := []*StepArg{&StepArg{Name: "arg1", Value: "value", ArgType: Static}}
	originalStep := &Step{
		LineNo:         12,
		Value:          "concept with {}",
		LineText:       "concept with \"value\"",
		Args:           argsInStep,
		IsConcept:      true,
		HasInlineTable: false}
	new(Specification).createConceptStep(dictionary.search("concept with {}").ConceptStep, originalStep)
	c.Assert(originalStep.IsConcept, Equals, true)
	c.Assert(len(originalStep.ConceptSteps), Equals, 1)
	c.Assert(originalStep.Args[0].Value, Equals, "value")

	c.Assert(originalStep.Lookup.getArg("foo").Value, Equals, "value")

	nestedConcept := originalStep.ConceptSteps[0]
	c.Assert(nestedConcept.IsConcept, Equals, true)
	c.Assert(len(nestedConcept.ConceptSteps), Equals, 1)

	c.Assert(nestedConcept.Args[0].ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.Args[0].Value, Equals, "foo")

	c.Assert(nestedConcept.ConceptSteps[0].Args[0].ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.ConceptSteps[0].Args[0].Value, Equals, "baz")

	c.Assert(nestedConcept.Lookup.getArg("baz").ArgType, Equals, Dynamic)
	c.Assert(nestedConcept.Lookup.getArg("baz").Value, Equals, "foo")
}

func (s *MySuite) TestRenameStep(c *C) {
	argsInStep := []*StepArg{&StepArg{Name: "arg1", Value: "value", ArgType: Static}, &StepArg{Name: "arg2", Value: "value1", ArgType: Static}}
	originalStep := &Step{
		LineNo:         12,
		Value:          "step with {}",
		Args:           argsInStep,
		IsConcept:      false,
		HasInlineTable: false}

	argsInStep = []*StepArg{&StepArg{Name: "arg2", Value: "value1", ArgType: Static}, &StepArg{Name: "arg1", Value: "value", ArgType: Static}}
	newStep := &Step{
		LineNo:         12,
		Value:          "step from {} {}",
		Args:           argsInStep,
		IsConcept:      false,
		HasInlineTable: false}
	orderMap := make(map[int]int)
	orderMap[0] = 1
	orderMap[1] = 0
	IsConcept := false
	isRefactored := originalStep.Rename(*originalStep, *newStep, false, orderMap, &IsConcept)

	c.Assert(isRefactored, Equals, true)
	c.Assert(originalStep.Value, Equals, "step from {} {}")
	c.Assert(originalStep.Args[0].Name, Equals, "arg2")
	c.Assert(originalStep.Args[1].Name, Equals, "arg1")
}

func (s *MySuite) TestGetLineTextForStep(c *C) {
	step := &Step{LineText: "foo"}

	c.Assert(step.getLineText(), Equals, "foo")
}

func (s *MySuite) TestGetLineTextForStepWithTable(c *C) {
	step := &Step{
		LineText:       "foo",
		HasInlineTable: true}

	c.Assert(step.getLineText(), Equals, "foo <table>")
}

func (s *MySuite) TestCreateInValidSpecialArgInStep(c *C) {
	tokens := []*Token{
		&Token{Kind: SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: TableHeader, Args: []string{"unknown:foo", "description"}, LineNo: 2},
		&Token{Kind: TableRow, Args: []string{"123", "Admin"}, LineNo: 3},
		&Token{Kind: ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: StepKind, Value: "Example {special} step", LineNo: 3, Args: []string{"unknown:foo"}},
	}
	spec, parseResults := new(SpecParser).CreateSpecification(tokens, new(ConceptDictionary))
	c.Assert(spec.Scenarios[0].Steps[0].Args[0].ArgType, Equals, Dynamic)
	c.Assert(len(parseResults.Warnings), Equals, 1)
	c.Assert(parseResults.Warnings[0].Message, Equals, "Could not resolve special param type <unknown:foo>. Treating it as dynamic param.")
}
