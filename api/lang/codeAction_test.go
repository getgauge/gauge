/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestGetCodeActionForUnimplementedStep(t *testing.T) {
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(lsp.DocumentURI("foo.spec"), "# spec heading\n## scenario heading\n* foo bar")

	stub := "a stub for unimplemented step"
	d := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 2, Character: 0},
				End:   lsp.Position{Line: 2, Character: 9},
			},
			Message:  "Step implantation not found",
			Severity: 1,
			Code:     stub,
		},
	}
	codeActionParams := lsp.CodeActionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"},
		Context:      lsp.CodeActionContext{Diagnostics: d},
		Range: lsp.Range{
			Start: lsp.Position{Line: 2, Character: 0},
			End:   lsp.Position{Line: 2, Character: 9},
		},
	}
	b, _ := json.Marshal(codeActionParams)
	p := json.RawMessage(b)

	want := []lsp.Command{
		{
			Command:   generateStepCommand,
			Title:     generateStubTitle,
			Arguments: []interface{}{stub},
		},
		{
			Command:   generateConceptCommand,
			Title:     generateConceptTitle,
			Arguments: []interface{}{concpetInfo{ConceptName: "# foo bar\n* "}},
		},
	}

	got, err := codeActions(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. \nGot : %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}

func TestGetCodeActionForUnimplementedStepWithParam(t *testing.T) {
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(lsp.DocumentURI("foo.spec"), "# spec heading\n## scenario heading\n* foo bar \"some\"")

	stub := "a stub for unimplemented step"
	d := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 2, Character: 0},
				End:   lsp.Position{Line: 2, Character: 9},
			},
			Message:  "Step implantation not found",
			Severity: 1,
			Code:     stub,
		},
	}
	codeActionParams := lsp.CodeActionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"},
		Context:      lsp.CodeActionContext{Diagnostics: d},
		Range: lsp.Range{
			Start: lsp.Position{Line: 2, Character: 0},
			End:   lsp.Position{Line: 2, Character: 9},
		},
	}
	b, _ := json.Marshal(codeActionParams)
	p := json.RawMessage(b)

	want := []lsp.Command{
		{
			Command:   generateStepCommand,
			Title:     generateStubTitle,
			Arguments: []interface{}{stub},
		},
		{
			Command:   generateConceptCommand,
			Title:     generateConceptTitle,
			Arguments: []interface{}{concpetInfo{ConceptName: "# foo bar <arg0>\n* "}},
		},
	}

	got, err := codeActions(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. \nGot : %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}

func TestGetCodeActionForUnimplementedStepWithTableParam(t *testing.T) {
	specText := `#Specification Heading

##Scenario Heading

* Step text
	|Head|
   	|----|
   	|some|`
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(lsp.DocumentURI("foo.spec"), specText)

	stub := "a stub for unimplemented step"
	d := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 4, Character: 0},
				End:   lsp.Position{Line: 4, Character: 12},
			},
			Message:  "Step implantation not found",
			Severity: 1,
			Code:     stub,
		},
	}
	codeActionParams := lsp.CodeActionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"},
		Context:      lsp.CodeActionContext{Diagnostics: d},
		Range: lsp.Range{
			Start: lsp.Position{Line: 4, Character: 0},
			End:   lsp.Position{Line: 4, Character: 12},
		},
	}
	b, _ := json.Marshal(codeActionParams)
	p := json.RawMessage(b)

	want := []lsp.Command{
		{
			Command:   generateStepCommand,
			Title:     generateStubTitle,
			Arguments: []interface{}{stub},
		},
		{
			Command:   generateConceptCommand,
			Title:     generateConceptTitle,
			Arguments: []interface{}{concpetInfo{ConceptName: "# Step text <arg0>\n* "}},
		},
	}

	got, err := codeActions(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. \nGot : %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}

func TestGetCodeActionForUnimplementedStepWithFileParameter(t *testing.T) {
	specText := `#Specification Heading

##Scenario Heading

* Step text <file:_testdata/dummyFile.txt>`
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(lsp.DocumentURI("foo.spec"), specText)

	stub := "a stub for unimplemented step"
	d := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 4, Character: 0},
				End:   lsp.Position{Line: 4, Character: 12},
			},
			Message:  "Step implantation not found",
			Severity: 1,
			Code:     stub,
		},
	}
	codeActionParams := lsp.CodeActionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"},
		Context:      lsp.CodeActionContext{Diagnostics: d},
		Range: lsp.Range{
			Start: lsp.Position{Line: 4, Character: 0},
			End:   lsp.Position{Line: 4, Character: 12},
		},
	}
	b, _ := json.Marshal(codeActionParams)
	p := json.RawMessage(b)

	want := []lsp.Command{
		{
			Command:   generateStepCommand,
			Title:     generateStubTitle,
			Arguments: []interface{}{stub},
		},
		{
			Command:   generateConceptCommand,
			Title:     generateConceptTitle,
			Arguments: []interface{}{concpetInfo{ConceptName: "# Step text <arg0>\n* "}},
		},
	}

	got, err := codeActions(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. \nGot : %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}

func TestNotToPanicForUnimplementedWithInvalidStartLine(t *testing.T) {
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(lsp.DocumentURI("foo.spec"), "# spec heading\n## scenario heading\n* foo bar")

	stub := "a stub for unimplemented step"
	d := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: 0},
			},
			Message:  "Step implantation not found",
			Severity: 1,
			Code:     stub,
		},
	}
	codeActionParams := lsp.CodeActionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"},
		Context:      lsp.CodeActionContext{Diagnostics: d},
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   lsp.Position{Line: 0, Character: 0},
		},
	}
	b, _ := json.Marshal(codeActionParams)
	p := json.RawMessage(b)

	want := []lsp.Command{
		{
			Command:   generateStepCommand,
			Title:     generateStubTitle,
			Arguments: []interface{}{stub},
		},
	}

	got, err := codeActions(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. \nGot : %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}
