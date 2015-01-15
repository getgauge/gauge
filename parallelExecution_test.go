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
