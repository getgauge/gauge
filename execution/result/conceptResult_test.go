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

package result

import (
	"testing"

	"github.com/getgauge/gauge/gauge_messages"
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
