package lang

import (
	"encoding/json"
	"testing"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"reflect"
)

func Test_getScenariosShouldGiveTheScenarioAtCurrentCursorPosition(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text

Scenario Heading2
-----------------

* Step text`

	uri := "foo.spec"
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	position := lsp.Position{Line: 5, Character: 1}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)

	got, err := getScenarios(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected errror to be nil. Got: \n%v",err.Error())
	}

	info := got.(ScenarioInfo)

	want := ScenarioInfo{
		Heading:"Scenario Heading",
		LineNo:4,
		ExecutionIdentifier:"foo.spec:4",
	}
	if ! reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v",info, want)
	}
}


func Test_getScenariosShouldGiveTheScenariosIfCursorPositionIsNotInSpan(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text

Scenario Heading2
-----------------

* Step text
`

	uri := "foo.spec"
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	position := lsp.Position{Line: 2, Character: 1}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)

	got, err := getScenarios(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected errror to be nil. Got: \n%v",err.Error())
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
		t.Errorf("expected %v to be equal %v",info, want)
	}
}
