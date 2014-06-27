package main

import (
	"code.google.com/p/goprotobuf/proto"
	"net"
)

type specExecutor struct {
	specification        *specification
	dataTableIndex       int
	connection           net.Conn
	conceptDictionary    *conceptDictionary
	pluginHandler        *pluginHandler
	currentExecutionInfo *ExecutionInfo
	specResult           *specResult
}

func (specExecutor *specExecutor) initialize(specificationToExecute *specification, connection net.Conn, pluginHandler *pluginHandler) {
	specExecutor.specification = specificationToExecute
	specExecutor.connection = connection
	specExecutor.pluginHandler = pluginHandler
}

type stepValidationError struct {
	step     *step
	message  string
	fileName string
}

func (e *stepValidationError) Error() string {
	return e.message
}

func (e *specExecutor) executeBeforeSpecHook() *ProtoExecutionResult {
	message := &Message{MessageType: Message_SpecExecutionStarting.Enum(),
		SpecExecutionStartingRequest: &SpecExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message, e.specResult)
}

func (e *specExecutor) executeAfterSpecHook() *ProtoExecutionResult {
	message := &Message{MessageType: Message_SpecExecutionEnding.Enum(),
		SpecExecutionEndingRequest: &SpecExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	return e.executeHook(message, e.specResult)
}

func (e *specExecutor) executeHook(message *Message, execTimeTracker execTimeTracker) *ProtoExecutionResult {
	e.pluginHandler.notifyPlugins(message)
	executionResult := executeAndGetStatus(e.connection, message)
	execTimeTracker.addExecTime(executionResult.GetExecutionTime())
	return executionResult
}

func (specExecutor *specExecutor) execute() *specResult {
	specInfo := &SpecInfo{Name: proto.String(specExecutor.specification.heading.value),
		FileName: proto.String(specExecutor.specification.fileName),
		IsFailed: proto.Bool(false), Tags: getTagValue(specExecutor.specification.tags)}
	specExecutor.currentExecutionInfo = &ExecutionInfo{CurrentSpec: specInfo}

	getCurrentConsole().writeSpecHeading(specInfo.GetName())

	specExecutor.specResult = newSpecResult(specExecutor.specification)
	specExecutor.specResult.addSpecItems(specExecutor.specification)

	beforeSpecHookStatus := specExecutor.executeBeforeSpecHook()
	if beforeSpecHookStatus.GetFailed() {
		addPreHook(specExecutor.specResult, beforeSpecHookStatus)
		specExecutor.currentExecutionInfo.setSpecFailure()
	} else {
		getCurrentConsole().writeSteps(specExecutor.specification.contexts)
		dataTableRowCount := specExecutor.specification.dataTable.getRowCount()
		if dataTableRowCount == 0 {
			scenarioResult := specExecutor.executeScenarios()
			specExecutor.specResult.addScenarioResults(scenarioResult)
		} else {
			specExecutor.executeTableDrivenScenarios()
		}
	}

	afterSpecHookStatus := specExecutor.executeAfterSpecHook()
	if afterSpecHookStatus.GetFailed() {
		addPostHook(specExecutor.specResult, afterSpecHookStatus)
		specExecutor.currentExecutionInfo.setSpecFailure()
	}
	return specExecutor.specResult
}

func (specExecutor *specExecutor) executeTableDrivenScenarios() {
	var dataTableScenarioExecutionResult [][]*scenarioResult
	dataTableRowCount := specExecutor.specification.dataTable.getRowCount()
	for specExecutor.dataTableIndex = 0; specExecutor.dataTableIndex < dataTableRowCount; specExecutor.dataTableIndex++ {
		dataTableScenarioExecutionResult = append(dataTableScenarioExecutionResult, specExecutor.executeScenarios())
	}
	specExecutor.specResult.addTableDrivenScenarioResult(dataTableScenarioExecutionResult)
}

func getTagValue(tags *tags) []string {
	tagValues := make([]string, 0)
	if tags != nil {
		tagValues = append(tagValues, tags.values...)
	}
	return tagValues
}

func (executor *specExecutor) validateSpecification() []*stepValidationError {
	validationErrors := make([]*stepValidationError, 0)

	contextSteps := executor.specification.contexts
	contextStepsValidationErrors := executor.validateSteps(contextSteps)
	validationErrors = append(validationErrors, contextStepsValidationErrors...)

	for _, scenario := range executor.specification.scenarios {
		stepValidationErrors := executor.validateSteps(scenario.steps)
		validationErrors = append(validationErrors, stepValidationErrors...)
	}
	return validationErrors
}

func (executor *specExecutor) validateSteps(steps []*step) []*stepValidationError {
	validationErrors := make([]*stepValidationError, 0)
	for _, step := range steps {
		if step.isConcept {
			errors := executor.validateConcept(step)
			validationErrors = append(validationErrors, errors...)
		} else {
			err := executor.validateStep(step)
			if err != nil {
				validationErrors = append(validationErrors, err)
			}
		}
	}
	return validationErrors
}

func (executor *specExecutor) validateConcept(concept *step) []*stepValidationError {
	validationErrors := make([]*stepValidationError, 0)
	for _, conceptStep := range concept.conceptSteps {
		if err := executor.validateStep(conceptStep); err != nil {
			err.fileName = executor.conceptDictionary.search(concept.value).fileName
			validationErrors = append(validationErrors, err)
		}
	}
	return validationErrors
}

func (executor *specExecutor) validateStep(step *step) *stepValidationError {
	message := &Message{MessageType: Message_StepValidateRequest.Enum(),
		StepValidateRequest: &StepValidateRequest{StepText: proto.String(step.value), NumberOfParameters: proto.Int(len(step.args))}}
	response, err := getResponse(executor.connection, message)
	if err != nil {
		return &stepValidationError{step: step, message: err.Error(), fileName: executor.specification.fileName}
	}

	if response.GetMessageType() == Message_StepValidateResponse {
		validateResponse := response.GetStepValidateResponse()
		if !validateResponse.GetIsValid() {
			return &stepValidationError{step: step, message: validateResponse.GetErrorMessage(), fileName: executor.specification.fileName}
		}
	} else {
		panic("Expected a validate step response")
	}

	return nil
}

func (executor *specExecutor) executeBeforeScenarioHook(scenarioResult *scenarioResult) *ProtoExecutionResult {
	message := &Message{MessageType: Message_ScenarioExecutionStarting.Enum(),
		ScenarioExecutionStartingRequest: &ScenarioExecutionStartingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	return executor.executeHook(message, scenarioResult)
}

func (executor *specExecutor) executeAfterScenarioHook(scenarioResult *scenarioResult) *ProtoExecutionResult {
	message := &Message{MessageType: Message_ScenarioExecutionEnding.Enum(),
		ScenarioExecutionEndingRequest: &ScenarioExecutionEndingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	return executor.executeHook(message, scenarioResult)
}

func (specExecutor *specExecutor) executeScenarios() []*scenarioResult {
	scenarioResults := make([]*scenarioResult, 0)
	for _, scenario := range specExecutor.specification.scenarios {
		scenarioResults = append(scenarioResults, specExecutor.executeScenario(scenario))
	}
	return scenarioResults
}

func (executor *specExecutor) executeScenario(scenario *scenario) *scenarioResult {
	executor.currentExecutionInfo.CurrentScenario = &ScenarioInfo{Name: proto.String(scenario.heading.value), Tags: getTagValue(scenario.tags), IsFailed: proto.Bool(false)}
	getCurrentConsole().writeScenarioHeading(scenario.heading.value)

	scenarioResult := &scenarioResult{newProtoScenario(scenario)}
	beforeHookExecutionStatus := executor.executeBeforeScenarioHook(scenarioResult)
	if beforeHookExecutionStatus.GetFailed() {
		addPreHook(scenarioResult, beforeHookExecutionStatus)
		executor.currentExecutionInfo.setScenarioFailure()
	} else {
		contextProtoItems, failed := executor.executeContext()
		scenarioResult.addContexts(contextProtoItems)
		if !failed {
			scenarioProtoItems, isFailure := executor.executeItems(scenario.items)
			if isFailure {
				scenarioResult.setFailure()
			}
			scenarioResult.addItems(scenarioProtoItems)
		}
	}

	afterHookExecutionStatus := executor.executeAfterScenarioHook(scenarioResult)
	addPostHook(scenarioResult, afterHookExecutionStatus)
	return scenarioResult
}

func (executor *specExecutor) executeContext() ([]*ProtoItem, bool) {
	contextSteps := executor.specification.contexts
	items := make([]item, len(contextSteps))
	for i, context := range contextSteps {
		items[i] = context
	}
	contextStepItems, failure := executor.executeItems(items)
	return contextStepItems, failure
}

func (executor *specExecutor) executeItems(items []item) ([]*ProtoItem, bool) {
	protoItems := make([]*ProtoItem, 0)
	argLookup := new(argLookup).fromDataTableRow(&executor.specification.dataTable, executor.dataTableIndex)
	for _, item := range items {
		protoItems = append(protoItems, executor.resolveToProtoItem(item, argLookup))
	}

	for _, protoItem := range protoItems {
		failure := executor.executeItem(protoItem)
		if failure == true {
			return protoItems, true
		}
	}
	return protoItems, false
}

func (executor *specExecutor) resolveToProtoItem(item item, lookup *argLookup) *ProtoItem {
	var protoItem *ProtoItem
	switch item.kind() {
	case stepKind:
		if (item.(*step)).isConcept {
			protoItem = executor.resolveToProtoConceptItem(item.(*step), lookup)
		} else {
			protoItem = executor.resolveToProtoStepItem(item.(*step), lookup)
		}
		break

	default : protoItem = convertToProtoItem(item)
	}
	return protoItem
}

func (executor *specExecutor) resolveToProtoStepItem(step *step, lookup *argLookup) *ProtoItem {
	protoStepItem := convertToProtoItem(step)
	paramResolver := new(paramResolver)
	parameters := paramResolver.getResolvedParams(step.args, lookup, executor.dataTableLookup())
	updateProtoStepParameters(protoStepItem.Step, parameters)
	return protoStepItem
}

func (executor *specExecutor) resolveToProtoConceptItem(concept *step, lookup *argLookup) *ProtoItem {
	protoConceptItem := convertToProtoItem(concept)

	conceptLookup := concept.lookup.getCopy()
	executor.populateConceptDynamicParams(conceptLookup, lookup)

	paramResolver := new(paramResolver)
	conceptStepParameters := paramResolver.getResolvedParams(concept.args, lookup, executor.dataTableLookup())
	updateProtoStepParameters(protoConceptItem.Concept.ConceptStep, conceptStepParameters)

	for stepIndex, step := range concept.conceptSteps {
		stepParameters := paramResolver.getResolvedParams(step.args, conceptLookup, executor.dataTableLookup())
		updateProtoStepParameters(protoConceptItem.Concept.Steps[stepIndex], stepParameters)
	}
	return protoConceptItem
}

func updateProtoStepParameters(protoStep *ProtoStep, parameters []*Parameter) {
	paramIndex := 0
	for fragmentIndex, fragment := range protoStep.Fragments {
		if (fragment.GetFragmentType() == Fragment_Parameter) {
			protoStep.Fragments[fragmentIndex].Parameter = parameters[paramIndex]
			paramIndex++
		}
	}
}

func (executor *specExecutor) dataTableLookup() *argLookup {
	return new(argLookup).fromDataTableRow(&executor.specification.dataTable, executor.dataTableIndex)
}


func (executor *specExecutor) executeItem(protoItem *ProtoItem) bool {
	if protoItem.GetItemType() == ProtoItem_Concept {
		return executor.executeConcept(protoItem.GetConcept())
	} else if protoItem.GetItemType() == ProtoItem_Step {
		return executor.executeStep(protoItem.GetStep())
	} else {
		return false
	}
}

func (executor *specExecutor) executeSteps(protoSteps []*ProtoStep) bool {
	for _, protoStep := range protoSteps {
		failure := executor.executeStep(protoStep)
		if failure {
			return true
		}
	}
	return false
}

func (executor *specExecutor) executeConcept(protoConcept *ProtoConcept) bool {
	executor.executeSteps(protoConcept.Steps)
	executor.setExecutionResultForConcept(protoConcept)
	return protoConcept.GetConceptExecutionResult().GetExecutionResult().GetFailed()
}

func (executor *specExecutor) setExecutionResultForConcept(protoConcept *ProtoConcept) {
	var conceptExecutionTime int64
	for _, step := range protoConcept.GetSteps() {
		stepExecResult := step.GetStepExecutionResult().GetExecutionResult()
		conceptExecutionTime += stepExecResult.GetExecutionTime()
		if stepExecResult.GetFailed() {
			conceptExecutionResult := &ProtoStepExecutionResult{ExecutionResult: stepExecResult}
			conceptExecutionResult.ExecutionResult.ExecutionTime = proto.Int64(conceptExecutionTime)
			protoConcept.ConceptExecutionResult = conceptExecutionResult
			protoConcept.ConceptStep.StepExecutionResult = conceptExecutionResult
			return
		}
	}
	protoConcept.ConceptExecutionResult = &ProtoStepExecutionResult{ExecutionResult: &ProtoExecutionResult{Failed: proto.Bool(false), ExecutionTime: proto.Int64(conceptExecutionTime)}}
	protoConcept.ConceptStep.StepExecutionResult = protoConcept.ConceptExecutionResult
}

func printStatus(executionResult *ProtoExecutionResult) {
	getCurrentConsole().writeError(executionResult.GetErrorMessage())
	getCurrentConsole().writeError(executionResult.GetStackTrace())
}

func (executor *specExecutor) executeStep(protoStep *ProtoStep) bool {

	stepRequest := executor.createStepRequest(protoStep)
	stepWithResolvedArgs := createStepFromStepRequest(stepRequest)
	console := getCurrentConsole()
	console.writeStep(stepWithResolvedArgs)

	protoStepExecResult := &ProtoStepExecutionResult{}
	executor.currentExecutionInfo.CurrentStep = &StepInfo{Step: stepRequest, IsFailed: proto.Bool(false)}

	beforeHookStatus := executor.executeBeforeStepHook()
	if beforeHookStatus.GetFailed() {
		protoStepExecResult.PreHookFailure = getProtoHookFailure(beforeHookStatus)
		protoStepExecResult.ExecutionResult = &ProtoExecutionResult{Failed: proto.Bool(true)}
		executor.currentExecutionInfo.setStepFailure()
		printStatus(beforeHookStatus)
	} else {
		executeStepMessage := &Message{MessageType: Message_ExecuteStep.Enum(), ExecuteStepRequest: stepRequest}
		stepExecutionStatus := executeAndGetStatus(executor.connection, executeStepMessage)
		if stepExecutionStatus.GetFailed() {
			executor.currentExecutionInfo.setStepFailure()
			printStatus(stepExecutionStatus)
		}
		protoStepExecResult.ExecutionResult = stepExecutionStatus
	}
	afterStepHookStatus := executor.executeAfterStepHook()
	addExecutionTimes(protoStepExecResult, beforeHookStatus, afterStepHookStatus)
	if afterStepHookStatus.GetFailed() {
		executor.currentExecutionInfo.setStepFailure()
		printStatus(afterStepHookStatus)
		protoStepExecResult.PostHookFailure = getProtoHookFailure(afterStepHookStatus)
	}

	console.writeStepFinished(stepWithResolvedArgs, protoStepExecResult.GetExecutionResult().GetFailed())
	protoStep.StepExecutionResult = protoStepExecResult
	return protoStep.GetStepExecutionResult().GetExecutionResult().GetFailed()
}

func addExecutionTimes(stepExecResult *ProtoStepExecutionResult, execResults ...*ProtoExecutionResult) {
	for _, execResult := range execResults {
		currentScenarioExecTime := stepExecResult.ExecutionResult.ExecutionTime
		if currentScenarioExecTime == nil {
			stepExecResult.ExecutionResult.ExecutionTime = proto.Int64(execResult.GetExecutionTime())
		} else {
			stepExecResult.ExecutionResult.ExecutionTime = proto.Int64(*currentScenarioExecTime+execResult.GetExecutionTime())
		}
	}
}

func (executor *specExecutor) executeBeforeStepHook() *ProtoExecutionResult {
	message := &Message{MessageType: Message_StepExecutionStarting.Enum(),
		StepExecutionStartingRequest: &StepExecutionStartingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	executor.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(executor.connection, message)
}

func (executor *specExecutor) executeAfterStepHook() *ProtoExecutionResult {
	message := &Message{MessageType: Message_StepExecutionEnding.Enum(),
		StepExecutionEndingRequest: &StepExecutionEndingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	executor.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(executor.connection, message)
}

func (executor *specExecutor) createStepRequest(protoStep *ProtoStep) *ExecuteStepRequest {
	stepRequest := &ExecuteStepRequest{ParsedStepText: proto.String(protoStep.GetParsedText()), ActualStepText: proto.String(protoStep.GetActualText())}
	stepRequest.Parameters = getParameters(protoStep.GetFragments())
	return stepRequest
}

func (executor *specExecutor) getCurrentDataTableValueFor(columnName string) string {
	return executor.specification.dataTable.get(columnName)[executor.dataTableIndex].value
}

func executeAndGetStatus(connection net.Conn, message *Message) *ProtoExecutionResult {
	response, err := getResponse(connection, message)
	if err != nil {
		return &ProtoExecutionResult{Failed: proto.Bool(true), ErrorMessage: proto.String(err.Error())}
	}

	if response.GetMessageType() == Message_ExecutionStatusResponse {
		executionResult := response.GetExecutionStatusResponse().GetExecutionResult()
		if executionResult == nil {
			panic("ProtoExecutionResult should not be nil")
		}
		return executionResult
	} else {
		panic("Expected ExecutionStatusResponse")
	}
}

func (executor *specExecutor) populateConceptDynamicParams(conceptLookup *argLookup, dataTableLookup *argLookup) {
	for key, _ := range conceptLookup.paramIndexMap {
		conceptLookupArg := conceptLookup.getArg(key)
		if conceptLookupArg.argType == dynamic {
			resolvedArg := dataTableLookup.getArg(conceptLookupArg.value)
			conceptLookup.addArgValue(key, resolvedArg)
		}
	}
}

func (executionInfo *ExecutionInfo) setSpecFailure() {
	executionInfo.CurrentSpec.IsFailed = proto.Bool(true)
}

func (executionInfo *ExecutionInfo) setScenarioFailure() {
	executionInfo.setSpecFailure()
	executionInfo.CurrentScenario.IsFailed = proto.Bool(true)
}

func (executionInfo *ExecutionInfo) setStepFailure() {
	executionInfo.setScenarioFailure()
	executionInfo.CurrentStep.IsFailed = proto.Bool(true)
}

func getParameters(fragments []*Fragment) []*Parameter {
	parameters := make([]*Parameter, 0)
	for _, fragment := range fragments {
		if fragment.GetFragmentType() == Fragment_Parameter {
			parameters = append(parameters, fragment.GetParameter())
		}
	}
	return parameters
}
