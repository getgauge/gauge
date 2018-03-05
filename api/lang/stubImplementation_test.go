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
	"reflect"
	"testing"

	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

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

func TestPutStubImplementationShouldReturnFileDiff(t *testing.T) {
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
			MessageType: gauge_messages.Message_FileDiff,
			FileDiff: &gauge_messages.FileDiff{
				FilePath: "file",
				TextDiffs: []*gauge_messages.TextDiff{
					{
						Span: &gauge_messages.Span{
							Start:     1,
							StartChar: 2,
							End:       3,
							EndChar:   4,
						},
						Content: "file content",
					},
				},
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
			Start: lsp.Position{Line: 1, Character: 2},
			End:   lsp.Position{Line: 3, Character: 4},
		},
	}
	want.Changes[string(uri)] = append(want.Changes[string(uri)], textEdit)

	if !reflect.DeepEqual(stubImplResponse, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, stubImplResponse)
	}
}
