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
	"encoding/json"
	"reflect"
	"testing"

	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

var placeHolderTests = []struct {
	input string
	args  []string
	want  string
}{
	{
		input: "say {} to {}",
		args:  []string{"hello", "gauge"},
		want:  `say "${1:hello}" to "${0:gauge}"`,
	},
	{
		input: "say {}",
		args:  []string{"hello"},
		want:  `say "${0:hello}"`,
	},
	{
		input: "say",
		args:  []string{},
		want:  `say`,
	},
}

func TestAddPlaceHolders(t *testing.T) {
	for _, test := range placeHolderTests {
		got := addPlaceHolders(test.input, test.args)
		if got != test.want {
			t.Errorf("Adding Autocomplete placeholder failed, got: `%s`, want: `%s`", got, test.want)
		}
	}
}

type dummyInfoProvider struct{}

func (p *dummyInfoProvider) Init() {}
func (p *dummyInfoProvider) Steps() []*gauge.StepValue {
	return []*gauge.StepValue{{
		Args:                   []string{"hello", "gauge"},
		StepValue:              "Say {} to {}",
		ParameterizedStepValue: "Say <hello> to <gauge>",
	}}
}
func (p *dummyInfoProvider) Concepts() []*gm.ConceptInfo {
	return []*gm.ConceptInfo{
		{
			StepValue: &gm.ProtoStepValue{
				StepValue:              "concept1",
				ParameterizedStepValue: "concept1",
				Parameters:             []string{},
			},
		},
	}
}

func (p *dummyInfoProvider) Params(file string, argType gauge.ArgType) []gauge.StepArg {
	return []gauge.StepArg{{Value: "hello", ArgType: gauge.Static}, {Value: "gauge", ArgType: gauge.Static}}
}

func (p *dummyInfoProvider) SearchConceptDictionary(stepValue string) *gauge.Concept {
	return &(gauge.Concept{FileName: "concept_uri", ConceptStep: &gauge.Step{
		Value:    "concept1",
		LineNo:   1,
		LineText: "concept1",
	}})
}

func (p *dummyInfoProvider) GetConceptDictionary() *gauge.ConceptDictionary {
	return nil
}

func TestCompletion(t *testing.T) {
	f = &files{cache: make(map[string][]string)}
	f.add("uri", " * ")
	position := lsp.Position{Line: 0, Character: len(" * ")}
	want := completionList{IsIncomplete: false, Items: []completionItem{
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "concept1",
				Detail:        "Concept",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: position, End: position}, NewText: `concept1`},
				FilterText:    `concept1`,
				Documentation: "concept1",
			},
			InsertTextFormat: snippet,
		},
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "Say <hello> to <gauge>",
				Detail:        "Step",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: position, End: position}, NewText: `Say "${1:hello}" to "${0:gauge}"`},
				FilterText:    "Say <hello> to <gauge>",
				Documentation: "Say <hello> to <gauge>",
			},
			InsertTextFormat: snippet,
		},
	},
	}
	provider = &dummyInfoProvider{}

	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri"}, Position: position})
	p := json.RawMessage(b)

	got, err := completion(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Expected error == nil in Completion, got %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Autocomplete request failed, got: `%s`, want: `%s`", got, want)
	}
}

func TestCompletionForLineWithText(t *testing.T) {
	f = &files{cache: make(map[string][]string)}
	f.add("uri", " * step")
	position := lsp.Position{Line: 0, Character: len(` *`)}
	wantStartPos := lsp.Position{Line: position.Line, Character: len(` *`)}
	wantEndPos := lsp.Position{Line: position.Line, Character: len(` * step`)}
	want := completionList{IsIncomplete: false, Items: []completionItem{
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "concept1",
				Detail:        "Concept",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: ` concept1`},
				FilterText:    ` concept1`,
				Documentation: "concept1",
			},
			InsertTextFormat: snippet,
		},
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "Say <hello> to <gauge>",
				Detail:        "Step",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: ` Say "${1:hello}" to "${0:gauge}"`},
				FilterText:    " Say <hello> to <gauge>",
				Documentation: "Say <hello> to <gauge>",
			},
			InsertTextFormat: snippet,
		},
	},
	}
	provider = &dummyInfoProvider{}

	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri"}, Position: position})
	p := json.RawMessage(b)

	got, err := completion(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Expected error == nil in Completion, got %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Autocomplete request failed, got: `%+v`, want: `%+v`", got, want)
	}
}

