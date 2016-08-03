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
	"errors"

	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestGetStatusForFailedScenario(c *C) {
	sce := &gauge_messages.ProtoScenario{
		Failed: proto.Bool(true),
	}
	res := result.NewScenarioResult(sce)

	status := getStatus(res)

	c.Assert(*status, Equals, *gauge_messages.ExecutionResponse_FAILED.Enum())
}

func (s *MySuite) TestGetStatusForPassedScenario(c *C) {
	sce := &gauge_messages.ProtoScenario{
		Failed: proto.Bool(false),
	}
	res := result.NewScenarioResult(sce)

	status := getStatus(res)

	c.Assert(*status, Equals, *gauge_messages.ExecutionResponse_PASSED.Enum())
}

func (s *MySuite) TestGetStatusForSkippedScenario(c *C) {
	sce := &gauge_messages.ProtoScenario{
		Skipped: proto.Bool(true),
	}
	res := result.NewScenarioResult(sce)

	status := getStatus(res)

	c.Assert(*status, Equals, *gauge_messages.ExecutionResponse_SKIPPED.Enum())
}

func (s *MySuite) TestGetErrorResponse(c *C) {
	errs := getErrorExecutionResponse(errors.New("error1"), errors.New("error2"))

	expected := &gauge_messages.ExecutionResponse{Type: gauge_messages.ExecutionResponse_ErrorResult.Enum(), Errors: []*gauge_messages.ExecutionResponse_ExecutionError{
		{ErrorMessage: proto.String("error1")}, {ErrorMessage: proto.String("error2")},
	}}

	c.Assert(errs, DeepEquals, expected)
}

func (s *MySuite) TestGetHookFailureWhenNoFailure(c *C) {
	failure := getHookFailure(nil)

	var expected *gauge_messages.ExecutionResponse_ExecutionError
	c.Assert(failure, DeepEquals, expected)
}

func (s *MySuite) TestGetHookFailureWhenHookFailure(c *C) {
	hookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: proto.String("err msg")}
	failure := getHookFailure(&hookFailure)

	expected := &gauge_messages.ExecutionResponse_ExecutionError{ErrorMessage: proto.String("err msg")}
	c.Assert(failure, DeepEquals, expected)
}

func (s *MySuite) TestGetErrors(c *C) {
	items := []*gauge_messages.ProtoItem{newFailedStep("msg1")}
	errors := getErrors(items)

	expected := []*gauge_messages.ExecutionResponse_ExecutionError{{ErrorMessage: proto.String("msg1")}}
	c.Assert(errors, DeepEquals, expected)
}

func (s *MySuite) TestGetErrorsWithMultipleStepFailures(c *C) {
	items := []*gauge_messages.ProtoItem{
		newFailedStep("msg1"),
		newFailedStep("msg2"),
	}
	errors := getErrors(items)

	expected := []*gauge_messages.ExecutionResponse_ExecutionError{{ErrorMessage: proto.String("msg1")}, {ErrorMessage: proto.String("msg2")}}
	c.Assert(errors, DeepEquals, expected)
}

func (s *MySuite) TestGetErrorsWithConceptFailures(c *C) {
	items := []*gauge_messages.ProtoItem{
		{
			ItemType: gauge_messages.ProtoItem_Concept.Enum(),
			Concept: &gauge_messages.ProtoConcept{
				Steps: []*gauge_messages.ProtoItem{newFailedStep("msg1")},
			},
		},
	}
	errors := getErrors(items)

	expected := []*gauge_messages.ExecutionResponse_ExecutionError{{ErrorMessage: proto.String("msg1")}}
	c.Assert(errors, DeepEquals, expected)
}

