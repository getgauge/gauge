/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package result

import (
	"testing"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type MySuite struct{}

var _ = gc.Suite(&MySuite{})

func (s *MySuite) TestUpdateConceptExecutionResultWithARecoverableStep(c *gc.C) {
	cptStep := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}
	item1 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}}
	step2Res := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{RecoverableError: true, Failed: true}}
	item2 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: step2Res}}
	item3 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}}
	cptRes := NewConceptResult(&gauge_messages.ProtoConcept{ConceptStep: cptStep, Steps: []*gauge_messages.ProtoItem{item1, item2, item3}})

	cptRes.UpdateConceptExecResult()

	c.Assert(cptRes.GetFailed(), gc.Equals, true)
	c.Assert(cptRes.GetRecoverable(), gc.Equals, true)
}

func (s *MySuite) TestUpdateConceptExecutionResultWithNonRecoverableFailure(c *gc.C) {
	cptStep := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}
	item1 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}}
	step2Res := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "step failure"}}
	item2 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: step2Res}}
	item3 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}}
	cptRes := NewConceptResult(&gauge_messages.ProtoConcept{ConceptStep: cptStep, Steps: []*gauge_messages.ProtoItem{item1, item2, item3}})

	cptRes.UpdateConceptExecResult()

	c.Assert(cptRes.GetFailed(), gc.Equals, true)
	c.Assert(cptRes.GetRecoverable(), gc.Equals, false)
	c.Assert(cptRes.ProtoConcept.GetConceptExecutionResult().GetExecutionResult().GetErrorMessage(), gc.Equals, "step failure")
}

func (s *MySuite) TestUpdateConceptExecutionResultWithRecoverableAndNonRecoverableSteps(c *gc.C) {
	cptStep := &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}
	step1Res := &gauge_messages.ProtoExecutionResult{Failed: true, RecoverableError: true, ErrorMessage: "a recoverable step"}
	item1 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: step1Res}}}
	step2Res := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: "step failure"}}
	item2 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: step2Res}}
	item3 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}}
	cptRes := NewConceptResult(&gauge_messages.ProtoConcept{ConceptStep: cptStep, Steps: []*gauge_messages.ProtoItem{item1, item2, item3}})

	cptRes.UpdateConceptExecResult()

	c.Assert(cptRes.GetFailed(), gc.Equals, true)
	c.Assert(cptRes.GetRecoverable(), gc.Equals, false)
	c.Assert(cptRes.ProtoConcept.GetConceptExecutionResult().GetExecutionResult().GetErrorMessage(), gc.Equals, "step failure")
}
