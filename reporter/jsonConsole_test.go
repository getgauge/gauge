/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package reporter

import (
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func setupJSONConsole() (*dummyWriter, *jsonConsole) {
	dw := newDummyWriter()
	console := newJSONConsole(dw, false, 0)
	return dw, console
}

func (s *MySuite) TestSpecStart_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()
	protoSpec := &gauge_messages.ProtoSpec{
		FileName:    "file",
		SpecHeading: "Specification",
	}
	scenarios := []*gauge.Scenario{
		{
			Heading: &gauge.Heading{
				Value:       "Scenario",
				LineNo:      2,
				HeadingType: 1,
			},
		},
	}
	spec := &gauge.Specification{
		FileName: "file",
		Heading: &gauge.Heading{
			Value:       "Specification",
			LineNo:      1,
			HeadingType: 0,
		},
		Scenarios: scenarios,
	}

	expected := `{"type":"specStart","id":"file","name":"Specification","filename":"file","line":1}
`
	jc.SpecStart(spec, &result.SpecResult{Skipped: false, ProtoSpec: protoSpec})
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioStart_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   3,
		},
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: false,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario",
			IsFailed: false,
		},
	}

	expected := `{"type":"scenarioStart","id":"file:2","parentId":"file","name":"Scenario","filename":"file","line":2,"result":{"time":0}}
`
	jc.ScenarioStart(scenario, info, &result.ScenarioResult{})
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioEnd_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	protoScenario := &gauge_messages.ProtoScenario{
		ScenarioHeading: "Scenario",
		Failed:          false,
	}

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   3,
		},
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: false,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario",
			IsFailed: false,
		},
	}

	expected := `{"type":"scenarioEnd","id":"file:2","parentId":"file","name":"Scenario","filename":"file","line":2,"result":{"status":"pass","time":0}}
`

	jc.ScenarioEnd(scenario, &result.ScenarioResult{ProtoScenario: protoScenario}, info)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioEndWithPreHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	protoScenario := &gauge_messages.ProtoScenario{
		ScenarioHeading: "Scenario",
		Failed:          true,
		PreHookFailure: &gauge_messages.ProtoHookFailure{
			StackTrace:   "stacktrace",
			ErrorMessage: "message",
		},
		ExecutionStatus: gauge_messages.ExecutionStatus_FAILED,
	}

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   3,
		},
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: true,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario",
			IsFailed: true,
		},
	}

	expected := `{"type":"scenarioEnd","id":"file:2","parentId":"file","name":"Scenario","filename":"file","line":2,"result":{"status":"fail","time":0,"beforeHookFailure":{"text":"Before Scenario","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"}}}
`

	jc.ScenarioEnd(scenario, &result.ScenarioResult{ProtoScenario: protoScenario}, info)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioEndWithPostHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	protoScenario := &gauge_messages.ProtoScenario{
		ScenarioHeading: "Scenario",
		Failed:          true,
		PostHookFailure: &gauge_messages.ProtoHookFailure{
			StackTrace:   "stacktrace",
			ErrorMessage: "message",
		},
		ExecutionStatus: gauge_messages.ExecutionStatus_PASSED,
	}

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   3,
		},
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: true,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario",
			IsFailed: true,
		},
	}

	expected := `{"type":"scenarioEnd","id":"file:2","parentId":"file","name":"Scenario","filename":"file","line":2,"result":{"status":"pass","time":0,"afterHookFailure":{"text":"After Scenario","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"}}}
`

	jc.ScenarioEnd(scenario, &result.ScenarioResult{ProtoScenario: protoScenario}, info)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioEndWithPreAndPostHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	protoScenario := &gauge_messages.ProtoScenario{
		ScenarioHeading: "Scenario",
		Failed:          true,
		PreHookFailure: &gauge_messages.ProtoHookFailure{
			StackTrace:   "stacktrace",
			ErrorMessage: "message",
		},
		PostHookFailure: &gauge_messages.ProtoHookFailure{
			StackTrace:   "stacktrace",
			ErrorMessage: "message",
		},
		ExecutionStatus: gauge_messages.ExecutionStatus_FAILED,
	}

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   3,
		},
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: true,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario",
			IsFailed: true,
		},
	}

	expected := `{"type":"scenarioEnd","id":"file:2","parentId":"file","name":"Scenario","filename":"file","line":2,"result":{"status":"fail","time":0,"beforeHookFailure":{"text":"Before Scenario","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"},"afterHookFailure":{"text":"After Scenario","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"}}}
`

	jc.ScenarioEnd(scenario, &result.ScenarioResult{ProtoScenario: protoScenario}, info)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioEndWithBeforeStepHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	protoScenario := &gauge_messages.ProtoScenario{
		ScenarioHeading: "Scenario",
		Failed:          true,
		ScenarioItems: []*gauge_messages.ProtoItem{
			{
				ItemType: gauge_messages.ProtoItem_Step,
				Step: &gauge_messages.ProtoStep{
					ActualText: "Step",
					ParsedText: "Step",
					StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{
						PreHookFailure: &gauge_messages.ProtoHookFailure{
							ErrorMessage: "message",
							StackTrace:   "stacktrace",
						},
					},
				},
			},
		},
		ExecutionStatus: gauge_messages.ExecutionStatus_FAILED,
	}

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   3,
		},
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: true,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario",
			IsFailed: true,
		},
	}

	expected := `{"type":"scenarioEnd","id":"file:2","parentId":"file","name":"Scenario","filename":"file","line":2,"result":{"status":"fail","time":0,"errors":[{"text":"BeforeStep hook for step: Step","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"}]}}
`

	jc.ScenarioEnd(scenario, &result.ScenarioResult{ProtoScenario: protoScenario}, info)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioEndWithStepFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	protoStep := &gauge_messages.ProtoStep{
		ActualText: "Step",
		ParsedText: "Step",
		StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{
			ExecutionResult: &gauge_messages.ProtoExecutionResult{
				Failed:       true,
				ErrorMessage: "message",
				StackTrace:   "stacktrace",
			},
		},
	}

	protoScenario := &gauge_messages.ProtoScenario{
		ScenarioHeading: "Scenario",
		Failed:          true,
		ScenarioItems: []*gauge_messages.ProtoItem{
			{
				ItemType: gauge_messages.ProtoItem_Step,
				Step:     protoStep,
			},
		},
		ExecutionStatus: gauge_messages.ExecutionStatus_FAILED,
	}

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   3,
		},
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: false,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario",
			IsFailed: false,
		},
		CurrentStep: &gauge_messages.StepInfo{
			Step: &gauge_messages.ExecuteStepRequest{
				ActualStepText:  "Step",
				ParsedStepText:  "Step",
				ScenarioFailing: false,
			},
			IsFailed:   false,
			StackTrace: "stacktrace",
		},
	}

	step := gauge.Step{
		Value:     "Step",
		LineNo:    4,
		LineText:  "Step",
		IsConcept: false,
		Parent:    nil,
	}

	res := result.NewStepResult(protoStep)
	res.StepFailed = false
	jc.StepEnd(step, res, info)
	jc.ScenarioEnd(scenario, &result.ScenarioResult{ProtoScenario: protoScenario}, info)

	expected := `{"type":"scenarioEnd","id":"file:2","parentId":"file","name":"Scenario","filename":"file","line":2,"result":{"status":"fail","time":0,"errors":[{"text":"Step","filename":"","message":"message","lineNo":"4","stackTrace":"stacktrace"}]}}
`

	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioEndWithConceptFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	protoStep := &gauge_messages.ProtoStep{
		ActualText: "Step",
		ParsedText: "Step",
		StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{
			ExecutionResult: &gauge_messages.ProtoExecutionResult{
				Failed:       true,
				ErrorMessage: "message",
				StackTrace:   "stacktrace",
			},
		},
	}

	protoScenario := &gauge_messages.ProtoScenario{
		ScenarioHeading: "Scenario",
		Failed:          true,
		ScenarioItems: []*gauge_messages.ProtoItem{
			{
				ItemType: gauge_messages.ProtoItem_Step,
				Step:     protoStep,
			},
		},
		ExecutionStatus: gauge_messages.ExecutionStatus_FAILED,
	}

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   3,
		},
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: false,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario",
			IsFailed: false,
		},
		CurrentStep: &gauge_messages.StepInfo{
			Step: &gauge_messages.ExecuteStepRequest{
				ActualStepText:  "Step",
				ParsedStepText:  "Step",
				ScenarioFailing: false,
			},
			IsFailed:   false,
			StackTrace: "stacktrace",
		},
	}

	step := gauge.Step{
		Value:     "Step",
		LineNo:    4,
		LineText:  "Step",
		IsConcept: true,
		Parent:    nil,
	}

	step2 := gauge.Step{
		Value:     "Step 2",
		LineNo:    2,
		LineText:  "Step 2",
		IsConcept: false,
		Parent:    &step,
	}

	step.ConceptSteps = []*gauge.Step{&step2}

	res := result.NewStepResult(protoStep)
	res.StepFailed = false
	jc.StepEnd(step2, res, info)
	jc.ScenarioEnd(scenario, &result.ScenarioResult{ProtoScenario: protoScenario}, info)

	expected := `{"type":"scenarioEnd","id":"file:2","parentId":"file","name":"Scenario","filename":"file","line":2,"result":{"status":"fail","time":0,"errors":[{"text":"Step","filename":"","message":"message","lineNo":"2","stackTrace":"stacktrace"}]}}
`

	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioEndWithAfterStepHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	protoScenario := &gauge_messages.ProtoScenario{
		ScenarioHeading: "Scenario",
		Failed:          true,
		ScenarioItems: []*gauge_messages.ProtoItem{
			{
				ItemType: gauge_messages.ProtoItem_Step,
				Step: &gauge_messages.ProtoStep{
					ActualText: "Step",
					ParsedText: "Step",
					StepExecutionResult: &gauge_messages.ProtoStepExecutionResult{
						PostHookFailure: &gauge_messages.ProtoHookFailure{
							ErrorMessage: "message",
							StackTrace:   "stacktrace",
						},
					},
				},
			},
		},
		ExecutionStatus: gauge_messages.ExecutionStatus_FAILED,
	}

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   3,
		},
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: true,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario",
			IsFailed: true,
		},
	}

	expected := `{"type":"scenarioEnd","id":"file:2","parentId":"file","name":"Scenario","filename":"file","line":2,"result":{"status":"fail","time":0,"errors":[{"text":"AfterStep hook for step: Step","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"}]}}
`

	jc.ScenarioEnd(scenario, &result.ScenarioResult{ProtoScenario: protoScenario}, info)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestSpecEnd_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()
	protoSpec := &gauge_messages.ProtoSpec{
		FileName:    "file",
		SpecHeading: "Specification",
	}
	scenarios := []*gauge.Scenario{
		{
			Heading: &gauge.Heading{
				Value:       "Scenario",
				LineNo:      2,
				HeadingType: 1,
			},
		},
	}
	spec := &gauge.Specification{
		FileName: "file",
		Heading: &gauge.Heading{
			Value:       "Specification",
			LineNo:      1,
			HeadingType: 0,
		},
		Scenarios: scenarios,
	}
	expected := `{"type":"specEnd","id":"file","name":"Specification","filename":"file","line":1,"result":{"status":"pass","time":0}}
`
	jc.SpecEnd(spec, &result.SpecResult{ProtoSpec: protoSpec})
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestSpecEndWithPreHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()
	protoSpec := &gauge_messages.ProtoSpec{
		FileName:    "file",
		SpecHeading: "Specification",
		PreHookFailures: []*gauge_messages.ProtoHookFailure{
			{
				StackTrace:   "stacktrace",
				ErrorMessage: "message",
			},
		},
	}
	scenarios := []*gauge.Scenario{
		{
			Heading: &gauge.Heading{
				Value:       "Scenario",
				LineNo:      2,
				HeadingType: 1,
			},
		},
	}
	spec := &gauge.Specification{
		FileName: "file",
		Heading: &gauge.Heading{
			Value:       "Specification",
			LineNo:      1,
			HeadingType: 0,
		},
		Scenarios: scenarios,
	}
	expected := `{"type":"specEnd","id":"file","name":"Specification","filename":"file","line":1,"result":{"status":"fail","time":0,"beforeHookFailure":{"text":"Before Specification","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"}}}
`
	res := &result.SpecResult{
		ProtoSpec: protoSpec,
		IsFailed:  true,
	}

	jc.SpecEnd(spec, res)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestSpecEndWithPostHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()
	protoSpec := &gauge_messages.ProtoSpec{
		FileName:    "file",
		SpecHeading: "Specification",
		PostHookFailures: []*gauge_messages.ProtoHookFailure{
			{
				StackTrace:   "stacktrace",
				ErrorMessage: "message",
			},
		},
	}
	scenarios := []*gauge.Scenario{
		{
			Heading: &gauge.Heading{
				Value:       "Scenario",
				LineNo:      2,
				HeadingType: 1,
			},
		},
	}
	spec := &gauge.Specification{
		FileName: "file",
		Heading: &gauge.Heading{
			Value:       "Specification",
			LineNo:      1,
			HeadingType: 0,
		},
		Scenarios: scenarios,
	}
	expected := `{"type":"specEnd","id":"file","name":"Specification","filename":"file","line":1,"result":{"status":"fail","time":0,"afterHookFailure":{"text":"After Specification","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"}}}
`
	res := &result.SpecResult{
		ProtoSpec: protoSpec,
		IsFailed:  true,
	}

	jc.SpecEnd(spec, res)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestSpecEndWithPreAndPostHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()
	protoSpec := &gauge_messages.ProtoSpec{
		FileName:    "file",
		SpecHeading: "Specification",
		PreHookFailures: []*gauge_messages.ProtoHookFailure{
			{
				StackTrace:   "stacktrace",
				ErrorMessage: "message",
			},
		},
		PostHookFailures: []*gauge_messages.ProtoHookFailure{
			{
				StackTrace:   "stacktrace",
				ErrorMessage: "message",
			},
		},
	}
	scenarios := []*gauge.Scenario{
		{
			Heading: &gauge.Heading{
				Value:       "Scenario",
				LineNo:      2,
				HeadingType: 1,
			},
		},
	}
	spec := &gauge.Specification{
		FileName: "file",
		Heading: &gauge.Heading{
			Value:       "Specification",
			LineNo:      1,
			HeadingType: 0,
		},
		Scenarios: scenarios,
	}
	expected := `{"type":"specEnd","id":"file","name":"Specification","filename":"file","line":1,"result":{"status":"fail","time":0,"beforeHookFailure":{"text":"Before Specification","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"},"afterHookFailure":{"text":"After Specification","filename":"","message":"message","lineNo":"","stackTrace":"stacktrace"}}}
`
	res := &result.SpecResult{
		ProtoSpec: protoSpec,
		IsFailed:  true,
	}

	jc.SpecEnd(spec, res)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestSpecEndWithNoScenariosInSpec_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()
	protoSpec := &gauge_messages.ProtoSpec{
		FileName:    "file",
		SpecHeading: "Specification",
	}

	scenarios := []*gauge.Scenario{}

	spec := &gauge.Specification{
		FileName: "file",
		Heading: &gauge.Heading{
			Value:       "Specification",
			LineNo:      1,
			HeadingType: 0,
		},
		Scenarios: scenarios,
	}
	expected := `{"type":"specEnd","id":"file","name":"Specification","filename":"file","line":1,"result":{"status":"skip","time":0}}
`
	res := &result.SpecResult{
		ProtoSpec: protoSpec,
		Skipped:   true,
	}

	jc.SpecEnd(spec, res)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestSuiteEnd_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	jc.SuiteEnd(&result.SuiteResult{})
	c.Assert(dw.output, Equals, "{\"type\":\"suiteEnd\",\"result\":{\"status\":\"pass\",\"time\":0}}\n")
}

