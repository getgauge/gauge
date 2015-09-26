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
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	. "gopkg.in/check.v1"
	"testing"
	"fmt"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestDistributionOfSpecs(c *C) {
	specs := createSpecsList(10)
	specCollections := filter.DistributeSpecs(specs, 10)
	c.Assert(len(specCollections), Equals, 10)
	verifySpecCollectionsForSize(c, 1, specCollections...)

	specCollections = filter.DistributeSpecs(specs, 5)
	c.Assert(len(specCollections), Equals, 5)
	verifySpecCollectionsForSize(c, 2, specCollections...)

	specCollections = filter.DistributeSpecs(specs, 4)
	c.Assert(len(specCollections), Equals, 4)
	verifySpecCollectionsForSize(c, 3, specCollections[:2]...)
	verifySpecCollectionsForSize(c, 2, specCollections[2:]...)

	specCollections = filter.DistributeSpecs(specs, 3)
	c.Assert(len(specCollections), Equals, 3)
	verifySpecCollectionsForSize(c, 4, specCollections[0])
	verifySpecCollectionsForSize(c, 3, specCollections[1:]...)
}

func (s *MySuite) TestDistributionOfSpecsWithMoreNumberOfDistributions(c *C) {
	specs := createSpecsList(6)
	e := parallelSpecExecution{numberOfExecutionStreams: 10, specifications: specs}
	specCollections := filter.DistributeSpecs(specs, e.getNumberOfStreams())
	c.Assert(len(specCollections), Equals, 6)
	verifySpecCollectionsForSize(c, 1, specCollections...)

	e.numberOfExecutionStreams = 17
	specCollections = filter.DistributeSpecs(specs, e.getNumberOfStreams())
	c.Assert(len(specCollections), Equals, 6)
	verifySpecCollectionsForSize(c, 1, specCollections...)

	e.numberOfExecutionStreams = 17
	specs = createSpecsList(0)
	e.specifications = specs
	specCollections = filter.DistributeSpecs(specs, e.getNumberOfStreams())
	c.Assert(len(specCollections), Equals, 0)
}

func createSpecsList(number int) []*parser.Specification {
	specs := make([]*parser.Specification, 0)
	for i := 0; i < number; i++ {
		specs = append(specs, &parser.Specification{FileName: fmt.Sprint("spec", i)})
	}
	return specs
}

func verifySpecCollectionsForSize(c *C, size int, specCollections ...*filter.SpecCollection) {
	for _, collection := range specCollections {
		c.Assert(len(collection.Specs), Equals, size)
	}
}

func (s *MySuite) TestAggregationOfSuiteResult(c *C) {
	e := parallelSpecExecution{}
	suiteRes1 := &result.SuiteResult{ExecutionTime: 1, SpecsFailedCount: 1, IsFailed: true, SpecResults: []*result.SpecResult{&result.SpecResult{}, &result.SpecResult{}}}
	suiteRes2 := &result.SuiteResult{ExecutionTime: 3, SpecsFailedCount: 0, IsFailed: false, SpecResults: []*result.SpecResult{&result.SpecResult{}, &result.SpecResult{}}}
	suiteRes3 := &result.SuiteResult{ExecutionTime: 5, SpecsFailedCount: 0, IsFailed: false, SpecResults: []*result.SpecResult{&result.SpecResult{}, &result.SpecResult{}}}
	suiteResults := make([]*result.SuiteResult, 0)
	suiteResults = append(suiteResults, suiteRes1, suiteRes2, suiteRes3)

	aggregatedRes := e.aggregateResults(suiteResults)
	c.Assert(aggregatedRes.ExecutionTime, Equals, int64(9))
	c.Assert(aggregatedRes.SpecsFailedCount, Equals, 1)
	c.Assert(aggregatedRes.IsFailed, Equals, true)
	c.Assert(len(aggregatedRes.SpecResults), Equals, 6)
}

func (s *MySuite) TestAggregationOfSuiteResultWithUnhandledErrors(c *C) {
	e := parallelSpecExecution{}
	suiteRes1 := &result.SuiteResult{IsFailed: true, UnhandledErrors: []error{streamExecError{specsSkipped: []string{"spec1", "spec2"}, message: "Runner failed to start"}}}
	suiteRes2 := &result.SuiteResult{IsFailed: false, UnhandledErrors: []error{streamExecError{specsSkipped: []string{"spec3", "spec4"}, message: "Runner failed to start"}}}
	suiteRes3 := &result.SuiteResult{IsFailed: false}
	suiteResults := make([]*result.SuiteResult, 0)
	suiteResults = append(suiteResults, suiteRes1, suiteRes2, suiteRes3)

	aggregatedRes := e.aggregateResults(suiteResults)
	c.Assert(len(aggregatedRes.UnhandledErrors), Equals, 2)
	c.Assert(aggregatedRes.UnhandledErrors[0].Error(), Equals, "The following specifications could not be executed:\n"+
		"spec1\n"+
		"spec2\n"+
		"Reason : Runner failed to start.")
	c.Assert(aggregatedRes.UnhandledErrors[1].Error(), Equals, "The following specifications could not be executed:\n"+
		"spec3\n"+
		"spec4\n"+
		"Reason : Runner failed to start.")
	err := (aggregatedRes.UnhandledErrors[0]).(streamExecError)
	c.Assert(len(err.specsSkipped), Equals, 2)
}

func (s *MySuite) TestAggregationOfSuiteResultWithHook(c *C) {
	e := parallelSpecExecution{}
	suiteRes1 := &result.SuiteResult{PreSuite: &gauge_messages.ProtoHookFailure{}}
	suiteRes2 := &result.SuiteResult{PreSuite: &gauge_messages.ProtoHookFailure{}}
	suiteRes3 := &result.SuiteResult{PostSuite: &gauge_messages.ProtoHookFailure{}}
	suiteResults := make([]*result.SuiteResult, 0)
	suiteResults = append(suiteResults, suiteRes1, suiteRes2, suiteRes3)

	aggregatedRes := e.aggregateResults(suiteResults)
	c.Assert(aggregatedRes.PreSuite, Equals, suiteRes2.PreSuite)
	c.Assert(aggregatedRes.PostSuite, Equals, suiteRes3.PostSuite)
}

func (s *MySuite) TestFunctionsOfTypeSpecList(c *C) {
	mySpecs := &specList{specs: createSpecsList(4)}
	c.Assert(mySpecs.getSpec().FileName, Equals, "spec0")
	c.Assert(mySpecs.getSpec().FileName, Equals, "spec1")
	c.Assert(mySpecs.isEmpty(), Equals, false)
	c.Assert(len(mySpecs.specs), Equals, 2)
	c.Assert(mySpecs.getSpec().FileName, Equals, "spec2")
	c.Assert(mySpecs.getSpec().FileName, Equals, "spec3")
	c.Assert(mySpecs.isEmpty(), Equals, true)
}

