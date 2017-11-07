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

package lang

import (
	"testing"

	"reflect"

	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/getgauge/gauge/validation"
	"github.com/getgauge/gauge/gauge_messages"
)

func TestDiagnostic(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
================

* Step text`

	uri := "foo.spec"
	provider = infoGatherer.NewSpecInfoGatherer(gauge.NewConceptDictionary())
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	want := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{0, 0},
				End:   lsp.Position{0, 21},
			},
			Message:  "Spec should have atleast one scenario",
			Severity: 1,
		},
		{
			Range: lsp.Range{
				Start: lsp.Position{3, 0},
				End:   lsp.Position{3, 16},
			},
			Message:  "Multiple spec headings found in same file",
			Severity: 1,
		},
	}
	validation.GetResponseFromRunner = func(m *gauge_messages.Message, v *validation.SpecValidator) (*gauge_messages.Message, error) {
		res := &gauge_messages.StepValidateResponse{IsValid: true}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse, StepValidateResponse: res}, nil
	}
	got := createDiagnostics(uri)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}

func TestDiagnosticWithUnimplementedStepError(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text
`

	uri := "foo.spec"
	provider = infoGatherer.NewSpecInfoGatherer(gauge.NewConceptDictionary())
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	want := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{6, 0},
				End:   lsp.Position{6, 11},
			},
			Message:  "Step implementation not found",
			Severity: 1,
		},

	}
	validation.GetResponseFromRunner = func(m *gauge_messages.Message, v *validation.SpecValidator) (*gauge_messages.Message, error) {
		res := &gauge_messages.StepValidateResponse{IsValid: false, ErrorType: gauge_messages.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse, StepValidateResponse: res}, nil
	}
	got := createDiagnostics(uri)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}


func TestDiagnosticWithDuplicateStepError(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text
`

	uri := "foo.spec"
	provider = infoGatherer.NewSpecInfoGatherer(gauge.NewConceptDictionary())
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	want := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{6, 0},
				End:   lsp.Position{6, 11},
			},
			Message:  "Duplicate step implementation",
			Severity: 1,
		},

	}
	validation.GetResponseFromRunner = func(m *gauge_messages.Message, v *validation.SpecValidator) (*gauge_messages.Message, error) {
		res := &gauge_messages.StepValidateResponse{IsValid: false, ErrorType: gauge_messages.StepValidateResponse_DUPLICATE_STEP_IMPLEMENTATION}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse, StepValidateResponse: res}, nil
	}
	got := createDiagnostics(uri)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}


func TestDiagnosticWithoutParseError(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text`

	uri := "foo.spec"

	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)
	provider = infoGatherer.NewSpecInfoGatherer(gauge.NewConceptDictionary())
	provider.GetConceptDictionary().ConceptsMap = map[string]*gauge.Concept{}

	validation.GetResponseFromRunner = func(m *gauge_messages.Message, v *validation.SpecValidator) (*gauge_messages.Message, error) {
		res := &gauge_messages.StepValidateResponse{IsValid: true}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse, StepValidateResponse: res}, nil
	}

	d := createDiagnostics(uri)

	if len(d) > 0{
		t.Errorf("want: `%s` errors,\n got: `%v`", 0, len(d))
	}
}

func TestParseConcept(t *testing.T) {
	cptText := `# concept
* foo
`
	uri := "foo.cpt"

	f = &files{cache: make(map[string][]string)}
	f.add(uri, cptText)

	provider = infoGatherer.NewSpecInfoGatherer(gauge.NewConceptDictionary())
	provider.GetConceptDictionary().ConceptsMap = map[string]*gauge.Concept{}
	res := validateConcept(uri, uri)

	if len(res.ParseErrors) > 0 {
		t.Errorf("parsing failed, %s", res.Errors())
	}
}

func TestParseConceptChanges(t *testing.T) {
	cptText := `# concept
* foo
`
	uri := "foo.cpt"

	f = &files{cache: make(map[string][]string)}
	f.add(uri, cptText)
	provider = infoGatherer.NewSpecInfoGatherer(gauge.NewConceptDictionary())
	res := validateConcept(uri, uri)

	if len(res.ParseErrors) > 0 {
		t.Errorf("parsing failed, %s", res.Errors())
	}

	cptText = `# concept
`
	f.add(uri, cptText)
	res = validateConcept(uri, uri)
	if len(res.ParseErrors) <= 0 {
		t.Errorf("expected parse errors")
	}

	expectedError := parser.ParseError{
		FileName: "foo.cpt",
		LineNo:   1,
		Message:  "Concept should have atleast one step",
		LineText: "concept",
	}

	if !reflect.DeepEqual(expectedError, res.ParseErrors[0]) {
		t.Errorf("want: `%s`,\n got: `%s`", expectedError, res.ParseErrors[0])
	}
}

func TestConceptDiagnostic(t *testing.T) {
	cptText := `# concept
`

	uri := "foo.cpt"

	f = &files{cache: make(map[string][]string)}
	f.add(uri, cptText)

	provider = infoGatherer.NewSpecInfoGatherer(gauge.NewConceptDictionary())
	provider.GetConceptDictionary().ConceptsMap = map[string]*gauge.Concept{
		"concept": {
			ConceptStep: &gauge.Step{Value: "concept", LineNo: 1, IsConcept: true, LineText: "concept"},
			FileName:    uri,
		},
	}

	want := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{0, 0},
				End:   lsp.Position{0, 9},
			},
			Severity: 1,
			Message:  "Concept should have atleast one step",
		},
	}

	got := createDiagnostics(uri)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}

func TestConceptDiagnosticWithCircularReference(t *testing.T) {
	cptText := `# concept
* concept
`

	uri := "foo.cpt"

	f = &files{cache: make(map[string][]string)}
	f.add(uri, cptText)

	provider = infoGatherer.NewSpecInfoGatherer(gauge.NewConceptDictionary())

	want := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{0, 0},
				End:   lsp.Position{0, 9},
			},
			Severity: 1,
			Message:  `Circular reference found in concept. "concept" => foo.cpt:2`,
		},
	}

	got := createDiagnostics(uri)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}
