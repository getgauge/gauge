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
	"github.com/getgauge/gauge/gauge_messages"
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

func (s *MySuite) TestAggregationOfSuiteResult(c *C) {
	e := parallelSpecExecution{}
	suiteRes1 := &suiteResult{executionTime: 1, specsFailedCount: 1, isFailed: true, specResults: []*specResult{&specResult{}, &specResult{}}}
	suiteRes2 := &suiteResult{executionTime: 3, specsFailedCount: 0, isFailed: false, specResults: []*specResult{&specResult{}, &specResult{}}}
	suiteRes3 := &suiteResult{executionTime: 5, specsFailedCount: 0, isFailed: false, specResults: []*specResult{&specResult{}, &specResult{}}}
	suiteResults := make([]*suiteResult, 0)
	suiteResults = append(suiteResults, suiteRes1, suiteRes2, suiteRes3)

	aggregatedRes := e.aggregateResults(suiteResults)
	c.Assert(aggregatedRes.executionTime, Equals, int64(9))
	c.Assert(aggregatedRes.specsFailedCount, Equals, 1)
	c.Assert(aggregatedRes.isFailed, Equals, true)
	c.Assert(len(aggregatedRes.specResults), Equals, 6)
}

func (s *MySuite) TestAggregationOfSuiteResultWithUnhandledErrors(c *C) {
	e := parallelSpecExecution{}
	suiteRes1 := &suiteResult{isFailed: true, unhandledErrors: []error{streamExecError{specsSkipped: []string{"spec1", "spec2"}, message: "Runner failed to start"}}}
	suiteRes2 := &suiteResult{isFailed: false, unhandledErrors: []error{streamExecError{specsSkipped: []string{"spec3", "spec4"}, message: "Runner failed to start"}}}
	suiteRes3 := &suiteResult{isFailed: false}
	suiteResults := make([]*suiteResult, 0)
	suiteResults = append(suiteResults, suiteRes1, suiteRes2, suiteRes3)

	aggregatedRes := e.aggregateResults(suiteResults)
	c.Assert(len(aggregatedRes.unhandledErrors), Equals, 2)
	c.Assert(aggregatedRes.unhandledErrors[0].Error(), Equals, "The following specifications are not executed: [spec1 spec2]. Reason: Runner failed to start")
	c.Assert(aggregatedRes.unhandledErrors[1].Error(), Equals, "The following specifications are not executed: [spec3 spec4]. Reason: Runner failed to start")
}

func (s *MySuite) TestAggregationOfSuiteResultWithHook(c *C) {
	e := parallelSpecExecution{}
	suiteRes1 := &suiteResult{preSuite: &gauge_messages.ProtoHookFailure{}}
	suiteRes2 := &suiteResult{preSuite: &gauge_messages.ProtoHookFailure{}}
	suiteRes3 := &suiteResult{postSuite: &gauge_messages.ProtoHookFailure{}}
	suiteResults := make([]*suiteResult, 0)
	suiteResults = append(suiteResults, suiteRes1, suiteRes2, suiteRes3)

	aggregatedRes := e.aggregateResults(suiteResults)
	c.Assert(aggregatedRes.preSuite, Equals, suiteRes2.preSuite)
	c.Assert(aggregatedRes.postSuite, Equals, suiteRes3.postSuite)
}
