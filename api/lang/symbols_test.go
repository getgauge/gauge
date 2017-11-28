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
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"reflect"
	"testing"
)

func TestWorkspaceSymbolsGetsFromAllSpecs(t *testing.T) {
	provider = &dummyInfoProvider{
		specsFunc: func(specs []string) []*infoGatherer.SpecDetail {
			return []*infoGatherer.SpecDetail{
				&infoGatherer.SpecDetail{
					Spec: &gauge.Specification{
						Heading:  &gauge.Heading{Value: "Specification 1", LineNo: 1},
						FileName: "foo1.spec",
					},
				},
				&infoGatherer.SpecDetail{
					Spec: &gauge.Specification{
						Heading:  &gauge.Heading{Value: "Specification 2", LineNo: 2},
						FileName: "foo2.spec",
					},
				},
			}
		},
	}

	want := []lsp.SymbolInformation{
		{
			ContainerName: "foo1.spec",
			Name:          "Specification 1",
			Kind:          lsp.SKClass,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo1.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 1, Character: 0},
					End:   lsp.Position{Line: 1, Character: len("Specification 1")},
				},
			},
		},
		{
			ContainerName: "foo2.spec",
			Name:          "Specification 2",
			Kind:          lsp.SKClass,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo2.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 2, Character: 0},
					End:   lsp.Position{Line: 2, Character: len("Specification 2")},
				},
			},
		},
	}
	b, _ := json.Marshal(lsp.WorkspaceSymbolParams{Limit: 5, Query: "Spec"})
	p := json.RawMessage(b)

	got, err := workspaceSymbols(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := got.([]lsp.SymbolInformation)

	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}
}

func TestWorkspaceSymbolsGetsFromScenarios(t *testing.T) {
	provider = &dummyInfoProvider{
		specsFunc: func(specs []string) []*infoGatherer.SpecDetail {
			return []*infoGatherer.SpecDetail{
				&infoGatherer.SpecDetail{
					Spec: &gauge.Specification{
						Heading:  &gauge.Heading{Value: "Sample 1", LineNo: 1},
						FileName: "foo1.spec",
						Scenarios: []*gauge.Scenario{
							{
								Heading: &gauge.Heading{Value: "Sample Scenario 1", LineNo: 10},
							},
							{
								Heading: &gauge.Heading{Value: "Sample Scenario 2", LineNo: 20},
							},
							{
								Heading: &gauge.Heading{Value: "Random Scenario 1", LineNo: 30},
							},
						},
					},
				},
				&infoGatherer.SpecDetail{
					Spec: &gauge.Specification{
						Heading:  &gauge.Heading{Value: "Sample 2", LineNo: 2},
						FileName: "foo2.spec",
						Scenarios: []*gauge.Scenario{
							{
								Heading: &gauge.Heading{Value: "Sample Scenario 5", LineNo: 10},
							},
							{
								Heading: &gauge.Heading{Value: "Sample Scenario 6", LineNo: 20},
							},
							{
								Heading: &gauge.Heading{Value: "Random Scenario 9", LineNo: 30},
							},
						},
					},
				},
			}
		},
	}

	want := []lsp.SymbolInformation{
		{
			ContainerName: "foo1.spec",
			Name:          "Sample 1",
			Kind:          lsp.SKClass,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo1.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 1, Character: 0},
					End:   lsp.Position{Line: 1, Character: len("Sample 1")},
				},
			},
		},
		{
			ContainerName: "foo1.spec",
			Name:          "Sample Scenario 1",
			Kind:          lsp.SKFunction,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo1.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 10, Character: 0},
					End:   lsp.Position{Line: 10, Character: len("Sample Scenario 1")},
				},
			},
		},
		{
			ContainerName: "foo1.spec",
			Name:          "Sample Scenario 2",
			Kind:          lsp.SKFunction,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo1.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 20, Character: 0},
					End:   lsp.Position{Line: 20, Character: len("Sample Scenario 2")},
				},
			},
		},
		{
			ContainerName: "foo2.spec",
			Name:          "Sample 2",
			Kind:          lsp.SKClass,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo2.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 2, Character: 0},
					End:   lsp.Position{Line: 2, Character: len("Sample 2")},
				},
			},
		},
		{
			ContainerName: "foo2.spec",
			Name:          "Sample Scenario 5",
			Kind:          lsp.SKFunction,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo2.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 10, Character: 0},
					End:   lsp.Position{Line: 10, Character: len("Sample Scenario 5")},
				},
			},
		},
		{
			ContainerName: "foo2.spec",
			Name:          "Sample Scenario 6",
			Kind:          lsp.SKFunction,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo2.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 20, Character: 0},
					End:   lsp.Position{Line: 20, Character: len("Sample Scenario 6")},
				},
			},
		},
	}
	b, _ := json.Marshal(lsp.WorkspaceSymbolParams{Limit: 5, Query: "Sample"})
	p := json.RawMessage(b)

	got, err := workspaceSymbols(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := got.([]lsp.SymbolInformation)

	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}
}

func TestDocumentSymbols(t *testing.T) {
	provider = &dummyInfoProvider{}
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text

Scenario Heading2
-----------------

* Step text`

	uri := "file:///foo.spec"
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)
	b, _ := json.Marshal(lsp.DocumentSymbolParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}})
	p := json.RawMessage(b)

	got, err := documentSymbols(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected errror to be nil. Got: \n%v", err.Error())
	}

	info := got.([]lsp.SymbolInformation)

	want := []lsp.SymbolInformation{
		{
			ContainerName: "foo.spec",
			Name:          "Specification Heading",
			Kind:          lsp.SKClass,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 1, Character: 0},
					End:   lsp.Position{Line: 1, Character: len("Specification Heading")},
				},
			},
		},
		{
			ContainerName: "foo.spec",
			Name:          "Scenario Heading",
			Kind:          lsp.SKFunction,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 4, Character: 0},
					End:   lsp.Position{Line: 4, Character: len("Scenario Heading")},
				},
			},
		},
		{
			ContainerName: "foo.spec",
			Name:          "Scenario Heading2",
			Kind:          lsp.SKFunction,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo.spec"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 9, Character: 0},
					End:   lsp.Position{Line: 9, Character: len("Scenario Heading2")},
				},
			},
		},
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}

	f.remove(uri)
}
