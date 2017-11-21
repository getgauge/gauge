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

package env

import (
	"os"
	"testing"

	"github.com/getgauge/gauge/config"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestLoadDefaultEnv(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("default")

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
}

// If default env dir is present, the values present in there should overwrite
// the default values (present in the code), even when env flag is passed
func (s *MySuite) TestLoadDefaultEnvFromDirIfPresent(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj2"

	e := LoadEnv("foo")

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports_dir")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "false")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "false")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
}

// If default env dir is present, the values present in there should overwrite
// the default values (present in the code), even when env flag is passed.
// If the passed env also has the same values, that should take precedence.
func (s *MySuite) TestLoadDefaultEnvFromDirAndOverwritePassedEnv(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj2"

	e := LoadEnv("bar")

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports_dir")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "false")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "bar/logs")
}

func (s *MySuite) TestLoadDefaultEnvEvenIfDefaultEnvNotPresent(c *C) {
	os.Clearenv()
	config.ProjectRoot = ""

	e := LoadEnv("default")

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
}

func (s *MySuite) TestLoadDefaultEnvWithOtherPropertiesSetInShell(c *C) {
	os.Clearenv()
	os.Setenv("foo", "bar")
	os.Setenv("logs_directory", "custom_logs_dir")
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("default")

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("foo"), Equals, "bar")
	c.Assert(os.Getenv("property1"), Equals, "value1")
	c.Assert(os.Getenv("logs_directory"), Equals, "custom_logs_dir")
}

func (s *MySuite) TestLoadDefaultEnvWithOtherPropertiesNotSetInShell(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("default")

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("property1"), Equals, "value1")
}

func (s *MySuite) TestLoadCustomEnvAlongWithDefaultEnv(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("foo")

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "false")
	c.Assert(os.Getenv("logs_directory"), Equals, "foo/logs")
}

func (s *MySuite) TestLoadCustomEnvAlongWithOtherPropertiesSetInShell(c *C) {
	os.Clearenv()
	os.Setenv("gauge_reports_dir", "custom_reports_dir")
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("foo")

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "custom_reports_dir")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "false")
	c.Assert(os.Getenv("logs_directory"), Equals, "foo/logs")
}

func (s *MySuite) TestEnvPropertyIsSet(c *C) {
	os.Clearenv()
	os.Setenv("foo", "bar")

	actual := isPropertySet("foo")

	c.Assert(actual, Equals, true)
}

func (s *MySuite) TestEnvPropertyIsNotSet(c *C) {
	os.Clearenv()

	actual := isPropertySet("foo")

	c.Assert(actual, Equals, false)
}

func (s *MySuite) TestFatalErrorIsThrownIfEnvNotFound(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("bar")
	c.Assert(e.Error(), Equals, "Failed to load env. bar environment does not exist")
}

func (s *MySuite) TestLoadDefaultEnvWithSubstitutedVariables(c *C) {
	os.Clearenv()
	os.Setenv("foo", "bar")
	os.Setenv("property1", "value1")

	config.ProjectRoot = "_testdata/proj3"

	e := LoadEnv("default")

	c.Assert(e, Equals, nil)

	c.Assert(os.Getenv("property1"), Equals, "value1")
	c.Assert(os.Getenv("property3"), Equals, "bar/value1")
	c.Assert(os.Getenv("property2"), Equals, "value1/value2")

}

func (s *MySuite) TestLoadDefaultEnvWithInvalidSubstitutedVariable(c *C) {
	config.ProjectRoot = "_testdata/proj3"
	e := LoadEnv("default")
	c.Assert(e, ErrorMatches, ".*env variable was not set.")
}