func TestCompletionInBetweenLine(t *testing.T) {
	f = &files{cache: make(map[string][]string)}
	f.add("uri", "* step")
	position := lsp.Position{Line: 0, Character: len(`* s`)}
	wantStartPos := lsp.Position{Line: position.Line, Character: len(`* `)}
	wantEndPos := lsp.Position{Line: position.Line, Character: len(`* step`)}
	want := completionList{IsIncomplete: false, Items: []completionItem{
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "concept1",
				Detail:        "Concept",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: `concept1`},
				FilterText:    `concept1`,
				Documentation: "concept1",
			},
			InsertTextFormat: snippet,
		},
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "Say <hello> to <gauge>",
				Detail:        "Step",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: `Say "${1:hello}" to "${0:gauge}"`},
				FilterText:    "Say <hello> to <gauge>",
				Documentation: "Say <hello> to <gauge>",
			},
			InsertTextFormat: snippet,
		},
	},
	}
	provider = &dummyInfoProvider{}

	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri"}, Position: position})
	p := json.RawMessage(b)

	got, err := completion(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Expected error == nil in Completion, got %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Autocomplete request failed, got: `%s`, want: `%s`", got, want)
	}
}

func TestCompletionInBetweenLineHavingParams(t *testing.T) {
	f = &files{cache: make(map[string][]string)}
	line := "*step with a <param> and more"
	f.add("uri", line)
	position := lsp.Position{Line: 0, Character: len(`*step with a <param> and`)}
	wantStartPos := lsp.Position{Line: position.Line, Character: len(`*`)}
	wantEndPos := lsp.Position{Line: position.Line, Character: len(line)}
	want := completionList{IsIncomplete: false, Items: []completionItem{
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "concept1",
				Detail:        "Concept",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: ` concept1`},
				FilterText:    ` concept1`,
				Documentation: "concept1",
			},
			InsertTextFormat: snippet,
		},
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "Say <hello> to <gauge>",
				Detail:        "Step",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: ` Say "${1:hello}" to "${0:gauge}"`},
				FilterText:    " Say <param> to <gauge>",
				Documentation: "Say <hello> to <gauge>",
			},
			InsertTextFormat: snippet,
		},
	},
	}
	provider = &dummyInfoProvider{}

	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri"}, Position: position})
	p := json.RawMessage(b)

	got, err := completion(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Expected error == nil in Completion, got %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Autocomplete request failed, got: `%+v`, want: `%+v`", got, want)
	}
}

func TestCompletionInBetweenLineHavingSpecialParams(t *testing.T) {
	f = &files{cache: make(map[string][]string)}
	line := "*step with a <file:test.txt> and more"
	f.add("uri", line)
	position := lsp.Position{Line: 0, Character: len(`*step with a <file:test.txt>`)}
	wantStartPos := lsp.Position{Line: position.Line, Character: len(`*`)}
	wantEndPos := lsp.Position{Line: position.Line, Character: len(line)}
	want := completionList{IsIncomplete: false, Items: []completionItem{
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "concept1",
				Detail:        "Concept",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: ` concept1`},
				FilterText:    ` concept1`,
				Documentation: "concept1",
			},
			InsertTextFormat: snippet,
		},
		{
			CompletionItem: lsp.CompletionItem{
				Label:         "Say <hello> to <gauge>",
				Detail:        "Step",
				Kind:          lsp.CIKFunction,
				TextEdit:      lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: ` Say "${1:hello}" to "${0:gauge}"`},
				FilterText:    " Say <file:test.txt> to <gauge>",
				Documentation: "Say <hello> to <gauge>",
			},
			InsertTextFormat: snippet,
		},
	},
	}
	provider = &dummyInfoProvider{}

	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri"}, Position: position})
	p := json.RawMessage(b)

	got, err := completion(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Expected error == nil in Completion, got %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Autocomplete request failed, got: `%+v`, want: `%+v`", got, want)
	}
}

