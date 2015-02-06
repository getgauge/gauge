// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type validator struct {
	manifest           *manifest
	specsToExecute     []*specification
	runner             *testRunner
	conceptsDictionary *conceptDictionary
}

type specValidator struct {
	specification        *specification
	runner               *testRunner
	conceptsDictionary   *conceptDictionary
	stepValidationErrors []*stepValidationError
	stepValidationCache  map[string]bool
}

type stepValidationError struct {
	step     *step
	message  string
	fileName string
}

func (e *stepValidationError) Error() string {
	return e.message
}

type executionValidationErrors map[*specification][]*stepValidationError

func newValidator(manifest *manifest, specsToExecute []*specification, runner *testRunner, conceptsDictionary *conceptDictionary) *validator {
	return &validator{manifest: manifest, specsToExecute: specsToExecute, runner: runner, conceptsDictionary: conceptsDictionary}
}

func (self *validator) validate() executionValidationErrors {
	validationStatus := make(executionValidationErrors)
	specValidator := &specValidator{runner: self.runner, conceptsDictionary: self.conceptsDictionary, stepValidationCache: make(map[string]bool)}
	for _, spec := range self.specsToExecute {
		specValidator.specification = spec
		validationErrors := specValidator.validate()
		if len(validationErrors) != 0 {
			validationStatus[spec] = validationErrors
		}
	}
	if len(validationStatus) > 0 {
		return validationStatus
	} else {
		return nil
	}
}

func (self *specValidator) validate() []*stepValidationError {
	self.specification.traverse(self)
	return self.stepValidationErrors
}

func (self *specValidator) step(step *step) {
	if step.isConcept {
		for _, conceptStep := range step.conceptSteps {
			if _, ok := self.stepValidationCache[conceptStep.value]; !ok {
				self.stepValidationCache[conceptStep.value] = true
				self.step(conceptStep)
			}
		}
	} else if _, ok := self.stepValidationCache[step.value]; !ok {
		self.stepValidationCache[step.value] = true
		self.validateStep(step)
	}
}

func (self *specValidator) validateStep(step *step) {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateRequest.Enum(),
		StepValidateRequest: &gauge_messages.StepValidateRequest{StepText: proto.String(step.value), NumberOfParameters: proto.Int(len(step.args))}}
	response, err := getResponseForMessageWithTimeout(message, self.runner.connection, config.RunnerRequestTimeout())
	if err != nil {
		self.stepValidationErrors = append(self.stepValidationErrors, &stepValidationError{step: step, message: err.Error(), fileName: self.specification.fileName})
	}
	if response.GetMessageType() == gauge_messages.Message_StepValidateResponse {
		validateResponse := response.GetStepValidateResponse()
		if !validateResponse.GetIsValid() {
			self.stepValidationErrors = append(self.stepValidationErrors, &stepValidationError{step: step, message: validateResponse.GetErrorMessage(), fileName: self.specification.fileName})
		}
	} else {
		self.stepValidationErrors = append(self.stepValidationErrors, &stepValidationError{step: step, message: "Invalid response from runner for Validation request", fileName: self.specification.fileName})
	}
}

func (self *specValidator) contextStep(step *step) {
	self.step(step)
}

func (self *specValidator) specHeading(heading *heading) {
	self.stepValidationErrors = make([]*stepValidationError, 0)
}

func (self *specValidator) specTags(tags *tags) {

}

func (self *specValidator) scenarioTags(tags *tags) {

}

func (self *specValidator) dataTable(dataTable *table) {

}

func (self *specValidator) scenario(scenario *scenario) {

}

func (self *specValidator) scenarioHeading(heading *heading) {
}

func (self *specValidator) comment(comment *comment) {

}