func (s *MySuite) TestGetErrorsWithNestedConceptFailures(c *C) {
	items := []*gauge_messages.ProtoItem{
		{
			ItemType: gauge_messages.ProtoItem_Concept.Enum(),
			Concept: &gauge_messages.ProtoConcept{
				Steps: []*gauge_messages.ProtoItem{
					newFailedStep("msg1"),
					{
						ItemType: gauge_messages.ProtoItem_Concept.Enum(),
						Concept: &gauge_messages.ProtoConcept{
							Steps: []*gauge_messages.ProtoItem{
								newFailedStep("msg2"),
							},
						},
					},
				},
			},
		},
	}
	errors := getErrors(items)

	expected := []*gauge_messages.ExecutionResponse_ExecutionError{{ErrorMessage: proto.String("msg1")}, {ErrorMessage: proto.String("msg2")}}
	c.Assert(errors, DeepEquals, expected)
}

func (s *MySuite) TestGetErrorsWithStepAndConceptFailures(c *C) {
	items := []*gauge_messages.ProtoItem{
		{
			ItemType: gauge_messages.ProtoItem_Concept.Enum(),
			Concept: &gauge_messages.ProtoConcept{
				Steps: []*gauge_messages.ProtoItem{
					newFailedStep("msg1"),
					{
						ItemType: gauge_messages.ProtoItem_Concept.Enum(),
						Concept: &gauge_messages.ProtoConcept{
							Steps: []*gauge_messages.ProtoItem{newFailedStep("msg2")},
						},
					},
				},
			},
		},
		newFailedStep("msg3"),
	}
	errors := getErrors(items)

	expected := []*gauge_messages.ExecutionResponse_ExecutionError{{ErrorMessage: proto.String("msg1")}, {ErrorMessage: proto.String("msg2")}, {ErrorMessage: proto.String("msg3")}}
	c.Assert(errors, DeepEquals, expected)
}

func (s *MySuite) TestListenSuiteStartExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gauge_messages.ExecutionResponse)

	listenExecutionEvents(&dummyServer{response: actual})
	event.Notify(event.NewExecutionEvent(event.SuiteStart, nil, nil, 0, gauge_messages.ExecutionInfo{}))
	defer sendSuiteEnd(actual)

	expected := &gauge_messages.ExecutionResponse{
		Type: gauge_messages.ExecutionResponse_SuiteStart.Enum(),
	}
	c.Assert(<-actual, DeepEquals, expected)

}

func (s *MySuite) TestListenSpecStartExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gauge_messages.ExecutionResponse)
	ei := gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{FileName: proto.String("example.spec")},
	}

	listenExecutionEvents(&dummyServer{response: actual})
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.SpecStart, nil, nil, 0, ei))

	expected := &gauge_messages.ExecutionResponse{
		Type: gauge_messages.ExecutionResponse_SpecStart.Enum(),
		ID:   proto.String("example.spec"),
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenScenarioStartExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gauge_messages.ExecutionResponse)
	ei := gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{FileName: proto.String("example.spec")},
	}

	listenExecutionEvents(&dummyServer{response: actual})
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.ScenarioStart, &gauge.Scenario{Heading: &gauge.Heading{LineNo: 1}}, nil, 0, ei))

	expected := &gauge_messages.ExecutionResponse{
		Type: gauge_messages.ExecutionResponse_ScenarioStart.Enum(),
		ID:   proto.String("example.spec:1"),
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenSpecEndExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gauge_messages.ExecutionResponse)
	ei := gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{FileName: proto.String("example.spec")},
	}
	hookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: proto.String("err msg")}

	listenExecutionEvents(&dummyServer{response: actual})
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.SpecEnd, nil, &result.SpecResult{
		ProtoSpec: &gauge_messages.ProtoSpec{PreHookFailure: hookFailure, PostHookFailure: hookFailure},
	}, 0, ei))

	expected := &gauge_messages.ExecutionResponse{
		Type:              gauge_messages.ExecutionResponse_SpecEnd.Enum(),
		ID:                proto.String("example.spec"),
		BeforeHookFailure: &gauge_messages.ExecutionResponse_ExecutionError{ErrorMessage: proto.String("err msg")},
		AfterHookFailure:  &gauge_messages.ExecutionResponse_ExecutionError{ErrorMessage: proto.String("err msg")},
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenSuiteEndExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gauge_messages.ExecutionResponse)
	hookFailure := &gauge_messages.ProtoHookFailure{ErrorMessage: proto.String("err msg")}

	listenExecutionEvents(&dummyServer{response: actual})
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, &result.SuiteResult{PreSuite: hookFailure, PostSuite: hookFailure}, 0, gauge_messages.ExecutionInfo{}))

	expected := &gauge_messages.ExecutionResponse{
		Type:              gauge_messages.ExecutionResponse_SuiteEnd.Enum(),
		BeforeHookFailure: &gauge_messages.ExecutionResponse_ExecutionError{ErrorMessage: proto.String("err msg")},
		AfterHookFailure:  &gauge_messages.ExecutionResponse_ExecutionError{ErrorMessage: proto.String("err msg")},
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenScenarioEndExecutionEvent(c *C) {
	event.InitRegistry()
	actual := make(chan *gauge_messages.ExecutionResponse)
	ei := gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{FileName: proto.String("example.spec")},
	}
	scn := &gauge_messages.ProtoScenario{
		ScenarioItems: []*gauge_messages.ProtoItem{},
		ExecutionTime: proto.Int64(1),
	}

	listenExecutionEvents(&dummyServer{response: actual})
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.ScenarioEnd, &gauge.Scenario{Heading: &gauge.Heading{LineNo: 1}}, result.NewScenarioResult(scn), 0, ei))

	expected := &gauge_messages.ExecutionResponse{
		Type:          gauge_messages.ExecutionResponse_ScenarioEnd.Enum(),
		ExecutionTime: proto.Int64(1),
		Status:        gauge_messages.ExecutionResponse_PASSED.Enum(),
		ID:            proto.String("example.spec:1"),
	}
	c.Assert(<-actual, DeepEquals, expected)
}

