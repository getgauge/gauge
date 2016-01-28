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
	"errors"
	"strconv"
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

func (v *validator) validate() validationErrors {
	validationStatus := make(validationErrors)
	specValidator := &specValidator{runner: v.runner, conceptsDictionary: v.conceptsDictionary, stepValidationCache: make(map[string]*stepValidationError)}
	for _, spec := range v.specsToExecute {
		specValidator.specification = spec
		validationErrors := specValidator.validate()
		if len(validationErrors) != 0 {
			validationStatus[spec] = validationErrors
		}
	}
	if len(validationStatus) > 0 {
		return validationStatus
	}
	return nil
}

func (v *specValidator) validate() []*stepValidationError {
	v.specification.Traverse(v)
	return v.stepValidationErrors
}

func (v *specValidator) Step(step *gauge.Step) {
	if step.IsConcept {
		for _, conceptStep := range step.ConceptSteps {
			v.Step(conceptStep)
		}
	} else {
		value, ok := v.stepValidationCache[step.Value]
		if !ok {
			err := v.validateStep(step)
			if err != nil {
				v.stepValidationErrors = append(v.stepValidationErrors, err)
			}
			v.stepValidationCache[step.Value] = err
		} else if value != nil {
			v.stepValidationErrors = append(v.stepValidationErrors,
				&stepValidationError{step: step, fileName: v.specification.FileName, errorType: value.errorType, message: value.message})
		}
	}
}

var invalidResponse gauge_messages.StepValidateResponse_ErrorType = -1

func (v *specValidator) validateStep(step *gauge.Step) *stepValidationError {
	message := &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateRequest.Enum(),
		StepValidateRequest: &gauge_messages.StepValidateRequest{StepText: proto.String(step.Value), NumberOfParameters: proto.Int(len(step.Args))}}
	response, err := conn.GetResponseForMessageWithTimeout(message, v.runner.Connection, config.RunnerRequestTimeout())
	if err != nil {
		return &stepValidationError{step: step, message: err.Error(), fileName: v.specification.FileName}
	}
	if response.GetMessageType() == gauge_messages.Message_StepValidateResponse {
		validateResponse := response.GetStepValidateResponse()
		if !validateResponse.GetIsValid() {
			message := getMessage(validateResponse.ErrorType.String())
			return &stepValidationError{step: step, fileName: v.specification.FileName, errorType: validateResponse.ErrorType, message: message}
		}
		return nil
	}
	return &stepValidationError{step: step, fileName: v.specification.FileName, errorType: &invalidResponse, message: "Invalid response from runner for Validation request"}
}

func getMessage(message string) string {
	lower := strings.ToLower(strings.Replace(message, "_", " ", -1))
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func (v *specValidator) ContextStep(step *gauge.Step) {
	v.Step(step)
}

func (v *specValidator) TearDown(step *gauge.TearDown) {
}

func (v *specValidator) SpecHeading(heading *gauge.Heading) {
	v.stepValidationErrors = make([]*stepValidationError, 0)
}

func (v *specValidator) SpecTags(tags *gauge.Tags) {
}

func (v *specValidator) ScenarioTags(tags *gauge.Tags) {

}

func (v *specValidator) DataTable(dataTable *gauge.Table) {

}

func (v *specValidator) Scenario(scenario *gauge.Scenario) {

}

func (v *specValidator) ScenarioHeading(heading *gauge.Heading) {
}

func (v *specValidator) Comment(comment *gauge.Comment) {
}

func (v *specValidator) ExternalDataTable(dataTable *gauge.DataTable) {

}

func validateTableRowsRange(start string, end string, rowCount int) (int, int, error) {
	message := "Table rows range validation failed."
	startRow, err := strconv.Atoi(start)
	if err != nil {
		return 0, 0, errors.New(message)
	}
	endRow, err := strconv.Atoi(end)
	if err != nil {
		return 0, 0, errors.New(message)
	}
	if startRow > endRow || endRow > rowCount || startRow < 1 || endRow < 1 {
		return 0, 0, errors.New(message)
	}
	return startRow - 1, endRow - 1, nil
}
