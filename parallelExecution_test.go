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

package main

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestDistributionOfSpecs(c *C) {
	e := parallelSpecExecution{specifications: createSpecsList(10)}
	specCollections := e.distributeSpecs(10)
	c.Assert(len(specCollections), Equals, 10)
	verifySpecCollectionsForSize(c, 1, specCollections...)

	specCollections = e.distributeSpecs(5)
	c.Assert(len(specCollections), Equals, 5)
	verifySpecCollectionsForSize(c, 2, specCollections...)

	specCollections = e.distributeSpecs(4)
	c.Assert(len(specCollections), Equals, 4)
	verifySpecCollectionsForSize(c, 3, specCollections[:2]...)
	verifySpecCollectionsForSize(c, 2, specCollections[2:]...)

	specCollections = e.distributeSpecs(3)
	c.Assert(len(specCollections), Equals, 3)
	verifySpecCollectionsForSize(c, 4, specCollections[0])
	verifySpecCollectionsForSize(c, 3, specCollections[1:]...)
}

func (s *MySuite) TestDistributionOfSpecsWithMoreNumberOfDistributions(c *C) {
	e := parallelSpecExecution{specifications: createSpecsList(6)}
	specCollections := e.distributeSpecs(10)
	c.Assert(len(specCollections), Equals, 6)
	verifySpecCollectionsForSize(c, 1, specCollections...)

	specCollections = e.distributeSpecs(17)
	c.Assert(len(specCollections), Equals, 6)
	verifySpecCollectionsForSize(c, 1, specCollections...)

	e = parallelSpecExecution{specifications: createSpecsList(0)}
	specCollections = e.distributeSpecs(17)
	c.Assert(len(specCollections), Equals, 0)
}

func createSpecsList(number int) []*specification {
	specs := make([]*specification, 0)
	for i := 0; i < number; i++ {
		specs = append(specs, &specification{})
	}
	return specs
}

func verifySpecCollectionsForSize(c *C, size int, specCollections ...*specCollection) {
	for _, collection := range specCollections {
		c.Assert(len(collection.specs), Equals, size)
	}
}
