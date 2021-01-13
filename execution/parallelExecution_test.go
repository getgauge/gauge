/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"testing"

	"net"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/runner"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestNumberOfStreams(c *C) {
	specs := createSpecsList(6)
	e := parallelExecution{numberOfExecutionStreams: 5, specCollection: gauge.NewSpecCollection(specs, false)}
	c.Assert(e.numberOfStreams(), Equals, 5)

	specs = createSpecsList(6)
	e = parallelExecution{numberOfExecutionStreams: 10, specCollection: gauge.NewSpecCollection(specs, false)}
	c.Assert(e.numberOfStreams(), Equals, 6)

	specs = createSpecsList(0)
	e = parallelExecution{numberOfExecutionStreams: 17, specCollection: gauge.NewSpecCollection(specs, false)}
	c.Assert(e.numberOfStreams(), Equals, 0)
}

func getValidationErrorMap() *gauge.BuildErrors {
	return &gauge.BuildErrors{
		SpecErrs:     make(map[*gauge.Specification][]error),
		ScenarioErrs: make(map[*gauge.Scenario][]error),
		StepErrs:     make(map[*gauge.Step]error),
	}
}

func (s *MySuite) TestAggregationOfSuiteResult(c *C) {
	e := parallelExecution{errMaps: getValidationErrorMap()}
	suiteRes1 := &result.SuiteResult{ExecutionTime: 1, SpecsFailedCount: 1, IsFailed: true, SpecResults: []*result.SpecResult{{}, {}}}
	suiteRes2 := &result.SuiteResult{ExecutionTime: 3, SpecsFailedCount: 0, IsFailed: false, SpecResults: []*result.SpecResult{{}, {}}}
	suiteRes3 := &result.SuiteResult{ExecutionTime: 5, SpecsFailedCount: 0, IsFailed: false, SpecResults: []*result.SpecResult{{}, {}}}
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
	suiteRes4 := &result.SuiteResult{SpecResults: []*result.SpecResult{{Skipped: true}}}
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

func (s *MySuite) TestIsMultiThreadedWithEnvSetToFalse(c *C) {
	e := parallelExecution{errMaps: getValidationErrorMap()}

	env.EnableMultiThreadedExecution = func() bool { return false }

	c.Assert(false, Equals, e.isMultithreaded())
}

func (s *MySuite) TestIsMultiThreadedWithRunnerWhenSupportsMultithreading(c *C) {
	e := parallelExecution{errMaps: getValidationErrorMap(), runners: []runner.Runner{&fakeRunner{isMultiThreaded: true}}}

	env.EnableMultiThreadedExecution = func() bool { return true }

	c.Assert(true, Equals, e.isMultithreaded())
}

func (s *MySuite) TestIsMultiThreadedWithRunnerWhenDoesNotSupportMultithreading(c *C) {
	e := parallelExecution{errMaps: getValidationErrorMap(), runners: []runner.Runner{&fakeRunner{isMultiThreaded: false}}}

	env.EnableMultiThreadedExecution = func() bool { return true }

	c.Assert(false, Equals, e.isMultithreaded())
}

type fakeRunner struct {
	isMultiThreaded bool
}

func (f *fakeRunner) ExecuteMessageWithTimeout(m *gauge_messages.Message) (*gauge_messages.Message, error) {
	return nil, nil
}

func (f *fakeRunner) ExecuteAndGetStatus(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
	return nil
}
func (f *fakeRunner) Alive() bool {
	return false
}
func (f *fakeRunner) Kill() error {
	return nil
}
func (f *fakeRunner) Connection() net.Conn {
	return nil
}
func (f *fakeRunner) IsMultithreaded() bool {
	return f.isMultiThreaded
}

func (f *fakeRunner) Info() *runner.RunnerInfo {
	return nil
}
func (f *fakeRunner) Pid() int {
	return 0
}
