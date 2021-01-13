/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package validation

import (
	"errors"
	"reflect"
	"testing"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/gauge"
)

var implNotFound = gm.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND
var dupImplFound = gm.StepValidateResponse_DUPLICATE_STEP_IMPLEMENTATION

func TestGroupErrors(t *testing.T) {
	implNotFoundError := StepValidationError{
		errorType:  &implNotFound,
		suggestion: "suggestion",
	}
	dupImplFoundError := StepValidationError{
		errorType:  &dupImplFound,
		suggestion: "suggestion1",
	}
	errs := map[*gauge.Specification][]error{
		&gauge.Specification{}: []error{implNotFoundError},
		&gauge.Specification{}: []error{dupImplFoundError},
		&gauge.Specification{}: []error{errors.New("error")},
		&gauge.Specification{}: []error{implNotFoundError, dupImplFoundError},
	}
	want := map[gm.StepValidateResponse_ErrorType][]StepValidationError{
		implNotFound: []StepValidationError{implNotFoundError, implNotFoundError},
		dupImplFound: []StepValidationError{dupImplFoundError, dupImplFoundError},
	}
	got := groupErrors(validationErrors(errs))

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Grouping of step validation errors failed. got: %v, want: %v", got, want)
	}
}

func TestGetSuggestionMessageForStepImplNotFoundError(t *testing.T) {
	want := message[implNotFound]

	got := getSuggestionMessage(implNotFound)

	if got != want {
		t.Errorf("Wrong suggestion message. got: %v, want: %v", got, want)
	}
}

func TestFilterDuplicateSuggestions(t *testing.T) {
	dupImplFoundError := StepValidationError{
		errorType:  &dupImplFound,
		suggestion: "suggestion1",
	}

	want := []string{"suggestion1"}
	got := filterDuplicateSuggestions([]StepValidationError{dupImplFoundError, dupImplFoundError})

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wrong suggestion message. got: %v, want: %v", got, want)
	}
}

func TestGetSuggestionMessageForOtherValidationErrors(t *testing.T) {
	want := "Suggestions for fixing `Duplicate step implementation` errors.\n"

	got := getSuggestionMessage(dupImplFound)

	if got != want {
		t.Errorf("Wrong suggestion message. got: %v, want: %v", got, want)
	}
}
