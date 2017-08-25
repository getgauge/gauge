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
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/runner"
)

var TableRows = ""
var HideSuggestion bool

type validator struct {
	manifest           *manifest.Manifest
	specsToExecute     []*gauge.Specification
	runner             runner.Runner
	conceptsDictionary *gauge.ConceptDictionary
}

type specValidator struct {
	specification       *gauge.Specification
	runner              runner.Runner
	conceptsDictionary  *gauge.ConceptDictionary
	validationErrors    []error
	stepValidationCache map[string]error
}

type StepValidationError struct {
	step       *gauge.Step
	message    string
	fileName   string
	errorType  *gm.StepValidateResponse_ErrorType
	suggestion string
}

type SpecValidationError struct {
	message  string
	fileName string
}

func (s StepValidationError) Error() string {
	return fmt.Sprintf("%s:%d %s => '%s'", s.fileName, s.step.LineNo, s.message, s.step.GetLineText())
}

func (s StepValidationError) Suggestion() string {
	return s.suggestion
}

func (s SpecValidationError) Error() string {
	return fmt.Sprintf("%s %s", s.fileName, s.message)
}

func NewSpecValidationError(m string, f string) SpecValidationError {
	return SpecValidationError{message: m, fileName: f}
}

func NewStepValidationError(s *gauge.Step, m string, f string, e *gm.StepValidateResponse_ErrorType) StepValidationError {
	return StepValidationError{step: s, message: m, fileName: f, errorType: e}
}

func Validate(args []string) {
	if len(args) == 0 {
		args = append(args, common.SpecsDirectoryName)
	}
	res := ValidateSpecs(args, false)
	if len(res.Errs) > 0 {
		os.Exit(1)
	}
	if res.SpecCollection.Size() < 1 {
		logger.Infof("No specifications found in %s.", strings.Join(args, ", "))
		res.Runner.Kill()
		if res.ParseOk {
			os.Exit(0)
		}
		os.Exit(1)
	}
	res.Runner.Kill()
	if res.ErrMap.HasErrors() {
		os.Exit(1)
	}
	logger.Infof("No error found.")
}

//TODO : duplicate in execute.go. Need to fix runner init.
func startAPI(debug bool) runner.Runner {
	sc := api.StartAPI(debug)
	select {
	case runner := <-sc.RunnerChan:
		return runner
	case err := <-sc.ErrorChan:
		logger.Fatalf("Failed to start gauge API: %s", err.Error())
	}
	return nil
}

type ValidationResult struct {
	SpecCollection *gauge.SpecCollection
	ErrMap         *gauge.BuildErrors
	Runner         runner.Runner
	Errs           []error
	ParseOk        bool
}

func NewValidationResult(s *gauge.SpecCollection, errMap *gauge.BuildErrors, r runner.Runner, parseOk bool, e ...error) *ValidationResult {
	return &ValidationResult{SpecCollection: s, ErrMap: errMap, Runner: r, ParseOk: parseOk, Errs: e}
}

func ValidateSpecs(args []string, debug bool) *ValidationResult {
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		logger.Errorf(err.Error())
		return NewValidationResult(nil, nil, nil, false, err)
	}
	conceptDict, res := parser.ParseConcepts()
	if len(res.CriticalErrors) > 0 {
		var errs []error
		for _, err := range res.CriticalErrors {
			errs = append(errs, err)
		}
		return NewValidationResult(nil, nil, nil, false, errs...)
	}
	errMap := gauge.NewBuildErrors()
	s, specsFailed := parser.ParseSpecs(args, conceptDict, errMap)
	r := startAPI(debug)
	vErrs := newValidator(manifest, s, r, conceptDict).validate()
	errMap = getErrMap(errMap, vErrs)
	s = parser.GetSpecsForDataTableRows(s, errMap)
	printValidationFailures(vErrs)
	if !res.Ok {
		r.Kill()
		return NewValidationResult(nil, nil, nil, false, errors.New("Parsing failed."))
	}
	if specsFailed {
		return NewValidationResult(gauge.NewSpecCollection(s, false), errMap, r, false)
	}
	showSuggestion(vErrs)
	return NewValidationResult(gauge.NewSpecCollection(s, false), errMap, r, true)
}

