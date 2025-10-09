/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import (
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/execution/result"
)

func ConvertToProtoItem(item Item) *gauge_messages.ProtoItem {
	switch item.Kind() {
	case ScenarioKind:
		return convertToProtoScenarioItem(item.(*Scenario))
	case StepKind:
		return convertToProtoStepItem(item.(*Step))
	case CommentKind:
		return convertToProtoCommentItem(item.(*Comment))
	case DataTableKind:
		return convertToProtoDataTableItem(item.(*DataTable))
	case TagKind:
		return convertToProtoTagItem(item.(*Tags))
	case TearDownKind:
		teardown := item.(*TearDown)
		return convertToProtoCommentItem(&Comment{LineNo: teardown.LineNo, Value: teardown.Value})
	}
	return nil
}

func convertToProtoTagItem(tags *Tags) *gauge_messages.ProtoItem {
	return &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Tags, Tags: convertToProtoTags(tags)}
}

func convertToProtoStepItem(step *Step) *gauge_messages.ProtoItem {
	if step.IsConcept {
		return convertToProtoConcept(step)
	}
	return &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Step, Step: convertToProtoStep(step)}
}

func convertToProtoStepItems(steps []*Step) []*gauge_messages.ProtoItem {
	protoItems := make([]*gauge_messages.ProtoItem, 0)
	for _, step := range steps {
		protoItems = append(protoItems, convertToProtoStepItem(step))
	}
	return protoItems
}

func convertToProtoScenarioItem(scenario *Scenario) *gauge_messages.ProtoItem {
	scenarioItems := make([]*gauge_messages.ProtoItem, 0)
	for _, item := range scenario.Items {
		scenarioItems = append(scenarioItems, ConvertToProtoItem(item))
	}
	protoScenario := NewProtoScenario(scenario)
	protoScenario.ScenarioItems = scenarioItems
	return &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Scenario, Scenario: protoScenario}
}

func convertToProtoConcept(concept *Step) *gauge_messages.ProtoItem {
	protoConcept := &gauge_messages.ProtoConcept{ConceptStep: convertToProtoStep(concept), Steps: convertToProtoStepItems(concept.ConceptSteps)}
	protoConceptItem := &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Concept, Concept: protoConcept}
	return protoConceptItem
}

func convertToProtoStep(step *Step) *gauge_messages.ProtoStep {
	return &gauge_messages.ProtoStep{ActualText: step.LineText, ParsedText: step.Value, Fragments: makeFragmentsCopy(step.Fragments)}
}

func convertToProtoTags(tags *Tags) *gauge_messages.ProtoTags {
	return &gauge_messages.ProtoTags{Tags: tags.Values()}
}

func makeFragmentsCopy(fragments []*gauge_messages.Fragment) []*gauge_messages.Fragment {
	copiedFragments := make([]*gauge_messages.Fragment, 0)
	for _, fragment := range fragments {
		copiedFragments = append(copiedFragments, makeFragmentCopy(fragment))
	}
	return copiedFragments
}

func makeFragmentCopy(fragment *gauge_messages.Fragment) *gauge_messages.Fragment {
	if fragment.GetFragmentType() == gauge_messages.Fragment_Text {
		return &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Text, Text: fragment.GetText()}
	} else {
		return &gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: makeParameterCopy(fragment.Parameter)}
	}
}

func makeParameterCopy(parameter *gauge_messages.Parameter) *gauge_messages.Parameter {
	switch parameter.GetParameterType() {
	case gauge_messages.Parameter_Static:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Static, Value: parameter.GetValue(), Name: parameter.GetName()}
	case gauge_messages.Parameter_Dynamic:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Dynamic, Value: parameter.GetValue(), Name: parameter.GetName()}
	case gauge_messages.Parameter_Table:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Table, Table: makeTableCopy(parameter.GetTable()), Name: parameter.GetName()}
	case gauge_messages.Parameter_Special_String:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Special_String, Value: parameter.GetValue(), Name: parameter.GetName()}
	case gauge_messages.Parameter_Special_Table:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Special_Table, Table: makeTableCopy(parameter.GetTable()), Name: parameter.GetName()}
	case gauge_messages.Parameter_Multiline_String:
    	return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Multiline_String, Value: parameter.GetValue(), Name: parameter.GetName()}	}
	return parameter
}

func makeTableCopy(table *gauge_messages.ProtoTable) *gauge_messages.ProtoTable {
	copiedTable := &gauge_messages.ProtoTable{}
	copiedTable.Headers = makeProtoTableRowCopy(table.GetHeaders())

	copiedRows := make([]*gauge_messages.ProtoTableRow, 0)
	for _, tableRow := range table.GetRows() {
		copiedRows = append(copiedRows, makeProtoTableRowCopy(tableRow))
	}
	copiedTable.Rows = copiedRows
	return copiedTable
}

