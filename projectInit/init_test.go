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

package projectInit

import (
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/config"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestGetTemplateLanguage(c *C) {
	c.Assert(getTemplateLanguage("java"), Equals, "java")
	c.Assert(getTemplateLanguage("java_maven"), Equals, "java")
	c.Assert(getTemplateLanguage("java_maven_selenium"), Equals, "java")
}

func (s *MySuite) TestIfGaugeProjectGivenEmptyDir(c *C) {
	path, _ := filepath.Abs("_testdata")
	config.ProjectRoot = path
	c.Assert(isGaugeProject(), Equals, false)
}

func (s *MySuite) TestIfGaugeProject(c *C) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "gaugeProject"))
	config.ProjectRoot = path
	c.Assert(isGaugeProject(), Equals, true)
}

func (s *MySuite) TestIfGaugeProjectGivenDirWithNonGaugeManifest(c *C) {
	path, _ := filepath.Abs(filepath.Join("_testdata", "foo"))
	config.ProjectRoot = path
	c.Assert(isGaugeProject(), Equals, false)
}