func (s *MySuite) TestSuiteEndWithBeforeHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	res := &result.SuiteResult{
		PreSuite: &gauge_messages.ProtoHookFailure{
			StackTrace:   "stack trace",
			ErrorMessage: "message",
		},
		IsFailed: true,
	}
	expected := `{"type":"suiteEnd","result":{"status":"fail","time":0,"beforeHookFailure":{"text":"Before Suite","filename":"","message":"message","lineNo":"","stackTrace":"stack trace"}}}
`
	jc.SuiteEnd(res)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestSuiteEndWithAfterHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	res := &result.SuiteResult{
		PostSuite: &gauge_messages.ProtoHookFailure{
			StackTrace:   "stack trace",
			ErrorMessage: "message",
		},
		PreSuite: &gauge_messages.ProtoHookFailure{
			StackTrace:   "stack trace",
			ErrorMessage: "message",
		},
		IsFailed: true,
	}
	expected := `{"type":"suiteEnd","result":{"status":"fail","time":0,"beforeHookFailure":{"text":"Before Suite","filename":"","message":"message","lineNo":"","stackTrace":"stack trace"},"afterHookFailure":{"text":"After Suite","filename":"","message":"message","lineNo":"","stackTrace":"stack trace"}}}
`
	jc.SuiteEnd(res)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestSuiteEndWithBeforeAndAfterHookFailure_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	res := &result.SuiteResult{
		PostSuite: &gauge_messages.ProtoHookFailure{
			StackTrace:   "stack trace",
			ErrorMessage: "message",
		},
		IsFailed: true,
	}
	expected := `{"type":"suiteEnd","result":{"status":"fail","time":0,"afterHookFailure":{"text":"After Suite","filename":"","message":"message","lineNo":"","stackTrace":"stack trace"}}}
`
	jc.SuiteEnd(res)
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioStartWithScenarioDataTable_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario with data table",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   5,
		},
		ScenarioDataTableRow:      *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{{{Value: "value1", CellType: gauge.Static}}}, 0),
		ScenarioDataTableRowIndex: 0,
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: false,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario with data table",
			IsFailed: false,
		},
	}

	expected := `{"type":"scenarioStart","id":"file:2","parentId":"file","name":"Scenario with data table","filename":"file","line":2,"result":{"time":0,"table":{"text":"\n   |header|\n   |------|\n   |value1|\n","rowIndex":0}}}
`
	jc.ScenarioStart(scenario, info, &result.ScenarioResult{})
	c.Assert(dw.output, Equals, expected)
}

