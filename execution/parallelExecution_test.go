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
	"testing"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/validation"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestNumberOfStreams(c *C) {
	specs := createSpecsList(6)
	e := parallelExecution{numberOfExecutionStreams: 5, specCollection: gauge.NewSpecCollection(specs)}
	c.Assert(e.numberOfStreams(), Equals, 5)

	specs = createSpecsList(6)
	e = parallelExecution{numberOfExecutionStreams: 10, specCollection: gauge.NewSpecCollection(specs)}
	c.Assert(e.numberOfStreams(), Equals, 6)

	specs = createSpecsList(0)
	e = parallelExecution{numberOfExecutionStreams: 17, specCollection: gauge.NewSpecCollection(specs)}
	c.Assert(e.numberOfStreams(), Equals, 0)
}

func getValidationErrorMap() *validation.ValidationErrMaps {
	return &validation.ValidationErrMaps{
		SpecErrs:     make(map[*gauge.Specification][]error),
		ScenarioErrs: make(map[*gauge.Scenario][]error),
		StepErrs:     make(map[*gauge.Step]error),
	}
}

func (s *MySuite) TestAggregationOfSuiteResult(c *C) {
	e := parallelExecution{errMaps: getValidationErrorMap()}
	suiteRes1 := &result.SuiteResult{ExecutionTime: 1, SpecsFailedCount: 1, IsFailed: true, SpecResults: []*result.SpecResult{&result.SpecResult{}, &result.SpecResult{}}}
	suiteRes2 := &result.SuiteResult{ExecutionTime: 3, SpecsFailedCount: 0, IsFailed: false, SpecResults: []*result.SpecResult{&result.SpecResult{}, &result.SpecResult{}}}
	suiteRes3 := &result.SuiteResult{ExecutionTime: 5, SpecsFailedCount: 0, IsFailed: false, SpecResults: []*result.SpecResult{&result.SpecResult{}, &result.SpecResult{}}}
	var suiteResults []*result.SuiteResult
	suiteResults = append(suiteResults, suiteRes1, suiteRes2, suiteRes3)
	e.aggregateResults(suiteResults)

	aggregatedRes := e.suiteResult
	c.Assert(aggregatedRes.SpecsFailedCount, Equals, 1)
	c.Assert(aggregatedRes.IsFailed, Equals, true)
	c.Assert(len(aggregatedRes.SpecResults), Equals, 6)
	c.Assert(aggregatedRes.SpecsSkippedCount, Equals, 0)
}

func (s *MySuite) TestAggregationOfSuiteResultWithUnhandledErrors(c *C) {
	e := parallelExecution{}
	suiteRes1 := &result.SuiteResult{IsFailed: true, UnhandledErrors: []error{streamExecError{specsSkipped: []string{"spec1", "spec2"}, message: "Runner failed to start"}}}
	suiteRes2 := &result.SuiteResult{IsFailed: false, UnhandledErrors: []error{streamExecError{specsSkipped: []string{"spec3", "spec4"}, message: "Runner failed to start"}}}
	suiteRes3 := &result.SuiteResult{IsFailed: false}
	suiteRes4 := &result.SuiteResult{SpecResults: []*result.SpecResult{&result.SpecResult{Skipped: true}}}
	var suiteResults []*result.SuiteResult
	suiteResults = append(suiteResults, suiteRes1, suiteRes2, suiteRes3, suiteRes4)
	e.aggregateResults(suiteResults)

	aggregatedRes := e.suiteResult
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
	c.Assert(aggregatedRes.SpecsSkippedCount, Equals, 1)
}

func (s *MySuite) TestAggregationOfSuiteResultWithHook(c *C) {
	e := parallelExecution{errMaps: getValidationErrorMap()}
	suiteRes1 := &result.SuiteResult{PreSuite: &gauge_messages.ProtoHookFailure{}}
	suiteRes2 := &result.SuiteResult{PreSuite: &gauge_messages.ProtoHookFailure{}}
	suiteRes3 := &result.SuiteResult{PostSuite: &gauge_messages.ProtoHookFailure{}}
	var suiteResults []*result.SuiteResult
	suiteResults = append(suiteResults, suiteRes1, suiteRes2, suiteRes3)
	e.aggregateResults(suiteResults)

	aggregatedRes := e.suiteResult
	c.Assert(aggregatedRes.PreSuite, Equals, suiteRes2.PreSuite)
	c.Assert(aggregatedRes.PostSuite, Equals, suiteRes3.PostSuite)
}