func makeProtoTableRowCopy(tableRow *gauge_messages.ProtoTableRow) *gauge_messages.ProtoTableRow {
	copiedCells := make([]string, 0)
	return &gauge_messages.ProtoTableRow{Cells: append(copiedCells, tableRow.GetCells()...)}
}

func convertToProtoCommentItem(comment *Comment) *gauge_messages.ProtoItem {
	return &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Comment, Comment: &gauge_messages.ProtoComment{Text: comment.Value}}
}

func convertToProtoDataTableItem(dataTable *DataTable) *gauge_messages.ProtoItem {
	return &gauge_messages.ProtoItem{ItemType: gauge_messages.ProtoItem_Table, Table: ConvertToProtoTable(dataTable.Table)}
}

func convertToProtoParameter(arg *StepArg) *gauge_messages.Parameter {
	switch arg.ArgType {
	case Static:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Static, Value: arg.Value, Name: arg.Name}
	case Dynamic:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Dynamic, Value: arg.Value, Name: arg.Name}
	case TableArg:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Table, Table: ConvertToProtoTable(&arg.Table), Name: arg.Name}
	case SpecialString:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Special_String, Value: arg.Value, Name: arg.Name}
	case SpecialTable:
		return &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Special_Table, Table: ConvertToProtoTable(&arg.Table), Name: arg.Name}
	case MultilineString: 
		return  &gauge_messages.Parameter{ParameterType: gauge_messages.Parameter_Multiline_String, Value: arg.Value, Name: arg.Name}
	}
    return nil
}

func ConvertToProtoTable(table *Table) *gauge_messages.ProtoTable {
	if table == nil {
		return nil
	}
	protoTableParam := &gauge_messages.ProtoTable{Rows: make([]*gauge_messages.ProtoTableRow, 0)}
	protoTableParam.Headers = &gauge_messages.ProtoTableRow{Cells: table.Headers}
	for _, row := range table.Rows() { // nolint
		protoTableParam.Rows = append(protoTableParam.Rows, &gauge_messages.ProtoTableRow{Cells: row})
	}
	return protoTableParam
}

func ConvertToProtoSuiteResult(suiteResult *result.SuiteResult) *gauge_messages.ProtoSuiteResult {
	protoSuiteResult := &gauge_messages.ProtoSuiteResult{
		PreHookFailure:          suiteResult.PreSuite,
		PostHookFailure:         suiteResult.PostSuite,
		Failed:                  suiteResult.IsFailed,
		SpecsFailedCount:        int32(suiteResult.SpecsFailedCount),
		ExecutionTime:           suiteResult.ExecutionTime,
		SpecResults:             convertToProtoSpecResults(suiteResult.SpecResults),
		SuccessRate:             getSuccessRate(len(suiteResult.SpecResults), suiteResult.SpecsFailedCount+suiteResult.SpecsSkippedCount),
		Environment:             suiteResult.Environment,
		Tags:                    suiteResult.Tags,
		ProjectName:             suiteResult.ProjectName,
		Timestamp:               suiteResult.Timestamp,
		TimestampISO:            suiteResult.TimestampISO,
		SpecsSkippedCount:       int32(suiteResult.SpecsSkippedCount),
		PreHookMessages:         suiteResult.PreHookMessages,
		PostHookMessages:        suiteResult.PostHookMessages,
		PreHookScreenshotFiles:  suiteResult.PreHookScreenshotFiles,
		PostHookScreenshotFiles: suiteResult.PostHookScreenshotFiles,
		PreHookScreenshots:      suiteResult.PreHookScreenshots,
		PostHookScreenshots:     suiteResult.PostHookScreenshots,
	}
	return protoSuiteResult
}

func ConvertToProtoSpecResult(specResult *result.SpecResult) *gauge_messages.ProtoSpecResult {
	return convertToProtoSpecResult(specResult)
}

func ConvertToProtoScenarioResult(scenarioResult *result.ScenarioResult) *gauge_messages.ProtoScenarioResult {
	return &gauge_messages.ProtoScenarioResult{
		ProtoItem: &gauge_messages.ProtoItem{
			ItemType: gauge_messages.ProtoItem_Scenario,
			Scenario: scenarioResult.ProtoScenario,
		},
		ExecutionTime: scenarioResult.ExecTime(),
		Timestamp:     time.Now().Format(time.RFC3339),
	}
}

