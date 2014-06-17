package main

import (
	"code.google.com/p/goprotobuf/proto"
)

type suiteResult struct {
	specResults      []*specResult
	preSuite         *ProtoHookFailure
	postSuite        *ProtoHookFailure
	isFailed         bool
	specsFailedCount int
}

type specResult struct {
	protoSpec           *ProtoSpec
	scenarioFailedCount int
	scenarioCount       int
	isFailed            bool
}

type scenarioResult struct {
	protoScenario *ProtoScenario
}

type result interface {
	getPreHook() **ProtoHookFailure
	getPostHook() **ProtoHookFailure
	setFailure()
}

func (suiteResult *suiteResult) getPreHook() **ProtoHookFailure {
	return &suiteResult.preSuite
}

func (suiteResult *suiteResult) getPostHook() **ProtoHookFailure {
	return &suiteResult.postSuite
}

func (suiteResult *suiteResult) setFailure() {
	suiteResult.isFailed = true
}

func (specResult *specResult) getPreHook() **ProtoHookFailure {
	return &specResult.protoSpec.PreHookFailure
}

func (specResult *specResult) getPostHook() **ProtoHookFailure {
	return &specResult.protoSpec.PostHookFailure
}

func (specResult *specResult) setFailure() {
	specResult.isFailed = true
}

func (scenarioResult *scenarioResult) getPreHook() **ProtoHookFailure {
	return &scenarioResult.protoScenario.PreHookFailure
}

func (scenarioResult *scenarioResult) getPostHook() **ProtoHookFailure {
	return &scenarioResult.protoScenario.PostHookFailure
}

func (scenarioResult *scenarioResult) setFailure() {
	scenarioResult.protoScenario.Failed = proto.Bool(true)
}

func (specResult *specResult) addSpecItems(spec *specification) {
	for _, item := range spec.items {
		if item.kind() != scenarioKind {
			specResult.protoSpec.Items = append(specResult.protoSpec.Items, convertToProtoItem(item))
		}
	}
}

func newSuiteResult() *suiteResult {
	result := new(suiteResult)
	result.specResults = make([]*specResult, 0)
	return result
}

func addPreHook(result result, executionResult *ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.getPreHook()) = getProtoHookFailure(executionResult)
		result.setFailure()
	}
}

func addPostHook(result result, executionResult *ProtoExecutionResult) {
	if executionResult.GetFailed() {
		*(result.getPostHook()) = getProtoHookFailure(executionResult)
		result.setFailure()
	}
}

func (suiteResult *suiteResult) addSpecResult(specResult *specResult) {
	suiteResult.isFailed = specResult.isFailed
	if specResult.isFailed {
		suiteResult.specsFailedCount++
	}
	suiteResult.specResults = append(suiteResult.specResults, specResult)

}

func getProtoHookFailure(executionResult *ProtoExecutionResult) *ProtoHookFailure {
	return &ProtoHookFailure{StackTrace: executionResult.StackTrace, ErrorMessage: executionResult.ErrorMessage, ScreenShot: executionResult.ScreenShot}
}

func (specResult *specResult) setFileName(fileName string) {
	specResult.protoSpec.FileName = proto.String(fileName)
}

func (specResult *specResult) addScenarioResults(scenarioResults []*scenarioResult) {
	for _, scenarioResult := range scenarioResults {
		if scenarioResult.protoScenario.GetFailed() {
			specResult.isFailed = true
			specResult.scenarioFailedCount++
		}
		specResult.protoSpec.Items = append(specResult.protoSpec.Items, &ProtoItem{ItemType: ProtoItem_Scenario.Enum(), Scenario: scenarioResult.protoScenario})
	}
	specResult.scenarioCount += len(scenarioResults)
}

func (specResult *specResult) addTableDrivenScenarioResult(scenarioResults [][](*scenarioResult)) {
	numberOfScenarios := len(scenarioResults[0])

	for i := 0; i < numberOfScenarios; i++ {
		protoTableDrivenScenario := &ProtoTableDrivenScenario{Scenarios: make([]*ProtoScenario, 0)}
		scenarioFailed := false
		for _, eachRow := range scenarioResults {
			protoTableDrivenScenario.Scenarios = append(protoTableDrivenScenario.GetScenarios(), eachRow[i].protoScenario)
			if eachRow[i].protoScenario.GetFailed() {
				scenarioFailed = true
			}
		}
		if scenarioFailed {
			specResult.scenarioFailedCount++
			specResult.isFailed = true
		}
		protoItem := &ProtoItem{ItemType: ProtoItem_TableDrivenScenario.Enum(), TableDrivenScenario: protoTableDrivenScenario}
		specResult.protoSpec.Items = append(specResult.protoSpec.Items, protoItem)
	}
	specResult.protoSpec.IsTableDriven = proto.Bool(true)
	specResult.scenarioCount += numberOfScenarios
}

func (scenarioResult *scenarioResult) addItems(protoItems []*ProtoItem) {
	scenarioResult.protoScenario.ScenarioItems = append(scenarioResult.protoScenario.ScenarioItems, protoItems...)
}
