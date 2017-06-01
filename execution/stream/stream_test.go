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

package stream

import (
	"errors"

	"testing"

	"os"

	"sync"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/order"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) SetUpTest(c *C) {
	wd, _ := os.Getwd()
	config.ProjectRoot = wd
}

func (s *MySuite) TestGetStatusForFailedScenario(c *C) {
	sce := &gm.ProtoScenario{
		ExecutionStatus: gm.ExecutionStatus_FAILED,
	}
	res := result.NewScenarioResult(sce)

	status := getStatus(res)

	c.Assert(status, Equals, gm.Result_FAILED)
}

func (s *MySuite) TestGetStatusForPassedScenario(c *C) {
	sce := &gm.ProtoScenario{
		ExecutionStatus: gm.ExecutionStatus_PASSED,
	}
	res := result.NewScenarioResult(sce)

	status := getStatus(res)

	c.Assert(status, Equals, gm.Result_PASSED)
}

func (s *MySuite) TestGetStatusForSkippedScenario(c *C) {
	sce := &gm.ProtoScenario{
		ExecutionStatus: gm.ExecutionStatus_SKIPPED,
	}
	res := result.NewScenarioResult(sce)

	status := getStatus(res)

	c.Assert(status, Equals, gm.Result_SKIPPED)
}

func (s *MySuite) TestGetErrorResponse(c *C) {
	errs := getErrorExecutionResponse(errors.New("error1"), errors.New("error2"))

	expected := &gm.ExecutionResponse{
		Type: gm.ExecutionResponse_ErrorResult,
		Result: &gm.Result{
			Errors: []*gm.Result_ExecutionError{
				{ErrorMessage: "error1"},
				{ErrorMessage: "error2"},
			},
		},
	}
	c.Assert(errs, DeepEquals, expected)
}

func (s *MySuite) TestGetHookFailureWhenNoFailure(c *C) {
	failure := getHookFailure(nil)

	var expected *gm.Result_ExecutionError
	c.Assert(failure, DeepEquals, expected)
}

func (s *MySuite) TestGetHookFailureWhenHookFailure(c *C) {
	hookFailure := []*gm.ProtoHookFailure{{ErrorMessage: "err msg"}}
	failure := getHookFailure(hookFailure)

	expected := &gm.Result_ExecutionError{ErrorMessage: "err msg"}
	c.Assert(failure, DeepEquals, expected)
}

func (s *MySuite) TestGetErrors(c *C) {
	items := []*gm.ProtoItem{newFailedStep("msg1")}
	errs := getErrors(items)

	expected := []*gm.Result_ExecutionError{{ErrorMessage: "msg1"}}
	c.Assert(errs, DeepEquals, expected)
}

func (s *MySuite) TestGetErrorsWithMultipleStepFailures(c *C) {
	items := []*gm.ProtoItem{
		newFailedStep("msg1"),
		newFailedStep("msg2"),
	}
	errs := getErrors(items)

	expected := []*gm.Result_ExecutionError{{ErrorMessage: "msg1"}, {ErrorMessage: "msg2"}}
	c.Assert(errs, DeepEquals, expected)
}

func (s *MySuite) TestGetErrorsWithConceptFailures(c *C) {
	items := []*gm.ProtoItem{
		{
			ItemType: gm.ProtoItem_Concept,
			Concept: &gm.ProtoConcept{
				Steps: []*gm.ProtoItem{newFailedStep("msg1")},
			},
		},
	}
	errs := getErrors(items)

	expected := []*gm.Result_ExecutionError{{ErrorMessage: "msg1"}}
	c.Assert(errs, DeepEquals, expected)
}

func (s *MySuite) TestGetErrorsWithNestedConceptFailures(c *C) {
	items := []*gm.ProtoItem{
		{
			ItemType: gm.ProtoItem_Concept,
			Concept: &gm.ProtoConcept{
				Steps: []*gm.ProtoItem{
					newFailedStep("msg1"),
					{
						ItemType: gm.ProtoItem_Concept,
						Concept: &gm.ProtoConcept{
							Steps: []*gm.ProtoItem{
								newFailedStep("msg2"),
							},
						},
					},
				},
			},
		},
	}
	errs := getErrors(items)

	expected := []*gm.Result_ExecutionError{{ErrorMessage: "msg1"}, {ErrorMessage: "msg2"}}
	c.Assert(errs, DeepEquals, expected)
}

func (s *MySuite) TestGetErrorsWithStepAndConceptFailures(c *C) {
	items := []*gm.ProtoItem{
		{
			ItemType: gm.ProtoItem_Concept,
			Concept: &gm.ProtoConcept{
				Steps: []*gm.ProtoItem{
					newFailedStep("msg1"),
					{
						ItemType: gm.ProtoItem_Concept,
						Concept: &gm.ProtoConcept{
							Steps: []*gm.ProtoItem{newFailedStep("msg2")},
						},
					},
				},
			},
		},
		newFailedStep("msg3"),
	}
	errs := getErrors(items)

	expected := []*gm.Result_ExecutionError{{ErrorMessage: "msg1"}, {ErrorMessage: "msg2"}, {ErrorMessage: "msg3"}}
	c.Assert(errs, DeepEquals, expected)
}