func ConvertToProtoStepResult(stepResult *result.StepResult) *gauge_messages.ProtoStepResult {
	return &gauge_messages.ProtoStepResult{
		ProtoItem: &gauge_messages.ProtoItem{
			ItemType: gauge_messages.ProtoItem_Step,
			Step:     stepResult.ProtoStep,
		},
		ExecutionTime: stepResult.ExecTime(),
		Timestamp:     time.Now().Format(time.RFC3339),
	}
}

func getSuccessRate(totalSpecs int, failedSpecs int) float32 {
	if totalSpecs == 0 {
		return 0
	}
	return (float32)(100.0 * (totalSpecs - failedSpecs) / totalSpecs)
}

func convertToProtoSpecResult(specResult *result.SpecResult) *gauge_messages.ProtoSpecResult {
	return &gauge_messages.ProtoSpecResult{
		ProtoSpec:            specResult.ProtoSpec,
		ScenarioCount:        int32(specResult.ScenarioCount),
		ScenarioFailedCount:  int32(specResult.ScenarioFailedCount),
		Failed:               specResult.IsFailed,
		FailedDataTableRows:  specResult.FailedDataTableRows,
		ExecutionTime:        specResult.ExecutionTime,
		Skipped:              specResult.Skipped,
		ScenarioSkippedCount: int32(specResult.ScenarioSkippedCount),
		Errors:               specResult.Errors,
		Timestamp:            time.Now().Format(time.RFC3339),
		TimestampISO:         time.Now().Format(time.RFC3339Nano),
	}
}

func convertToProtoSpecResults(specResults []*result.SpecResult) []*gauge_messages.ProtoSpecResult {
	protoSpecResults := make([]*gauge_messages.ProtoSpecResult, 0)
	for _, specResult := range specResults {
		protoSpecResults = append(protoSpecResults, convertToProtoSpecResult(specResult))
	}
	return protoSpecResults
}

func ConvertToProtoSpec(spec *Specification) *gauge_messages.ProtoSpec {
	protoSpec := newProtoSpec(spec)
	if spec.DataTable.IsInitialized() {
		protoSpec.IsTableDriven = true
	}
	var protoItems []*gauge_messages.ProtoItem
	for _, item := range spec.Items {
		protoItems = append(protoItems, ConvertToProtoItem(item))
	}
	protoSpec.Items = protoItems
	return protoSpec
}

func ConvertToProtoStepValue(stepValue *StepValue) *gauge_messages.ProtoStepValue {
	return &gauge_messages.ProtoStepValue{
		StepValue:              stepValue.StepValue,
		ParameterizedStepValue: stepValue.ParameterizedStepValue,
		Parameters:             stepValue.Args,
	}
}

func newProtoSpec(specification *Specification) *gauge_messages.ProtoSpec {
	return &gauge_messages.ProtoSpec{
		Items:         make([]*gauge_messages.ProtoItem, 0),
		SpecHeading:   specification.Heading.Value,
		IsTableDriven: specification.DataTable.IsInitialized(),
		FileName:      specification.FileName,
		Tags:          getTags(specification.Tags),
	}

}

func NewSpecResult(specification *Specification) *result.SpecResult {
	return &result.SpecResult{
		ProtoSpec:           newProtoSpec(specification),
		FailedDataTableRows: make([]int32, 0),
	}
}

func NewProtoScenario(scenario *Scenario) *gauge_messages.ProtoScenario {
	return &gauge_messages.ProtoScenario{
		ScenarioHeading: scenario.Heading.Value,
		Failed:          false,
		Skipped:         false,
		Tags:            getTags(scenario.Tags),
		Contexts:        make([]*gauge_messages.ProtoItem, 0),
		ExecutionTime:   0,
		TearDownSteps:   make([]*gauge_messages.ProtoItem, 0),
		SkipErrors:      make([]string, 0),
		Span:            &gauge_messages.Span{Start: int64(scenario.Span.Start), End: int64(scenario.Span.End)},
		ExecutionStatus: gauge_messages.ExecutionStatus_NOTEXECUTED,
	}
}

func getTags(tags *Tags) []string {
	if tags != nil {
		return tags.Values()
	}
	return make([]string, 0)
}

func ConvertToProtoExecutionArg(args []*ExecutionArg) []*gauge_messages.ExecutionArg {
	execArgs := []*gauge_messages.ExecutionArg{}
	for _, arg := range args {
		execArgs = append(execArgs, &gauge_messages.ExecutionArg{
			FlagName:  arg.Name,
			FlagValue: arg.Value,
		})
	}
	return execArgs
}
