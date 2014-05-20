package main

import (
	"code.google.com/p/goprotobuf/proto"
	"net"
)

type itemExecutor func(item, *specExecutor) *stepExecutionStatus

type specExecutor struct {
	specification        *specification
	dataTableIndex       int
	connection           net.Conn
	conceptDictionary    *conceptDictionary
	pluginHandler        *pluginHandler
	currentExecutionInfo *ExecutionInfo
}

func (specExecutor *specExecutor) initialize(specificationToExecute *specification, connection net.Conn, pluginHandler *pluginHandler) {
	specExecutor.specification = specificationToExecute
	specExecutor.connection = connection
	specExecutor.pluginHandler = pluginHandler
}

type specExecutionStatus struct {
	specification *specification
	// Key will be the datatable index
	// if no datatable, 0th key points to the execution status
	scenariosExecutionStatuses map[int][]*scenarioExecutionStatus
	hooksExecutionStatuses     []*ExecutionStatus
}

type stepValidationError struct {
	step     *step
	message  string
	fileName string
}

func (e *stepValidationError) Error() string {
	return e.message
}

func (status *specExecutionStatus) isFailed() bool {
	if status.hooksExecutionStatuses != nil {
		for _, s := range status.hooksExecutionStatuses {
			if !s.GetPassed() {
				return true
			}
		}
	}

	for _, v := range status.scenariosExecutionStatuses {
		for _, scenario := range v {
			if scenario.isFailed() {
				return true
			}
		}
	}

	return false
}

func (e *specExecutor) executeBeforeSpecHook() *ExecutionStatus {
	message := &Message{MessageType: Message_SpecExecutionStarting.Enum(),
		SpecExecutionStartingRequest: &SpecExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}

	e.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(e.connection, message)
}

