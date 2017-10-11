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
	"testing"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestConceptDefinitionInSpecFile(t *testing.T) {
	f = &files{cache: make(map[string][]string)}
	f.add("uri.spec", "# Specification \n## Scenario \n * concept1")
	provider = &dummyInfoProvider{}
	position := lsp.Position{Line: 2, Character: len(" * conce")}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri.spec"}, Position: position})
	p := json.RawMessage(b)

	got, err := definition(&jsonrpc2.Request{Params: &p})
	if err != nil {
		t.Errorf("Failed to find definition, err: `%v`", err)
	}

	want := lsp.Location{URI: "file://concept_uri", Range: lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 0}}}
	if got != want {
		t.Errorf("Wrong definition found, got: `%v`, want: `%v`", got, want)
	}
}

func TestConceptDefinitionInConceptFile(t *testing.T) {
	f = &files{cache: make(map[string][]string)}
	f.add("uri.cpt", "# Concept \n* a step \n \n # Another Concept \n*concept1")
	provider = &dummyInfoProvider{}
	position := lsp.Position{Line: 4, Character: len("*conce")}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "uri.cpt"}, Position: position})
	p := json.RawMessage(b)

	got, err := definition(&jsonrpc2.Request{Params: &p})
	if err != nil {
		t.Errorf("Failed to find definition, err: `%v`", err)
	}

	want := lsp.Location{URI: "file://concept_uri", Range: lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 0}}}
	if got != want {
		t.Errorf("Wrong definition found, got: `%v`, want: `%v`", got, want)
	}
}
