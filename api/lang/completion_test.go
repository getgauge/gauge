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
	f = &files{cache: make(map[string][]string)}
	f.add("uri", " * ")
	position := lsp.Position{Line: 0, Character: 3}
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
	provider = &dummyCompletionProvider{}

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
	position := lsp.Position{Line: 0, Character: 2}
	wantStartPos := lsp.Position{Line: position.Line, Character: 2}
	wantEndPos := lsp.Position{Line: position.Line, Character: 7}
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
	provider = &dummyCompletionProvider{}

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
	position := lsp.Position{Line: 0, Character: 5}
	wantStartPos := lsp.Position{Line: position.Line, Character: 2}
	wantEndPos := lsp.Position{Line: position.Line, Character: 6}
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
	provider = &dummyCompletionProvider{}

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
	f.add("uri", "*step with a <param> and more")
	position := lsp.Position{Line: 0, Character: 25}
	wantStartPos := lsp.Position{Line: position.Line, Character: 1}
	wantEndPos := lsp.Position{Line: position.Line, Character: 29}
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
	provider = &dummyCompletionProvider{}

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
	f.add("uri", "*step with a <file:test.txt> and more")
	position := lsp.Position{Line: 0, Character: 30}
	wantStartPos := lsp.Position{Line: position.Line, Character: 1}
	wantEndPos := lsp.Position{Line: position.Line, Character: 37}
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
	provider = &dummyCompletionProvider{}

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

func TestGetPrefix(t *testing.T) {
	want := " "
	got := getPrefix("line1\n*")

	if got != want {
		t.Errorf("GetPrefix failed for autocomplete, want: `%s`, got: `%s`", want, got)
	}
}

func TestGetPrefixWithSpace(t *testing.T) {
	want := ""
	got := getPrefix("* ")

	if got != want {
		t.Errorf("GetPrefix failed for autocomplete, want: `%s`, got: `%s`", want, got)
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

func TestIsInStepCompletionWithParams(t *testing.T) {
	if isStepCompletion(`* Step with "static" and <dynamic> params`, 13) {
		t.Errorf("isStepCompletion not recognizing step context")
	}
	if !isStepCompletion(`* Step with "static" and <dynamic> params`, 24) {
		t.Errorf("isStepCompletion not recognizing step context")
	}
	if isStepCompletion(`* Step with "static" and <dynamic> params`, 28) {
		t.Errorf("isStepCompletion not recognizing step context")
	}
}

func TestGetFilterTextWithStaticParam(t *testing.T) {
	got := getFilterText("Text with {}", []string{"param1"}, []gauge.StepArg{{Name: "Args1", Value: "Args1", ArgType: gauge.Static}})
	want := `Text with "Args1"`
	if got != want {
		t.Errorf("The parameters are not replaced correctly")
	}
}

func TestGetFilterTextWithDynamicParam(t *testing.T) {
	got := getFilterText("Text with {}", []string{"param1"}, []gauge.StepArg{{Name: "Args1", Value: "Args1", ArgType: gauge.Dynamic}})
	want := `Text with <Args1>`
	if got != want {
		t.Errorf("The parameters are not replaced correctly")
	}
}

func TestGetFilterTextShouldNotReplaceIfNoStepArgsGiven(t *testing.T) {
	got := getFilterText("Text with {}", []string{"param1"}, []gauge.StepArg{})
	want := `Text with <param1>`
	if got != want {
		t.Errorf("The parameters are not replaced correctly")
	}
}

func TestGetFilterTextWithLesserNumberOfStepArgsGiven(t *testing.T) {
	stepArgs := []gauge.StepArg{
		{Name: "Args1", Value: "Args1", ArgType: gauge.Dynamic},
		{Name: "Args2", Value: "Args2", ArgType: gauge.Static},
	}
	got := getFilterText("Text with {} {} and {}", []string{"param1", "param2", "param3"}, stepArgs)
	want := `Text with <Args1> "Args2" and <param3>`
	if got != want {
		t.Errorf("The parameters are not replaced correctly")
	}
}

var testEditPosition = []struct {
	input     string
	cursorPos lsp.Position
	wantStart lsp.Position
	wantEnd   lsp.Position
}{
	{
		input:     "*",
		cursorPos: lsp.Position{Line: 0, Character: 1},
		wantStart: lsp.Position{Line: 0, Character: 1},
		wantEnd:   lsp.Position{Line: 0, Character: 1},
	},
	{
		input:     "* ",
		cursorPos: lsp.Position{Line: 0, Character: 1},
		wantStart: lsp.Position{Line: 0, Character: 1},
		wantEnd:   lsp.Position{Line: 0, Character: 2},
	},
	{
		input:     "* Step",
		cursorPos: lsp.Position{Line: 10, Character: 1},
		wantStart: lsp.Position{Line: 10, Character: 1},
		wantEnd:   lsp.Position{Line: 10, Character: 6},
	},
	{
		input:     "* Step",
		cursorPos: lsp.Position{Line: 0, Character: 2},
		wantStart: lsp.Position{Line: 0, Character: 2},
		wantEnd:   lsp.Position{Line: 0, Character: 6},
	},
	{
		input:     "* Step",
		cursorPos: lsp.Position{Line: 0, Character: 4},
		wantStart: lsp.Position{Line: 0, Character: 2},
		wantEnd:   lsp.Position{Line: 0, Character: 6},
	},
	{
		input:     "    * Step",
		cursorPos: lsp.Position{Line: 0, Character: 7},
		wantStart: lsp.Position{Line: 0, Character: 6},
		wantEnd:   lsp.Position{Line: 0, Character: 10},
	},
	{
		input:     " * Step ",
		cursorPos: lsp.Position{Line: 0, Character: 10},
		wantStart: lsp.Position{Line: 0, Character: 3},
		wantEnd:   lsp.Position{Line: 0, Character: 10},
	},
}

func TestGetEditPosition(t *testing.T) {
	for _, test := range testEditPosition {
		gotStart, gotEnd := getEditPosition(test.input, test.cursorPos)
		if gotStart.Line != test.wantStart.Line || gotStart.Character != test.wantStart.Character {
			t.Errorf(`Incorrect Edit Start Position got: %+v , want : %+v, input : "%s"`, gotStart, test.wantStart, test.input)
		}
		if gotEnd.Line != test.wantEnd.Line || gotEnd.Character != test.wantEnd.Character {
			t.Errorf(`Incorrect Edit End Position got: %+v , want : %+v, input : "%s"`, gotEnd, test.wantEnd, test.input)
		}
	}
}
