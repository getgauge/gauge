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

type dummyCompletionProvider struct{}

func (p *dummyCompletionProvider) Init() {}
func (p *dummyCompletionProvider) Steps() []*gauge.StepValue {
	return []*gauge.StepValue{{
		Args:                   []string{"hello", "gauge"},
		StepValue:              "Say {} to {}",
		ParameterizedStepValue: "Say <hello> to <gauge>",
	}}
}
func (p *dummyCompletionProvider) Concepts() []*gm.ConceptInfo {
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

func TestCompletion(t *testing.T) {
	want := completionList{IsIncomplete: false, Items: []completionItem{
		{
			CompletionItem: lsp.CompletionItem{
				Label:      "concept1",
				Detail:     "Concept",
				Kind:       lsp.CIKFunction,
				TextEdit:   lsp.TextEdit{Range: lsp.Range{}, NewText: `concept1`},
				FilterText: `concept1`,
			},
			InsertTextFormat: snippet,
		},
		{
			CompletionItem: lsp.CompletionItem{
				Label:      "Say <hello> to <gauge>",
				Detail:     "Step",
				Kind:       lsp.CIKFunction,
				TextEdit:   lsp.TextEdit{Range: lsp.Range{}, NewText: `Say "${1:hello}" to "${0:gauge}"`},
				FilterText: "Say <hello> to <gauge>",
			},
			InsertTextFormat: snippet,
		},
	},
	}
	provider = &dummyCompletionProvider{}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{})
	p := json.RawMessage(b)

	got, err := completion(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("Expected error == nil in Completion, got %s", err.Error())
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Autocomplete request failed, got: `%s`, want: `%s`", got, want)
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

func TestGetPrefix(t *testing.T) {
	want := " "
	f = &files{cache: make(map[string][]string)}
	f.add("uri", "line1\n*")
	params := lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri"}, Position: lsp.Position{Line: 1, Character: 1}}
	got := getPrefix(params)

	if got != want {
		t.Errorf("GetPrefix failed for autocomplete, want: `%s`, got: `%s`", want, got)
	}
}

func TestGetPrefixWithSpace(t *testing.T) {
	want := ""
	f = &files{cache: make(map[string][]string)}
	f.add("uri", "* ")
	params := lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri"}, Position: lsp.Position{Line: 0, Character: 2}}
	got := getPrefix(params)

	if got != want {
		t.Errorf("GetPrefix failed for autocomplete, want: `%s`, got: `%s`", want, got)
	}
}

func TestGetPrefixWithNoCharsInLine(t *testing.T) {
	want := ""
	f = &files{cache: make(map[string][]string)}
	params := lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri"}, Position: lsp.Position{Line: 1, Character: 0}}
	got := getPrefix(params)

	if got != want {
		t.Errorf("GetPrefix failed for autocomplete, want: `%s`, got: `%s`", want, got)
	}
}
