/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package result

import (
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	gc "gopkg.in/check.v1"
)

func (s *MySuite) TestAddScenarioResults(c *gc.C) {
	specItems := []*gauge_messages.ProtoItem{}
	protoSpec := &gauge_messages.ProtoSpec{
		Items: specItems,
	}
	specResult := SpecResult{
		ProtoSpec: protoSpec,
	}
	heading := "Scenario heading"
	item1 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}}
	item2 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}}
	step3Res := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{RecoverableError: true, Failed: false}}
	item3 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: step3Res}}
	items := []*gauge_messages.ProtoItem{item1, item2, item3}
	scenarioResult := NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: heading, ScenarioItems: items})
	results := make([]Result, 0)
	results = append(results, scenarioResult)

	specResult.AddScenarioResults(results)

	c.Assert(specResult.GetFailed(), gc.Equals, false)
	c.Assert(specResult.ScenarioCount, gc.Equals, 1)
	c.Assert(specResult.ProtoSpec.IsTableDriven, gc.Equals, false)
	c.Assert(specResult.ScenarioFailedCount, gc.Equals, 0)

}

func (s *MySuite) TestAddTableRelatedScenarioResult(c *gc.C) {
	specItems := []*gauge_messages.ProtoItem{}
	protoSpec := &gauge_messages.ProtoSpec{
		Items: specItems,
	}
	specResult := SpecResult{
		ProtoSpec: protoSpec,
	}
	heading1 := "Scenario heading 1"
	heading2 := "Scenario heading 2"
	item1 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}}
	item2 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{}}}}
	step3Res := &gauge_messages.ProtoStepExecutionResult{ExecutionResult: &gauge_messages.ProtoExecutionResult{RecoverableError: true, Failed: false}}
	item3 := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: &gauge_messages.ProtoStep{StepExecutionResult: step3Res}}
	items := []*gauge_messages.ProtoItem{item1, item2, item3}
	scenarioResult1 := NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: heading1, ScenarioItems: items})
	scenarioResult2 := NewScenarioResult(&gauge_messages.ProtoScenario{ScenarioHeading: heading2, ScenarioItems: items})
	scenarioResultsForIndex0 := []Result{scenarioResult1, scenarioResult2}
	scenarioResultsForIndex1 := []Result{scenarioResult1, scenarioResult2}
	results := make([][]Result, 0)
	results = append(results, scenarioResultsForIndex0)
	results = append(results, scenarioResultsForIndex1)

	specResult.AddTableRelatedScenarioResult(results, 1)

	c.Assert(specResult.GetFailed(), gc.Equals, false)
	c.Assert(specResult.ScenarioCount, gc.Equals, 2)
	c.Assert(specResult.ProtoSpec.IsTableDriven, gc.Equals, true)
	c.Assert(specResult.ScenarioFailedCount, gc.Equals, 0)
	c.Assert(specResult.ExecutionTime, gc.Equals, int64(0))
}
