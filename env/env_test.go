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
	config.ProjectRoot = "testdata"
	ProjectEnv = "default"

	LoadEnv(true)

	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
}

func (s *MySuite) TestLoadCustomEnvAlongWithDefaultEnv(c *C) {
	config.ProjectRoot = "testdata"
	ProjectEnv = "foo"

	LoadEnv(true)

	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "false")
	c.Assert(os.Getenv("logs_directory"), Equals, "foo/logs")
}
