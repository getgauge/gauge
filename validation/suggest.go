/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package validation

import (
	"fmt"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/logger"
)

var message = map[gm.StepValidateResponse_ErrorType]string{
	gm.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND: "Add the following missing implementations to fix `Step implementation not found` errors.\n",
}

func showSuggestion(validationErrors validationErrors) {
	if !HideSuggestion {
		for t, errs := range groupErrors(validationErrors) {
			logger.Infof(true, getSuggestionMessage(t))
			suggestions := filterDuplicateSuggestions(errs)
			for _, suggestion := range suggestions {
				logger.Infof(true, suggestion)
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
