/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"encoding/json"
	"testing"

	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/gauge"

	"reflect"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestGetScenariosShouldGiveTheScenarioAtCurrentCursorPosition(t *testing.T) {
	provider = &dummyInfoProvider{}
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text

Scenario Heading2
-----------------

* Step text`

	uri := lsp.DocumentURI("foo.spec")
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(uri, specText)

	position := lsp.Position{Line: 5, Character: 1}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)

	got, err := scenarios(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected errror to be nil. Got: \n%v", err.Error())
	}

	info := got.(ScenarioInfo)

	want := ScenarioInfo{
		Heading:             "Scenario Heading",
		LineNo:              4,
		ExecutionIdentifier: "foo.spec:4",
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}
	openFilesCache.remove(uri)
}

func TestGetScenariosShouldGiveTheScenariosIfCursorPositionIsNotInSpan(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text

Scenario Heading2
-----------------

* Step text
`

	uri := lsp.DocumentURI("foo.spec")
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(uri, specText)

	position := lsp.Position{Line: 2, Character: 1}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)

	got, err := scenarios(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected errror to be nil. Got: \n%v", err.Error())
	}

	info := got.([]ScenarioInfo)

	want := []ScenarioInfo{
		{
			Heading:             "Scenario Heading",
			LineNo:              4,
			ExecutionIdentifier: "foo.spec:4",
		},
		{
			Heading:             "Scenario Heading2",
			LineNo:              9,
			ExecutionIdentifier: "foo.spec:9",
		},
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}
	openFilesCache.remove(uri)
}

func TestGetScenariosShouldGiveTheScenariosIfDocumentIsNotOpened(t *testing.T) {
	provider = &dummyInfoProvider{
		specsFunc: func(specs []string) []*infoGatherer.SpecDetail {
			return []*infoGatherer.SpecDetail{
				&infoGatherer.SpecDetail{
					Spec: &gauge.Specification{
						Heading:  &gauge.Heading{Value: "Specification 1"},
						FileName: "foo.spec",
						Scenarios: []*gauge.Scenario{
							&gauge.Scenario{Heading: &gauge.Heading{Value: "Scenario 1", LineNo: 4}, Span: &gauge.Span{Start: 4, End: 7}},
							&gauge.Scenario{Heading: &gauge.Heading{Value: "Scenario 2", LineNo: 9}, Span: &gauge.Span{Start: 9, End: 12}},
						},
					},
				},
			}
		},
	}

	position := lsp.Position{Line: 2, Character: 1}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"}, Position: position})
	p := json.RawMessage(b)

	got, err := scenarios(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := got.([]ScenarioInfo)

	want := []ScenarioInfo{
		{
			Heading:             "Scenario 1",
			LineNo:              4,
			ExecutionIdentifier: "foo.spec:4",
		},
		{
			Heading:             "Scenario 2",
			LineNo:              9,
			ExecutionIdentifier: "foo.spec:9",
		},
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}
}

func TestGetSpecsShouldReturnAllSpecsInDirectory(t *testing.T) {
	provider = &dummyInfoProvider{
		specsFunc: func(specs []string) []*infoGatherer.SpecDetail {
			return []*infoGatherer.SpecDetail{
				&infoGatherer.SpecDetail{
					Spec: &gauge.Specification{
						Heading:  &gauge.Heading{Value: "Specification 1"},
						FileName: "foo1.spec",
					},
				},
				&infoGatherer.SpecDetail{
					Spec: &gauge.Specification{
						Heading:  &gauge.Heading{Value: "Specification 2"},
						FileName: "foo2.spec",
					},
				},
			}
		},
	}

	want := []specInfo{
		{
			Heading:             "Specification 1",
			ExecutionIdentifier: "foo1.spec",
		},
		{
			Heading:             "Specification 2",
			ExecutionIdentifier: "foo2.spec",
		},
	}
	got, err := specs()

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := got.([]specInfo)

	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}
}
