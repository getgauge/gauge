/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestLookupaddArg(c *C) {
	lookup := new(ArgLookup)
	lookup.AddArgName("param1")
	lookup.AddArgName("param2")

	c.Assert(lookup.ParamIndexMap["param1"], Equals, 0)
	c.Assert(lookup.ParamIndexMap["param2"], Equals, 1)
	c.Assert(len(lookup.paramValue), Equals, 2)
	c.Assert(lookup.paramValue[0].name, Equals, "param1")
	c.Assert(lookup.paramValue[1].name, Equals, "param2")

}

func (s *MySuite) TestLookupContainsArg(c *C) {
	lookup := new(ArgLookup)
	lookup.AddArgName("param1")
	lookup.AddArgName("param2")

	c.Assert(lookup.ContainsArg("param1"), Equals, true)
	c.Assert(lookup.ContainsArg("param2"), Equals, true)
	c.Assert(lookup.ContainsArg("param3"), Equals, false)
}

func (s *MySuite) TestAddArgValue(c *C) {
	lookup := new(ArgLookup)
	lookup.AddArgName("param1")
	err := lookup.AddArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	c.Assert(err, IsNil)
	lookup.AddArgName("param2")
	err = lookup.AddArgValue("param2", &StepArg{Value: "value2", ArgType: Dynamic})
	c.Assert(err, IsNil)
	stepArg, err := lookup.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "value1")
	stepArg, err = lookup.GetArg("param2")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "value2")
	c.Assert(stepArg.Name, Equals, "param2")
}

func (s *MySuite) TestErrorForInvalidArg(c *C) {
	lookup := new(ArgLookup)
	err := lookup.AddArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	c.Assert(err.Error(), Equals, "Accessing an invalid parameter (param1)")
	_, err = lookup.GetArg("param1")
	c.Assert(err.Error(), Equals, "Accessing an invalid parameter (param1)")
}

func (s *MySuite) TestGetLookupCopy(c *C) {
	originalLookup := new(ArgLookup)
	originalLookup.AddArgName("param1")
	err := originalLookup.AddArgValue("param1", &StepArg{Value: "oldValue", ArgType: Dynamic})
	c.Assert(err, IsNil)

	copiedLookup, _ := originalLookup.GetCopy()
	err = copiedLookup.AddArgValue("param1", &StepArg{Value: "new value", ArgType: Static})
	c.Assert(err, IsNil)

	stepArg, err := copiedLookup.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "new value")
	c.Assert(stepArg.Name, Equals, "param1")
	stepArg, err = originalLookup.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "oldValue")
}

func (s *MySuite) TestGetLookupFromTableRow(c *C) {
	dataTable := new(Table)
	dataTable.AddHeaders([]string{"id", "name"})
	dataTable.AddRowValues(dataTable.CreateTableCells([]string{"1", "admin"}))
	dataTable.AddRowValues(dataTable.CreateTableCells([]string{"2", "root"}))

	emptyLookup := new(ArgLookup)
	err := emptyLookup.ReadDataTableRow(new(Table), 0)
	c.Assert(err, IsNil)
	c.Assert(emptyLookup.ParamIndexMap, IsNil)

	lookup1 := new(ArgLookup)
	err = lookup1.ReadDataTableRow(dataTable, 0)
	c.Assert(err, IsNil)
	idArg1, err := lookup1.GetArg("id")
	c.Assert(err, IsNil)
	nameArg1, err := lookup1.GetArg("name")
	c.Assert(err, IsNil)
	c.Assert(idArg1.Value, Equals, "1")
	c.Assert(idArg1.ArgType, Equals, Static)
	c.Assert(nameArg1.Value, Equals, "admin")
	c.Assert(nameArg1.ArgType, Equals, Static)

	lookup2 := new(ArgLookup)
	err = lookup2.ReadDataTableRow(dataTable, 1)
	c.Assert(err, IsNil)
	idArg2, err := lookup2.GetArg("id")
	c.Assert(err, IsNil)
	nameArg2, err := lookup2.GetArg("name")
	c.Assert(err, IsNil)
	c.Assert(idArg2.Value, Equals, "2")
	c.Assert(idArg2.ArgType, Equals, Static)
	c.Assert(nameArg2.Value, Equals, "root")
	c.Assert(nameArg2.ArgType, Equals, Static)
}

