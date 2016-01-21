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
	"path/filepath"

	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestThrowsErrorForMultipleSpecHeading(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 3},
		&Token{Kind: gauge.SpecKind, Value: "Another Heading", LineNo: 4},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseError.Message, Equals, "Parse error: Multiple spec headings found in same file")
	c.Assert(result.ParseError.LineNo, Equals, 4)
}

func (s *MySuite) TestThrowsErrorForScenarioWithoutSpecHeading(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 1},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 2},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseError.Message, Equals, "Parse error: Scenario should be defined after the spec heading")
	c.Assert(result.ParseError.LineNo, Equals, 1)
}

func (s *MySuite) TestThrowsErrorForDuplicateScenariosWithinTheSameSpec(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseError.Message, Equals, "Parse error: Duplicate scenario definition 'Scenario Heading' found in the same specification")
	c.Assert(result.ParseError.LineNo, Equals, 4)
}

func (s *MySuite) TestSpecWithHeadingAndSimpleSteps(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 3},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(len(spec.Items), Equals, 1)
	c.Assert(spec.Items[0], Equals, spec.Scenarios[0])
	scenarioItems := (spec.Items[0]).(*gauge.Scenario).Items
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
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.CommentKind, Value: "A comment with some text and **bold** characters", LineNo: 2},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&Token{Kind: gauge.CommentKind, Value: "Another comment", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 5},
		&Token{Kind: gauge.CommentKind, Value: "Third comment", LineNo: 6},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(len(spec.Items), Equals, 2)
	c.Assert(spec.Items[0], Equals, spec.Comments[0])
	c.Assert(spec.Items[1], Equals, spec.Scenarios[0])

	scenarioItems := (spec.Items[1]).(*gauge.Scenario).Items
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
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"id"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"1"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "enter {static} with {dynamic}", LineNo: 5, Args: []string{"user \\n foo", "id"}},
		&Token{Kind: gauge.StepKind, Value: "sample \\{static\\}", LineNo: 6},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(result.Ok, Equals, true)
	step := spec.Scenarios[0].Steps[0]
	c.Assert(step.Value, Equals, "enter {} with {}")
	c.Assert(step.LineNo, Equals, 5)
	c.Assert(len(step.Args), Equals, 2)
	c.Assert(step.Args[0].Value, Equals, "user \\n foo")
	c.Assert(step.Args[0].ArgType, Equals, gauge.Static)
	c.Assert(step.Args[1].Value, Equals, "id")
	c.Assert(step.Args[1].ArgType, Equals, gauge.Dynamic)
	c.Assert(step.Args[1].Name, Equals, "id")

	escapedStep := spec.Scenarios[0].Steps[1]
	c.Assert(escapedStep.Value, Equals, "sample \\{static\\}")
	c.Assert(len(escapedStep.Args), Equals, 0)
}

func (s *MySuite) TestStepsWithKeywords(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "sample {static} and {dynamic}", LineNo: 3, Args: []string{"name"}},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result, NotNil)
	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}'")
	c.Assert(result.ParseError.LineNo, Equals, 3)
}

func (s *MySuite) TestContextWithKeywords(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.StepKind, Value: "sample {static} and {dynamic}", LineNo: 3, Args: []string{"name"}},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result, NotNil)
	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}'")
	c.Assert(result.ParseError.LineNo, Equals, 3)
}

func (s *MySuite) TestSpecWithDataTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading"},
		&Token{Kind: gauge.CommentKind, Value: "Comment before data table"},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
		&Token{Kind: gauge.CommentKind, Value: "Comment before data table"},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(len(spec.Items), Equals, 3)
	c.Assert(spec.Items[0], Equals, spec.Comments[0])
	c.Assert(spec.Items[1], DeepEquals, &spec.DataTable)
	c.Assert(spec.Items[2], Equals, spec.Comments[1])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.DataTable, NotNil)
	c.Assert(len(spec.DataTable.Table.Get("id")), Equals, 2)
	c.Assert(len(spec.DataTable.Table.Get("name")), Equals, 2)
	c.Assert(spec.DataTable.Table.Get("id")[0].Value, Equals, "1")
	c.Assert(spec.DataTable.Table.Get("id")[0].CellType, Equals, gauge.Static)
	c.Assert(spec.DataTable.Table.Get("id")[1].Value, Equals, "2")
	c.Assert(spec.DataTable.Table.Get("id")[1].CellType, Equals, gauge.Static)
	c.Assert(spec.DataTable.Table.Get("name")[0].Value, Equals, "foo")
	c.Assert(spec.DataTable.Table.Get("name")[0].CellType, Equals, gauge.Static)
	c.Assert(spec.DataTable.Table.Get("name")[1].Value, Equals, "bar")
	c.Assert(spec.DataTable.Table.Get("name")[1].CellType, Equals, gauge.Static)
}

