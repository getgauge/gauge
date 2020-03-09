// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either Version 3 of the License, or
// (at your option) any later Version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package version

import (
	"fmt"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestParsingVersion(c *C) {
	Version, err := ParseVersion("1.5.9")
	c.Assert(err, Equals, nil)
	c.Assert(Version.Major, Equals, 1)
	c.Assert(Version.Minor, Equals, 5)
	c.Assert(Version.Patch, Equals, 9)
}

func (s *MySuite) TestParsingErrorForIncorrectNumberOfDotCharacters(c *C) {
	_, err := ParseVersion("1.5.9.9")
	c.Assert(err, ErrorMatches, "incorrect version format. version should be in the form 1.5.7")

	_, err = ParseVersion("0.")
	c.Assert(err, ErrorMatches, "incorrect version format. version should be in the form 1.5.7")
}

func (s *MySuite) TestParsingErrorForNonIntegerVersion(c *C) {
	_, err := ParseVersion("a.9.0")
	c.Assert(err, ErrorMatches, `error parsing major version a to integer. strconv.*: parsing "a": invalid syntax`)

	_, err = ParseVersion("0.ffhj.78")
	c.Assert(err, ErrorMatches, `error parsing minor version ffhj to integer. strconv.*: parsing "ffhj": invalid syntax`)

	_, err = ParseVersion("8.9.opl")
	c.Assert(err, ErrorMatches, `error parsing patch version opl to integer. strconv.*: parsing "opl": invalid syntax`)
}

func (s *MySuite) TestVersionComparisonGreaterLesser(c *C) {
	higherVersion, _ := ParseVersion("0.0.7")
	lowerVersion, _ := ParseVersion("0.0.3")
	c.Assert(lowerVersion.IsLesserThan(higherVersion), Equals, true)
	c.Assert(higherVersion.IsGreaterThan(lowerVersion), Equals, true)

	higherVersion, _ = ParseVersion("0.7.2")
	lowerVersion, _ = ParseVersion("0.5.7")
	c.Assert(lowerVersion.IsLesserThan(higherVersion), Equals, true)
	c.Assert(higherVersion.IsGreaterThan(lowerVersion), Equals, true)

	higherVersion, _ = ParseVersion("4.7.2")
	lowerVersion, _ = ParseVersion("3.8.7")
	c.Assert(lowerVersion.IsLesserThan(higherVersion), Equals, true)
	c.Assert(higherVersion.IsGreaterThan(lowerVersion), Equals, true)

	version1, _ := ParseVersion("4.7.2")
	version2, _ := ParseVersion("4.7.2")
	c.Assert(version1.IsEqualTo(version2), Equals, true)
}

func (s *MySuite) TestVersionComparisonGreaterThanEqual(c *C) {
	higherVersion, _ := ParseVersion("0.0.7")
	lowerVersion, _ := ParseVersion("0.0.3")
	c.Assert(higherVersion.IsGreaterThanEqualTo(lowerVersion), Equals, true)

	higherVersion, _ = ParseVersion("0.7.2")
	lowerVersion, _ = ParseVersion("0.5.7")
	c.Assert(higherVersion.IsGreaterThan(lowerVersion), Equals, true)

	higherVersion, _ = ParseVersion("4.7.2")
	lowerVersion, _ = ParseVersion("3.8.7")
	c.Assert(lowerVersion.IsLesserThan(higherVersion), Equals, true)
	c.Assert(higherVersion.IsGreaterThan(lowerVersion), Equals, true)

	version1, _ := ParseVersion("6.7.2")
	version2, _ := ParseVersion("6.7.2")
	c.Assert(version1.IsGreaterThanEqualTo(version2), Equals, true)
}

func (s *MySuite) TestVersionComparisonLesserThanEqual(c *C) {
	higherVersion, _ := ParseVersion("0.0.7")
	lowerVersion, _ := ParseVersion("0.0.3")
	c.Assert(lowerVersion.IsLesserThanEqualTo(higherVersion), Equals, true)

	higherVersion, _ = ParseVersion("0.7.2")
	lowerVersion, _ = ParseVersion("0.5.7")
	c.Assert(lowerVersion.IsLesserThanEqualTo(higherVersion), Equals, true)

	higherVersion, _ = ParseVersion("5.8.2")
	lowerVersion, _ = ParseVersion("2.9.7")
	c.Assert(lowerVersion.IsLesserThanEqualTo(higherVersion), Equals, true)

	version1, _ := ParseVersion("6.7.2")
	version2, _ := ParseVersion("6.7.2")
	c.Assert(version1.IsLesserThanEqualTo(version2), Equals, true)
}

func (s *MySuite) TestVersionIsBetweenTwoVersions(c *C) {
	higherVersion, _ := ParseVersion("0.0.9")
	lowerVersion, _ := ParseVersion("0.0.7")
	middleVersion, _ := ParseVersion("0.0.8")
	c.Assert(middleVersion.IsBetween(lowerVersion, higherVersion), Equals, true)

	higherVersion, _ = ParseVersion("0.7.2")
	lowerVersion, _ = ParseVersion("0.5.7")
	middleVersion, _ = ParseVersion("0.6.9")
	c.Assert(middleVersion.IsBetween(lowerVersion, higherVersion), Equals, true)

	higherVersion, _ = ParseVersion("4.7.2")
	lowerVersion, _ = ParseVersion("3.8.7")
	middleVersion, _ = ParseVersion("4.0.1")
	c.Assert(middleVersion.IsBetween(lowerVersion, higherVersion), Equals, true)

	higherVersion, _ = ParseVersion("4.7.2")
	lowerVersion, _ = ParseVersion("4.0.1")
	middleVersion, _ = ParseVersion("4.0.1")
	c.Assert(middleVersion.IsBetween(lowerVersion, higherVersion), Equals, true)

	higherVersion, _ = ParseVersion("0.0.2")
	lowerVersion, _ = ParseVersion("0.0.1")
	middleVersion, _ = ParseVersion("0.0.2")
	c.Assert(middleVersion.IsBetween(lowerVersion, higherVersion), Equals, true)
}

func (s *MySuite) TestGetLatestVersion(c *C) {
	highestVersion := &Version{2, 2, 2}
	versions := []*Version{&Version{0, 0, 1}, &Version{1, 2, 2}, highestVersion, &Version{0, 0, 2}, &Version{0, 2, 2}, &Version{0, 0, 3}, &Version{0, 2, 1}, &Version{0, 1, 2}}
	latestVersion := GetLatestVersion(versions)

	c.Assert(latestVersion, DeepEquals, highestVersion)
}

func (s *MySuite) TestCheckVersionCompatibilitySuccess(c *C) {
	versionSupported := &VersionSupport{"0.6.5", "1.8.5"}
	gaugeVersion := &Version{0, 6, 7}
	c.Assert(CheckCompatibility(gaugeVersion, versionSupported), Equals, nil)

	versionSupported = &VersionSupport{"0.0.1", "0.0.1"}
	gaugeVersion = &Version{0, 0, 1}
	c.Assert(CheckCompatibility(gaugeVersion, versionSupported), Equals, nil)

	versionSupported = &VersionSupport{Minimum: "0.0.1"}
	gaugeVersion = &Version{1, 5, 2}
	c.Assert(CheckCompatibility(gaugeVersion, versionSupported), Equals, nil)

	versionSupported = &VersionSupport{Minimum: "0.5.1"}
	gaugeVersion = &Version{0, 5, 1}
	c.Assert(CheckCompatibility(gaugeVersion, versionSupported), Equals, nil)

}

func (s *MySuite) TestCheckVersionCompatibilityFailure(c *C) {
	versionsSupported := &VersionSupport{"0.6.5", "1.8.5"}
	gaugeVersion := &Version{1, 9, 9}
	c.Assert(CheckCompatibility(gaugeVersion, versionsSupported), NotNil)

	versionsSupported = &VersionSupport{"0.0.1", "0.0.1"}
	gaugeVersion = &Version{0, 0, 2}
	c.Assert(CheckCompatibility(gaugeVersion, versionsSupported), NotNil)

	versionsSupported = &VersionSupport{Minimum: "1.3.1"}
	gaugeVersion = &Version{1, 3, 0}
	c.Assert(CheckCompatibility(gaugeVersion, versionsSupported), NotNil)

	versionsSupported = &VersionSupport{Minimum: "0.5.1"}
	gaugeVersion = &Version{0, 0, 9}
	c.Assert(CheckCompatibility(gaugeVersion, versionsSupported), NotNil)

}

func (s *MySuite) TestFullVersionWithBuildMetadata(c *C) {
	c.Assert(FullVersion(), Equals, CurrentGaugeVersion.String())
	BuildMetadata = "nightly-2016-02-21"
	c.Assert(FullVersion(), Equals, fmt.Sprintf("%s.%s", CurrentGaugeVersion.String(), BuildMetadata))
}
