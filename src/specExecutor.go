package main

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	"net"
)

type specExecutor struct {
	specification     *specification
	dataTableIndex    int
	connection        net.Conn
	conceptDictionary *conceptDictionary
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
		SpecExecutionStartingRequest: &SpecExecutionStartingRequest{SpecName: proto.String(e.specification.heading.value),
			SpecFile: proto.String(e.specification.fileName)}}

	return executeAndGetStatus(e.connection, message)
}

func (e *specExecutor) executeAfterSpecHook() *ExecutionStatus {
	message := &Message{MessageType: Message_SpecExecutionEnding.Enum(),
		SpecExecutionEndingRequest: &SpecExecutionEndingRequest{SpecName: proto.String(e.specification.heading.value),
			SpecFile: proto.String(e.specification.fileName)}}

	return executeAndGetStatus(e.connection, message)
}

func (executor *specExecutor) execute() *specExecutionStatus {
	specExecutionStatus := &specExecutionStatus{specification: executor.specification, scenariosExecutionStatuses: make(map[int][]*scenarioExecutionStatus)}

	beforeSpecHookStatus := executor.executeBeforeSpecHook()
	if beforeSpecHookStatus.GetPassed() {
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
	}

	afterSpecHookStatus := executor.executeAfterSpecHook()
	specExecutionStatus.hooksExecutionStatuses = append(specExecutionStatus.hooksExecutionStatuses, beforeSpecHookStatus, afterSpecHookStatus)

	return specExecutionStatus
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
		StepValidateRequest: &StepValidateRequest{StepText: proto.String(step.value)}}
	response, err := getResponse(executor.connection, message)
	if err != nil {
		return &stepValidationError{step: step, message: err.Error(), fileName: executor.specification.fileName}
	}

	if response.GetMessageType() == Message_StepValidateResponse {
		validateResponse := response.GetStepValidateResponse()
		if !validateResponse.GetIsValid() {
			return &stepValidationError{step: step, message: "", fileName: executor.specification.fileName}
		}
	} else {
		panic("Expected a validate step response")
	}

	return nil
}

func (executor *specExecutor) executeBeforeScenarioHook() *ExecutionStatus {
	message := &Message{MessageType: Message_ScenarioExecutionStarting.Enum(),
		ScenarioExecutionStartingRequest: &ScenarioExecutionStartingRequest{}}
	return executeAndGetStatus(executor.connection, message)
}

func (executor *specExecutor) executeAfterScenarioHook() *ExecutionStatus {
	message := &Message{MessageType: Message_ScenarioExecutionEnding.Enum(),
		ScenarioExecutionEndingRequest: &ScenarioExecutionEndingRequest{}}
	return executeAndGetStatus(executor.connection, message)
}

func (executor *specExecutor) executeScenarios() []*scenarioExecutionStatus {
	var scenarioExecutionStatuses []*scenarioExecutionStatus
	argLookup := new(argLookup).fromDataTableRow(&executor.specification.dataTable, executor.dataTableIndex)
	for _, scenario := range executor.specification.scenarios {
		status := executor.executeScenario(scenario, argLookup)
		scenarioExecutionStatuses = append(scenarioExecutionStatuses, status)
	}
	return scenarioExecutionStatuses
}

