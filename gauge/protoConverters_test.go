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

package gauge

import (
	"github.com/getgauge/gauge/gauge_messages"
	. "github.com/go-check/check"
)

func (s *MySuite) TestCopyingFragments(c *C) {
	text1 := &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text, Text: "step with"}
	staticParam := &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Static, Value: "param0"}}
	text2 := &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text, Text: "and"}
	dynamicParam := &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Dynamic, Value: "param1"}}
	fragments := []*gauge_messages.Fragment{text1, staticParam, text2, dynamicParam}

	copiedFragments := makeFragmentsCopy(fragments)

	compareFragments(fragments, copiedFragments, c)

	fragments[1].Parameter.Value = "changedParam"
	fragments[2].Text = "changed text"

	c.Assert(copiedFragments[1].Parameter.GetValue(), Equals, "param0")
	c.Assert(copiedFragments[2].GetText(), Equals, "and")
}

func (s *MySuite) TestCopyingProtoTable(c *C) {
	headers := &gauge_messages.ProtoTableRow{Cells: []string{"id", "name", "description"}}
	row1 := &gauge_messages.ProtoTableRow{Cells: []string{"123", "abc", "first description"}}
	row2 := &gauge_messages.ProtoTableRow{Cells: []string{"456", "def", "second description"}}

	table := &gauge_messages.ProtoTable{Headers: headers, Rows: []*gauge_messages.ProtoTableRow{row1, row2}}
	copiedTable := makeTableCopy(table)

	compareTable(table, copiedTable, c)
	table.Headers.Cells[0] = "new id"
	table.Rows[0].Cells[0] = "789"
	table.Rows[1].Cells[1] = "xyz"

	c.Assert(copiedTable.Headers.Cells[0], Equals, "id")
	c.Assert(copiedTable.Rows[0].Cells[0], Equals, "123")
	c.Assert(copiedTable.Rows[1].Cells[1], Equals, "def")

}

func (s *MySuite) TestCopyingStepValue(c *C) {
	stepValue := &StepValue{[]string{"param1"}, "foo with {}", "foo with <param>"}
	protoStepValue := ConvertToProtoStepValue(stepValue)

	c.Assert(protoStepValue.GetStepValue(), Equals, stepValue.StepValue)
	c.Assert(protoStepValue.GetParameterizedStepValue(), Equals, stepValue.ParameterizedStepValue)
	c.Assert(protoStepValue.GetParameters(), DeepEquals, stepValue.Args)
}

func (s *MySuite) TestNewProtoScenario(c *C) {
	sceHeading := "sce heading"
	sce := &Scenario{Heading: &Heading{Value: sceHeading}, Span: &Span{Start: 1, End: 4}}

	protoSce := NewProtoScenario(sce)

	c.Assert(protoSce.GetScenarioHeading(), Equals, sceHeading)
	c.Assert(protoSce.GetExecutionStatus(), Equals, gauge_messages.ExecutionStatus_NOTEXECUTED)
	c.Assert(protoSce.Span.Start, Equals, int64(1))
	c.Assert(protoSce.Span.End, Equals, int64(4))
}

func (s *MySuite) TestConvertToProtoSpecWithDataTable(c *C) {
	spec := &Specification{
		Heading: &Heading{
			Value: "Spec Heading",
		},
		FileName:  "example.spec",
		DataTable: DataTable{Table: Table{headerIndexMap: make(map[string]int)}},
	}
	protoSpec := ConvertToProtoSpec(spec)

	c.Assert(protoSpec.GetIsTableDriven(), Equals, true)
}

func (s *MySuite) TestConvertToProtoSpecWithoutDataTable(c *C) {
	spec := &Specification{
		Heading: &Heading{
			Value: "Spec Heading",
		},
		FileName: "example.spec",
	}
	protoSpec := ConvertToProtoSpec(spec)

	c.Assert(protoSpec.GetIsTableDriven(), Equals, false)
}

