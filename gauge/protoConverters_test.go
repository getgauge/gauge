package main

import (. "launchpad.net/gocheck"
	"code.google.com/p/goprotobuf/proto"
)

func (s *MySuite) TestCopyingFragments(c *C) {
	text1 := &Fragment{FragmentType:Fragment_Text.Enum(), Text:proto.String("step with")}
	staticParam := &Fragment{FragmentType:Fragment_Parameter.Enum(), Parameter:&Parameter{ParameterType:Parameter_Static.Enum(), Value:proto.String("param0") }}
	text2 := &Fragment{FragmentType:Fragment_Text.Enum(), Text:proto.String("and")}
	dynamicParam := &Fragment{FragmentType:Fragment_Parameter.Enum(), Parameter: &Parameter{ParameterType:Parameter_Dynamic.Enum(), Value:proto.String("param1") }}
	fragments := []*Fragment{text1, staticParam, text2, dynamicParam}

	copiedFragments := makeFragmentsCopy(fragments)

	compareFragments(fragments, copiedFragments, c)

	fragments[1].Parameter.Value = proto.String("changedParam")
	fragments[2].Text = proto.String("changed text")

	c.Assert(copiedFragments[1].Parameter.GetValue(), Equals, "param0")
	c.Assert(copiedFragments[2].GetText(), Equals, "and")
}

func (s *MySuite) TestCopyingProtoTable(c *C) {
	headers := &ProtoTableRow{Cells: []string{"id", "name", "description"}}
	row1 := &ProtoTableRow{Cells: []string{"123", "abc", "first description"}}
	row2 := &ProtoTableRow{Cells: []string{"456", "def", "second description"}}

	table := &ProtoTable{Headers:headers, Rows:[]*ProtoTableRow{row1, row2}}
	copiedTable := makeTableCopy(table)

	compareTable(table, copiedTable, c)
	table.Headers.Cells[0] = "new id"
	table.Rows[0].Cells[0] = "789"
	table.Rows[1].Cells[1] = "xyz"

	c.Assert(copiedTable.Headers.Cells[0], Equals, "id")
	c.Assert(copiedTable.Rows[0].Cells[0], Equals, "123")
	c.Assert(copiedTable.Rows[1].Cells[1], Equals, "def")

}

func compareFragments(fragmentList1 []*Fragment, fragmentList2 []*Fragment, c *C) {
	c.Assert(len(fragmentList1), Equals, len(fragmentList2))
	for i, _ := range fragmentList1 {
		compareFragment(fragmentList1[i], fragmentList2[i], c)
	}
}

func compareFragment(fragment1 *Fragment, fragment2 *Fragment, c *C) {
	c.Assert(fragment1.GetFragmentType(), Equals, fragment2.GetFragmentType())
	c.Assert(fragment1.GetText(), Equals, fragment2.GetText())
	parameter1 := fragment1.GetParameter()
	parameter2 := fragment2.GetParameter()
	compareParameter(parameter1, parameter2, c)
}

func compareParameter(parameter1 *Parameter, parameter2 *Parameter, c *C) {
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


func compareTable(table1 *ProtoTable, table2 *ProtoTable, c *C) {
	compareTableRow(table1.GetHeaders(), table2.GetHeaders(), c)
	c.Assert(len(table1.GetRows()), Equals, len(table2.GetRows()))
	for i,row := range table1.GetRows() {
		compareTableRow(row, table2.GetRows()[i], c)
	}
}

func compareTableRow(row1 *ProtoTableRow, row2 *ProtoTableRow, c *C) {
	c.Assert(row1.GetCells(), DeepEquals, row2.GetCells() )
}
