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

package execution

import (
	"strings"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/runner"
	"github.com/golang/protobuf/proto"
)

type validator struct {
	manifest           *manifest.Manifest
	specsToExecute     []*gauge.Specification
	runner             *runner.TestRunner
	conceptsDictionary *gauge.ConceptDictionary
}

type specValidator struct {
	specification        *gauge.Specification
	runner               *runner.TestRunner
	conceptsDictionary   *gauge.ConceptDictionary
	stepValidationErrors []*stepValidationError
	stepValidationCache  map[string]*stepValidationError
}

type stepValidationError struct {
	step      *gauge.Step
	message   string
	fileName  string
	errorType *gauge_messages.StepValidateResponse_ErrorType
}

func (e *stepValidationError) Error() string {
	return e.message
}

type validationErrors map[*gauge.Specification][]*stepValidationError

func newValidator(manifest *manifest.Manifest, specsToExecute []*gauge.Specification, runner *runner.TestRunner, conceptsDictionary *gauge.ConceptDictionary) *validator {
	return &validator{manifest: manifest, specsToExecute: specsToExecute, runner: runner, conceptsDictionary: conceptsDictionary}
}

func (self *validator) validate() validationErrors {
	validationStatus := make(validationErrors)
	specValidator := &specValidator{runner: self.runner, conceptsDictionary: self.conceptsDictionary, stepValidationCache: make(map[string]*stepValidationError)}
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
	self.specification.Traverse(self)
	return self.stepValidationErrors
}

func (self *specValidator) Step(step *gauge.Step) {
	if step.IsConcept {
		for _, conceptStep := range step.ConceptSteps {
			self.Step(conceptStep)
		}
	} else {
		value, ok := self.stepValidationCache[step.Value]
		if !ok {
			err := self.validateStep(step)
			if err != nil {
				self.stepValidationErrors = append(self.stepValidationErrors, err)
			}
			self.stepValidationCache[step.Value] = err
		} else if value != nil {
			self.stepValidationErrors = append(self.stepValidationErrors,
				&stepValidationError{step: step, fileName: self.specification.FileName, errorType: value.errorType, message: value.message})
		}
	}
}

var invalidResponse gauge_messages.StepValidateResponse_ErrorType = -1

func (self *specValidator) validateStep(step *gauge.Step) *stepValidationError {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateRequest.Enum(),
		StepValidateRequest: &gauge_messages.StepValidateRequest{StepText: proto.String(step.Value), NumberOfParameters: proto.Int(len(step.Args))}}
	response, err := conn.GetResponseForMessageWithTimeout(message, self.runner.Connection, config.RunnerRequestTimeout())
	if err != nil {
		return &stepValidationError{step: step, message: err.Error(), fileName: self.specification.FileName}
	}
	if response.GetMessageType() == gauge_messages.Message_StepValidateResponse {
		validateResponse := response.GetStepValidateResponse()
		if !validateResponse.GetIsValid() {
			message := getMessage(validateResponse.ErrorType.String())
			return &stepValidationError{step: step, fileName: self.specification.FileName, errorType: validateResponse.ErrorType, message: message}
		}
		return nil
	} else {
		return &stepValidationError{step: step, fileName: self.specification.FileName, errorType: &invalidResponse, message: "Invalid response from runner for Validation request"}
	}
}

func getMessage(message string) string {
	lower := strings.ToLower(strings.Replace(message, "_", " ", -1))
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func (self *specValidator) ContextStep(step *gauge.Step) {
	self.Step(step)
}

func (self *specValidator) TearDown(step *gauge.TearDown) {
}

func (self *specValidator) SpecHeading(heading *gauge.Heading) {
	self.stepValidationErrors = make([]*stepValidationError, 0)
}

func (self *specValidator) SpecTags(tags *gauge.Tags) {
}

func (self *specValidator) ScenarioTags(tags *gauge.Tags) {

}

func (self *specValidator) DataTable(dataTable *gauge.Table) {

}

func (self *specValidator) Scenario(scenario *gauge.Scenario) {

}

func (self *specValidator) ScenarioHeading(heading *gauge.Heading) {
}

func (self *specValidator) Comment(comment *gauge.Comment) {
}

func (self *specValidator) ExternalDataTable(dataTable *gauge.DataTable) {

}