func getErrMap(errMap *gauge.BuildErrors, validationErrors validationErrors) *gauge.BuildErrors {
	for spec, valErrors := range validationErrors {
		for _, err := range valErrors {
			switch err.(type) {
			case StepValidationError:
				errMap.StepErrs[err.(StepValidationError).step] = err.(StepValidationError)
			case SpecValidationError:
				errMap.SpecErrs[spec] = append(errMap.SpecErrs[spec], err.(SpecValidationError))
			}
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
	return errMap
}

func fillScenarioErrors(scenario *gauge.Scenario, errMap *gauge.BuildErrors, steps []*gauge.Step) {
	for _, step := range steps {
		if step.IsConcept {
			fillScenarioErrors(scenario, errMap, step.ConceptSteps)
		}
		if err, ok := errMap.StepErrs[step]; ok {
			errMap.ScenarioErrs[scenario] = append(errMap.ScenarioErrs[scenario], err)
		}
	}
}

func fillSpecErrors(spec *gauge.Specification, errMap *gauge.BuildErrors, steps []*gauge.Step) {
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

type validationErrors map[*gauge.Specification][]error

func newValidator(m *manifest.Manifest, s []*gauge.Specification, r runner.Runner, c *gauge.ConceptDictionary) *validator {
	return &validator{manifest: m, specsToExecute: s, runner: r, conceptsDictionary: c}
}

func (v *validator) validate() validationErrors {
	validationStatus := make(validationErrors)
	specValidator := &specValidator{runner: v.runner, conceptsDictionary: v.conceptsDictionary, stepValidationCache: make(map[string]error)}
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

func (v *specValidator) validate() []error {
	queue := &gauge.ItemQueue{Items: v.specification.AllItems()}
	v.specification.Traverse(v, queue)
	return v.validationErrors
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
			v.validationErrors = append(v.validationErrors, err)
		}
		v.stepValidationCache[s.Value] = err
		return
	}
	if val != nil {
		valErr := val.(StepValidationError)
		if s.Parent == nil {
			v.validationErrors = append(v.validationErrors,
				NewStepValidationError(s, valErr.message, v.specification.FileName, valErr.errorType))
		} else {
			cpt := v.conceptsDictionary.Search(s.Parent.Value)
			v.validationErrors = append(v.validationErrors,
				NewStepValidationError(s, valErr.message, cpt.FileName, valErr.errorType))
		}
	}
}

var invalidResponse gm.StepValidateResponse_ErrorType = -1

var getResponseFromRunner = func(m *gm.Message, v *specValidator) (*gm.Message, error) {
	return conn.GetResponseForMessageWithTimeout(m, v.runner.Connection(), config.RunnerRequestTimeout())
}

func (v *specValidator) validateStep(s *gauge.Step) error {
	stepValue, _ := parser.ExtractStepValueAndParams(s.LineText, s.HasInlineTable)
	protoStepValue := gauge.ConvertToProtoStepValue(stepValue)

	m := &gm.Message{MessageType: gm.Message_StepValidateRequest,
		StepValidateRequest: &gm.StepValidateRequest{StepText: s.Value, NumberOfParameters: int32(len(s.Args)), StepValue: protoStepValue}}

	r, err := getResponseFromRunner(m, v)
	if err != nil {
		return NewStepValidationError(s, err.Error(), v.specification.FileName, nil)
	}
	if r.GetMessageType() == gm.Message_StepValidateResponse {
		res := r.GetStepValidateResponse()
		if !res.GetIsValid() {
			msg := getMessage(res.GetErrorType().String())
			suggestion := res.GetSuggestion()
			if s.Parent == nil {
				vErr := NewStepValidationError(s, msg, v.specification.FileName, &res.ErrorType)
				vErr.suggestion = suggestion
				return vErr
			}
			cpt := v.conceptsDictionary.Search(s.Parent.Value)
			vErr := NewStepValidationError(s, msg, cpt.FileName, &res.ErrorType)
			vErr.suggestion = suggestion
			return vErr

		}
		return nil
	}
	return NewStepValidationError(s, "Invalid response from runner for Validation request", v.specification.FileName, &invalidResponse)
}

func getMessage(message string) string {
	lower := strings.ToLower(strings.Replace(message, "_", " ", -1))
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func (v *specValidator) TearDown(step *gauge.TearDown) {
}

func (v *specValidator) Heading(heading *gauge.Heading) {
}

func (v *specValidator) Tags(tags *gauge.Tags) {
}

func (v *specValidator) Table(dataTable *gauge.Table) {

}

func (v *specValidator) Scenario(scenario *gauge.Scenario) {

}

func (v *specValidator) Comment(comment *gauge.Comment) {
}

func (v *specValidator) DataTable(dataTable *gauge.DataTable) {

}

func (v *specValidator) Specification(specification *gauge.Specification) {
	v.validationErrors = make([]error, 0)
	err := validateDataTableRange(specification.DataTable.Table.GetRowCount())
	if err != nil {
		v.validationErrors = append(v.validationErrors, NewSpecValidationError(err.Error(), specification.FileName))
	}
}

func validateDataTableRange(rowCount int) error {
	if TableRows == "" {
		return nil
	}
	if strings.Contains(TableRows, "-") {
		indexes := strings.Split(TableRows, "-")
		if len(indexes) > 2 {
			return fmt.Errorf("Table rows range '%s' is invalid => Table rows range should be of format rowNumber-rowNumber", TableRows)
		}
		if err := validateTableRow(indexes[0], rowCount); err != nil {
			return err
		}
		if err := validateTableRow(indexes[1], rowCount); err != nil {
			return err
		}
	} else {
		indexes := strings.Split(TableRows, ",")
		for _, i := range indexes {
			if err := validateTableRow(i, rowCount); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateTableRow(rowNumber string, rowCount int) error {
	if rowNumber = strings.TrimSpace(rowNumber); rowNumber == "" {
		return fmt.Errorf("Table rows range validation failed => Row number cannot be empty")
	}
	row, err := strconv.Atoi(rowNumber)
	if err != nil {
		return fmt.Errorf("Table rows range validation failed => Failed to parse '%s' to row number", rowNumber)
	}
	if row < 1 || row > rowCount {
		return fmt.Errorf("Table rows range validation failed => Table row number '%d' is out of range", row)
	}
	return nil
}
