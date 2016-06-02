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

package validation

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/runner"
	"github.com/golang/protobuf/proto"
)

type ValidationErrMaps struct {
	SpecErrs     map[*gauge.Specification][]*StepValidationError
	ScenarioErrs map[*gauge.Scenario][]*StepValidationError
	StepErrs     map[*gauge.Step]*StepValidationError
}

type validator struct {
	manifest           *manifest.Manifest
	specsToExecute     []*gauge.Specification
	runner             runner.Runner
	conceptsDictionary *gauge.ConceptDictionary
}

type specValidator struct {
	specification        *gauge.Specification
	runner               runner.Runner
	conceptsDictionary   *gauge.ConceptDictionary
	stepValidationErrors []*StepValidationError
	stepValidationCache  map[string]*StepValidationError
}

type StepValidationError struct {
	step      *gauge.Step
	message   string
	fileName  string
	errorType *gauge_messages.StepValidateResponse_ErrorType
}

func (s *StepValidationError) Error() string {
	return fmt.Sprintf("%s:%d: %s => '%s'", s.fileName, s.step.LineNo, s.message, s.step.LineText)
}

func Validate(args []string) {
	if len(args) == 0 {
		args = append(args, common.SpecsDirectoryName)
	}
	_, errMap, runner := ValidateSpecs(args)
	runner.Kill()
	if len(errMap.StepErrs) > 0 {
		os.Exit(1)
	}
	logger.Info("No error found.")
}

//TODO : duplicate in execute.go. Need to fix runner init.
func startAPI() runner.Runner {
	sc := api.StartAPI()
	select {
	case runner := <-sc.RunnerChan:
		return runner
	case err := <-sc.ErrorChan:
		logger.Fatalf("Failed to start gauge API: %s", err.Error())
	}
	return nil
}

func ValidateSpecs(args []string) (*gauge.SpecCollection, *ValidationErrMaps, runner.Runner) {
	s, c, f := parseSpecs(args)
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	r := startAPI()
	v := newValidator(manifest, s, r, c)
	vErrs := v.validate()
	errMap := &ValidationErrMaps{
		SpecErrs:     make(map[*gauge.Specification][]*StepValidationError),
		ScenarioErrs: make(map[*gauge.Scenario][]*StepValidationError),
		StepErrs:     make(map[*gauge.Step]*StepValidationError),
	}
	if len(vErrs) > 0 {
		printValidationFailures(vErrs)
		fillErrors(errMap, vErrs)
	}
	if f {
		r.Kill()
		os.Exit(1)
	}
	if len(s) == 0 {
		logger.Info("No specifications found in %s.", strings.Join(args, ", "))
		r.Kill()
		os.Exit(0)
	}
	return gauge.NewSpecCollection(s), errMap, r
}

func parseSpecs(args []string) ([]*gauge.Specification, *gauge.ConceptDictionary, bool) {
	conceptsDictionary, conceptParseResult := parser.CreateConceptsDictionary()
	conceptFailed := parser.HandleParseResult(conceptParseResult)
	specsToExecute, _, specFailed := filter.GetSpecsToExecute(conceptsDictionary, args)
	return specsToExecute, conceptsDictionary, conceptFailed || specFailed
}

func fillErrors(errMap *ValidationErrMaps, validationErrors validationErrors) {
	for spec, errors := range validationErrors {
		for _, err := range errors {
			errMap.StepErrs[err.step] = err
		}
		skippedScnInSpec := 0
		for _, scenario := range spec.Scenarios {
			fillScenarioErrors(scenario, errMap, scenario.Steps)
			if _, ok := errMap.ScenarioErrs[scenario]; ok {
				skippedScnInSpec++
			}
		}
		if len(spec.Scenarios) > 0 && skippedScnInSpec == len(spec.Scenarios) {
			errMap.SpecErrs[spec] = append(errMap.SpecErrs[spec], errMap.ScenarioErrs[spec.Scenarios[0]]...)
		}
		fillSpecErrors(spec, errMap, append(spec.Contexts, spec.TearDownSteps...))
	}
}