func (s *MySuite) TestStepWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Step with inline table", LineNo: 3},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, true)
	step := spec.Scenarios[0].Steps[0]

	c.Assert(step.Args[0].ArgType, Equals, gauge.TableArg)
	inlineTable := step.Args[0].Table
	c.Assert(inlineTable, NotNil)

	c.Assert(step.Value, Equals, "Step with inline table {}")
	c.Assert(step.HasInlineTable, Equals, true)
	c.Assert(len(inlineTable.Get("id")), Equals, 2)
	c.Assert(len(inlineTable.Get("name")), Equals, 2)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[0].Value, Equals, "foo")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "bar")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, gauge.Static)
}

func (s *MySuite) TestStepWithInlineTableWithDynamicParam(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"type1", "type2"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "2"}},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Step with inline table", LineNo: 3},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "<type1>"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "<type2>"}},
		&Token{Kind: gauge.TableRow, Args: []string{"<2>", "<type3>"}},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, true)
	step := spec.Scenarios[0].Steps[0]

	c.Assert(step.Args[0].ArgType, Equals, gauge.TableArg)
	inlineTable := step.Args[0].Table
	c.Assert(inlineTable, NotNil)

	c.Assert(step.Value, Equals, "Step with inline table {}")
	c.Assert(len(inlineTable.Get("id")), Equals, 3)
	c.Assert(len(inlineTable.Get("name")), Equals, 3)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[2].Value, Equals, "<2>")
	c.Assert(inlineTable.Get("id")[2].CellType, Equals, gauge.Static)

	c.Assert(inlineTable.Get("name")[0].Value, Equals, "type1")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, gauge.Dynamic)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "type2")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, gauge.Dynamic)
	c.Assert(inlineTable.Get("name")[2].Value, Equals, "<type3>")
	c.Assert(inlineTable.Get("name")[2].CellType, Equals, gauge.Static)
}

func (s *MySuite) TestStepWithInlineTableWithUnResolvableDynamicParam(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"type1", "type2"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "2"}},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Step with inline table", LineNo: 3},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "<invalid>"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "<type2>"}},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.Scenarios[0].Steps[0].Args[0].Table.Get("id")[0].Value, Equals, "1")
	c.Assert(spec.Scenarios[0].Steps[0].Args[0].Table.Get("name")[0].Value, Equals, "<invalid>")
	c.Assert(result.Warnings[0].Message, Equals, "Dynamic param <invalid> could not be resolved, Treating it as static param")
}

func (s *MySuite) TestContextWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading"},
		&Token{Kind: gauge.StepKind, Value: "Context with inline table"},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
		&Token{Kind: gauge.TableRow, Args: []string{"3", "not a <dynamic>"}},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading"},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(len(spec.Items), Equals, 2)
	c.Assert(spec.Items[0], DeepEquals, spec.Contexts[0])
	c.Assert(spec.Items[1], Equals, spec.Scenarios[0])

	c.Assert(result.Ok, Equals, true)
	context := spec.Contexts[0]

	c.Assert(context.Args[0].ArgType, Equals, gauge.TableArg)
	inlineTable := context.Args[0].Table

	c.Assert(inlineTable, NotNil)
	c.Assert(context.Value, Equals, "Context with inline table {}")
	c.Assert(len(inlineTable.Get("id")), Equals, 3)
	c.Assert(len(inlineTable.Get("name")), Equals, 3)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[2].Value, Equals, "3")
	c.Assert(inlineTable.Get("id")[2].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[0].Value, Equals, "foo")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "bar")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[2].Value, Equals, "not a <dynamic>")
	c.Assert(inlineTable.Get("name")[2].CellType, Equals, gauge.Static)
}

