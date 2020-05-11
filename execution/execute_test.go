/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"fmt"

	"github.com/getgauge/gauge/gauge"

	. "gopkg.in/check.v1"
)

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