func (s *MySuite) TestListenSuiteStartExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gm.ExecutionResponse)
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	listenExecutionEvents(&dummyServer{response: actual}, 1234, wg)
	event.Notify(event.NewExecutionEvent(event.SuiteStart, nil, nil, 0, gm.ExecutionInfo{}))
	defer sendSuiteEnd(actual)

	expected := &gm.ExecutionResponse{
		Type:            gm.ExecutionResponse_SuiteStart,
		RunnerProcessId: 1234,
	}
	c.Assert(<-actual, DeepEquals, expected)

}

func (s *MySuite) TestListenSpecStartExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gm.ExecutionResponse)
	ei := gm.ExecutionInfo{
		CurrentSpec: &gm.SpecInfo{FileName: "example.spec"},
	}
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	listenExecutionEvents(&dummyServer{response: actual}, 1234, wg)
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.SpecStart, nil, nil, 0, ei))

	expected := &gm.ExecutionResponse{
		Type:            gm.ExecutionResponse_SpecStart,
		ID:              "example.spec",
		RunnerProcessId: 1234,
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenScenarioStartExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gm.ExecutionResponse)
	ei := gm.ExecutionInfo{
		CurrentSpec: &gm.SpecInfo{FileName: "example.spec"},
	}
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	listenExecutionEvents(&dummyServer{response: actual}, 1234, wg)
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.ScenarioStart, &gauge.Scenario{Heading: &gauge.Heading{LineNo: 1}}, nil, 0, ei))

	expected := &gm.ExecutionResponse{
		Type:            gm.ExecutionResponse_ScenarioStart,
		ID:              "example.spec:1",
		RunnerProcessId: 1234,
		Result: &gm.Result{
			TableRowNumber: 0,
		},
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenSpecEndExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gm.ExecutionResponse)
	ei := gm.ExecutionInfo{
		CurrentSpec: &gm.SpecInfo{FileName: "example.spec"},
	}
	hookFailure := []*gm.ProtoHookFailure{{ErrorMessage: "err msg"}}
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	listenExecutionEvents(&dummyServer{response: actual}, 1234, wg)
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.SpecEnd, nil, &result.SpecResult{
		ProtoSpec: &gm.ProtoSpec{PreHookFailures: hookFailure, PostHookFailures: hookFailure},
	}, 0, ei))

	expected := &gm.ExecutionResponse{
		Type:            gm.ExecutionResponse_SpecEnd,
		ID:              "example.spec",
		RunnerProcessId: 1234,
		Result: &gm.Result{
			BeforeHookFailure: &gm.Result_ExecutionError{ErrorMessage: "err msg"},
			AfterHookFailure:  &gm.Result_ExecutionError{ErrorMessage: "err msg"},
		},
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenSuiteEndExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gm.ExecutionResponse)
	hookFailure := &gm.ProtoHookFailure{ErrorMessage: "err msg"}
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	listenExecutionEvents(&dummyServer{response: actual}, 1234, wg)
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, &result.SuiteResult{PreSuite: hookFailure, PostSuite: hookFailure}, 0, gm.ExecutionInfo{}))

	expected := &gm.ExecutionResponse{
		Type:            gm.ExecutionResponse_SuiteEnd,
		RunnerProcessId: 1234,
		Result: &gm.Result{
			BeforeHookFailure: &gm.Result_ExecutionError{ErrorMessage: "err msg"},
			AfterHookFailure:  &gm.Result_ExecutionError{ErrorMessage: "err msg"},
		},
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenScenarioEndExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gm.ExecutionResponse)
	ei := gm.ExecutionInfo{
		CurrentSpec: &gm.SpecInfo{FileName: "example.spec"},
	}
	scn := &gm.ProtoScenario{
		ScenarioItems: []*gm.ProtoItem{},
		ExecutionTime: 1,
	}
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	listenExecutionEvents(&dummyServer{response: actual}, 1234, wg)
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.ScenarioEnd, &gauge.Scenario{Heading: &gauge.Heading{LineNo: 1}}, result.NewScenarioResult(scn), 0, ei))

	expected := &gm.ExecutionResponse{
		Type:            gm.ExecutionResponse_ScenarioEnd,
		RunnerProcessId: 1234,
		Result: &gm.Result{
			ExecutionTime:  1,
			Status:         gm.Result_PASSED,
			TableRowNumber: 0,
		},
		ID: "example.spec:1",
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenScenarioEndExecutionEventForFailedScenario(c *C) {
	event.InitRegistry()
	actual := make(chan *gm.ExecutionResponse)
	ei := gm.ExecutionInfo{
		CurrentSpec: &gm.SpecInfo{FileName: "example.spec"},
	}
	scn := &gm.ProtoScenario{
		ScenarioItems: []*gm.ProtoItem{
			newFailedStep("error message"),
		},
		ExecutionTime:   1,
		ExecutionStatus: gm.ExecutionStatus_FAILED,
	}
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	listenExecutionEvents(&dummyServer{response: actual}, 1234, wg)
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.ScenarioEnd, &gauge.Scenario{Heading: &gauge.Heading{LineNo: 1}}, result.NewScenarioResult(scn), 0, ei))

	expected := &gm.ExecutionResponse{
		Type:            gm.ExecutionResponse_ScenarioEnd,
		ID:              "example.spec:1",
		RunnerProcessId: 1234,
		Result: &gm.Result{
			ExecutionTime: 1,
			Status:        gm.Result_FAILED,
			Errors: []*gm.Result_ExecutionError{
				{
					ErrorMessage: "error message",
				},
			},
			TableRowNumber: 0,
		},
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestGetDataTableRowNumberWhenDataTableIsNotPresent(c *C) {
	scn := &gauge.Scenario{}

	number := getDataTableRowNumber(scn)

	c.Assert(number, Equals, 0)
}

func (s *MySuite) TestGetDataTableRowNumberWhenDataTableIsPresent(c *C) {
	scn := &gauge.Scenario{
		DataTableRow:      gauge.Table{},
		DataTableRowIndex: 2,
	}
	scn.DataTableRow.AddHeaders([]string{})

	number := getDataTableRowNumber(scn)

	c.Assert(number, Equals, 3)
}

func (s *MySuite) TestSetFlags(c *C) {
	req := &gm.ExecutionRequest{
		IsParallel:      true,
		Sort:            true,
		ParallelStreams: 3,
		Strategy:        gm.ExecutionRequest_EAGER,
		LogLevel:        gm.ExecutionRequest_DEBUG,
		TableRows:       "1-2",
		Tags:            "tag1 & tag2",
	}
	errs := setFlags(req)

	c.Assert(len(errs), Equals, 0)
	c.Assert(execution.Strategy, Equals, "eager")
	c.Assert(execution.NumberOfExecutionStreams, Equals, 3)
	c.Assert(reporter.NumberOfExecutionStreams, Equals, 3)
	c.Assert(filter.NumberOfExecutionStreams, Equals, 3)
	c.Assert(order.Sorted, Equals, true)
	c.Assert(filter.ExecuteTags, Equals, "tag1 & tag2")
	c.Assert(reporter.Verbose, Equals, true)
	c.Assert(reporter.IsParallel, Equals, true)
	c.Assert(execution.InParallel, Equals, true)
}

func (s *MySuite) TestSetFlagsWithInvalidNumberOfExecStreams(c *C) {
	req := &gm.ExecutionRequest{
		ParallelStreams: -3,
	}
	nCores := util.NumberOfCores()
	errs := setFlags(req)

	c.Assert(len(errs), Equals, 0)
	c.Assert(execution.NumberOfExecutionStreams, Equals, nCores)
	c.Assert(reporter.NumberOfExecutionStreams, Equals, nCores)
	c.Assert(filter.NumberOfExecutionStreams, Equals, nCores)
}

func (s *MySuite) TestResetFlags(c *C) {
	execution.Strategy = "HAHAH"
	reporter.IsParallel = true
	execution.InParallel = false
	reporter.Verbose = true
	filter.ExecuteTags = "sdfdsf"
	execution.NumberOfExecutionStreams = 1
	reporter.NumberOfExecutionStreams = 2
	filter.NumberOfExecutionStreams = 3
	order.Sorted = true
	resetFlags()

	cores := util.NumberOfCores()

	c.Assert(execution.Strategy, Equals, "lazy")
	c.Assert(execution.NumberOfExecutionStreams, Equals, cores)
	c.Assert(reporter.NumberOfExecutionStreams, Equals, cores)
	c.Assert(filter.NumberOfExecutionStreams, Equals, cores)
	c.Assert(order.Sorted, Equals, false)
	c.Assert(filter.ExecuteTags, Equals, "")
	c.Assert(reporter.Verbose, Equals, false)
	c.Assert(reporter.IsParallel, Equals, false)
	c.Assert(execution.InParallel, Equals, false)
}

type dummyServer struct {
	response chan *gm.ExecutionResponse
}

func (d *dummyServer) Send(r *gm.ExecutionResponse) error {
	d.response <- r
	return nil
}

func (d *dummyServer) Context() context.Context {
	return nil
}
func (d *dummyServer) SendMsg(m interface{}) error {
	return nil
}
func (d *dummyServer) RecvMsg(m interface{}) error {
	return nil
}

func (d *dummyServer) SendHeader(metadata.MD) error {
	return nil
}
func (d *dummyServer) SetTrailer(metadata.MD) {

}

func (d *dummyServer) SetHeader(md metadata.MD) error {
	return nil
}

func newFailedStep(msg string) *gm.ProtoItem {
	return &gm.ProtoItem{
		ItemType: gm.ProtoItem_Step,
		Step: &gm.ProtoStep{
			StepExecutionResult: &gm.ProtoStepExecutionResult{
				ExecutionResult: &gm.ProtoExecutionResult{
					Failed:       true,
					ErrorMessage: msg,
				},
			},
		},
	}
}

func sendSuiteEnd(actual chan *gm.ExecutionResponse) {
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, &result.SuiteResult{PreSuite: nil, PostSuite: nil}, 0, gm.ExecutionInfo{}))
	<-actual
}
