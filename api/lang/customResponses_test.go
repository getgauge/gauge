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
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"

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

func TestGetImplementationFilesShouldReturnFilePaths(t *testing.T) {
	var params = struct {
		Concept bool
	}{}

	b, _ := json.Marshal(params)
	p := json.RawMessage(b)

	GetResponseFromRunner = func(m *gauge_messages.Message) (*gauge_messages.Message, error) {
		response := &gauge_messages.Message{
			MessageType: gauge_messages.Message_ImplementationFileListResponse,
			ImplementationFileListResponse: &gauge_messages.ImplementationFileListResponse{
				ImplementationFilePaths: []string{"file"},
			},
		}
		return response, nil
	}
	implFiles, err := getImplFiles(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	want := []string{"file"}

	if !reflect.DeepEqual(implFiles, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, implFiles)
	}
}

func TestGetImplementationFilesShouldReturnFilePathsForConcept(t *testing.T) {
	type implFileParam struct {
		Concept bool `json:"concept"`
	}

	params := implFileParam{Concept: true}

	b, _ := json.Marshal(params)
	p := json.RawMessage(b)

	util.GetConceptFiles = func() []string {
		return []string{"file.cpt"}
	}

	implFiles, err := getImplFiles(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	want := []string{"file.cpt"}

	if !reflect.DeepEqual(implFiles, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, implFiles)
	}
}

func TestPutStubImplementationShouldReturnNewFileContent(t *testing.T) {
	type stubImpl struct {
		ImplementationFilePath string   `json:"implementationFilePath"`
		Codes                  []string `json:"codes"`
	}
	cwd, _ := os.Getwd()
	dummyFilePath := filepath.Join(filepath.Join(cwd, "_testdata"), "dummyFile.txt")
	stubImplParams := stubImpl{ImplementationFilePath: dummyFilePath, Codes: []string{"code"}}

	b, _ := json.Marshal(stubImplParams)
	p := json.RawMessage(b)

	GetResponseFromRunner = func(m *gauge_messages.Message) (*gauge_messages.Message, error) {
		response := &gauge_messages.Message{
			MessageType: gauge_messages.Message_FileChanges,
			FileChanges: &gauge_messages.FileChanges{
				FileName:    "file",
				FileContent: "file content",
			},
		}
		return response, nil
	}

	stubImplResponse, err := putStubImpl(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	var want lsp.WorkspaceEdit
	want.Changes = make(map[string][]lsp.TextEdit, 0)
	uri := util.ConvertPathToURI(lsp.DocumentURI("file"))
	textEdit := lsp.TextEdit{
		NewText: "file content",
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   lsp.Position{Line: 1, Character: 0},
		},
	}
	want.Changes[string(uri)] = append(want.Changes[string(uri)], textEdit)

	if !reflect.DeepEqual(stubImplResponse, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, stubImplResponse)
	}
}
