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

func convertToProtoStepItems(steps []*step) []*ProtoItem {
	protoItems := make([]*ProtoItem, 0)
	for _, step := range steps {
		protoItems = append(protoItems, convertToProtoStepItem(step))
	}
	return protoItems
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
	protoConcept := &ProtoConcept{ConceptStep: convertToProtoStep(concept), Steps: convertToProtoStepItems(concept.conceptSteps)}
	protoConceptItem := &ProtoItem{ItemType: ProtoItem_Concept.Enum(), Concept: protoConcept}
	return protoConceptItem
}

func convertToProtoStep(step *step) *ProtoStep {
	return &ProtoStep{ActualText: proto.String(step.lineText), ParsedText: proto.String(step.value), Fragments: makeFragmentsCopy(step.fragments)}
}

func makeFragmentsCopy(fragments []*Fragment) []*Fragment {
	copiedFragments := make([]*Fragment, 0)
	for _, fragment := range fragments {
		copiedFragments = append(copiedFragments, makeFragmentCopy(fragment))
	}
	return copiedFragments
}

func makeFragmentCopy(fragment *Fragment) *Fragment {
	if fragment.GetFragmentType() == Fragment_Text {
		return &Fragment{FragmentType: Fragment_Text.Enum(), Text: proto.String(fragment.GetText())}
	} else {
		return &Fragment{FragmentType: Fragment_Parameter.Enum(), Parameter: makeParameterCopy(fragment.Parameter)}
	}
}

func makeParameterCopy(parameter *Parameter) *Parameter {
	switch parameter.GetParameterType() {
	case Parameter_Static:
		return &Parameter{ParameterType: Parameter_Static.Enum(), Value: proto.String(parameter.GetValue()), Name: proto.String(parameter.GetName())}
	case Parameter_Dynamic:
		return &Parameter{ParameterType: Parameter_Dynamic.Enum(), Value: proto.String(parameter.GetValue()), Name: proto.String(parameter.GetName())}
	case Parameter_Table:
		return &Parameter{ParameterType: Parameter_Table.Enum(), Table: makeTableCopy(parameter.GetTable()), Name: proto.String(parameter.GetName())}
	case Parameter_Special_String:
		return &Parameter{ParameterType: Parameter_Special_String.Enum(), Value: proto.String(parameter.GetValue()), Name: proto.String(parameter.GetName())}
	case Parameter_Special_Table:
		return &Parameter{ParameterType: Parameter_Special_Table.Enum(), Table: makeTableCopy(parameter.GetTable()), Name: proto.String(parameter.GetName())}
	}
	return parameter
}

func makeTableCopy(table *ProtoTable) *ProtoTable {
	copiedTable := &ProtoTable{}
	copiedTable.Headers = makeProtoTableRowCopy(table.GetHeaders())

	copiedRows := make([]*ProtoTableRow, 0)
	for _, tableRow := range table.GetRows() {
		copiedRows = append(copiedRows, makeProtoTableRowCopy(tableRow))
	}
	copiedTable.Rows = copiedRows
	return copiedTable
}

func makeProtoTableRowCopy(tableRow *ProtoTableRow) *ProtoTableRow {
	copiedCells := make([]string, 0)
	return &ProtoTableRow{Cells: append(copiedCells, tableRow.GetCells()...)}
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
		return &Parameter{ParameterType: Parameter_Static.Enum(), Value: proto.String(arg.value), Name: proto.String(arg.name)}
	case dynamic:
		return &Parameter{ParameterType: Parameter_Dynamic.Enum(), Value: proto.String(arg.value), Name: proto.String(arg.name)}
	case tableArg:
		return &Parameter{ParameterType: Parameter_Table.Enum(), Table: convertToProtoTableParam(&arg.table), Name: proto.String(arg.name)}
	case specialString:
		return &Parameter{ParameterType: Parameter_Special_String.Enum(), Value: proto.String(arg.value), Name: proto.String(arg.name)}
	case specialTable:
		return &Parameter{ParameterType: Parameter_Special_Table.Enum(), Table: convertToProtoTableParam(&arg.table), Name: proto.String(arg.name)}
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
		ExecutionTime:    proto.Int64(suiteResult.executionTime),
		SpecResults:      convertToProtoSpecResult(suiteResult.specResults),
		SuccessRate:      proto.Float32(getSuccessRate(len(suiteResult.specResults), suiteResult.specsFailedCount)),
	}
	return protoSuiteResult
}

func getSuccessRate(totalSpecs int, failedSpecs int) float32 {
	if totalSpecs == 0 {
		return 0
	}
	return (float32)(100.0 * (totalSpecs - failedSpecs) / totalSpecs)
}

func convertToProtoSpecResult(specResults []*specResult) []*ProtoSpecResult {
	protoSpecResults := make([]*ProtoSpecResult, 0)
	for _, specResult := range specResults {
		protoSpecResult := &ProtoSpecResult{
			ProtoSpec:           specResult.protoSpec,
			ScenarioCount:       proto.Int32(int32(specResult.scenarioCount)),
			ScenarioFailedCount: proto.Int32(int32(specResult.scenarioFailedCount)),
			Failed:              proto.Bool(specResult.isFailed),
			FailedDataTableRows: specResult.failedDataTableRows,
			ExecutionTime:       proto.Int64(specResult.executionTime),
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

func convertToProtoStepValue(stepValue *stepValue) *ProtoStepValue {
	return &ProtoStepValue{
		StepValue:              proto.String(stepValue.stepValue),
		ParameterizedStepValue: proto.String(stepValue.parameterizedStepValue),
		Parameters:             stepValue.args,
	}
}

func newSpecResult(specification *specification) *specResult {
	return &specResult{
		protoSpec:           newProtoSpec(specification),
		failedDataTableRows: make([]int32, 0),
	}
}

func newProtoSpec(specification *specification) *ProtoSpec {

	return &ProtoSpec{
		Items:         make([]*ProtoItem, 0),
		SpecHeading:   proto.String(specification.heading.value),
		IsTableDriven: proto.Bool(false),
		FileName:      proto.String(specification.fileName),
		Tags:          getTags(specification.tags),
	}

}

func newProtoScenario(scenario *scenario) *ProtoScenario {
	return &ProtoScenario{
		ScenarioHeading: proto.String(scenario.heading.value),
		Failed:          proto.Bool(false),
		Tags:            getTags(scenario.tags),
		Contexts:        make([]*ProtoItem, 0),
		ExecutionTime:   proto.Int64(0),
	}
}

func getTags(tags *tags) []string {
	if tags != nil {
		return tags.values
	} else {
		return make([]string, 0)
	}
}
