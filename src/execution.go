// This file is part of twist
package main

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/twist2/common"
	"io/ioutil"
	"net"
	"os"
	"fmt"
)

type execution struct {
	tokens     []*token
	manifest   *manifest
	connection net.Conn
}

func newExecution(manifest *manifest, tokens []*token, conn net.Conn) *execution {
	e := execution{manifest: manifest, tokens: tokens, connection: conn}
	return &e
}

func (e *execution) startScenarioExecution() error {
	message := &Message{MessageType: Message_ExecutionStarting.Enum(),
		ExecutionStartingRequest: &ExecutionStartingRequest{ScenarioFile: proto.String("sample.scn")}}

	_, err := getResponse(e.connection, message)
	if err != nil {
		return err
	}

	return nil
}

func (e *execution) startStepExecution(token *token) (bool, error) {
	message := &Message{MessageType: Message_ExecuteStep.Enum(),
		ExecuteStepRequest: &ExecuteStepRequest{StepText: proto.String(token.value), Args: token.args}}

	common.PrintExecutionStart(token.line)

	response, err := getResponse(e.connection, message)
	if err != nil {
		return false, err
	}

	if response.GetMessageType() == Message_ExecuteStepResponse {
		stepResponse := response.GetExecuteStepResponse()
		if stepResponse.GetPassed() != true {
			ioutil.WriteFile("/tmp/twist-screenshot.png", stepResponse.GetScreenShot(), 0644)
			common.PrintExecutionFailed(token.line, stepResponse.GetErrorMessage(), stepResponse.GetStackTrace())
			return false, nil
		} else {
			common.PrintExecutionPassed(token.line)
		}
	}

	return true, nil
}

func (e *execution) validateStep(token *token) (bool, error) {
	message := &Message{MessageType: Message_StepValidateRequest.Enum(),
		StepValidateRequest: &StepValidateRequest{StepText: proto.String(token.value)}}
	response, err := getResponse(e.connection, message)
	if err != nil {
		return false, err
	}

	if response.GetMessageType() == Message_StepValidateResponse {
		validateResponse := response.GetStepValidateResponse()
		return validateResponse.GetIsValid(), nil
	} else {
		panic("Expected a validate step response")
	}
}

func (e *execution) stopScenarioExecution() error {
	message := &Message{MessageType: Message_ExecutionEnding.Enum(),
		ExecutionEndingRequest: &ExecutionEndingRequest{}}

	_, err := getResponse(e.connection, message)
	if err != nil {
		return err
	}

	return nil
}

func (e *execution) start() error {
	for _, token := range e.tokens {
		var err error
		quit := false

		switch token.kind {
		case typeScenario:
			err = e.startScenarioExecution()
			break
		case typeWorkflowStep:
			valid, err := e.validateStep(token)
			if !valid {
				common.PrintError(fmt.Sprintf("Error: Unimplemented step: %s", token.line))
				quit = true
				err = err
			} else {
				passed, err := e.startStepExecution(token)
				quit = !passed
				err = err
			}
			break
		}

		if err != nil {
			common.PrintError(fmt.Sprintf("Failed to execute step. %s", err.Error()))
			os.Exit(1)
		}

		if quit {
			break
		}
	}

	return e.stopScenarioExecution()
}