func (s *MySuite) TestListenScenarioEndExecutionEventForFailedScenario(c *C) {
	event.InitRegistry()
	actual := make(chan *gauge_messages.ExecutionResponse)
	ei := gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{FileName: proto.String("example.spec")},
	}
	scn := &gauge_messages.ProtoScenario{
		ScenarioItems: []*gauge_messages.ProtoItem{
			newFailedStep("error message"),
		},
		ExecutionTime: proto.Int64(1),
		Failed:        proto.Bool(true),
	}

	listenExecutionEvents(&dummyServer{response: actual})
	defer sendSuiteEnd(actual)
	event.Notify(event.NewExecutionEvent(event.ScenarioEnd, &gauge.Scenario{Heading: &gauge.Heading{LineNo: 1}}, result.NewScenarioResult(scn), 0, ei))

	expected := &gauge_messages.ExecutionResponse{
		Type:          gauge_messages.ExecutionResponse_ScenarioEnd.Enum(),
		ExecutionTime: proto.Int64(1),
		Status:        gauge_messages.ExecutionResponse_FAILED.Enum(),
		ID:            proto.String("example.spec:1"),
		Errors: []*gauge_messages.ExecutionResponse_ExecutionError{
			&gauge_messages.ExecutionResponse_ExecutionError{
				ErrorMessage: proto.String("error message"),
			},
		},
	}
	c.Assert(<-actual, DeepEquals, expected)
}

type dummyServer struct {
	response chan *gauge_messages.ExecutionResponse
}

func (d *dummyServer) Send(r *gauge_messages.ExecutionResponse) error {
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

func newFailedStep(msg string) *gauge_messages.ProtoItem {
	return &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Step.Enum(),
		Step: &gauge_messages.ProtoStep{
			StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{
				ExecutionResult: &gauge_messages.ProtoExecutionResult{
					Failed:       proto.Bool(true),
					ErrorMessage: proto.String(msg),
				},
			},
		},
	}
}

func sendSuiteEnd(actual chan *gauge_messages.ExecutionResponse) {
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, &result.SuiteResult{PreSuite: nil, PostSuite: nil}, 0, gauge_messages.ExecutionInfo{}))
	<-actual
}
