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

package execution

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/getgauge/gauge/gauge"

	. "gopkg.in/check.v1"
)

func (s *MySuite) TestFunctionsOfTypeSpecList(c *C) {
	mySpecs := &specStore{specs: createSpecsList(4)}
	c.Assert(mySpecs.next().FileName, Equals, "spec0")
	c.Assert(mySpecs.next().FileName, Equals, "spec1")
	c.Assert(mySpecs.hasNext(), Equals, true)
	c.Assert(mySpecs.size(), Equals, 4)
	c.Assert(mySpecs.next().FileName, Equals, "spec2")
	c.Assert(mySpecs.next().FileName, Equals, "spec3")
	c.Assert(mySpecs.hasNext(), Equals, false)
}

func createSpecsList(number int) []*gauge.Specification {
	var specs []*gauge.Specification
	for i := 0; i < number; i++ {
		specs = append(specs, &gauge.Specification{FileName: fmt.Sprint("spec", i)})
	}
	return specs
}

func (s *MySuite) TestValidateFlagsIfNotParallel(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = false
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsIfNotParallel")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	err := cmd.Run()
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestValidateFlagsWithStartegyEager(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = true
		Strategy = "eager"
		NumberOfExecutionStreams = 1
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsWithStartegyEager")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	err := cmd.Run()
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestValidateFlagsWithStartegyLazy(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = true
		Strategy = "lazy"
		NumberOfExecutionStreams = 1
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsWithStartegyLazy")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	err := cmd.Run()
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestValidateFlagsWithInvalidStrategy(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = true
		Strategy = "sdf"
		NumberOfExecutionStreams = 1
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsWithInvalidStrategy")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	err := cmd.Run()
	e, ok := err.(*exec.ExitError)
	c.Assert(ok, Equals, true)
	c.Assert(e.Success(), Equals, false)
}

func (s *MySuite) TestValidateFlagsWithInvalidStream(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = true
		NumberOfExecutionStreams = -1
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsWithInvalidStream")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	err := cmd.Run()
	e, ok := err.(*exec.ExitError)
	c.Assert(ok, Equals, true)
	c.Assert(e.Success(), Equals, false)
}
