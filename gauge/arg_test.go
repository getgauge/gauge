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

import . "gopkg.in/check.v1"

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
	lookup.AddArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	lookup.AddArgName("param2")
	lookup.AddArgValue("param2", &StepArg{Value: "value2", ArgType: Dynamic})
	stepArg, err := lookup.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "value1")
	stepArg, err = lookup.GetArg("param2")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "value2")
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
	originalLookup.AddArgValue("param1", &StepArg{Value: "oldValue", ArgType: Dynamic})

	copiedLookup, _ := originalLookup.GetCopy()
	copiedLookup.AddArgValue("param1", &StepArg{Value: "new value", ArgType: Static})
	stepArg, err := copiedLookup.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "new value")
	stepArg, err = originalLookup.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "oldValue")
}

func (s *MySuite) TestGetLookupFromTableRow(c *C) {
	dataTable := new(Table)
	dataTable.AddHeaders([]string{"id", "name"})
	dataTable.AddRowValues(dataTable.CreateTableCells([]string{"1", "admin"}))
	dataTable.AddRowValues(dataTable.CreateTableCells([]string{"2", "root"}))

	emptyLookup, err := new(ArgLookup).FromDataTableRow(new(Table), 0)
	c.Assert(err, IsNil)
	c.Assert(emptyLookup.ParamIndexMap, IsNil)

	lookup1, err := new(ArgLookup).FromDataTableRow(dataTable, 0)
	c.Assert(err, IsNil)
	idArg1, err := lookup1.GetArg("id")
	c.Assert(err, IsNil)
	nameArg1, err := lookup1.GetArg("name")
	c.Assert(err, IsNil)
	c.Assert(idArg1.Value, Equals, "1")
	c.Assert(idArg1.ArgType, Equals, Static)
	c.Assert(nameArg1.Value, Equals, "admin")
	c.Assert(nameArg1.ArgType, Equals, Static)

	lookup2, err := new(ArgLookup).FromDataTableRow(dataTable, 1)
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
