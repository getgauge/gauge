/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

/*
Validation invokes language runner for every step in serial fashion with the StepValidateRequest and runner gets back with the StepValidateResponse.

Step Level validation
	1. Duplicate step implementation
	2. Step implementation not found : Prints a step implementation stub for every unimplemented step

If there is a validation error it skips that scenario and executes other scenarios in the spec.
*/
package validation

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
)

// TableRows is used to check for table rows range validation.
var TableRows = ""

// HideSuggestion is used decide whether suggestion should be given for the unimplemented step or not based on the flag : --hide-suggestion.
var HideSuggestion bool

type validator struct {
	specsToExecute     []*gauge.Specification
	runner             runner.Runner
	conceptsDictionary *gauge.ConceptDictionary
}

type SpecValidator struct {
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

func (s StepValidationError) Message() string {
	return s.message
}

func (s StepValidationError) Step() *gauge.Step {
	return s.step
}

func (s StepValidationError) FileName() string {
	return s.fileName
}

func (s StepValidationError) ErrorType() gm.StepValidateResponse_ErrorType {
	return *s.errorType
}

// Error prints a step validation error with filename, line number, error message, step text and suggestion in case of step implementation not found.
func (s StepValidationError) Error() string {
	return fmt.Sprintf("%s:%d %s => '%s'", s.fileName, s.step.LineNo, s.message, s.step.GetLineText())
}

func (s StepValidationError) Suggestion() string {
	return s.suggestion
}

// Error prints a spec validation error with filename and error message.
func (s SpecValidationError) Error() string {
	return fmt.Sprintf("%s %s", s.fileName, s.message)
}

// NewSpecValidationError generates new spec validation error with error message and filename.
func NewSpecValidationError(m string, f string) SpecValidationError {
	return SpecValidationError{message: m, fileName: f}
}

// NewStepValidationError generates new step validation error with error message, filename and error type.
func NewStepValidationError(s *gauge.Step, m string, f string, e *gm.StepValidateResponse_ErrorType, suggestion string) StepValidationError {
	return StepValidationError{step: s, message: m, fileName: f, errorType: e, suggestion: suggestion}
}

// Validate validates specs and if it has any errors, it exits.
func Validate(args []string) {
	if len(args) == 0 {
		args = append(args, util.GetSpecDirs()...)
	}
	res := ValidateSpecs(args, false)
	if len(res.Errs) > 0 {
		os.Exit(1)
	}
	if res.SpecCollection.Size() < 1 {
		logger.Infof(true, "No specifications found in %s.", strings.Join(args, ", "))
		err := res.Runner.Kill()
		if err != nil {
			logger.Errorf(false, "unable to kill runner: %s", err.Error())
		}
		if res.ParseOk {
			os.Exit(0)
		}
		os.Exit(1)
	}
	err := res.Runner.Kill()
	if err != nil {
		logger.Errorf(false, "unable to kill runner: %s", err.Error())
	}

	if res.ErrMap.HasErrors() {
		os.Exit(1)
	}
	logger.Infof(true, "No errors found.")
}

//TODO : duplicate in execute.go. Need to fix runner init.
func startAPI(debug bool) runner.Runner {
	sc := api.StartAPI(debug)
	select {
	case runner := <-sc.RunnerChan:
		return runner
	case err := <-sc.ErrorChan:
		logger.Fatalf(true, "Failed to start gauge API: %s", err.Error())
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

// NewValidationResult creates a new Validation result
func NewValidationResult(s *gauge.SpecCollection, errMap *gauge.BuildErrors, r runner.Runner, parseOk bool, e ...error) *ValidationResult {
	return &ValidationResult{SpecCollection: s, ErrMap: errMap, Runner: r, ParseOk: parseOk, Errs: e}
}

// ValidateSpecs parses the specs, creates a new validator and call the runner to get the validation result.
func ValidateSpecs(args []string, debug bool) *ValidationResult {
	logger.Debug(true, "Parsing started.")
	conceptDict, res, err := parser.ParseConcepts()
	if err != nil {
		logger.Fatalf(true, "Unable to parse : %s", err.Error())
	}
	errMap := gauge.NewBuildErrors()
	s, specsFailed := parser.ParseSpecs(args, conceptDict, errMap)
	logger.Debug(true, "Parsing completed.")
	r := startAPI(debug)
	vErrs := NewValidator(s, r, conceptDict).Validate()
	errMap = getErrMap(errMap, vErrs)
	s = parser.GetSpecsForDataTableRows(s, errMap)
	printValidationFailures(vErrs)
	showSuggestion(vErrs)
	if !res.Ok {
		err := r.Kill()
		if err != nil {
			logger.Errorf(true, "unable to kill runner: %s", err.Error())
		}
		return NewValidationResult(nil, nil, nil, false, errors.New("Parsing failed"))
	}
	if specsFailed {
		return NewValidationResult(gauge.NewSpecCollection(s, false), errMap, r, false)
	}
	return NewValidationResult(gauge.NewSpecCollection(s, false), errMap, r, true)
}

func getErrMap(errMap *gauge.BuildErrors, validationErrors validationErrors) *gauge.BuildErrors {
	for spec, valErrors := range validationErrors {
		for _, err := range valErrors {
			switch e := err.(type) {
			case StepValidationError:
				errMap.StepErrs[e.step] = e
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
		if err, ok := errMap.StepErrs[step]; ok { // nolint
			errMap.ScenarioErrs[scenario] = append(errMap.ScenarioErrs[scenario], err)
		}
	}
}

func fillSpecErrors(spec *gauge.Specification, errMap *gauge.BuildErrors, steps []*gauge.Step) {
	for _, context := range steps {
		if context.IsConcept {
			fillSpecErrors(spec, errMap, context.ConceptSteps)
		}
		if err, ok := errMap.StepErrs[context]; ok { // nolint
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
	for _, e := range FilterDuplicates(validationErrors) {
		logger.Errorf(true, "[ValidationError] %s", e.Error())
	}
}

func FilterDuplicates(validationErrors validationErrors) []error {
	filteredErrs := make([]error, 0)
	exists := make(map[string]bool)
	for _, errs := range validationErrors {
		for _, e := range errs {
			var val string
			if vErr, ok := e.(StepValidationError); ok {
				val = vErr.step.Value + vErr.step.FileName + strconv.Itoa(e.(StepValidationError).step.LineNo)
			} else if vErr, ok := e.(SpecValidationError); ok {
				val = vErr.message + vErr.fileName
			} else {
				continue
			}
			if _, ok := exists[val]; !ok {
				exists[val] = true
				filteredErrs = append(filteredErrs, e)
			}
		}
	}
	return filteredErrs
}

type validationErrors map[*gauge.Specification][]error

func NewValidator(s []*gauge.Specification, r runner.Runner, c *gauge.ConceptDictionary) *validator {
	return &validator{specsToExecute: s, runner: r, conceptsDictionary: c}
}

func (v *validator) Validate() validationErrors {
	validationStatus := make(validationErrors)
	logger.Debug(true, "Validation started.")
	specValidator := &SpecValidator{runner: v.runner, conceptsDictionary: v.conceptsDictionary, stepValidationCache: make(map[string]error)}
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
	logger.Debug(true, "Validation completed.")
	return nil
}

func (v *SpecValidator) validate() []error {
	queue := &gauge.ItemQueue{Items: v.specification.AllItems()}
	v.specification.Traverse(v, queue)
	return v.validationErrors
}

// Validates a step. If validation result from runner is not valid then it creates a new validation error.
// If the error type is StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND then gives suggestion with step implementation stub.
func (v *SpecValidator) Step(s *gauge.Step) {
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
				NewStepValidationError(s, valErr.message, v.specification.FileName, valErr.errorType, valErr.suggestion))
		} else {
			cpt := v.conceptsDictionary.Search(s.Parent.Value)
			v.validationErrors = append(v.validationErrors,
				NewStepValidationError(s, valErr.message, cpt.FileName, valErr.errorType, valErr.suggestion))
		}
	}
}

var invalidResponse gm.StepValidateResponse_ErrorType = -1

func (v *SpecValidator) validateStep(s *gauge.Step) error {
	stepValue, err := parser.ExtractStepValueAndParams(s.LineText, s.HasInlineTable)
	if err != nil {
		return nil
	}
	protoStepValue := gauge.ConvertToProtoStepValue(stepValue)

	m := &gm.Message{MessageType: gm.Message_StepValidateRequest,
		StepValidateRequest: &gm.StepValidateRequest{StepText: s.Value, NumberOfParameters: int32(len(s.Args)), StepValue: protoStepValue}}

	r, err := v.runner.ExecuteMessageWithTimeout(m)
	if err != nil {
		return NewStepValidationError(s, err.Error(), v.specification.FileName, &invalidResponse, "")
	}
	if r.GetMessageType() == gm.Message_StepValidateResponse {
		res := r.GetStepValidateResponse()
		if !res.GetIsValid() {
			msg := getMessage(res.GetErrorType().String())
			suggestion := res.GetSuggestion()
			if s.Parent == nil {
				vErr := NewStepValidationError(s, msg, v.specification.FileName, &res.ErrorType, suggestion)
				return vErr
			}
			cpt := v.conceptsDictionary.Search(s.Parent.Value)
			vErr := NewStepValidationError(s, msg, cpt.FileName, &res.ErrorType, suggestion)
			return vErr

		}
		return nil
	}
	return NewStepValidationError(s, "Invalid response from runner for Validation request", v.specification.FileName, &invalidResponse, "")
}

func getMessage(message string) string {
	lower := strings.ToLower(strings.Replace(message, "_", " ", -1))
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func (v *SpecValidator) TearDown(step *gauge.TearDown) {
}

func (v *SpecValidator) Heading(heading *gauge.Heading) {
}

func (v *SpecValidator) Tags(tags *gauge.Tags) {
}

func (v *SpecValidator) Table(dataTable *gauge.Table) {

}

func (v *SpecValidator) Scenario(scenario *gauge.Scenario) {

}

func (v *SpecValidator) Comment(comment *gauge.Comment) {
}

func (v *SpecValidator) DataTable(dataTable *gauge.DataTable) {

}

// Validates data table for the range, if any error found append to the validation errors
func (v *SpecValidator) Specification(specification *gauge.Specification) {
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