func fillScenarioErrors(scenario *gauge.Scenario, errMap *ValidationErrMaps, steps []*gauge.Step) {
	for _, step := range steps {
		if step.IsConcept {
			fillScenarioErrors(scenario, errMap, step.ConceptSteps)
		}
		if err, ok := errMap.StepErrs[step]; ok {
			errMap.ScenarioErrs[scenario] = append(errMap.ScenarioErrs[scenario], err)
		}
	}
}

func fillSpecErrors(spec *gauge.Specification, errMap *ValidationErrMaps, steps []*gauge.Step) {
	for _, context := range steps {
		if context.IsConcept {
			fillSpecErrors(spec, errMap, context.ConceptSteps)
		}
		if err, ok := errMap.StepErrs[context]; ok {
			errMap.SpecErrs[spec] = append(errMap.SpecErrs[spec], err)
			for _, scenario := range spec.Scenarios {
				if _, ok := errMap.ScenarioErrs[scenario]; !ok {
					errMap.ScenarioErrs[scenario] = append(errMap.ScenarioErrs[scenario], err)
				}
			}
		}
	}
}

func printValidationFailures(validationErrors validationErrors) {
	for _, errs := range validationErrors {
		for _, e := range errs {
			logger.Errorf("[ValidationError] %s", e.Error())
		}
	}
}

func NewValidationError(s *gauge.Step, m string, f string, e *gauge_messages.StepValidateResponse_ErrorType) *StepValidationError {
	return &StepValidationError{step: s, message: m, fileName: f, errorType: e}
}

type validationErrors map[*gauge.Specification][]*StepValidationError

func newValidator(m *manifest.Manifest, s []*gauge.Specification, r runner.Runner, c *gauge.ConceptDictionary) *validator {
	return &validator{manifest: m, specsToExecute: s, runner: r, conceptsDictionary: c}
}

func (v *validator) validate() validationErrors {
	validationStatus := make(validationErrors)
	specValidator := &specValidator{runner: v.runner, conceptsDictionary: v.conceptsDictionary, stepValidationCache: make(map[string]*StepValidationError)}
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

func (v *specValidator) validate() []*StepValidationError {
	v.specification.Traverse(v)
	return v.stepValidationErrors
}

func (v *specValidator) Step(s *gauge.Step) {
	if s.IsConcept {
		for _, c := range s.ConceptSteps {
			v.Step(c)
		}
		return
	}
	val, ok := v.stepValidationCache[s.Value]
	if !ok {
		err := v.validateStep(s)
		if err != nil {
			v.stepValidationErrors = append(v.stepValidationErrors, err)
		}
		v.stepValidationCache[s.Value] = err
		return
	}
	if val != nil {
		v.stepValidationErrors = append(v.stepValidationErrors,
			NewValidationError(s, val.message, v.specification.FileName, val.errorType))
	}
}

var invalidResponse gauge_messages.StepValidateResponse_ErrorType = -1

func (v *specValidator) validateStep(s *gauge.Step) *StepValidationError {
	m := &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateRequest.Enum(),
		StepValidateRequest: &gauge_messages.StepValidateRequest{StepText: proto.String(s.Value), NumberOfParameters: proto.Int(len(s.Args))}}
	r, err := conn.GetResponseForMessageWithTimeout(m, v.runner.Connection(), config.RunnerRequestTimeout())
	if err != nil {
		return NewValidationError(s, err.Error(), v.specification.FileName, nil)
	}
	if r.GetMessageType() == gauge_messages.Message_StepValidateResponse {
		res := r.GetStepValidateResponse()
		if !res.GetIsValid() {
			msg := getMessage(res.ErrorType.String())
			return NewValidationError(s, msg, v.specification.FileName, res.ErrorType)
		}
		return nil
	}
	return NewValidationError(s, "Invalid response from runner for Validation request", v.specification.FileName, &invalidResponse)
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
	v.stepValidationErrors = make([]*StepValidationError, 0)
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

func ValidateTableRowsRange(start string, end string, rowCount int) (int, int, error) {
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
