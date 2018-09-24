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

	"github.com/getgauge/gauge/gauge"

	. "gopkg.in/check.v1"
)

type testLogger struct {
	output string
}

func (l *testLogger) Write(b []byte) (int, error) {
	l.output = string(b)
	return len(b), nil
}

func (s *MySuite) TestFunctionsOfTypeSpecList(c *C) {
	mySpecs := gauge.NewSpecCollection(createSpecsList(4), false)
	c.Assert(mySpecs.Next()[0].FileName, Equals, "spec0")
	c.Assert(mySpecs.Next()[0].FileName, Equals, "spec1")
	c.Assert(mySpecs.HasNext(), Equals, true)
	c.Assert(mySpecs.Size(), Equals, 4)
	c.Assert(mySpecs.Next()[0].FileName, Equals, "spec2")
	c.Assert(mySpecs.Next()[0].FileName, Equals, "spec3")
	c.Assert(mySpecs.HasNext(), Equals, false)
}

func createSpecsList(number int) []*gauge.Specification {
	var specs []*gauge.Specification
	for i := 0; i < number; i++ {
		specs = append(specs, &gauge.Specification{FileName: fmt.Sprint("spec", i)})
	}
	return specs
}

func (s *MySuite) TestValidateFlagsIfNotParallel(c *C) {
	InParallel = false
	err := validateFlags()
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestValidateFlagsWithStartegyEager(c *C) {
	InParallel = true
	Strategy = "eager"
	NumberOfExecutionStreams = 1
	err := validateFlags()
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestValidateFlagsWithStartegyLazy(c *C) {
	InParallel = true
	Strategy = "lazy"
	NumberOfExecutionStreams = 1
	err := validateFlags()
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestValidateFlagsWithInvalidStrategy(c *C) {
	InParallel = true
	Strategy = "sdf"
	NumberOfExecutionStreams = 1
	err := validateFlags()
	c.Assert(err.Error(), Equals, "invalid input(sdf) to --strategy flag")
}

func (s *MySuite) TestValidateFlagsWithInvalidStream(c *C) {
	InParallel = true
	NumberOfExecutionStreams = -1
	err := validateFlags()
	c.Assert(err.Error(), Equals, "invalid input(-1) to --n flag")
}
