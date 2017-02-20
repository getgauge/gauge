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

package config

import (
	"os"
	"testing"

	. "github.com/go-check/check"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func stubGetFromConfig(propertyName string) string {
	return ""
}

func stub2GetFromConfig(propertyName string) string {
	return "10000"
}

func stub3GetFromConfig(propertyName string) string {
	return "false"
}

func stub4GetFromConfig(propertyName string) string {
	return "true	"
}

func (s *MySuite) TestRunnerRequestTimeout(c *C) {
	getFromConfig = stubGetFromConfig
	c.Assert(RunnerRequestTimeout(), Equals, defaultRunnerRequestTimeout)

	getFromConfig = stub2GetFromConfig
	c.Assert(RunnerRequestTimeout().Seconds(), Equals, float64(10))

	os.Setenv(runnerRequestTimeout, "1000")
	c.Assert(RunnerRequestTimeout().Seconds(), Equals, float64(1))
}

func (s *MySuite) TestAllowUpdates(c *C) {
	getFromConfig = stubGetFromConfig
	c.Assert(CheckUpdates(), Equals, true)

	getFromConfig = stub2GetFromConfig
	c.Assert(CheckUpdates(), Equals, true)

	getFromConfig = stub3GetFromConfig
	c.Assert(CheckUpdates(), Equals, false)

	getFromConfig = stub4GetFromConfig
	c.Assert(CheckUpdates(), Equals, true)
}
