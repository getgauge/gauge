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

	want := []lsp.Diagnostic{}

	got := createDiagnostics(uri)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
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
