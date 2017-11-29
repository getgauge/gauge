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

	"errors"

	"bytes"

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
	spec, _, _ := p.Parse(specText, gauge.NewConceptDictionary(), "")
	err := gauge_messages.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND
	errs := validationErrors{spec: []error{
		NewStepValidationError(spec.Scenarios[0].Steps[0], "", "", &err, ""),
		NewStepValidationError(spec.Scenarios[1].Steps[0], "", "", &err, ""),
	}}

	errMap := getErrMap(gauge.NewBuildErrors(), errs)

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
	spec, _, _ := p.Parse(specText, gauge.NewConceptDictionary(), "")
	err := gauge_messages.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND

	errs := validationErrors{spec: []error{
		NewStepValidationError(spec.Scenarios[0].Steps[0], "", "", &err, ""),
	}}

	errMap := getErrMap(gauge.NewBuildErrors(), errs)

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
	spec, _, _ := p.Parse(specText, gauge.NewConceptDictionary(), "")

	errs := validationErrors{spec: []error{}}

	errMap := getErrMap(gauge.NewBuildErrors(), errs)

	c.Assert(len(errMap.SpecErrs), Equals, 0)
	c.Assert(len(errMap.ScenarioErrs), Equals, 0)
	c.Assert(len(errMap.StepErrs), Equals, 0)
}

func (s *MySuite) TestSkipSpecIfTableRowOutOfRange(c *C) {
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
	spec, _, _ := p.Parse(specText, gauge.NewConceptDictionary(), "")

	errs := validationErrors{spec: []error{
		NewSpecValidationError("Table row out of range", spec.FileName),
	}}

	errMap := getErrMap(gauge.NewBuildErrors(), errs)

	c.Assert(len(errMap.SpecErrs), Equals, 1)
	c.Assert(len(errMap.ScenarioErrs), Equals, 0)
	c.Assert(len(errMap.StepErrs), Equals, 0)
}

func (s *MySuite) TestValidateStep(c *C) {
	HideSuggestion = false
	var suggestion bytes.Buffer
	myStep := &gauge.Step{Value: "my step", LineText: "my step", IsConcept: false, LineNo: 3}
	GetResponseFromRunner = func(m *gauge_messages.Message, v *SpecValidator) (*gauge_messages.Message, error) {
		suggestion.WriteString("\n\t@Step(\"my step\")\n\tpublic void implementation1(){\n\t\t// your code here...\n\t}")
		res := &gauge_messages.StepValidateResponse{IsValid: false, ErrorMessage: "my err msg", ErrorType: gauge_messages.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND, Suggestion: suggestion.String()}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse, StepValidateResponse: res}, nil
	}
	specVal := &SpecValidator{specification: &gauge.Specification{FileName: "foo.spec"}}
	valErr := specVal.validateStep(myStep)

	c.Assert(valErr, Not(Equals), nil)
	c.Assert(valErr.Error(), Equals, "foo.spec:3 Step implementation not found => 'my step'")
	c.Assert(valErr.(StepValidationError).Suggestion(), Equals, "\n\t"+
		"@Step(\"my step\")\n\t"+
		"public void implementation1(){\n\t"+
		"\t// your code here...\n\t"+
		"}")
}

func (s *MySuite) TestShouldNotGiveSuggestionWhenHideSuggestionFlagIsFalse(c *C) {
	HideSuggestion = true
	myStep := &gauge.Step{Value: "my step", LineText: "my step", IsConcept: false, LineNo: 3}
	GetResponseFromRunner = func(m *gauge_messages.Message, v *SpecValidator) (*gauge_messages.Message, error) {
		res := &gauge_messages.StepValidateResponse{IsValid: false, ErrorMessage: "my err msg", ErrorType: gauge_messages.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse, StepValidateResponse: res}, nil
	}
	specVal := &SpecValidator{specification: &gauge.Specification{FileName: "foo.spec"}}
	valErr := specVal.validateStep(myStep)

	c.Assert(valErr, Not(Equals), nil)
	c.Assert(valErr.Error(), Equals, "foo.spec:3 Step implementation not found => 'my step'")
	c.Assert(valErr.(StepValidationError).suggestion, Equals, "")
}

