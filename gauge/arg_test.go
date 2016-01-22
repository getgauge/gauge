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

	c.Assert(lookup.GetArg("param1").Value, Equals, "value1")
	c.Assert(lookup.GetArg("param2").Value, Equals, "value2")
}

func (s *MySuite) TestPanicForInvalidArg(c *C) {
	lookup := new(ArgLookup)

	c.Assert(func() { lookup.AddArgValue("param1", &StepArg{Value: "value1", ArgType: Static}) }, Panics, "Accessing an invalid parameter (param1)")
	c.Assert(func() { lookup.GetArg("param1") }, Panics, "Accessing an invalid parameter (param1)")
}

func (s *MySuite) TestGetLookupCopy(c *C) {
	originalLookup := new(ArgLookup)
	originalLookup.AddArgName("param1")
	originalLookup.AddArgValue("param1", &StepArg{Value: "oldValue", ArgType: Dynamic})

	copiedLookup := originalLookup.GetCopy()
	copiedLookup.AddArgValue("param1", &StepArg{Value: "new value", ArgType: Static})

	c.Assert(copiedLookup.GetArg("param1").Value, Equals, "new value")
	c.Assert(originalLookup.GetArg("param1").Value, Equals, "oldValue")
}

func (s *MySuite) TestGetLookupFromTableRow(c *C) {
	dataTable := new(Table)
	dataTable.AddHeaders([]string{"id", "name"})
	dataTable.AddRowValues([]string{"1", "admin"})
	dataTable.AddRowValues([]string{"2", "root"})

	emptyLookup := new(ArgLookup).FromDataTableRow(new(Table), 0)
	lookup1 := new(ArgLookup).FromDataTableRow(dataTable, 0)
	lookup2 := new(ArgLookup).FromDataTableRow(dataTable, 1)

	c.Assert(emptyLookup.ParamIndexMap, IsNil)

	c.Assert(lookup1.GetArg("id").Value, Equals, "1")
	c.Assert(lookup1.GetArg("id").ArgType, Equals, Static)
	c.Assert(lookup1.GetArg("name").Value, Equals, "admin")
	c.Assert(lookup1.GetArg("name").ArgType, Equals, Static)

	c.Assert(lookup2.GetArg("id").Value, Equals, "2")
	c.Assert(lookup2.GetArg("id").ArgType, Equals, Static)
	c.Assert(lookup2.GetArg("name").Value, Equals, "root")
	c.Assert(lookup2.GetArg("name").ArgType, Equals, Static)
}
