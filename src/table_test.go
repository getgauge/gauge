package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestIsInitalized(c *C) {
	var table table
	c.Assert(table.isInitialized(), Equals, false)
	c.Assert(table.getRowCount(), Equals, 0)

	table.addHeaders([]string{"one", "two", "three"})

	c.Assert(table.isInitialized(), Equals, true)
}

func (s *MySuite) TestShouldAddHeaders(c *C) {
	var table table

	table.addHeaders([]string{"one", "two", "three"})

	c.Assert(len(table.headers), Equals, 3)
	c.Assert(table.headerIndexMap["one"], Equals, 0)
	c.Assert(table.headers[0], Equals, "one")
	c.Assert(table.headerIndexMap["two"], Equals, 1)
	c.Assert(table.headers[1], Equals, "two")
	c.Assert(table.headerIndexMap["three"], Equals, 2)
	c.Assert(table.headers[2], Equals, "three")
}

func (s *MySuite) TestShouldAddRows(c *C) {
	var table table

	table.addHeaders([]string{"one", "two", "three"})
	table.addRowValues([]string{"foo", "bar", "baz"})
	table.addRowValues([]string{"john", "jim", "jack"})

	c.Assert(table.getRowCount(), Equals, 2)
	column1 := table.get("one")
	c.Assert(len(column1), Equals, 2)
	c.Assert(column1[0], Equals, "foo")
	c.Assert(column1[1], Equals, "john")

	column2 := table.get("two")
	c.Assert(len(column2), Equals, 2)
	c.Assert(column2[0], Equals, "bar")
	c.Assert(column2[1], Equals, "jim")

	column3 := table.get("three")
	c.Assert(len(column3), Equals, 2)
	c.Assert(column3[0], Equals, "baz")
	c.Assert(column3[1], Equals, "jack")

}
