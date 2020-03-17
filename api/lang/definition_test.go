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
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/runner"
	"testing"
	"time"

	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestConceptDefinitionInSpecFile(t *testing.T) {
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	uri := lsp.DocumentURI(util.ConvertPathToURI("uri.spec"))
	openFilesCache.add(uri, "# Specification \n## Scenario \n * concept1")

	conUri := lsp.DocumentURI(util.ConvertPathToURI("concept_uri.cpt"))
	openFilesCache.add(conUri, "# Concept \n* a step \n \n # Another Concept \n*concept1")

	provider = &dummyInfoProvider{}
	position := lsp.Position{Line: 2, Character: len(" * conce")}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)

	got, err := definition(&jsonrpc2.Request{Params: &p})
	if err != nil {
		t.Errorf("Failed to find definition, err: `%v`", err)
	}

	want := lsp.Location{URI: conUri, Range: lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 10}}}
	if got != want {
		t.Errorf("Wrong definition found, got: `%v`, want: `%v`", got, want)
	}
}

func TestConceptDefinitionInConceptFile(t *testing.T) {
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	uri := lsp.DocumentURI(util.ConvertPathToURI("concept_uri.cpt"))
	openFilesCache.add(uri, "# Concept \n* a step \n \n # Another Concept \n*concept1")
	provider = &dummyInfoProvider{}
	position := lsp.Position{Line: 4, Character: len("*conce")}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)

	got, err := definition(&jsonrpc2.Request{Params: &p})
	if err != nil {
		t.Errorf("Failed to find definition, err: `%v`", err)
	}
	want := lsp.Location{URI: uri, Range: lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 10}}}
	if got != want {
		t.Errorf("Wrong definition found, got: `%v`, want: `%v`", got, want)
	}
}

func TestExternalStepDefinition(t *testing.T) {
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	uri := lsp.DocumentURI(util.ConvertPathToURI("spec_uri.spec"))
	openFilesCache.add(uri, "# Specification \n\n## Scenario\n\n* a step")
	provider = &dummyInfoProvider{}
	position := lsp.Position{Line: 4, Character: len("* a step")}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)
	responses := map[gm.Message_MessageType]interface{}{}
	responses[gm.Message_StepNameResponse] = &gm.StepNameResponse{
		HasAlias:      false,
		IsExternal:    true,
		IsStepPresent: true,
	}

	lRunner.runner = &runner.GrpcRunner{LegacyClient: &mockClient{responses: responses}, Timeout: time.Second * 30}
	_, err := definition(&jsonrpc2.Request{Params: &p})
	if err == nil {
		t.Errorf("expected error to not be nil.")
	}
	expected := `implementation source not found: Step implementation referred from an external project or library`
	if err.Error() != expected {
		t.Errorf("Expected: `%s`\nGot: `%s`", expected, err.Error())
	}
}
