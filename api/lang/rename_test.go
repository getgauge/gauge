/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestRenameStep(t *testing.T) {
	specText := `# Specification Heading

## Scenario Heading
	
* Step text

* concept heading

* with a step
`

	cwd, _ := os.Getwd()
	specFile := filepath.Join(cwd, "_testdata", "test.spec")
	specURI := util.ConvertPathToURI(specFile)
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(specURI, specText)

	util.GetSpecFiles = func(paths []string) []string {
		return []string{specFile}
	}
	util.GetConceptFiles = func() []string {
		return []string{}
	}
	responses := map[gauge_messages.Message_MessageType]interface{}{}
	responses[gauge_messages.Message_StepNameResponse] = &gauge_messages.StepNameResponse{}
	responses[gauge_messages.Message_RefactorResponse] = &gauge_messages.RefactorResponse{}
	lRunner.runner = &runner.GrpcRunner{Timeout: time.Second * 30, LegacyClient: &mockClient{responses: responses}}

	renameParams := lsp.RenameParams{
		NewName: `* Step text with <param>`,
		Position: lsp.Position{
			Line:      4,
			Character: 3,
		},
		TextDocument: lsp.TextDocumentIdentifier{URI: specURI},
	}

	b, _ := json.Marshal(renameParams)
	p := json.RawMessage(b)

	got, err := renameStep(&jsonrpc2.Request{Params: &p})
	want := lsp.WorkspaceEdit{
		Changes: map[string][]lsp.TextEdit{
			string(specURI): []lsp.TextEdit{
				lsp.TextEdit{
					NewText: `* Step text with "param"`,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      4,
							Character: 0,
						},
						End: lsp.Position{
							Line:      4,
							Character: 11,
						},
					},
				},
			},
		},
	}

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("get step references failed, want: `%v`, got: `%v`", want, got)
	}
}

func TestRenameConceptStep(t *testing.T) {
	specText := `# Specification Heading

## Scenario Heading

* Step text

* concept heading

* with a step
`

	cwd, _ := os.Getwd()
	specFile := filepath.Join(cwd, "_testdata", "test.spec")
	conceptFile := filepath.Join(cwd, "_testdata", "some.cpt")
	specURI := util.ConvertPathToURI(specFile)
	conceptURI := util.ConvertPathToURI(conceptFile)
	openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}
	openFilesCache.add(specURI, specText)

	util.GetSpecFiles = func(paths []string) []string {
		return []string{specFile}
	}
	util.GetConceptFiles = func() []string {
		return []string{conceptFile}
	}
	responses := map[gauge_messages.Message_MessageType]interface{}{}
	responses[gauge_messages.Message_StepNameResponse] = &gauge_messages.StepNameResponse{}
	responses[gauge_messages.Message_RefactorResponse] = &gauge_messages.RefactorResponse{}
	lRunner.runner = &runner.GrpcRunner{Timeout: time.Second * 30, LegacyClient: &mockClient{responses: responses}}

	renameParams := lsp.RenameParams{
		NewName: `* concpet heading with "params"`,
		Position: lsp.Position{
			Line:      6,
			Character: 3,
		},
		TextDocument: lsp.TextDocumentIdentifier{URI: specURI},
	}

	b, _ := json.Marshal(renameParams)
	p := json.RawMessage(b)

	want := lsp.WorkspaceEdit{
		Changes: map[string][]lsp.TextEdit{
			string(specURI): []lsp.TextEdit{
				lsp.TextEdit{
					NewText: `* concpet heading with "params"`,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      6,
							Character: 0,
						},
						End: lsp.Position{
							Line:      6,
							Character: 17,
						},
					},
				},
			},
			string(conceptURI): []lsp.TextEdit{
				lsp.TextEdit{
					NewText: `# concpet heading with <params>`,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      0,
							Character: 17,
						},
					},
				},
			},
		},
	}

	got, err := renameStep(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}
	we := got.(lsp.WorkspaceEdit)
	for file, edits := range we.Changes {
		if !reflect.DeepEqual(edits, want.Changes[file]) {
			t.Errorf("refacotoring failed, want: `%v`, got: `%v`", want.Changes[file], edits)
		}
	}
}