func TestParamCompletion(t *testing.T) {
	f = &files{cache: make(map[string][]string)}
	line := ` * step with a "param`
	f.add("uri", line)
	position := lsp.Position{Line: 0, Character: len(` * step with a "pa`)}
	wantStartPos := lsp.Position{Line: position.Line, Character: len(` * step with a "`)}
	wantEndPos := lsp.Position{Line: position.Line, Character: len(` * step with a "param`)}
	want := completionList{IsIncomplete: false, Items: []completionItem{
		{
			CompletionItem: lsp.CompletionItem{
				Label:      "hello",
				FilterText: "hello\"",
				Detail:     "static",
				Kind:       lsp.CIKVariable,
				TextEdit:   lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: "hello\""},
			},
			InsertTextFormat: text,
		},
		{
			CompletionItem: lsp.CompletionItem{
				Label:      "gauge",
				FilterText: "gauge\"",
				Detail:     "static",
				Kind:       lsp.CIKVariable,
				TextEdit:   lsp.TextEdit{Range: lsp.Range{Start: wantStartPos, End: wantEndPos}, NewText: "gauge\""},
			},
			InsertTextFormat: text,
		},
	},
	}
	provider = &dummyInfoProvider{}

	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri"}, Position: position})
	p := json.RawMessage(b)

	got, err := completion(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Expected error == nil in Completion, got %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Autocomplete request failed, got: `%+v`, want: `%+v`", got, want)
	}
}

func TestCompletionWithError(t *testing.T) {
	p := json.RawMessage("sfdf")
	_, err := completion(&jsonrpc2.Request{Params: &p})

	if err == nil {
		t.Error("Expected error != nil in Completion, got nil")
	}
}

func TestCompletionResolve(t *testing.T) {
	want := completionItem{CompletionItem: lsp.CompletionItem{Label: "step"}}
	b, _ := json.Marshal(want)
	p := json.RawMessage(b)
	got, err := resolveCompletion(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("Expected error == nil in Completion resolve, got %s", err.Error())
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Autocomplete resolve request failed, got: `%s`, want: `%s`", got, want)
	}
}

func TestCompletionResolveWithError(t *testing.T) {
	p := json.RawMessage("sfdf")
	_, err := resolveCompletion(&jsonrpc2.Request{Params: &p})

	if err == nil {
		t.Error("Expected error != nil in Completion, got nil")
	}
}

func TestIsInStepCompletionAtStartOfLine(t *testing.T) {
	if !isStepCompletion("* ", 1) {
		t.Errorf("isStepCompletion not recognizing step context")
	}
}

func TestIsInStepCompletionAtEndOfLine(t *testing.T) {
	if !isStepCompletion("* Step without params", 21) {
		t.Errorf("isStepCompletion not recognizing step context")
	}
}

var paramContextTest = []struct {
	input   string
	charPos int
	want    bool
}{
	{
		input:   `* Step with "static" and <dynamic> params`,
		charPos: len(`* Step with "`),
		want:    true,
	},
	{
		input:   `* Step with "static" and <dynamic> params`,
		charPos: len(`* Step with "static" an`),
		want:    false,
	},
	{
		input:   `* Step with "static" and <dynamic> params`,
		charPos: len(`* Step with "static" and <d`),
		want:    true,
	},
}

func TestIsInParamContext(t *testing.T) {
	for _, test := range paramContextTest {
		got := inParameterContext(test.input, test.charPos)
		if test.want != got {
			t.Errorf("got : %s, want : %s", got, test.want)
		}
	}
}
