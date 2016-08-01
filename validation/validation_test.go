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

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	"github.com/golang/protobuf/proto"

	. "gopkg.in/check.v1"
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
	spec, _ := p.Parse(specText, gauge.NewConceptDictionary(), "")

	errs := validationErrors{spec: []*StepValidationError{
		{message: "", fileName: "", step: spec.Scenarios[0].Steps[0]},
		{message: "", fileName: "", step: spec.Scenarios[1].Steps[0]},
	}}

	errMap := getErrMap(errs)

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
	spec, _ := p.Parse(specText, gauge.NewConceptDictionary(), "")

	errs := validationErrors{spec: []*StepValidationError{
		{message: "", fileName: "", step: spec.Scenarios[0].Steps[0]},
	}}

	errMap := getErrMap(errs)

	c.Assert(len(errMap.SpecErrs), Equals, 0)
	c.Assert(len(errMap.ScenarioErrs), Equals, 1)
	c.Assert(len(errMap.StepErrs), Equals, 1)
}

func (s *MySuite) TestSkipSpecIfNoScenariosPresent(c *C) {
	specText := `Specification Heading
=====================
* say hello1
* say hello2
`
	p := new(parser.SpecParser)
	spec, _ := p.Parse(specText, gauge.NewConceptDictionary(), "")

	errs := validationErrors{spec: []*StepValidationError{}}

	errMap := getErrMap(errs)

	c.Assert(len(errMap.SpecErrs), Equals, 0)
	c.Assert(len(errMap.ScenarioErrs), Equals, 0)
	c.Assert(len(errMap.StepErrs), Equals, 0)
}

func (s *MySuite) TestValidateStep(c *C) {
	myStep := &gauge.Step{Value: "my step", LineText: "my step", IsConcept: false, LineNo: 3}
	getResponseFromRunner = func(m *gauge_messages.Message, v *specValidator) (*gauge_messages.Message, error) {
		res := &gauge_messages.StepValidateResponse{IsValid: proto.Bool(false), ErrorMessage: proto.String("my err msg"), ErrorType: gauge_messages.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND.Enum()}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse.Enum(), StepValidateResponse: res}, nil
	}
	specVal := &specValidator{specification: &gauge.Specification{FileName: "foo.spec"}}
	valErr := specVal.validateStep(myStep)

	c.Assert(valErr, Not(Equals), nil)
	c.Assert(valErr.Error(), Equals, "foo.spec:3: Step implementation not found => 'my step'")
}

func (s *MySuite) TestValidateStepInConcept(c *C) {
	parentStep := &gauge.Step{Value: "my concept", LineNo: 2, IsConcept: true, LineText: "my concept"}
	myStep := &gauge.Step{Value: "my step", LineText: "my step", IsConcept: false, LineNo: 3, Parent: parentStep}
	getResponseFromRunner = func(m *gauge_messages.Message, v *specValidator) (*gauge_messages.Message, error) {
		res := &gauge_messages.StepValidateResponse{IsValid: proto.Bool(false), ErrorMessage: proto.String("my err msg"), ErrorType: gauge_messages.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND.Enum()}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse.Enum(), StepValidateResponse: res}, nil
	}
	cptDict := gauge.NewConceptDictionary()
	cptDict.ConceptsMap["my concept"] = &gauge.Concept{ConceptStep: parentStep, FileName: "concept.cpt"}
	specVal := &specValidator{specification: &gauge.Specification{FileName: "foo.spec"}, conceptsDictionary: cptDict}
	valErr := specVal.validateStep(myStep)

	c.Assert(valErr, Not(Equals), nil)
	c.Assert(valErr.Error(), Equals, "concept.cpt:3: Step implementation not found => 'my step'")
}
