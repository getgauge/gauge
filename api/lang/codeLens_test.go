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
	"testing"

	"encoding/json"
	"reflect"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestGetCodeLens(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text`

	f = &files{cache: make(map[string][]string)}
	f.add("foo.spec", specText)

	b, _ := json.Marshal(lsp.CodeLensParams{TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"}})
	p := json.RawMessage(b)

	got, err := codeLenses(&jsonrpc2.Request{Params: &p})
	if err != nil {
		t.Errorf("Expected error to be nil. got : %s", err.Error())
	}

	specCodeLens := lsp.CodeLens{
		Command: lsp.Command{
			Command:   "gauge.execute",
			Title:     "Run Spec",
			Arguments: getExecutionArgs("foo.spec"),
		},
		Range: lsp.Range{
			Start: lsp.Position{0, 0},
			End:   lsp.Position{0, 8},
		},
	}

	scenCodeLens := lsp.CodeLens{
		Command: lsp.Command{
			Command:   "gauge.execute",
			Title:     "Run Scenario",
			Arguments: getExecutionArgs("foo.spec:4"),
		},
		Range: lsp.Range{
			Start: lsp.Position{3, 0},
			End:   lsp.Position{3, 12},
		},
	}
	want := []lsp.CodeLens{scenCodeLens, specCodeLens}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}

func TestGetCodeLensWithMultipleScenario(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text

Another Scenario
----------------

* another step
`

	f = &files{cache: make(map[string][]string)}
	f.add("foo.spec", specText)

	b, _ := json.Marshal(lsp.CodeLensParams{TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"}})
	p := json.RawMessage(b)
	got, err := codeLenses(&jsonrpc2.Request{Params: &p})
	if err != nil {
		t.Errorf("Expected error to be nil. got : %s", err.Error())
	}

	specCodeLens := lsp.CodeLens{
		Command: lsp.Command{
			Command:   "gauge.execute",
			Title:     "Run Spec",
			Arguments: getExecutionArgs("foo.spec"),
		},
		Range: lsp.Range{
			Start: lsp.Position{0, 0},
			End:   lsp.Position{0, 8},
		},
	}

	scenCodeLens1 := lsp.CodeLens{
		Command: lsp.Command{
			Command:   "gauge.execute",
			Title:     "Run Scenario",
			Arguments: getExecutionArgs("foo.spec:4"),
		},
		Range: lsp.Range{
			Start: lsp.Position{3, 0},
			End:   lsp.Position{3, 12},
		},
	}

	scenCodeLens2 := lsp.CodeLens{
		Command: lsp.Command{
			Command:   "gauge.execute",
			Title:     "Run Scenario",
			Arguments: getExecutionArgs("foo.spec:9"),
		},
		Range: lsp.Range{
			Start: lsp.Position{8, 0},
			End:   lsp.Position{8, 12},
		},
	}

	want := []lsp.CodeLens{scenCodeLens1, scenCodeLens2, specCodeLens}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}

func TestGetCodeLensWithDataTable(t *testing.T) {
	specText := `Specification Heading
=====================

	|Word  |Vowel Count|
	|------|-----------|
	|Mingle|2          |
	|Snap  |1          |
	|GoCD  |1          |
	|Rhythm|0          |


Scenario Heading
----------------

* The word <Word> has <Vowel Count> vowels.

`

	f = &files{cache: make(map[string][]string)}
	f.add("foo.spec", specText)

	b, _ := json.Marshal(lsp.CodeLensParams{TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"}})
	p := json.RawMessage(b)

	got, err := codeLenses(&jsonrpc2.Request{Params: &p})
	if err != nil {
		t.Errorf("Expected error to be nil. got : %s", err.Error())
	}

	specCodeLens := lsp.CodeLens{
		Command: lsp.Command{
			Command:   "gauge.execute",
			Title:     "Run Spec",
			Arguments: getExecutionArgs("foo.spec"),
		},
		Range: lsp.Range{
			Start: lsp.Position{0, 0},
			End:   lsp.Position{0, 8},
		},
	}

	specCodeLens2 := lsp.CodeLens{
		Command: lsp.Command{
			Command:   "gauge.execute.inParallel",
			Title:     "Run in parallel",
			Arguments: getExecutionArgs("foo.spec"),
		},
		Range: lsp.Range{
			Start: lsp.Position{0, 0},
			End:   lsp.Position{0, 15},
		},
	}

	scenCodeLens2 := lsp.CodeLens{
		Command: lsp.Command{
			Command:   "gauge.execute",
			Title:     "Run Scenario",
			Arguments: getExecutionArgs("foo.spec:12"),
		},
		Range: lsp.Range{
			Start: lsp.Position{11, 0},
			End:   lsp.Position{11, 12},
		},
	}

	want := []lsp.CodeLens{scenCodeLens2, specCodeLens, specCodeLens2}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}
