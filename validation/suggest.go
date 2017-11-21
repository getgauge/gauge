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
	"fmt"

	gm "github.com/getgauge/gauge/gauge_messages"
)

var message = map[gm.StepValidateResponse_ErrorType]string{
	gm.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND: "Add the following missing implementations to fix `Step implementation not found` errors.\n",
}

func showSuggestion(validationErrors validationErrors) {
	if !HideSuggestion {
		for t, errs := range groupErrors(validationErrors) {
			fmt.Println(getSuggestionMessage(t))
			suggestions := filterDuplicateSuggestions(errs)
			for _, suggestion := range suggestions {
				fmt.Println(suggestion)
			}
		}
	}
}

func filterDuplicateSuggestions(errors []StepValidationError) []string {
	suggestionMap := make(map[string]error)
	filteredSuggestions := make([]string, 0)
	for _, err := range errors {
		if _, ok := suggestionMap[err.Suggestion()]; !ok {
			suggestionMap[err.Suggestion()] = err
			filteredSuggestions = append(filteredSuggestions, err.Suggestion())
		}
	}
	return filteredSuggestions
}

func getSuggestionMessage(t gm.StepValidateResponse_ErrorType) string {
	if msg, ok := message[t]; ok {
		return msg
	}
	return fmt.Sprintf("Suggestions for fixing `%s` errors.\n", getMessage(t.String()))
}

func groupErrors(validationErrors validationErrors) map[gm.StepValidateResponse_ErrorType][]StepValidationError {
	errMap := make(map[gm.StepValidateResponse_ErrorType][]StepValidationError)
	for _, errs := range validationErrors {
		for _, v := range errs {
			if e, ok := v.(StepValidationError); ok && e.suggestion != "" {
				errType := *(v.(StepValidationError).errorType)
				errMap[errType] = append(errMap[errType], e)
			}
		}
	}
	return errMap
}
