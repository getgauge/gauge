/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
