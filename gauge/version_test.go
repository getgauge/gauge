package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestParsingVersion(c *C) {
	version, err := parseVersion("1.5.9")
	c.Assert(err, Equals, nil)
	c.Assert(version.major, Equals, 1)
	c.Assert(version.minor, Equals, 5)
	c.Assert(version.patch, Equals, 9)
}

func (s *MySuite) TestParsingErrorForIncorrectNumberOfDotCharacters(c *C) {
	_, err := parseVersion("1.5.9.9")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Incorrect number of '.' characters in Version. Version should be of the form 1.5.7")

	_, err = parseVersion("0.")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Incorrect number of '.' characters in Version. Version should be of the form 1.5.7")
}

func (s *MySuite) TestParsingErrorForNonIntegerVersion(c *C) {
	_, err := parseVersion("a.9.0")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Error parsing major version number a to integer. strconv.ParseInt: parsing \"a\": invalid syntax")

	_, err = parseVersion("0.ffhj.78")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Error parsing minor version number 0 to integer. strconv.ParseInt: parsing \"ffhj\": invalid syntax")

	_, err = parseVersion("8.9.opl")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Error parsing patch version number 8 to integer. strconv.ParseInt: parsing \"opl\": invalid syntax")
}
