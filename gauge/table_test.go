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
	"testing"

	. "github.com/go-check/check"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestIsInitalized(c *C) {
	var table Table
	c.Assert(table.IsInitialized(), Equals, false)
	c.Assert(table.GetRowCount(), Equals, 0)

	table.AddHeaders([]string{"one", "two", "three"})

	c.Assert(table.IsInitialized(), Equals, true)
}

func (s *MySuite) TestShouldAddHeaders(c *C) {
	var table Table

	table.AddHeaders([]string{"one", "two", "three"})

	c.Assert(len(table.Headers), Equals, 3)
	c.Assert(table.headerIndexMap["one"], Equals, 0)
	c.Assert(table.Headers[0], Equals, "one")
	c.Assert(table.headerIndexMap["two"], Equals, 1)
	c.Assert(table.Headers[1], Equals, "two")
	c.Assert(table.headerIndexMap["three"], Equals, 2)
	c.Assert(table.Headers[2], Equals, "three")
}

func (s *MySuite) TestShouldAddRowValues(c *C) {
	var table Table

	table.AddHeaders([]string{"one", "two", "three"})
	table.AddRowValues([]string{"foo", "bar", "baz"})
	table.AddRowValues([]string{"john", "jim"})

	c.Assert(table.GetRowCount(), Equals, 2)
	column1 := table.Get("one")
	c.Assert(len(column1), Equals, 2)
	c.Assert(column1[0].Value, Equals, "foo")
	c.Assert(column1[0].CellType, Equals, Static)
	c.Assert(column1[1].Value, Equals, "john")
	c.Assert(column1[1].CellType, Equals, Static)

	column2 := table.Get("two")
	c.Assert(len(column2), Equals, 2)
	c.Assert(column2[0].Value, Equals, "bar")
	c.Assert(column2[0].CellType, Equals, Static)
	c.Assert(column2[1].Value, Equals, "jim")
	c.Assert(column2[1].CellType, Equals, Static)

	column3 := table.Get("three")
	c.Assert(len(column3), Equals, 2)
	c.Assert(column3[0].Value, Equals, "baz")
	c.Assert(column3[0].CellType, Equals, Static)
	c.Assert(column3[1].Value, Equals, "")
	c.Assert(column3[1].CellType, Equals, Static)
}

func (s *MySuite) TestShouldAddRows(c *C) {
	var table Table

	table.AddHeaders([]string{"one", "two", "three"})
	table.addRows([]TableCell{TableCell{"foo", Static}, TableCell{"bar", Static}, TableCell{"baz", Static}})
	table.addRows([]TableCell{TableCell{"john", Static}, TableCell{"jim", Static}})

	c.Assert(table.GetRowCount(), Equals, 2)
	column1 := table.Get("one")
	c.Assert(len(column1), Equals, 2)
	c.Assert(column1[0].Value, Equals, "foo")
	c.Assert(column1[0].CellType, Equals, Static)
	c.Assert(column1[1].Value, Equals, "john")
	c.Assert(column1[1].CellType, Equals, Static)

	column2 := table.Get("two")
	c.Assert(len(column2), Equals, 2)
	c.Assert(column2[0].Value, Equals, "bar")
	c.Assert(column2[0].CellType, Equals, Static)
	c.Assert(column2[1].Value, Equals, "jim")
	c.Assert(column2[1].CellType, Equals, Static)

	column3 := table.Get("three")
	c.Assert(len(column3), Equals, 2)
	c.Assert(column3[0].Value, Equals, "baz")
	c.Assert(column3[0].CellType, Equals, Static)
	c.Assert(column3[1].Value, Equals, "")
	c.Assert(column3[1].CellType, Equals, Static)
}

func (s *MySuite) TestCoulmnNameExists(c *C) {
	var table Table

	table.AddHeaders([]string{"one", "two", "three"})
	table.AddRowValues([]string{"foo", "bar", "baz"})
	table.AddRowValues([]string{"john", "jim", "jack"})

	c.Assert(table.headerExists("one"), Equals, true)
	c.Assert(table.headerExists("two"), Equals, true)
	c.Assert(table.headerExists("four"), Equals, false)
}

func (s *MySuite) TestGetInvalidColumn(c *C) {
	var table Table

	table.AddHeaders([]string{"one", "two", "three"})
	table.AddRowValues([]string{"foo", "bar", "baz"})
	table.AddRowValues([]string{"john", "jim", "jack"})

	c.Assert(func() { table.Get("four") }, Panics, "Table column four not found")
}

func (s *MySuite) TestGetRows(c *C) {
	var table Table

	table.AddHeaders([]string{"one", "two", "three"})
	table.AddRowValues([]string{"foo", "bar", "baz"})
	table.AddRowValues([]string{"john", "jim", "jack"})

	rows := table.Rows()
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
	table.AddHeaders([]string{"id", "name"})

	firstRow := table.toHeaderSizeRow([]TableCell{TableCell{"123", Static}, TableCell{"foo", Static}})
	secondRow := table.toHeaderSizeRow([]TableCell{TableCell{"jim", Static}, TableCell{"jack", Static}})
	thirdRow := table.toHeaderSizeRow([]TableCell{TableCell{"789", Static}})

	c.Assert(len(firstRow), Equals, 2)
	c.Assert(firstRow[0].Value, Equals, "123")
	c.Assert(firstRow[1].Value, Equals, "foo")

	c.Assert(len(secondRow), Equals, 2)
	c.Assert(secondRow[0].Value, Equals, "jim")
	c.Assert(secondRow[1].Value, Equals, "jack")

	c.Assert(len(thirdRow), Equals, 2)
	c.Assert(thirdRow[0].Value, Equals, "789")
	c.Assert(thirdRow[1].Value, Equals, "")
}

func (s *MySuite) TestCreateTableCells(c *C) {
	var table Table
	table.AddHeaders([]string{"id", "name"})
	table.AddRowValues([]string{"cell 1", "cell 2   "})

	c.Assert(table.Columns[0][0].Value, Equals, "cell 1")
	c.Assert(table.Columns[1][0].Value, Equals, "cell 2   ")
}