func (s *MySuite) TestScenarioEndWithScenarioDataTable_JSONConsole(c *C) {
	dw, jc := setupJSONConsole()

	protoScenario := &gauge_messages.ProtoScenario{
		ScenarioHeading: "Scenario with data table",
		Failed:          false,
	}

	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{
			Value:       "Scenario with data table",
			LineNo:      2,
			HeadingType: 1,
		},
		Span: &gauge.Span{
			Start: 2,
			End:   5,
		},
		ScenarioDataTableRow:      *gauge.NewTable([]string{"header"}, [][]gauge.TableCell{{{Value: "value1", CellType: gauge.Static}}}, 0),
		ScenarioDataTableRowIndex: 0,
	}

	info := &gauge_messages.ExecutionInfo{
		CurrentSpec: &gauge_messages.SpecInfo{
			Name:     "Specification",
			FileName: "file",
			IsFailed: false,
		},
		CurrentScenario: &gauge_messages.ScenarioInfo{
			Name:     "Scenario with data table",
			IsFailed: false,
		},
	}

	expected := `{"type":"scenarioEnd","id":"file:2","parentId":"file","name":"Scenario with data table","filename":"file","line":2,"result":{"status":"pass","time":0,"table":{"text":"\n   |header|\n   |------|\n   |value1|\n","rowIndex":0}}}
`
	jc.ScenarioEnd(scenario, &result.ScenarioResult{ProtoScenario: protoScenario}, info)
	c.Assert(dw.output, Equals, expected)
}
