/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

type BuildErrors struct {
	SpecErrs     map[*Specification][]error
	ScenarioErrs map[*Scenario][]error
	StepErrs     map[*Step]error
}

func (e *BuildErrors) HasErrors() bool {
	return (len(e.SpecErrs) + len(e.ScenarioErrs) + len(e.StepErrs)) > 0
}

func NewBuildErrors() *BuildErrors {
	return &BuildErrors{
		SpecErrs:     make(map[*Specification][]error),
		ScenarioErrs: make(map[*Scenario][]error),
		StepErrs:     make(map[*Step]error),
	}
}
