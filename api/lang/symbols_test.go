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

	want := []string{
		"# Specification 1",
		"# Specification 2",
	}

	b, _ := json.Marshal(lsp.WorkspaceSymbolParams{Limit: 5, Query: "Spec"})
	p := json.RawMessage(b)
	got, err := workspaceSymbols(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := mapName(got.([]*lsp.SymbolInformation))

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

	want := []string{
		"# Sample 1",
		"# Sample 2",
		"## Sample Scenario 1",
		"## Sample Scenario 2",
		"## Sample Scenario 5",
		"## Sample Scenario 6",
	}

	b, _ := json.Marshal(lsp.WorkspaceSymbolParams{Limit: 5, Query: "Sample"})
	p := json.RawMessage(b)
	got, err := workspaceSymbols(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := mapName(got.([]*lsp.SymbolInformation))

	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}
}

func TestWorkspaceSymbolsEmptyWhenLessThanTwoCharsGiven(t *testing.T) {
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

	b, _ := json.Marshal(lsp.WorkspaceSymbolParams{Limit: 5, Query: "S"})
	p := json.RawMessage(b)
	got, err := workspaceSymbols(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	if got != nil {
		t.Errorf("expected %v to be nil", got)
	}
}

func TestWorkspaceSymbolsSortsAndGroupsBySpecsAndScenarios(t *testing.T) {
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
								Heading: &gauge.Heading{Value: "Scenario Sample 2", LineNo: 20},
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
								Heading: &gauge.Heading{Value: "Scenario Sample 5", LineNo: 10},
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

	want := []string{
		"# Sample 1",
		"# Sample 2",
		"## Sample Scenario 1",
		"## Sample Scenario 6",
		"## Scenario Sample 2",
		"## Scenario Sample 5",
	}

	b, _ := json.Marshal(lsp.WorkspaceSymbolParams{Limit: 5, Query: "Sample"})
	p := json.RawMessage(b)
	got, err := workspaceSymbols(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := mapName(got.([]*lsp.SymbolInformation))

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

	uri := util.ConvertPathToURI("foo.spec")
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)
	b, _ := json.Marshal(lsp.DocumentSymbolParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}})
	p := json.RawMessage(b)

	got, err := documentSymbols(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected errror to be nil. Got: \n%v", err.Error())
	}

	info := mapName(got.([]*lsp.SymbolInformation))

	want := []string{
		"# Specification Heading",
		"## Scenario Heading",
		"## Scenario Heading2",
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}

	f.remove(uri)
}

func TestDocumentSymbolsForConcept(t *testing.T) {
	provider = &dummyInfoProvider{}
	cptText := `
	# Concept 1
	
	* foo
	* bar
	
	Concept 2 <param1>
	==================
	
	* baz
	`

	uri := util.ConvertPathToURI("foo.cpt")
	f = &files{cache: make(map[string][]string)}
	f.add(uri, cptText)
	b, _ := json.Marshal(lsp.DocumentSymbolParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}})
	p := json.RawMessage(b)

	got, err := documentSymbols(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected errror to be nil. Got: \n%v", err.Error())
	}

	info := mapName(got.([]*lsp.SymbolInformation))

	want := []string{
		"# Concept 1",
		"# Concept 2 <param1>",
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}

	f.remove(uri)
}

func TestGetSpecSymbol(t *testing.T) {
	spec := &gauge.Specification{
		Heading:  &gauge.Heading{Value: "Sample 1", LineNo: 1},
		FileName: "foo1.spec",
	}

	want := &lsp.SymbolInformation{
		Name: "# Sample 1",
		Kind: lsp.SKNamespace,
		Location: lsp.Location{
			URI: util.ConvertPathToURI("foo1.spec"),
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: 0, Character: len("Sample 1")},
			},
		},
	}

	got := getSpecSymbol(spec)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v to be equal %v", got, want)
	}
}

func TestGetScenarioSymbol(t *testing.T) {
	scenario := &gauge.Scenario{
		Heading: &gauge.Heading{Value: "Sample Scenario 5", LineNo: 10},
	}

	want := &lsp.SymbolInformation{
		Name: "## Sample Scenario 5",
		Kind: lsp.SKNamespace,
		Location: lsp.Location{
			URI: util.ConvertPathToURI("foo.spec"),
			Range: lsp.Range{
				Start: lsp.Position{Line: 9, Character: 0},
				End:   lsp.Position{Line: 9, Character: len("Scenario Heading2")},
			},
		},
	}

	got := getScenarioSymbol(scenario, "foo.spec")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v to be equal %v", got, want)
	}
}

func TestGetConceptSymbols(t *testing.T) {
	conceptText := `
	# Concept 1
	
	* foo
	* bar
	
	Concept 2 <param1>
	==================
	
	* baz
	`
	want := []*lsp.SymbolInformation{
		{
			Name: "# Concept 1",
			Kind: lsp.SKNamespace,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo.cpt"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 1, Character: 0},
					End:   lsp.Position{Line: 1, Character: len("Concept 1")},
				},
			},
		},
		{
			Name: "# Concept 2 <param1>",
			Kind: lsp.SKNamespace,
			Location: lsp.Location{
				URI: util.ConvertPathToURI("foo.cpt"),
				Range: lsp.Range{
					Start: lsp.Position{Line: 6, Character: 0},
					End:   lsp.Position{Line: 6, Character: len("Concept 2 <param1>")},
				},
			},
		},
	}
	got := getConceptSymbols(conceptText, "foo.cpt")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v to be equal %v", got, want)
	}
}

func mapName(vs []*lsp.SymbolInformation) []string {
	val := make([]string, len(vs))
	for i, v := range vs {
		val[i] = v.Name
	}
	return val
}
