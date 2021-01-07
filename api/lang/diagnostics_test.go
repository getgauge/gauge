/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"os"
	"testing"
	"time"

	"reflect"

	"strings"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

var conceptFile = "foo.cpt"
var specFile = "foo.spec"

func TestMain(m *testing.M) {
	exitCode := m.Run()
	tearDown()
	os.Exit(exitCode)
}

func setup() {
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(util.ConvertPathToURI(conceptFile), "")
	openFilesCache.add(util.ConvertPathToURI(specFile), "")
	responses := map[gauge_messages.Message_MessageType]interface{}{}
	responses[gauge_messages.Message_StepValidateResponse] = &gauge_messages.StepValidateResponse{IsValid: true}
	lRunner.runner = &runner.GrpcRunner{LegacyClient: &mockClient{responses: responses}, Timeout: time.Second * 30}

	util.GetConceptFiles = func() []string {
		return []string{conceptFile}
	}

	util.GetSpecFiles = func(paths []string) []string {
		return []string{specFile}
	}
}
func tearDown() {
	lRunner.runner = nil
}

func TestDiagnosticWithParseErrorsInSpec(t *testing.T) {
	setup()
	specText := `Specification Heading
=====================

Scenario Heading
================

* Step text`

	uri := util.ConvertPathToURI(specFile)
	openFilesCache.add(uri, specText)

	want := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 1, Character: 21},
			},
			Message:  "Spec should have atleast one scenario",
			Severity: 1,
		},
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 3, Character: 0},
				End:   lsp.Position{Line: 4, Character: 16},
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
	uri := util.ConvertPathToURI(specFile)
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
	uri := util.ConvertPathToURI(conceptFile)
	openFilesCache.add(uri, cptText)

	diagnostics := make(map[lsp.DocumentURI][]lsp.Diagnostic)

	dictionary, err := validateConcepts(diagnostics)
	if err != nil {
		t.Errorf("expected no error.\n Got: %s", err.Error())
	}

	if len(dictionary.ConceptsMap) == 0 {
		t.Errorf("Concept dictionary is empty")
	}

	if len(diagnostics[uri]) > 0 {
		t.Errorf("Parsing failed, got : %+v", diagnostics)
	}
}

func TestDiagnosticsForConceptParseErrors(t *testing.T) {
	setup()
	cptText := `# concept`

	uri := util.ConvertPathToURI(conceptFile)

	openFilesCache.add(uri, cptText)

	diagnostics := make(map[lsp.DocumentURI][]lsp.Diagnostic)

	_, err := validateConcepts(diagnostics)

	if err != nil {
		t.Error(err)
	}
	if len(diagnostics[uri]) <= 0 {
		t.Errorf("expected parse errors")
	}

	want := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 9},
			},
			Message:  "Concept should have atleast one step",
			Severity: 1,
		},
	}

	if !reflect.DeepEqual(want, diagnostics[uri]) {
		t.Errorf("want: `%v`,\n got: `%v`", want, diagnostics[uri])
	}
}

func TestDiagnosticOfConceptsWithCircularReference(t *testing.T) {
	setup()
	cptText := `# concept
* concept
`
	uri := util.ConvertPathToURI(conceptFile)
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
