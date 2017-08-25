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
	"reflect"
	"testing"

	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
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

func TestGetSuggestionMessageForOtherValidationErrors(t *testing.T) {
	want := "Suggestions for fixing `Duplicate step implementation` errors.\n"

	got := getSuggestionMessage(dupImplFound)

	if got != want {
		t.Errorf("Wrong suggestion message. got: %v, want: %v", got, want)
	}
}
