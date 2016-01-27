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
	"os/exec"
	"testing"

	"github.com/getgauge/gauge/config"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestLoadDefaultEnv(c *C) {
	os.Clearenv()
	config.ProjectRoot = "testdata"

	LoadEnv("default")

	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
}

func (s *MySuite) TestLoadDefaultEnvEvenIfDefaultEnvNotPresent(c *C) {
	os.Clearenv()
	config.ProjectRoot = ""

	LoadEnv("default")

	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
}

func (s *MySuite) TestLoadCustomEnvAlongWithDefaultEnv(c *C) {
	os.Clearenv()
	config.ProjectRoot = "testdata"

	LoadEnv("foo")

	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
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

func (s *MySuite) TestDefaultPropertiesAreLoaded(c *C) {
	os.Clearenv()

	err := loadDefaultProperties()

	c.Assert(err, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
}

func (s *MySuite) TestPropertyCanBeOverwrittenIfNotSet(c *C) {
	os.Clearenv()

	canOverwrite := canOverwriteProperty("foo")

	c.Assert(canOverwrite, Equals, true)
}

func (s *MySuite) TestPropertyCanBeOverwrittenIfSetToDefault(c *C) {
	os.Clearenv()
	loadDefaultProperties()

	canOverwrite := canOverwriteProperty("gauge_reports_dir")

	c.Assert(canOverwrite, Equals, true)
}

func (s *MySuite) TestPropertyCantBeOverwrittenIfNotSetToDefault(c *C) {
	os.Clearenv()
	loadDefaultProperties()
	os.Setenv("gauge_reports_dir", "execution_reports")

	canOverwrite := canOverwriteProperty("gauge_reports_dir")

	c.Assert(canOverwrite, Equals, false)
}

// If env passed by user is not found, Gauge should exit non-zero error code.
func TestFatalErrorIsThrownIfEnvNotFound(t *testing.T) {
	if os.Getenv("NO_ENV") == "1" {
		os.Clearenv()
		config.ProjectRoot = "testdata"

		LoadEnv("bar")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatalErrorIsThrownIfEnvNotFound")
	cmd.Env = append(os.Environ(), "NO_ENV=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("Expected: Fatal Error\nGot: Error %v ", err)
}
