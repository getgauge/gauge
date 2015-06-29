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
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestIsInitalized(c *C) {
	var table Table
	c.Assert(table.isInitialized(), Equals, false)
	c.Assert(table.getRowCount(), Equals, 0)

	table.addHeaders([]string{"one", "two", "three"})

	c.Assert(table.isInitialized(), Equals, true)
}

func (s *MySuite) TestShouldAddHeaders(c *C) {
	var table Table

	table.addHeaders([]string{"one", "two", "three"})

	c.Assert(len(table.headers), Equals, 3)
	c.Assert(table.headerIndexMap["one"], Equals, 0)
	c.Assert(table.headers[0], Equals, "one")
	c.Assert(table.headerIndexMap["two"], Equals, 1)
	c.Assert(table.headers[1], Equals, "two")
	c.Assert(table.headerIndexMap["three"], Equals, 2)
	c.Assert(table.headers[2], Equals, "three")
}

func (s *MySuite) TestShouldAddRowValues(c *C) {
	var table Table

	table.addHeaders([]string{"one", "two", "three"})
	table.addRowValues([]string{"foo", "bar", "baz"})
	table.addRowValues([]string{"john", "jim"})

	c.Assert(table.getRowCount(), Equals, 2)
	column1 := table.get("one")
	c.Assert(len(column1), Equals, 2)
	c.Assert(column1[0].value, Equals, "foo")
	c.Assert(column1[0].cellType, Equals, Static)
	c.Assert(column1[1].value, Equals, "john")
	c.Assert(column1[1].cellType, Equals, Static)

	column2 := table.get("two")
	c.Assert(len(column2), Equals, 2)
	c.Assert(column2[0].value, Equals, "bar")
	c.Assert(column2[0].cellType, Equals, Static)
	c.Assert(column2[1].value, Equals, "jim")
	c.Assert(column2[1].cellType, Equals, Static)

	column3 := table.get("three")
	c.Assert(len(column3), Equals, 2)
	c.Assert(column3[0].value, Equals, "baz")
	c.Assert(column3[0].cellType, Equals, Static)
	c.Assert(column3[1].value, Equals, "")
	c.Assert(column3[1].cellType, Equals, Static)

}

func (s *MySuite) TestShouldAddRows(c *C) {
	var table Table

	table.addHeaders([]string{"one", "two", "three"})
	table.addRows([]TableCell{TableCell{"foo", Static}, TableCell{"bar", Static}, TableCell{"baz", Static}})
	table.addRows([]TableCell{TableCell{"john", Static}, TableCell{"jim", Static}})

	c.Assert(table.getRowCount(), Equals, 2)
	column1 := table.get("one")
	c.Assert(len(column1), Equals, 2)
	c.Assert(column1[0].value, Equals, "foo")
	c.Assert(column1[0].cellType, Equals, Static)
	c.Assert(column1[1].value, Equals, "john")
	c.Assert(column1[1].cellType, Equals, Static)

	column2 := table.get("two")
	c.Assert(len(column2), Equals, 2)
	c.Assert(column2[0].value, Equals, "bar")
	c.Assert(column2[0].cellType, Equals, Static)
	c.Assert(column2[1].value, Equals, "jim")
	c.Assert(column2[1].cellType, Equals, Static)

	column3 := table.get("three")
	c.Assert(len(column3), Equals, 2)
	c.Assert(column3[0].value, Equals, "baz")
	c.Assert(column3[0].cellType, Equals, Static)
	c.Assert(column3[1].value, Equals, "")
	c.Assert(column3[1].cellType, Equals, Static)

}

func (s *MySuite) TestCoulmnNameExists(c *C) {
	var table Table

	table.addHeaders([]string{"one", "two", "three"})
	table.addRowValues([]string{"foo", "bar", "baz"})
	table.addRowValues([]string{"john", "jim", "jack"})

	c.Assert(table.headerExists("one"), Equals, true)
	c.Assert(table.headerExists("two"), Equals, true)
	c.Assert(table.headerExists("four"), Equals, false)

}

func (s *MySuite) TestGetInvalidColumn(c *C) {
	var table Table

	table.addHeaders([]string{"one", "two", "three"})
	table.addRowValues([]string{"foo", "bar", "baz"})
	table.addRowValues([]string{"john", "jim", "jack"})

	c.Assert(func() { table.get("four") }, Panics, "Table column four not found")
}

func (s *MySuite) TestGetRows(c *C) {
	var table Table

	table.addHeaders([]string{"one", "two", "three"})
	table.addRowValues([]string{"foo", "bar", "baz"})
	table.addRowValues([]string{"john", "jim", "jack"})

	rows := table.getRows()
	c.Assert(len(rows), Equals, 2)
	firstRow := rows[0]
	c.Assert(firstRow[0], Equals, "foo")
	c.Assert(firstRow[1], Equals, "bar")
	c.Assert(firstRow[2], Equals, "baz")

	secondRow := rows[1]
	c.Assert(secondRow[0], Equals, "john")
	c.Assert(secondRow[1], Equals, "jim")
	c.Assert(secondRow[2], Equals, "jack")
}

func (s *MySuite) TestValuesBasedOnHeaders(c *C) {
	var table Table
	table.addHeaders([]string{"id", "name"})

	firstRow := table.toHeaderSizeRow([]TableCell{TableCell{"123", Static}, TableCell{"foo", Static}})
	secondRow := table.toHeaderSizeRow([]TableCell{TableCell{"jim", Static}, TableCell{"jack", Static}})
	thirdRow := table.toHeaderSizeRow([]TableCell{TableCell{"789", Static}})

	c.Assert(len(firstRow), Equals, 2)
	c.Assert(firstRow[0].value, Equals, "123")
	c.Assert(firstRow[1].value, Equals, "foo")

	c.Assert(len(secondRow), Equals, 2)
	c.Assert(secondRow[0].value, Equals, "jim")
	c.Assert(secondRow[1].value, Equals, "jack")

	c.Assert(len(thirdRow), Equals, 2)
	c.Assert(thirdRow[0].value, Equals, "789")
	c.Assert(thirdRow[1].value, Equals, "")
}