func (executor *specExecutor) executeScenario(scenario *scenario, argLookup *argLookup) *scenarioExecutionStatus {
	scenarioExecutionStatus := &scenarioExecutionStatus{scenario: scenario}
	beforeHookExecutionStatus := executor.executeBeforeScenarioHook()
	if beforeHookExecutionStatus.GetPassed() {
		contextStepsExecutionStatuses := executor.executeSteps(executor.specification.contexts, argLookup)
		scenarioExecutionStatus.stepExecutionStatuses = append(scenarioExecutionStatus.stepExecutionStatuses, contextStepsExecutionStatuses...)
		contextFailed := false
		for _, s := range contextStepsExecutionStatuses {
			if !s.passed {
				contextFailed = true
				break
			}
		}

		if !contextFailed {
			stepExecutionStatuses := executor.executeSteps(scenario.steps, argLookup)
			scenarioExecutionStatus.stepExecutionStatuses = append(scenarioExecutionStatus.stepExecutionStatuses, stepExecutionStatuses...)
		}
	}

	afterHookExecutionStatus := executor.executeAfterScenarioHook()
	scenarioExecutionStatus.hooksExecutionStatuses = append(scenarioExecutionStatus.hooksExecutionStatuses, afterHookExecutionStatus)
	return scenarioExecutionStatus
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
	var status *stepExecutionStatus
	var statuses []*stepExecutionStatus
	for _, step := range steps {
		if step.isConcept {
			status = executor.executeConcept(step, argLookup)
		} else {
			status = executor.executeStep(step, argLookup)
		}
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
func (executor *specExecutor) executeStep(step *step, argLookup *argLookup) *stepExecutionStatus {
	stepExecStatus := &stepExecutionStatus{passed: true}
	printStatus := func(execStatus *ExecutionStatus) {
		fmt.Printf("\x1b[31;1m%s\n\x1b[0m", execStatus.GetErrorMessage())
		fmt.Printf("\x1b[31;1m%s\n\x1b[0m", execStatus.GetStackTrace())
	}

	message := &Message{MessageType: Message_StepExecutionStarting.Enum(),
		StepExecutionStartingRequest: &StepExecutionStartingRequest{}}
	var status *ExecutionStatus
	status = executeAndGetStatus(executor.connection, message)
	if status.GetPassed() {
		stepRequest := executor.createStepRequest(step, argLookup)
		message = &Message{MessageType: Message_ExecuteStep.Enum(),
			ExecuteStepRequest: stepRequest}
		status = executeAndGetStatus(executor.connection, message)
		if !status.GetPassed() {
			printStatus(status)
			stepExecStatus.addExecutionStatus(status)
		}
	} else {
		printStatus(status)
		stepExecStatus.addExecutionStatus(status)
	}

	message = &Message{MessageType: Message_StepExecutionEnding.Enum(),
		StepExecutionEndingRequest: &StepExecutionEndingRequest{}}
	status = executeAndGetStatus(executor.connection, message)
	if !status.GetPassed() {
		printStatus(status)
		stepExecStatus.executionStatus = append(stepExecStatus.executionStatus, status)
	}

	if stepExecStatus.passed {
		fmt.Printf("=> \x1b[32;1m%s\n\x1b[0m", step.lineText)
	} else {
		fmt.Printf("\x1b[31;1m%s\n\x1b[0m", step.lineText)
	}

	return stepExecStatus
}

func (executor *specExecutor) createStepRequest(step *step, argLookup *argLookup) *ExecuteStepRequest {
	stepRequest := &ExecuteStepRequest{StepText: proto.String(step.value)}
	stepRequest.Args = executor.createStepArgs(step.args, step.inlineTable, argLookup)
	return stepRequest
}

func (executor *specExecutor) createStepArgs(args []*stepArg, inlineTable table, argLookup *argLookup) []*Argument {
	arguments := make([]*Argument, 0)
	for _, arg := range args {
		argument := new(Argument)
		if arg.argType == static || arg.argType == specialString {
			argument.Type = proto.String("string")
			argument.Value = proto.String(arg.value)
		} else if arg.argType == dynamic {
			resolvedArg := argLookup.getArg(arg.value)
			//In case a special table used in a concept, you will get a dynamic table value which has to be resolved from the concept lookup
			if resolvedArg.table.isInitialized() {
				argument.Type = proto.String("table")
				argument.Table = executor.createStepTable(&resolvedArg.table)
			} else {
				argument.Type = proto.String("string")
				argument.Value = proto.String(resolvedArg.value)
			}
		} else {
			argument.Type = proto.String("table")
			argument.Table = executor.createStepTable(&arg.table)
		}
		arguments = append(arguments, argument)
	}

	if inlineTable.isInitialized() {
		inlineTableArg := executor.createStepTable(&inlineTable)
		arguments = append(arguments, &Argument{Type: proto.String("table"), Table: inlineTableArg})
	}
	return arguments
}

func (executor *specExecutor) getCurrentDataTableValueFor(columnName string) string {
	return executor.specification.dataTable.get(columnName)[executor.dataTableIndex]
}

func (executor *specExecutor) createStepTable(table *table) *ProtoTable {
	protoTable := new(ProtoTable)
	tableRows := make([]*TableRow, 0)
	tableRows = append(tableRows, &TableRow{Cells: table.headers})
	for i := 0; i < len(table.columns[0]); i++ {
		row := make([]string, 0)
		for _, header := range table.headers {
			row = append(row, table.get(header)[i])
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
