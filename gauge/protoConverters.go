package main

import "code.google.com/p/goprotobuf/proto"

func convertToProtoItem(item item) *ProtoItem {
	switch item.kind() {
	case scenarioKind:
		return convertToProtoScenarioItem(item.(*scenario))
	case stepKind:
		return convertToProtoStepItem(item.(*step))
	case commentKind:
		return convertToProtoCommentItem(item.(*comment))
	case tagKind:
		return convertToProtoTagsItem(item.(*tags))
	case tableKind:
		return convertToProtoTableItem(item.(*table))
	}
	return nil
}

func convertToProtoStepItem(step *step) *ProtoItem {
	if step.isConcept {
		return convertToProtoConcept(step)
	}
	return &ProtoItem{ItemType: ProtoItem_Step.Enum(), Step: convertToProtoStep(step)}
}

func convertToProtoScenarioItem(scenario *scenario) *ProtoItem {
	scenarioItems := make([]*ProtoItem, 0)
	for _, item := range scenario.items {
		scenarioItems = append(scenarioItems, convertToProtoItem(item))
	}
	protoScenario := newProtoScenario(scenario)
	return &ProtoItem{ItemType: ProtoItem_Scenario.Enum(), Scenario: protoScenario}
}

func convertToProtoConcept(concept *step) *ProtoItem {
	protoConcept := &ProtoConcept{ConceptStep: convertToProtoStep(concept), Steps: convertToProtoSteps(concept.conceptSteps)}
	protoConceptItem := &ProtoItem{ItemType: ProtoItem_Concept.Enum(), Concept: protoConcept}
	return protoConceptItem
}

func convertToProtoStep(step *step) *ProtoStep {
	return &ProtoStep{ActualText: proto.String(step.lineText), ParsedText: proto.String(step.value), Fragments: step.fragments}
}

func convertToProtoSteps(steps []*step) []*ProtoStep {
	protoSteps := make([]*ProtoStep, 0)
	for _, step := range steps {
		protoSteps = append(protoSteps, convertToProtoStep(step))
	}
	return protoSteps
}

func convertToProtoCommentItem(comment *comment) *ProtoItem {
	return &ProtoItem{ItemType: ProtoItem_Comment.Enum(), Comment: &ProtoComment{Text: proto.String(comment.value)}}
}

func convertToProtoTagsItem(tags *tags) *ProtoItem {
	return &ProtoItem{ItemType: ProtoItem_Tags.Enum(), Tags: &ProtoTags{Tags: tags.values}}
}

func convertToProtoTableItem(table *table) *ProtoItem {
	return &ProtoItem{ItemType: ProtoItem_Table.Enum(), Table: convertToProtoTableParam(table)}
}

func convertToProtoParameters(args []*stepArg) []*Parameter {
	params := make([]*Parameter, 0)
	for _, arg := range args {
		params = append(params, convertToProtoParameter(arg))
	}
	return params
}

func convertToProtoParameter(arg *stepArg) *Parameter {
	switch arg.argType {
	case static:
		return &Parameter{ParameterType: Parameter_Static.Enum(), Value: proto.String(arg.value)}
	case dynamic:
		return &Parameter{ParameterType: Parameter_Dynamic.Enum(), Value: proto.String(arg.value)}
	case tableArg:
		return &Parameter{ParameterType: Parameter_Table.Enum(), Table: convertToProtoTableParam(&arg.table)}
	}
	return nil
}

func convertToProtoTableParam(table *table) *ProtoTable {
	protoTableParam := &ProtoTable{Rows: make([]*ProtoTableRow, 0)}
	protoTableParam.Headers = &ProtoTableRow{Cells: table.headers}
	for _, row := range table.getRows() {
		protoTableParam.Rows = append(protoTableParam.Rows, &ProtoTableRow{Cells: row})
	}
	return protoTableParam
}

func addExecutionResult(protoItem *ProtoItem, protoStepExecutionResult *ProtoStepExecutionResult) {
	if protoStepExecutionResult != nil {
		protoItem.Step.StepExecutionResult = protoStepExecutionResult
	}
}

func convertToProtoSuiteResult(suiteResult *suiteResult) *ProtoSuiteResult {
	protoSuiteResult := &ProtoSuiteResult{
		PreHookFailure:   suiteResult.preSuite,
		PostHookFailure:  suiteResult.postSuite,
		Failed:           proto.Bool(suiteResult.isFailed),
		SpecsFailedCount: proto.Int32(int32(suiteResult.specsFailedCount)),
		SpecResults:      convertToProtoSpecResult(suiteResult.specResults),
	}
	return protoSuiteResult
}

func convertToProtoSpecResult(specResults []*specResult) []*ProtoSpecResult {
	protoSpecResults := make([]*ProtoSpecResult, 0)
	for _, specResult := range specResults {
		protoSpecResult := &ProtoSpecResult{
			ProtoSpec:           specResult.protoSpec,
			ScenarioCount:       proto.Int32(int32(specResult.scenarioCount)),
			ScenarioFailedCount: proto.Int32(int32(specResult.scenarioFailedCount)),
			Failed:              proto.Bool(specResult.isFailed),
		}
		protoSpecResults = append(protoSpecResults, protoSpecResult)
	}
	return protoSpecResults
}

func convertToProtoSpec(spec *specification) *ProtoSpec {
	protoSpec := newProtoSpec(spec)
	protoItems := make([]*ProtoItem, 0)
	for _, item := range spec.items {
		protoItems = append(protoItems, convertToProtoItem(item))
	}
	protoSpec.Items = protoItems
	return protoSpec
}

func newProtoSpec(specification *specification) *ProtoSpec {
	return &ProtoSpec{
		Items:         make([]*ProtoItem, 0),
		SpecHeading:   proto.String(specification.heading.value),
		IsTableDriven: proto.Bool(false),
		FileName:      proto.String(specification.fileName)}
}

func newProtoScenario(scenario *scenario) *ProtoScenario {
	return &ProtoScenario{ScenarioHeading: proto.String(scenario.heading.value), Failed: proto.Bool(false)}
}