func (s *MySuite) TestGetLookupFromTables(c *C) {
	t1 := new(Table)
	t1.AddHeaders([]string{"id1", "name1"})
	t1.AddRowValues(t1.CreateTableCells([]string{"1", "admin"}))
	t1.AddRowValues(t1.CreateTableCells([]string{"2", "root"}))

	t2 := new(Table)
	t2.AddHeaders([]string{"id2", "name2"})
	t2.AddRowValues(t2.CreateTableCells([]string{"1", "admin"}))
	t2.AddRowValues(t2.CreateTableCells([]string{"2", "root"}))

	l := new(ArgLookup).FromDataTables(t1, t2)

	c.Assert(l.ContainsArg("id1"), Equals, true)
	c.Assert(l.ContainsArg("name1"), Equals, true)
	c.Assert(l.ContainsArg("id2"), Equals, true)
	c.Assert(l.ContainsArg("name2"), Equals, true)
}

func (s *MySuite) TestAddMultilineStringArg(c *C) {
	lookup := new(ArgLookup)
	lookup.AddArgName("param1")

	mlValue := `line1
line2
line3`
	err := lookup.AddArgValue("param1", &StepArg{Value: mlValue, ArgType: MultilineString})
	c.Assert(err, IsNil)

	stepArg, err := lookup.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.ArgType, Equals, MultilineString)
	c.Assert(stepArg.Value, Equals, mlValue)
}

func (s *MySuite) TestMultilineStringArgValueMethod(c *C) {
	mlValue := `first line
second line`
	stepArg := &StepArg{
		Name:    "mlParam",
		Value:   mlValue,
		ArgType: MultilineString,
	}
	c.Assert(stepArg.ArgValue(), Equals, mlValue)
}

func (s *MySuite) TestMultilineStringInCopy(c *C) {
	originalLookup := new(ArgLookup)
	mlValue := `a
b
c`
	originalLookup.AddArgName("param1")
	err := originalLookup.AddArgValue("param1", &StepArg{Value: mlValue, ArgType: MultilineString})
	c.Assert(err, IsNil)

	copiedLookup, err := originalLookup.GetCopy()
	c.Assert(err, IsNil)

	stepArg, err := copiedLookup.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.ArgType, Equals, MultilineString)
	c.Assert(stepArg.Value, Equals, mlValue)

	// Ensure modifying copy does not affect original
	stepArg.Value = "modified"
	origStepArg, err := originalLookup.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(origStepArg.Value, Equals, mlValue)
}

func (s *MySuite) TestMultilineStringWithEmptyValue(c *C) {
	lookup := new(ArgLookup)
	lookup.AddArgName("emptyParam")

	err := lookup.AddArgValue("emptyParam", &StepArg{Value: "", ArgType: MultilineString})
	c.Assert(err, IsNil)

	stepArg, err := lookup.GetArg("emptyParam")
	c.Assert(err, IsNil)
	c.Assert(stepArg.ArgType, Equals, MultilineString)
	c.Assert(stepArg.Value, Equals, "")
}

func (s *MySuite) TestMultilineStringConstantDefinition(c *C) {
	// Test that MultilineString constant is properly defined
	c.Assert(MultilineString, Equals, ArgType("multiline_string"))
}

func (s *MySuite) TestStepArgWithMultilineStringType(c *C) {
	// Test basic multiline string functionality
	multilineContent := `First line
Second line
Third line`
	
	stepArg := &StepArg{
		Value:   multilineContent,
		ArgType: MultilineString,
	}

	c.Assert(stepArg.ArgType, Equals, MultilineString)
	c.Assert(stepArg.Value, Equals, multilineContent)
	c.Assert(stepArg.ArgValue(), Equals, multilineContent)
}

func (s *MySuite) TestArgLookupStoresMultilineString(c *C) {
	// Test that ArgLookup can store and retrieve multiline strings
	lookup := new(ArgLookup)
	lookup.AddArgName("config")
	
	multilineValue := `host: localhost
port: 8080`

	err := lookup.AddArgValue("config", &StepArg{
		Value:   multilineValue,
		ArgType: MultilineString,
	})
	c.Assert(err, IsNil)

	arg, err := lookup.GetArg("config")
	c.Assert(err, IsNil)
	c.Assert(arg.ArgType, Equals, MultilineString)
	c.Assert(arg.Value, Equals, multilineValue)
}