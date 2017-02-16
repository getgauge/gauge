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

package runner

import (
	"testing"

	"github.com/getgauge/common"
	. "github.com/go-check/check"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestGetCleanEnvRemovesGAUGE_INTERNAL_PORTAndSetsPortNumber(c *C) {
	HELLO := "HELLO"
	portVariable := common.GaugeInternalPortEnvName + "=1234"
	PORT_NAME_WITH_EXTRA_WORD := "b" + common.GaugeInternalPortEnvName
	PORT_NAME_WITH_SPACES := "      " + common.GaugeInternalPortEnvName + "         "
	env := getCleanEnv("1234", []string{HELLO, common.GaugeInternalPortEnvName + "=", common.GaugeInternalPortEnvName,
		PORT_NAME_WITH_SPACES, PORT_NAME_WITH_EXTRA_WORD}, false)

	c.Assert(env[0], Equals, HELLO)
	c.Assert(env[1], Equals, portVariable)
	c.Assert(env[2], Equals, portVariable)
	c.Assert(env[3], Equals, portVariable)
	c.Assert(env[4], Equals, PORT_NAME_WITH_EXTRA_WORD)
}

func (s *MySuite) TestGetCleanEnvWithDebugging(c *C) {
	env := getCleanEnv("1234", []string{}, true)

	c.Assert(env[1], Equals, "debugging=true")
}
