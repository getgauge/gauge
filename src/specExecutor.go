package main

import (
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"fmt"
	"net"
)

type specExecutor struct {
	specification  *specification
	dataTableIndex int
	connection     net.Conn
}

type specExecutionStatus struct {
	specification *specification
	// Key will be the datatable index
	// if no datatable, 0th key points to the execution status
	scenariosExecutionStatuses map[int][]*scenarioExecutionStatus
	hooksExecutionStatuses     []*ExecutionStatus
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
	for _, scenario := range executor.specification.scenarios {
		status := executor.executeScenario(scenario)
		scenarioExecutionStatuses = append(scenarioExecutionStatuses, status)
	}
	return scenarioExecutionStatuses
}

func (executor *specExecutor) executeScenario(scenario *scenario) *scenarioExecutionStatus {
	scenarioExecutionStatus := &scenarioExecutionStatus{scenario: scenario}
	beforeHookExecutionStatus := executor.executeBeforeScenarioHook()
	if beforeHookExecutionStatus.GetPassed() {
		contextStepsExecutionStatuses := executor.executeSteps(executor.specification.contexts)
		scenarioExecutionStatus.stepExecutionStatuses = append(scenarioExecutionStatus.stepExecutionStatuses, contextStepsExecutionStatuses...)
		contextFailed := false
		for _, s := range contextStepsExecutionStatuses {
			if !s.passed {
				contextFailed = true
				break
			}
		}

		if !contextFailed {
			stepExecutionStatuses := executor.executeSteps(scenario.steps)
			scenarioExecutionStatus.stepExecutionStatuses = append(scenarioExecutionStatus.stepExecutionStatuses, stepExecutionStatuses...)
		}
	}

	afterHookExecutionStatus := executor.executeAfterScenarioHook()
	scenarioExecutionStatus.hooksExecutionStatuses = append(scenarioExecutionStatus.hooksExecutionStatuses, afterHookExecutionStatus)
	return scenarioExecutionStatus
}

type stepExecutionStatus struct {
	step            *step
	resolvedArgs    []*Argument
	executionStatus []*ExecutionStatus
	passed          bool
}

func (s *stepExecutionStatus) addExecutionStatus(executionStatus *ExecutionStatus) {
	if !executionStatus.GetPassed() {
		s.passed = false
	}
	s.executionStatus = append(s.executionStatus, executionStatus)
}

func (executor *specExecutor) executeSteps(steps []*step) []*stepExecutionStatus {
	var statuses []*stepExecutionStatus
	for _, step := range steps {
		if validationErr := executor.validateStep(step); validationErr != nil {
			//TODO: this will be moved from here when bug #15 is fixed
			return nil
		}

		status := executor.executeStep(step)
		statuses = append(statuses, status)
		// TODO: handle recoverable error when verification API is done
		if !status.passed {
			break
		}
	}
	return statuses
}

func (executor *specExecutor) executeStep(step *step) *stepExecutionStatus {
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
		stepRequest := executor.createStepRequest(step)
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

func (executor *specExecutor) validateStep(step *step) error {
	message := &Message{MessageType: Message_StepValidateRequest.Enum(),
		StepValidateRequest: &StepValidateRequest{StepText: proto.String(step.value)}}
	response, err := getResponse(executor.connection, message)
	if err != nil {
		return err
	}

	if response.GetMessageType() == Message_StepValidateResponse {
		validateResponse := response.GetStepValidateResponse()
		if !validateResponse.GetIsValid() {
			fmt.Println("Not implemented")
			return errors.New("Step is not implemented")
		}
		return nil
	} else {
		panic("Expected a validate step response")
	}
}

func (executor *specExecutor) createStepRequest(step *step) *ExecuteStepRequest {
	stepRequest := &ExecuteStepRequest{StepText: proto.String(step.value)}
	stepRequest.Args = executor.createStepArgs(step.args, step.inlineTable)
	return stepRequest
}

func (executor *specExecutor) createStepArgs(args []*stepArg, inlineTable table) []*Argument {
	arguments := make([]*Argument, 0)
	for _, arg := range args {
		argument := new(Argument)
		if arg.argType == static || arg.argType == specialString {
			argument.Type = proto.String("string")
			argument.Value = proto.String(arg.value)
		} else if arg.argType == dynamic {
			argument.Type = proto.String("string")
			value := executor.getCurrentDataTableValueFor(arg.value)
			argument.Value = proto.String(value)
		} else {
			argument.Type = proto.String("table")
			argument.Table = executor.createStepTable(arg.table)
		}
		arguments = append(arguments, argument)
	}

	if inlineTable.isInitialized() {
		inlineTableArg := executor.createStepTable(inlineTable)
		arguments = append(arguments, &Argument{Type: proto.String("table"), Table: inlineTableArg})
	}
	return arguments
}

func (executor *specExecutor) getCurrentDataTableValueFor(columnName string) string {
	return executor.specification.dataTable.get(columnName)[executor.dataTableIndex]
}

func (executor *specExecutor) createStepTable(table table) *ProtoTable {
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
