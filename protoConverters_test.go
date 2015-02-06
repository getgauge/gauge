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

import (
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestCopyingFragments(c *C) {
	text1 := &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("step with")}
	staticParam := &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter.Enum(), Parameter: &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Static.Enum(), Value: proto.String("param0")}}
	text2 := &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text.Enum(), Text: proto.String("and")}
	dynamicParam := &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter.Enum(), Parameter: &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Dynamic.Enum(), Value: proto.String("param1")}}
	fragments := []*gauge_messages.Fragment{text1, staticParam, text2, dynamicParam}

	copiedFragments := makeFragmentsCopy(fragments)

	compareFragments(fragments, copiedFragments, c)

	fragments[1].Parameter.Value = proto.String("changedParam")
	fragments[2].Text = proto.String("changed text")

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
	stepValue := &stepValue{[]string{"param1"}, "foo with {}", "foo with <param>"}
	protoStepValue := convertToProtoStepValue(stepValue)

	c.Assert(protoStepValue.GetStepValue(), Equals, stepValue.stepValue)
	c.Assert(protoStepValue.GetParameterizedStepValue(), Equals, stepValue.parameterizedStepValue)
	c.Assert(protoStepValue.GetParameters(), DeepEquals, stepValue.args)
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
