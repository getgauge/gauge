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

// A scenario that owns a scenario-level data table runs once per row. The
// caller (executeScenarioTableDrivenScenarios) is responsible for counting the
// scenario as failed at most once, so AddTableDrivenScenarioResult must not bump
// ScenarioFailedCount per row - otherwise passed = executed - failed goes
// negative and passing scenarios vanish from the summary (issue #1802).
func (s *MySuite) TestAddTableDrivenScenarioResultDoesNotBumpFailedCountPerRow(c *gc.C) {
	specResult := SpecResult{
		ProtoSpec: &gauge_messages.ProtoSpec{Items: []*gauge_messages.ProtoItem{}},
	}
	heading := "Scenario heading"
	table := &gauge_messages.ProtoTable{}

	// Two failing rows of the same scenario-data-table scenario.
	for rowIndex := 0; rowIndex < 2; rowIndex++ {
		failedScenario := &gauge_messages.ProtoScenario{
			ScenarioHeading: heading,
			ExecutionStatus: gauge_messages.ExecutionStatus_FAILED,
		}
		r := NewScenarioResult(failedScenario)
		specResult.AddTableDrivenScenarioResult(r, table, rowIndex, 0, false)
	}

	// The spec is marked failed, but the per-row failed count is NOT incremented
	// here - aggregation is owned by executeScenarioTableDrivenScenarios.
	c.Assert(specResult.GetFailed(), gc.Equals, true)
	c.Assert(specResult.ScenarioFailedCount, gc.Equals, 0)
	c.Assert(len(specResult.ProtoSpec.Items), gc.Equals, 2)
}