func (s *MySuite) TestErrorWhenDataTableHasOnlyHeader(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading"},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading"},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Data table should have at least 1 data row")
	c.Assert(result.ParseError.LineNo, Equals, 3)
}

func (s *MySuite) TestWarningWhenParsingMultipleDataTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading"},
		&Token{Kind: gauge.CommentKind, Value: "Comment before data table"},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
		&Token{Kind: gauge.CommentKind, Value: "Comment before data table"},
		&Token{Kind: gauge.TableHeader, Args: []string{"phone"}, LineNo: 7},
		&Token{Kind: gauge.TableRow, Args: []string{"1"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2"}},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.Warnings), Equals, 1)
	c.Assert(result.Warnings[0].String(), Equals, "line no: 7, Multiple data table present, ignoring table")

}

func (s *MySuite) TestWarningWhenParsingTableOccursWithoutStep(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}, LineNo: 3},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}, LineNo: 4},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}, LineNo: 5},
		&Token{Kind: gauge.StepKind, Value: "Step", LineNo: 6},
		&Token{Kind: gauge.CommentKind, Value: "comment in between", LineNo: 7},
		&Token{Kind: gauge.TableHeader, Args: []string{"phone"}, LineNo: 8},
		&Token{Kind: gauge.TableRow, Args: []string{"1"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2"}},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.Warnings), Equals, 2)
	c.Assert(result.Warnings[0].String(), Equals, "line no: 3, Table not associated with a step, ignoring table")
	c.Assert(result.Warnings[1].String(), Equals, "line no: 8, Table not associated with a step, ignoring table")

}

func (s *MySuite) TestAddSpecTags(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Tags.Values), Equals, 2)
	c.Assert(spec.Tags.Values[0], Equals, "tag1")
	c.Assert(spec.Tags.Values[1], Equals, "tag2")
}

func (s *MySuite) TestAddSpecTagsAndScenarioTags(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&Token{Kind: gauge.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 2},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

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
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Step with a {dynamic}", Args: []string{"foo"}, LineNo: 3, LineText: "*Step with a <foo>"},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Dynamic parameter <foo> could not be resolved")
	c.Assert(result.ParseError.LineNo, Equals, 3)

}

func (s *MySuite) TestErrorOnAddingDynamicParamterWithoutDataTableHeaderValue(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"id, name"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"123, hello"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "Step with a {dynamic}", Args: []string{"foo"}, LineNo: 5, LineText: "*Step with a <foo>"},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseError.Message, Equals, "Dynamic parameter <foo> could not be resolved")
	c.Assert(result.ParseError.LineNo, Equals, 5)

}

func (s *MySuite) TestCreateStepFromSimpleConcept(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "test concept step 1", LineNo: 3},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Scenarios[0].Steps), Equals, 1)
	specConceptStep := spec.Scenarios[0].Steps[0]
	c.Assert(specConceptStep.IsConcept, Equals, true)
	assertStepEqual(c, &gauge.Step{LineNo: 2, Value: "step 1", LineText: "step 1"}, specConceptStep.ConceptSteps[0])
}

