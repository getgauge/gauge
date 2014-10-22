package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestParsingFileSpecialType(c *C) {
	resolver := newSpecialTypeResolver()
	resolver.predefinedResolvers["file"] = func(value string) (*stepArg, error) {
		return &stepArg{value: "dummy", argType: static}, nil
	}

	stepArg, _ := resolver.resolve("file:foo")
	c.Assert(stepArg.value, Equals, "dummy")
	c.Assert(stepArg.argType, Equals, static)
	c.Assert(stepArg.name, Equals, "file:foo")
}

func (s *MySuite) TestConvertCsvToTable(c *C) {
	table, _ := convertCsvToTable("id,name \n1,foo\n2,bar")

	idColumn := table.get("id")
	c.Assert(idColumn[0].value, Equals, "1")
	c.Assert(idColumn[1].value, Equals, "2")

	nameColumn := table.get("name")
	c.Assert(nameColumn[0].value, Equals, "foo")
	c.Assert(nameColumn[1].value, Equals, "bar")
}

func (s *MySuite) TestParsingUnknownSpecialType(c *C) {
	resolver := newSpecialTypeResolver()

	_, err := resolver.resolve("unknown:foo")
	c.Assert(err.Error(), Equals, "Resolver not found for special param <unknown:foo>")
}