func (s *MySuite) TestValidateStepInConcept(c *C) {
	HideSuggestion = false
	var suggestion bytes.Buffer
	parentStep := &gauge.Step{Value: "my concept", LineNo: 2, IsConcept: true, LineText: "my concept"}
	myStep := &gauge.Step{Value: "my step", LineText: "my step", IsConcept: false, LineNo: 3, Parent: parentStep}
	GetResponseFromRunner = func(m *gauge_messages.Message, v *SpecValidator) (*gauge_messages.Message, error) {
		suggestion.WriteString("\n\t@Step(\"my step\")\n\tpublic void implementation1(){\n\t\t// your code here...\n\t}")
		res := &gauge_messages.StepValidateResponse{IsValid: false, ErrorMessage: "my err msg", ErrorType: gauge_messages.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND, Suggestion: suggestion.String()}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse, StepValidateResponse: res}, nil
	}
	cptDict := gauge.NewConceptDictionary()
	cptDict.ConceptsMap["my concept"] = &gauge.Concept{ConceptStep: parentStep, FileName: "concept.cpt"}
	specVal := &SpecValidator{specification: &gauge.Specification{FileName: "foo.spec"}, conceptsDictionary: cptDict}
	valErr := specVal.validateStep(myStep)

	c.Assert(valErr, Not(Equals), nil)
	c.Assert(valErr.Error(), Equals, "concept.cpt:3 Step implementation not found => 'my step'")
	c.Assert(valErr.(StepValidationError).Suggestion(), Equals, "\n\t@Step(\"my step\")\n\t"+
		"public void implementation1(){\n\t"+
		"\t// your code here...\n\t"+
		"}")
}

type tableRow struct {
	name           string
	input          string
	tableRowsCount int
	err            error
}

var tableRowTests = []*tableRow{
	{"Valid single row number", "3", 5, nil},
	{"Invalid single row number", "2", 1, errors.New("Table rows range validation failed => Table row number '2' is out of range")},
	{"Valid row numbers list", "2,3,4", 4, nil},
	{"Invalid list with empty value", ",3,4", 4, errors.New("Table rows range validation failed => Row number cannot be empty")},
	{"Invalid row numbers list", "2,3,4", 3, errors.New("Table rows range validation failed => Table row number '4' is out of range")},
	{"Invalid row numbers list with special chars", "2*&", 3, errors.New("Table rows range validation failed => Failed to parse '2*&' to row number")},
	{"Valid table rows range", "2-5", 5, nil},
	{"Invalid table rows range", "2-5", 4, errors.New("Table rows range validation failed => Table row number '5' is out of range")},
	{"Invalid table rows range", "2-2", 4, nil},
	{"Invalid table rows with character", "a", 4, errors.New("Table rows range validation failed => Failed to parse 'a' to row number")},
	{"Invalid table rows range with character", "a-5", 5, errors.New("Table rows range validation failed => Failed to parse 'a' to row number")},
	{"Invalid table rows range with string", "a-qwerty", 4, errors.New("Table rows range validation failed => Failed to parse 'a' to row number")},
	{"Empty table rows range", "", 4, nil},
	{"Table rows range with multiple -", "2-3-4", 4, errors.New("Table rows range '2-3-4' is invalid => Table rows range should be of format rowNumber-rowNumber")},
	{"Table rows range with different separator", "2:4", 4, errors.New("Table rows range validation failed => Failed to parse '2:4' to row number")},
	{"Table rows list with spaces", "2, 4 ", 4, nil},
	{"Row count is zero with empty input", "", 0, nil},
	{"Row count is zero with non empty input", "1", 0, errors.New("Table rows range validation failed => Table row number '1' is out of range")},
	{"Row count is non-zero with empty input", "", 2, nil},
	{"Row count is non-zero with non-empty input", "2", 2, nil},
}

func (s *MySuite) TestToValidateDataTableRowsRangeFromInputFlag(c *C) {
	for _, test := range tableRowTests {
		TableRows = test.input
		got := validateDataTableRange(test.tableRowsCount)
		want := test.err
		c.Assert(got, DeepEquals, want, Commentf(test.name))
	}
}