func (s *MySuite) TestConvertToProtoStep(c *C) {
	step := &Step{
		LineText: "line text",
		Value:    "value",
	}
	actual := convertToProtoStep(step)

	expected := &gauge_messages.ProtoStep{ActualText: step.LineText, ParsedText: step.Value, Fragments: []*gauge_messages.Fragment{}}
	c.Assert(actual, DeepEquals, expected)
}

func (s *MySuite) TestConvertToProtoConcept(c *C) {
	step := &Step{
		LineText:  "line text",
		Value:     "value",
		IsConcept: true,
		ConceptSteps: []*Step{
			{LineText: "line text1", Value: "value1", ConceptSteps: []*Step{}},
			{LineText: "line text2", Value: "value2", IsConcept: true,
				ConceptSteps: []*Step{{LineText: "line text3", Value: "value3", ConceptSteps: []*Step{}}},
			},
		},
	}
	actual := convertToProtoConcept(step)

	expected := &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Concept,
		Concept: &gauge_messages.ProtoConcept{
			ConceptStep: newProtoStep("line text", "value"),
			Steps: []*gauge_messages.ProtoItem{
				newStepItem("line text1", "value1"),
				{
					ItemType: gauge_messages.ProtoItem_Concept,
					Concept: &gauge_messages.ProtoConcept{
						ConceptStep: newProtoStep("line text2", "value2"),
						Steps:       []*gauge_messages.ProtoItem{newStepItem("line text3", "value3")},
					},
				},
			},
		},
	}

	c.Assert(actual, DeepEquals, expected)
}

func newStepItem(lineText, value string) *gauge_messages.ProtoItem {
	return &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Step,
		Step:     newProtoStep(lineText, value),
	}

}

func newProtoStep(lineText, value string) *gauge_messages.ProtoStep {
	return &gauge_messages.ProtoStep{
		ActualText: lineText,
		ParsedText: value,
		Fragments:  []*gauge_messages.Fragment{},
	}
}

func compareFragments(fragmentList1 []*gauge_messages.Fragment, fragmentList2 []*gauge_messages.Fragment, c *C) {
	c.Assert(len(fragmentList1), Equals, len(fragmentList2))
	for i, _ := range fragmentList1 {
		compareFragment(fragmentList1[i], fragmentList2[i], c)
	}
}

func compareFragment(fragment1 *gauge_messages.Fragment, fragment2 *gauge_messages.Fragment, c *C) {
	c.Assert(fragment1.GetFragmentType(), Equals, fragment2.GetFragmentType())
	c.Assert(fragment1.GetText(), Equals, fragment2.GetText())
	parameter1 := fragment1.GetParameter()
	parameter2 := fragment2.GetParameter()
	compareParameter(parameter1, parameter2, c)
}

func compareParameter(parameter1 *gauge_messages.Parameter, parameter2 *gauge_messages.Parameter, c *C) {
	if parameter1 != nil && parameter2 != nil {
		c.Assert(parameter1.GetParameterType(), Equals, parameter2.GetParameterType())
		c.Assert(parameter1.GetName(), Equals, parameter2.GetName())
		c.Assert(parameter1.GetValue(), Equals, parameter2.GetValue())
		compareTable(parameter1.GetTable(), parameter2.GetTable(), c)

	} else if (parameter1 == nil && parameter2 != nil) || (parameter1 != nil && parameter2 == nil) {
		c.Log("One parameter was nil and the other wasn't")
		c.Fail()
	}
}

func compareTable(table1 *gauge_messages.ProtoTable, table2 *gauge_messages.ProtoTable, c *C) {
	compareTableRow(table1.GetHeaders(), table2.GetHeaders(), c)
	c.Assert(len(table1.GetRows()), Equals, len(table2.GetRows()))
	for i, row := range table1.GetRows() {
		compareTableRow(row, table2.GetRows()[i], c)
	}
}

func compareTableRow(row1 *gauge_messages.ProtoTableRow, row2 *gauge_messages.ProtoTableRow, c *C) {
	c.Assert(row1.GetCells(), DeepEquals, row2.GetCells())
}
