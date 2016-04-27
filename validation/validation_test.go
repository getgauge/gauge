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
	"testing"

	. "gopkg.in/check.v1"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/gauge"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestSkipSpecIfAllScenariosAreSkipped(c *C) {
	specText := `Specification Heading
=====================
Scenario 1
----------
* say hello1

Scenario 2
----------
* say hello2
`
	p := new(parser.SpecParser)
	spec, _ := p.Parse(specText, gauge.NewConceptDictionary())

	errMap := &ValidationErrMaps{
		SpecErrs:     make(map[*gauge.Specification][]*StepValidationError),
		ScenarioErrs: make(map[*gauge.Scenario][]*StepValidationError),
		StepErrs:     make(map[*gauge.Step]*StepValidationError),
	}
	errs := validationErrors{spec: []*StepValidationError{
		&StepValidationError{message: "", fileName: "", step: spec.Scenarios[0].Steps[0]},
		&StepValidationError{message: "", fileName: "", step: spec.Scenarios[1].Steps[0]},
	}}

	fillErrors(errMap, errs)

	c.Assert(len(errMap.SpecErrs), Equals, 1)
	c.Assert(len(errMap.ScenarioErrs), Equals, 2)
	c.Assert(len(errMap.StepErrs), Equals, 2)
}


func (s *MySuite) TestDoesNotSkipSpecIfAllScenariosAreNotSkipped(c *C) {
	specText := `Specification Heading
=====================
Scenario 1
----------
* say hello1

Scenario 2
----------
* say hello2
`
	p := new(parser.SpecParser)
	spec, _ := p.Parse(specText, gauge.NewConceptDictionary())

	errMap := &ValidationErrMaps{
		SpecErrs:     make(map[*gauge.Specification][]*StepValidationError),
		ScenarioErrs: make(map[*gauge.Scenario][]*StepValidationError),
		StepErrs:     make(map[*gauge.Step]*StepValidationError),
	}
	errs := validationErrors{spec: []*StepValidationError{
		&StepValidationError{message: "", fileName: "", step: spec.Scenarios[0].Steps[0]},
	}}

	fillErrors(errMap, errs)

	c.Assert(len(errMap.SpecErrs), Equals, 0)
	c.Assert(len(errMap.ScenarioErrs), Equals, 1)
	c.Assert(len(errMap.StepErrs), Equals, 1)
}