func (s *MySuite) TestCreateStepFromConceptWithParameters(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "assign id {static} and name {static}", Args: []string{"foo", "foo1"}, LineNo: 3},
		&Token{Kind: gauge.StepKind, Value: "assign id {static} and name {static}", Args: []string{"bar", "bar1"}, LineNo: 4},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)

	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Scenarios[0].Steps), Equals, 2)

	firstConceptStep := spec.Scenarios[0].Steps[0]
	c.Assert(firstConceptStep.IsConcept, Equals, true)
	c.Assert(firstConceptStep.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(firstConceptStep.ConceptSteps[0].Args[0].Value, Equals, "userid")
	c.Assert(firstConceptStep.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(firstConceptStep.ConceptSteps[1].Args[0].Value, Equals, "username")
	c.Assert(firstConceptStep.GetArg("username").Value, Equals, "foo1")
	c.Assert(firstConceptStep.GetArg("userid").Value, Equals, "foo")

	secondConceptStep := spec.Scenarios[0].Steps[1]
	c.Assert(secondConceptStep.IsConcept, Equals, true)
	c.Assert(secondConceptStep.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(secondConceptStep.ConceptSteps[0].Args[0].Value, Equals, "userid")
	c.Assert(secondConceptStep.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(secondConceptStep.ConceptSteps[1].Args[0].Value, Equals, "username")
	c.Assert(secondConceptStep.GetArg("username").Value, Equals, "bar1")
	c.Assert(secondConceptStep.GetArg("userid").Value, Equals, "bar")

}

func (s *MySuite) TestCreateStepFromConceptWithDynamicParameters(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "description"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"123", "Admin fellow"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "assign id {dynamic} and name {dynamic}", Args: []string{"id", "description"}, LineNo: 5},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Items), Equals, 2)
	c.Assert(spec.Items[0], DeepEquals, &spec.DataTable)
	c.Assert(spec.Items[1], Equals, spec.Scenarios[0])

	scenarioItems := (spec.Items[1]).(*gauge.Scenario).Items
	c.Assert(scenarioItems[0], Equals, spec.Scenarios[0].Steps[0])

	c.Assert(len(spec.Scenarios[0].Steps), Equals, 1)

	firstConcept := spec.Scenarios[0].Steps[0]
	c.Assert(firstConcept.IsConcept, Equals, true)
	c.Assert(firstConcept.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(firstConcept.ConceptSteps[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(firstConcept.ConceptSteps[0].Args[0].Value, Equals, "userid")
	c.Assert(firstConcept.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(firstConcept.ConceptSteps[1].Args[0].Value, Equals, "username")
	c.Assert(firstConcept.ConceptSteps[1].Args[0].ArgType, Equals, gauge.Dynamic)

	arg1 := firstConcept.Lookup.GetArg("userid")
	c.Assert(arg1.Value, Equals, "id")
	c.Assert(arg1.ArgType, Equals, gauge.Dynamic)

	arg2 := firstConcept.Lookup.GetArg("username")
	c.Assert(arg2.Value, Equals, "description")
	c.Assert(arg2.ArgType, Equals, gauge.Dynamic)
}

func (s *MySuite) TestCreateStepFromConceptWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "assign id {static} and name", Args: []string{"sdf"}, LineNo: 3},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "description"}, LineNo: 4},
		&Token{Kind: gauge.TableRow, Args: []string{"123", "Admin"}, LineNo: 5},
		&Token{Kind: gauge.TableRow, Args: []string{"456", "normal fellow"}, LineNo: 6},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	steps := spec.Scenarios[0].Steps
	c.Assert(len(steps), Equals, 1)
	c.Assert(steps[0].IsConcept, Equals, true)
	c.Assert(steps[0].Value, Equals, "assign id {} and name {}")
	c.Assert(len(steps[0].Args), Equals, 2)
	c.Assert(steps[0].Args[1].ArgType, Equals, gauge.TableArg)
	c.Assert(len(steps[0].ConceptSteps), Equals, 2)
}

func (s *MySuite) TestCreateStepFromConceptWithInlineTableHavingDynamicParam(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "description"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"123", "Admin"}, LineNo: 3},
		&Token{Kind: gauge.TableRow, Args: []string{"456", "normal fellow"}, LineNo: 4},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 5},
		&Token{Kind: gauge.StepKind, Value: "assign id {static} and name", Args: []string{"sdf"}, LineNo: 6},
		&Token{Kind: gauge.TableHeader, Args: []string{"user-id", "description", "name"}, LineNo: 7},
		&Token{Kind: gauge.TableRow, Args: []string{"<id>", "<description>", "root"}, LineNo: 8},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	steps := spec.Scenarios[0].Steps
	c.Assert(len(steps), Equals, 1)
	c.Assert(steps[0].IsConcept, Equals, true)
	c.Assert(steps[0].Value, Equals, "assign id {} and name {}")
	c.Assert(len(steps[0].Args), Equals, 2)
	c.Assert(steps[0].Args[1].ArgType, Equals, gauge.TableArg)
	table := steps[0].Args[1].Table
	c.Assert(table.Get("user-id")[0].Value, Equals, "id")
	c.Assert(table.Get("user-id")[0].CellType, Equals, gauge.Dynamic)
	c.Assert(table.Get("description")[0].Value, Equals, "description")
	c.Assert(table.Get("description")[0].CellType, Equals, gauge.Dynamic)
	c.Assert(table.Get("name")[0].Value, Equals, "root")
	c.Assert(table.Get("name")[0].CellType, Equals, gauge.Static)
	c.Assert(len(steps[0].ConceptSteps), Equals, 2)
}