func (e *specExecutor) executeAfterSpecHook() *ExecutionStatus {
	message := &Message{MessageType: Message_SpecExecutionEnding.Enum(),
		SpecExecutionEndingRequest: &SpecExecutionEndingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	e.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(e.connection, message)
}

func (executor *specExecutor) execute() *specExecutionStatus {
	specInfo := &SpecInfo{Name: proto.String(executor.specification.heading.value),
		FileName: proto.String(executor.specification.fileName),
		IsFailed: proto.Bool(false), Tags: getTagValue(executor.specification.tags)}
	executor.currentExecutionInfo = &ExecutionInfo{CurrentSpec: specInfo}
	getCurrentConsole().writeSpecHeading(executor.specification)

	specExecutionStatus := &specExecutionStatus{specification: executor.specification, scenariosExecutionStatuses: make(map[int][]*scenarioExecutionStatus)}
	beforeSpecHookStatus := executor.executeBeforeSpecHook()
	if beforeSpecHookStatus.GetPassed() {
		getCurrentConsole().writeItems(executor.specification.items)
		dataTableRowCount := executor.specification.dataTable.getRowCount()
		if dataTableRowCount == 0 {
			scenariosExecutionStatuses := executor.executeScenarios()
			specExecutionStatus.scenariosExecutionStatuses[0] = scenariosExecutionStatuses
		} else {
			for executor.dataTableIndex = 0; executor.dataTableIndex < dataTableRowCount; executor.dataTableIndex++ {
				scenariosExecutionStatuses := executor.executeScenarios()
				specExecutionStatus.scenariosExecutionStatuses[executor.dataTableIndex] = scenariosExecutionStatuses
			}
		}
	} else {
		executor.currentExecutionInfo.setSpecFailure()
	}

	afterSpecHookStatus := executor.executeAfterSpecHook()
	specExecutionStatus.hooksExecutionStatuses = append(specExecutionStatus.hooksExecutionStatuses, beforeSpecHookStatus, afterSpecHookStatus)

	return specExecutionStatus
}

func getTagValue(tags *tags) []string {
	tagValues := make([]string, 0)
	if tags != nil {
		tagValues = append(tagValues, tags.values...)
	}
	return tagValues
}

type scenarioExecutionStatus struct {
	scenario               *scenario
	stepExecutionStatuses  []*stepExecutionStatus
	hooksExecutionStatuses []*ExecutionStatus
}

func (status *scenarioExecutionStatus) isFailed() bool {
	if status.hooksExecutionStatuses != nil {
		for _, hook := range status.hooksExecutionStatuses {
			if !hook.GetPassed() {
				return true
			}
		}
	}

	for _, step := range status.stepExecutionStatuses {
		if !step.passed {
			return true
		}
	}

	return false
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
		StepValidateRequest: &StepValidateRequest{StepText: proto.String(step.value), NumberOfArguments: proto.Int(len(step.args))}}
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

func (e *specExecutor) executeBeforeScenarioHook(scenario *scenario) *ExecutionStatus {
	message := &Message{MessageType: Message_ScenarioExecutionStarting.Enum(),
		ScenarioExecutionStartingRequest: &ScenarioExecutionStartingRequest{CurrentExecutionInfo: e.currentExecutionInfo}}
	e.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(e.connection, message)
}

func (executor *specExecutor) executeAfterScenarioHook() *ExecutionStatus {
	message := &Message{MessageType: Message_ScenarioExecutionEnding.Enum(),
		ScenarioExecutionEndingRequest: &ScenarioExecutionEndingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	executor.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(executor.connection, message)
}

func (executor *specExecutor) executeScenarios() []*scenarioExecutionStatus {
	var scenarioExecutionStatuses []*scenarioExecutionStatus
	for _, scenario := range executor.specification.scenarios {
		status := executor.executeScenario(scenario)
		scenarioExecutionStatuses = append(scenarioExecutionStatuses, status)
	}
	return scenarioExecutionStatuses
}

func (executor *specExecutor) executeScenario(scenario *scenario) *scenarioExecutionStatus {
	scenarioExecutionStatus := &scenarioExecutionStatus{scenario: scenario}
	executor.currentExecutionInfo.CurrentScenario = &ScenarioInfo{Name: proto.String(scenario.heading.value), Tags: getTagValue(scenario.tags), IsFailed: proto.Bool(false)}
	getCurrentConsole().writeScenarioHeading(scenario.heading.value)

	beforeHookExecutionStatus := executor.executeBeforeScenarioHook(scenario)
	if beforeHookExecutionStatus.GetPassed() {
		contextStepsExecutionStatuses, passed := executor.executeContext()
		scenarioExecutionStatus.stepExecutionStatuses = append(scenarioExecutionStatus.stepExecutionStatuses, contextStepsExecutionStatuses...)

		if passed {
			stepExecutionStatuses, _ := executor.executeItems(scenario.items)
			scenarioExecutionStatus.stepExecutionStatuses = append(scenarioExecutionStatus.stepExecutionStatuses, stepExecutionStatuses...)
		}
	} else {
		executor.currentExecutionInfo.setScenarioFailure()
	}

	afterHookExecutionStatus := executor.executeAfterScenarioHook()
	scenarioExecutionStatus.hooksExecutionStatuses = append(scenarioExecutionStatus.hooksExecutionStatuses, afterHookExecutionStatus)
	return scenarioExecutionStatus
}

func (executor *specExecutor) executeContext() ([]*stepExecutionStatus, bool) {
	contextSteps := executor.specification.contexts
	items := make([]item, len(contextSteps))
	for i, context := range contextSteps {
		items[i] = context
	}
	return executor.executeItems(items)
}

func (executor *specExecutor) executeItems(items []item) ([]*stepExecutionStatus, bool) {
	isFailure := false
	executionStatuses := make([]*stepExecutionStatus, 0)
	for _, item := range items {
		executionStatus := executor.executeItem(item)
		if executionStatus != nil {
			executionStatuses = append(executionStatuses, executionStatus)
			if !executionStatus.passed {
				isFailure = true
				break
			}
		}
	}
	return executionStatuses, !isFailure
}

func (executor *specExecutor) executeItem(item item) *stepExecutionStatus {
	if item.kind() != stepKind {
		getCurrentConsole().writeItem(item)
		return nil
	}

	argLookup := new(argLookup).fromDataTableRow(&executor.specification.dataTable, executor.dataTableIndex)
	step := item.(*step)
	if step.isConcept {
		return executor.executeConcept(step, argLookup)
	} else {
		return executor.executeStep(step, argLookup)
	}
}

type stepExecutionStatus struct {
	step                  *step
	resolvedArgs          []*Argument
	executionStatus       []*ExecutionStatus
	passed                bool
	isConcept             bool
	stepExecutionStatuses []*stepExecutionStatus
}

func (s *stepExecutionStatus) addExecutionStatus(executionStatus *ExecutionStatus) {
	if !executionStatus.GetPassed() {
		s.passed = false
	}
	s.executionStatus = append(s.executionStatus, executionStatus)
}

func (executor *specExecutor) executeSteps(steps []*step, argLookup *argLookup) []*stepExecutionStatus {
	var statuses []*stepExecutionStatus
	for _, step := range steps {
		status := executor.executeStep(step, argLookup)
		statuses = append(statuses, status)
		// TODO: handle recoverable error when verification API is done
		if !status.passed {
			break
		}
	}
	return statuses
}
func (executor *specExecutor) executeConcept(concept *step, dataTableLookup *argLookup) *stepExecutionStatus {
	conceptExecutionStatus := &stepExecutionStatus{passed: true, isConcept: true}
	conceptLookup := concept.lookup.getCopy()
	executor.populateConceptDynamicParams(conceptLookup, dataTableLookup)
	conceptExecutionStatus.stepExecutionStatuses = executor.executeSteps(concept.conceptSteps, conceptLookup)
	for _, status := range conceptExecutionStatus.stepExecutionStatuses {
		if !status.passed {
			conceptExecutionStatus.passed = false
			break
		}
	}
	return conceptExecutionStatus

}

func printStatus(execStatus *ExecutionStatus) {
	getCurrentConsole().writeError(execStatus.GetErrorMessage())
	getCurrentConsole().writeError(execStatus.GetStackTrace())
}

func (executor *specExecutor) executeStep(step *step, argLookup *argLookup) *stepExecutionStatus {
	stepRequest := executor.createStepRequest(step, argLookup)
	stepWithResolvedArgs := createStepFromStepRequest(stepRequest)
	console := getCurrentConsole()
	console.writeStep(stepWithResolvedArgs)

	stepExecStatus := &stepExecutionStatus{passed: true}
	executor.currentExecutionInfo.CurrentStep = &StepInfo{Step: stepRequest, IsFailed: proto.Bool(false)}

	beforeHookStatus := executor.executeBeforeStepHook()
	if beforeHookStatus.GetPassed() {
		executeStepMessage := &Message{MessageType: Message_ExecuteStep.Enum(), ExecuteStepRequest: stepRequest}
		stepExecutionStatus := executeAndGetStatus(executor.connection, executeStepMessage)
		if !stepExecutionStatus.GetPassed() {
			executor.currentExecutionInfo.setStepFailure()
			printStatus(stepExecutionStatus)
			stepExecStatus.addExecutionStatus(stepExecutionStatus)
		}
	} else {
		executor.currentExecutionInfo.setStepFailure()
		printStatus(beforeHookStatus)
		stepExecStatus.addExecutionStatus(beforeHookStatus)
	}

	afterStepHookStatus := executor.executeAfterStepHook()
	if !afterStepHookStatus.GetPassed() {
		executor.currentExecutionInfo.setStepFailure()
		printStatus(afterStepHookStatus)
		stepExecStatus.addExecutionStatus(afterStepHookStatus)
	}

	console.writeStepFinished(stepWithResolvedArgs, stepExecStatus.passed)
	return stepExecStatus
}

func (executor *specExecutor) executeBeforeStepHook() *ExecutionStatus {
	message := &Message{MessageType: Message_StepExecutionStarting.Enum(),
		StepExecutionStartingRequest: &StepExecutionStartingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	executor.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(executor.connection, message)
}

func (executor *specExecutor) executeAfterStepHook() *ExecutionStatus {
	message := &Message{MessageType: Message_StepExecutionEnding.Enum(),
		StepExecutionEndingRequest: &StepExecutionEndingRequest{CurrentExecutionInfo: executor.currentExecutionInfo}}
	executor.pluginHandler.notifyPlugins(message)
	return executeAndGetStatus(executor.connection, message)
}

func (executor *specExecutor) createStepRequest(step *step, argLookup *argLookup) *ExecuteStepRequest {
	stepRequest := &ExecuteStepRequest{ParsedStepText: proto.String(step.value), ActualStepText: proto.String(step.lineText)}
	stepRequest.Args = executor.createStepArgs(step.args, argLookup)
	return stepRequest
}

func (executor *specExecutor) createStepArgs(args []*stepArg, argLookup *argLookup) []*Argument {
	arguments := make([]*Argument, 0)
	for _, arg := range args {
		argument := new(Argument)
		if arg.argType == static {
			argument.Type = proto.String("string")
			argument.Value = proto.String(arg.value)
		} else if arg.argType == dynamic {
			resolvedArg := argLookup.getArg(arg.value)
			//In case a special table used in a concept, you will get a dynamic table value which has to be resolved from the concept lookup
			if resolvedArg.table.isInitialized() {
				argument.Type = proto.String("table")
				argument.Table = executor.createStepTable(&resolvedArg.table, argLookup)
			} else {
				argument.Type = proto.String("string")
				argument.Value = proto.String(resolvedArg.value)
			}
		} else {
			argument.Type = proto.String("table")
			argument.Table = executor.createStepTable(&arg.table, argLookup)
		}
		arguments = append(arguments, argument)
	}

	return arguments
}

func (executor *specExecutor) getCurrentDataTableValueFor(columnName string) string {
	return executor.specification.dataTable.get(columnName)[executor.dataTableIndex].value
}

func (executor *specExecutor) createStepTable(table *table, lookup *argLookup) *ProtoTable {
	protoTable := new(ProtoTable)
	tableRows := make([]*TableRow, 0)
	tableRows = append(tableRows, &TableRow{Cells: table.headers})
	for i := 0; i < len(table.columns[0]); i++ {
		row := make([]string, 0)
		for _, header := range table.headers {
			tableCell := table.get(header)[i]
			value := tableCell.value
			if tableCell.cellType == dynamic {
				if lookup.containsArg(tableCell.value) {
					value = lookup.getArg(tableCell.value).value
				} else {
					//if concept has a table with dynamic cell, arglookup won't have the table value, so fetch from datatable itself
					//todo cleanup
					tableLookup := new(argLookup).fromDataTableRow(&executor.specification.dataTable, executor.dataTableIndex)
					value = tableLookup.getArg(tableCell.value).value
				}
			}
			row = append(row, value)
		}
		tableRows = append(tableRows, &TableRow{Cells: row})
	}
	protoTable.Rows = tableRows
	return protoTable
}

func executeAndGetStatus(connection net.Conn, message *Message) *ExecutionStatus {
	response, err := getResponse(connection, message)
	if err != nil {
		return &ExecutionStatus{Passed: proto.Bool(false), ErrorMessage: proto.String(err.Error())}
	}

	if response.GetMessageType() == Message_ExecutionStatusResponse {
		status := response.GetExecutionStatusResponse().GetExecutionStatus()
		if status == nil {
			panic("ExecutionStatus should not be nil")
		}
		return status
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
