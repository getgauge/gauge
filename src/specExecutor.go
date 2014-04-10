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

func (executor *specExecutor) execute() error {
	//todo should return a proper result
	if err := executor.startSpecExecution(); err != nil {
		return err
	}

	dataTableRowCount := executor.specification.dataTable.getRowCount()
	if dataTableRowCount == 0 {
		return executor.executeScenarios();
	} else {
		for executor.dataTableIndex = 0; executor.dataTableIndex < dataTableRowCount; executor.dataTableIndex++ {
			executor.executeScenarios();
		}
	}

	return nil
}

func (executor *specExecutor) executeScenarios() error {
	for _, scenario := range executor.specification.scenarios {
		if err := executor.executeContext(); err != nil {
			return err
		}

		if err := executor.executeScenario(scenario); err != nil {
			return err
		}
	}
	return nil
}

func (executor *specExecutor) executeContext() error {
	return executor.executeSteps(executor.specification.contexts)
}

func (executor *specExecutor) executeScenario(scenario *scenario) error {
	return executor.executeSteps(scenario.steps)
}

func (executor *specExecutor) executeSteps(steps []*step) error {

	for _, step := range steps {
		if validationErr := executor.validateStep(step); validationErr != nil {
			return validationErr
		}

		shouldContinue, err := executor.executeStep(step)

		if !shouldContinue {
			return err
		}

	}
	return nil
}

func (e *specExecutor) startSpecExecution() error {
	message := &Message{MessageType: Message_ExecutionStarting.Enum(),
		ExecutionStartingRequest: &ExecutionStartingRequest{SpecFile: proto.String(e.specification.fileName)}}

	if _, err := getResponse(e.connection, message); err != nil {
		return err
	}

	return nil
}

func (executor *specExecutor) executeStep(step *step) (bool, error) {

	stepRequest := executor.createStepRequest(step)
	message := &Message{MessageType: Message_ExecuteStep.Enum(),
		ExecuteStepRequest: stepRequest}

	response, err := getResponse(executor.connection, message)
	if err != nil {
		return false, err
	}

	if response.GetMessageType() == Message_ExecuteStepResponse {
		stepResponse := response.GetExecuteStepResponse()
		if stepResponse.GetPassed() != true {
			fmt.Printf("\x1b[31;1m%s\n\x1b[0m", stepResponse.GetErrorMessage())
			fmt.Printf("\x1b[31;1m%s\n\x1b[0m", stepResponse.GetStackTrace())
			return false, nil
		} else {
			fmt.Printf("=> \x1b[32;1m%s\n\x1b[0m", step.value)
		}
	}

	return true, nil
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
