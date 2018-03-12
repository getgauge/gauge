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

	"strings"

	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/validation"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

var conceptFile = "foo.cpt"
var specFile = "foo.spec"

func setup() {
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(util.ConvertPathToURI(lsp.DocumentURI(conceptFile)), "")
	openFilesCache.add(util.ConvertPathToURI(lsp.DocumentURI(specFile)), "")

	validation.GetResponseFromRunner = func(m *gauge_messages.Message, v *validation.SpecValidator) (*gauge_messages.Message, error) {
		res := &gauge_messages.StepValidateResponse{IsValid: true}
		return &gauge_messages.Message{MessageType: gauge_messages.Message_StepValidateResponse, StepValidateResponse: res}, nil
	}

	util.GetConceptFiles = func() []string {
		return []string{conceptFile}
	}

	util.GetSpecFiles = func(path string) []string {
		return []string{specFile}
	}
}

func TestDiagnosticWithParseErrorsInSpec(t *testing.T) {
	setup()
	specText := `Specification Heading
=====================

Scenario Heading
================

* Step text`

	uri := util.ConvertPathToURI(lsp.DocumentURI(specFile))
	openFilesCache.add(uri, specText)

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

	diagnostics, err := getDiagnostics()
	if err != nil {
		t.Errorf("Expected no error, got : %s", err.Error())
	}

	got := diagnostics[uri]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%+v`,\n got: `%+v`", want, got)
	}
}

func TestDiagnosticWithNoErrors(t *testing.T) {
	setup()
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text
`
	uri := util.ConvertPathToURI(lsp.DocumentURI(specFile))
	openFilesCache.add(uri, specText)
	d, err := getDiagnostics()
	if err != nil {
		t.Errorf("expected no error.\n Got: %s", err.Error())
	}
	if len(d[uri]) > 0 {
		t.Errorf("expected no error.\n Got: %+v", d)
	}
}

func TestParseConcept(t *testing.T) {
	setup()
	cptText := `# concept
* foo
`
	uri := util.ConvertPathToURI(lsp.DocumentURI(conceptFile))
	openFilesCache.add(uri, cptText)

	diagnostics := diagnosticsMap{diagnostics: make(map[lsp.DocumentURI][]lsp.Diagnostic, 0)}

	dictionary, err := validateConcepts(diagnostics)
	if err != nil {
		t.Errorf("expected no error.\n Got: %s", err.Error())
	}

	if len(dictionary.ConceptsMap) == 0 {
		t.Errorf("Concept dictionary is empty")
	}

	if len(diagnostics.diagnostics[uri]) > 0 {
		t.Errorf("Parsing failed, got : %+v", diagnostics)
	}
}

func TestDiagnosticsForConceptParseErrors(t *testing.T) {
	setup()
	cptText := `# concept`

	uri := util.ConvertPathToURI(lsp.DocumentURI(conceptFile))

	openFilesCache.add(uri, cptText)

	diagnostics := diagnosticsMap{diagnostics: make(map[lsp.DocumentURI][]lsp.Diagnostic, 0)}

	validateConcepts(diagnostics)
	if len(diagnostics.diagnostics[uri]) <= 0 {
		t.Errorf("expected parse errors")
	}

	want := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{0, 0},
				End:   lsp.Position{0, 9},
			},
			Message:  "Concept should have atleast one step",
			Severity: 1,
		},
	}

	if !reflect.DeepEqual(want, diagnostics.diagnostics[uri]) {
		t.Errorf("want: `%s`,\n got: `%s`", want, diagnostics.diagnostics[uri])
	}
}

func TestDiagnosticOfConceptsWithCircularReference(t *testing.T) {
	setup()
	cptText := `# concept
* concept
`
	uri := util.ConvertPathToURI(lsp.DocumentURI(conceptFile))
	openFilesCache.add(uri, cptText)

	diagnostics, err := getDiagnostics()
	if err != nil {
		t.Errorf("expected no error.\n Got: %s", err.Error())
	}

	got := diagnostics[uri]
	containsDiagnostics(got, 1, 0, "Circular reference found in concept.", t)
}

var containsDiagnostics = func(diagnostics []lsp.Diagnostic, line1, line2 int, startMessage string, t *testing.T) {
	for _, diagnostic := range diagnostics {
		if !strings.Contains(diagnostic.Message, startMessage) {
			t.Errorf("Invalid error message, got : %s : ", diagnostic.Message)
		}
		if (diagnostic.Range.Start.Line != line1 || diagnostic.Range.Start.Line != line2) && diagnostic.Range.Start.Character != 0 {
			t.Errorf("Invalid start in range, got : %+v : ", diagnostic.Range.Start)
		}
		if (diagnostic.Range.End.Line != line1 || diagnostic.Range.End.Line != line2) && diagnostic.Range.End.Character != 9 {
			t.Errorf("Invalid end in range, got : %+v : ", diagnostic.Range.End)
		}
		if diagnostic.Severity != 1 {
			t.Errorf("Invalid diagnostic severity, want : 1, got : %d : ", diagnostic.Severity)
		}
	}
}
