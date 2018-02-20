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

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestGetCodeActionForUnimplementedStep(t *testing.T) {

	stub := "a stub for unimplemented step"
	d := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{3, 0},
				End:   lsp.Position{0, 10},
			},
			Message:  "Step implantation not found",
			Severity: 1,
			Code:     stub,
		},
	}

	b, _ := json.Marshal(lsp.CodeActionParams{TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"}, Context: lsp.CodeActionContext{Diagnostics: d}})
	p := json.RawMessage(b)

	want := []lsp.Command{
		{
			Command:   generateStubCommand,
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

func TestGetCodeActionForExtractConcept(t *testing.T) {
	selection := lsp.Range{
		Start: lsp.Position{Line: 3, Character: 0},
		End:   lsp.Position{Line: 5, Character: 0},
	}
	params := lsp.CodeActionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"},
		Range:        selection,
		Context:      lsp.CodeActionContext{Diagnostics: []lsp.Diagnostic{}},
	}
	b, _ := json.Marshal(params)
	p := json.RawMessage(b)

	want := []lsp.Command{
		{
			Command: extractConceptCommand,
			Title:   extractConceptTitle,
			Arguments: []interface{}{extractConceptInfo{
				Uri:   "foo.spec",
				Range: selection,
			}},
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

func TestGetCodeActionForExtractConceptShouldNotComeIfSelectionHasErrors(t *testing.T) {
	selection := lsp.Range{
		Start: lsp.Position{Line: 3, Character: 0},
		End:   lsp.Position{Line: 5, Character: 0},
	}
	params := lsp.CodeActionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"},
		Range:        selection,
		Context: lsp.CodeActionContext{Diagnostics: []lsp.Diagnostic{
			lsp.Diagnostic{
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      3,
						Character: 1,
					},
					End: lsp.Position{
						Line:      3,
						Character: 10,
					},
				},
			},
		}},
	}
	b, _ := json.Marshal(params)
	p := json.RawMessage(b)

	got, err := codeActions(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. \nGot : %s", err)
	}

	if len(got.([]lsp.Command)) > 0 {
		t.Errorf("want: `%s` to be empty but got `%s`", got, got)
	}
}
