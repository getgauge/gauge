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
	. "gopkg.in/check.v1"
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

func (s *MySuite) TestVersionComparisonGreaterLesser(c *C) {
	higherVersion, _ := parseVersion("0.0.7")
	lowerVersion, _ := parseVersion("0.0.3")
	c.Assert(lowerVersion.isLesserThan(higherVersion), Equals, true)
	c.Assert(higherVersion.isGreaterThan(lowerVersion), Equals, true)

	higherVersion, _ = parseVersion("0.7.2")
	lowerVersion, _ = parseVersion("0.5.7")
	c.Assert(lowerVersion.isLesserThan(higherVersion), Equals, true)
	c.Assert(higherVersion.isGreaterThan(lowerVersion), Equals, true)

	higherVersion, _ = parseVersion("4.7.2")
	lowerVersion, _ = parseVersion("3.8.7")
	c.Assert(lowerVersion.isLesserThan(higherVersion), Equals, true)
	c.Assert(higherVersion.isGreaterThan(lowerVersion), Equals, true)

	version1, _ := parseVersion("4.7.2")
	version2, _ := parseVersion("4.7.2")
	c.Assert(version1.isEqualTo(version2), Equals, true)
}

func (s *MySuite) TestVersionComparisonGreaterThanEqual(c *C) {
	higherVersion, _ := parseVersion("0.0.7")
	lowerVersion, _ := parseVersion("0.0.3")
	c.Assert(higherVersion.isGreaterThanEqualTo(lowerVersion), Equals, true)

	higherVersion, _ = parseVersion("0.7.2")
	lowerVersion, _ = parseVersion("0.5.7")
	c.Assert(higherVersion.isGreaterThan(lowerVersion), Equals, true)

	higherVersion, _ = parseVersion("4.7.2")
	lowerVersion, _ = parseVersion("3.8.7")
	c.Assert(lowerVersion.isLesserThan(higherVersion), Equals, true)
	c.Assert(higherVersion.isGreaterThan(lowerVersion), Equals, true)

	version1, _ := parseVersion("6.7.2")
	version2, _ := parseVersion("6.7.2")
	c.Assert(version1.isGreaterThanEqualTo(version2), Equals, true)
}

func (s *MySuite) TestVersionComparisonLesserThanEqual(c *C) {
	higherVersion, _ := parseVersion("0.0.7")
	lowerVersion, _ := parseVersion("0.0.3")
	c.Assert(lowerVersion.isLesserThanEqualTo(higherVersion), Equals, true)

	higherVersion, _ = parseVersion("0.7.2")
	lowerVersion, _ = parseVersion("0.5.7")
	c.Assert(lowerVersion.isLesserThanEqualTo(higherVersion), Equals, true)

	higherVersion, _ = parseVersion("5.8.2")
	lowerVersion, _ = parseVersion("2.9.7")
	c.Assert(lowerVersion.isLesserThanEqualTo(higherVersion), Equals, true)

	version1, _ := parseVersion("6.7.2")
	version2, _ := parseVersion("6.7.2")
	c.Assert(version1.isLesserThanEqualTo(version2), Equals, true)
}

func (s *MySuite) TestVersionIsBetweenTwoVersions(c *C) {
	higherVersion, _ := parseVersion("0.0.9")
	lowerVersion, _ := parseVersion("0.0.7")
	middleVersion, _ := parseVersion("0.0.8")
	c.Assert(middleVersion.isBetween(lowerVersion, higherVersion), Equals, true)

	higherVersion, _ = parseVersion("0.7.2")
	lowerVersion, _ = parseVersion("0.5.7")
	middleVersion, _ = parseVersion("0.6.9")
	c.Assert(middleVersion.isBetween(lowerVersion, higherVersion), Equals, true)

	higherVersion, _ = parseVersion("4.7.2")
	lowerVersion, _ = parseVersion("3.8.7")
	middleVersion, _ = parseVersion("4.0.1")
	c.Assert(middleVersion.isBetween(lowerVersion, higherVersion), Equals, true)

	higherVersion, _ = parseVersion("4.7.2")
	lowerVersion, _ = parseVersion("4.0.1")
	middleVersion, _ = parseVersion("4.0.1")
	c.Assert(middleVersion.isBetween(lowerVersion, higherVersion), Equals, true)

	higherVersion, _ = parseVersion("0.0.2")
	lowerVersion, _ = parseVersion("0.0.1")
	middleVersion, _ = parseVersion("0.0.2")
	c.Assert(middleVersion.isBetween(lowerVersion, higherVersion), Equals, true)
}