func (s *MySuite) TestCreateConceptStep(c *C) {
	dictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "param_nested_concept.cpt"))
	AddConcepts(path, dictionary)

	argsInStep := []*gauge.StepArg{&gauge.StepArg{Name: "bar", Value: "first name", ArgType: gauge.Static}, &gauge.StepArg{Name: "far", Value: "last name", ArgType: gauge.Static}}
	originalStep := &gauge.Step{
		LineNo:         12,
		Value:          "create user {} {}",
		LineText:       "create user \"first name\" \"last name\"",
		Args:           argsInStep,
		IsConcept:      false,
		HasInlineTable: false}

	createConceptStep(new(gauge.Specification), dictionary.Search("create user {} {}").ConceptStep, originalStep)

	c.Assert(originalStep.IsConcept, Equals, true)
	c.Assert(len(originalStep.ConceptSteps), Equals, 1)
	c.Assert(originalStep.Args[0].Value, Equals, "first name")
	c.Assert(originalStep.Lookup.GetArg("bar").Value, Equals, "first name")
	c.Assert(originalStep.Args[1].Value, Equals, "last name")
	c.Assert(originalStep.Lookup.GetArg("far").Value, Equals, "last name")

	nestedConcept := originalStep.ConceptSteps[0]
	c.Assert(nestedConcept.IsConcept, Equals, true)
	c.Assert(len(nestedConcept.ConceptSteps), Equals, 1)

	c.Assert(nestedConcept.Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(nestedConcept.Args[0].Name, Equals, "bar")

	c.Assert(nestedConcept.Args[1].ArgType, Equals, gauge.Dynamic)
	c.Assert(nestedConcept.Args[1].Name, Equals, "far")

	c.Assert(nestedConcept.ConceptSteps[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(nestedConcept.ConceptSteps[0].Args[0].Name, Equals, "baz")

	c.Assert(nestedConcept.Lookup.GetArg("baz").ArgType, Equals, gauge.Dynamic)
	c.Assert(nestedConcept.Lookup.GetArg("baz").Value, Equals, "bar")
}

func (s *MySuite) TestCreateInValidSpecialArgInStep(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"unknown:foo", "description"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"123", "Admin"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Example {special} step", LineNo: 3, Args: []string{"unknown:foo"}},
	}
	spec, parseResults := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(spec.Scenarios[0].Steps[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(len(parseResults.Warnings), Equals, 1)
	c.Assert(parseResults.Warnings[0].Message, Equals, "Could not resolve special param type <unknown:foo>. Treating it as dynamic param.")
}

func (s *MySuite) TestTearDownSteps(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.CommentKind, Value: "A comment with some text and **bold** characters", LineNo: 2},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&Token{Kind: gauge.CommentKind, Value: "Another comment", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 5},
		&Token{Kind: gauge.CommentKind, Value: "Third comment", LineNo: 6},
		&Token{Kind: gauge.TearDownKind, Value: "____", LineNo: 7},
		&Token{Kind: gauge.StepKind, Value: "Example step1", LineNo: 8},
		&Token{Kind: gauge.CommentKind, Value: "Fourth comment", LineNo: 9},
		&Token{Kind: gauge.StepKind, Value: "Example step2", LineNo: 10},
	}

	spec, _ := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(len(spec.TearDownSteps), Equals, 2)
	c.Assert(spec.TearDownSteps[0].Value, Equals, "Example step1")
	c.Assert(spec.TearDownSteps[0].LineNo, Equals, 8)
	c.Assert(spec.TearDownSteps[1].Value, Equals, "Example step2")
	c.Assert(spec.TearDownSteps[1].LineNo, Equals, 10)
}
