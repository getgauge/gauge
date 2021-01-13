/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/getgauge/common"
	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/runner"
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
	responses := map[gm.Message_MessageType]interface{}{}
	responses[gm.Message_ImplementationFileListResponse] = &gm.ImplementationFileListResponse{
		ImplementationFilePaths: []string{"file"},
	}
	lRunner.runner = &runner.GrpcRunner{LegacyClient: &mockClient{responses: responses}, Timeout: time.Second * 30}
	implFiles, err := getImplFiles(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	want := []string{"file"}

	if !reflect.DeepEqual(implFiles, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, implFiles)
	}
}

func TestGetImplementationFilesShouldReturnFilePathsIfParamIsNil(t *testing.T) {
	responses := map[gm.Message_MessageType]interface{}{}
	responses[gm.Message_ImplementationFileListResponse] = &gm.ImplementationFileListResponse{
		ImplementationFilePaths: []string{"file"},
	}
	lRunner.runner = &runner.GrpcRunner{LegacyClient: &mockClient{responses: responses}, Timeout: time.Second * 30}
	implFiles, err := getImplFiles(&jsonrpc2.Request{})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	want := []string{"file"}

	if !reflect.DeepEqual(implFiles, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, implFiles)
	}
}

func TestGetImplementationFilesShouldReturnEmptyArrayForNoImplementationFiles(t *testing.T) {
	var params = struct {
		Concept bool
	}{}

	b, _ := json.Marshal(params)
	p := json.RawMessage(b)

	responses := map[gm.Message_MessageType]interface{}{}
	responses[gm.Message_ImplementationFileListResponse] = &gm.ImplementationFileListResponse{
		ImplementationFilePaths: nil,
	}

	lRunner.runner = &runner.GrpcRunner{LegacyClient: &mockClient{responses: responses}, Timeout: time.Second * 30}

	implFiles, err := getImplFiles(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	want := []string{}

	if !reflect.DeepEqual(implFiles, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, implFiles)
	}
}

func TestGetImplementationFilesShouldReturnEmptyArrayForNoConceptFiles(t *testing.T) {
	type cptParam struct {
		Concept bool `json:"concept"`
	}

	params := cptParam{Concept: true}

	b, _ := json.Marshal(params)
	p := json.RawMessage(b)

	util.GetConceptFiles = func() []string {
		return nil
	}

	cptFiles, err := getImplFiles(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	want := []string{}

	if !reflect.DeepEqual(cptFiles, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, cptFiles)
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
	responses := map[gm.Message_MessageType]interface{}{}
	responses[gm.Message_FileDiff] = &gm.FileDiff{
		FilePath: "file",
		TextDiffs: []*gm.TextDiff{
			{
				Span: &gm.Span{
					Start:     1,
					StartChar: 2,
					End:       3,
					EndChar:   4,
				},
				Content: "file content",
			},
		},
	}
	lRunner.runner = &runner.GrpcRunner{LegacyClient: &mockClient{responses: responses}, Timeout: time.Second * 30}

	stubImplResponse, err := putStubImpl(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	var want lsp.WorkspaceEdit
	want.Changes = make(map[string][]lsp.TextEdit)
	uri := util.ConvertPathToURI("file")
	textEdit := lsp.TextEdit{
		NewText: "file content",
		Range: lsp.Range{
			Start: lsp.Position{Line: 1, Character: 2},
			End:   lsp.Position{Line: 3, Character: 4},
		},
	}
	want.Changes[string(uri)] = append(want.Changes[string(uri)], textEdit)

	if !reflect.DeepEqual(stubImplResponse, want) {
		t.Errorf("want: `%v`,\n got: `%v`", want, stubImplResponse)
	}
}

func TestGenerateConceptShouldReturnFileDiff(t *testing.T) {
	cwd, _ := os.Getwd()
	testData := filepath.Join(cwd, "_testdata")

	extractConcpetParam := concpetInfo{
		ConceptName: "# foo bar\n* ",
		ConceptFile: "New File",
		Dir:         testData,
	}
	b, _ := json.Marshal(extractConcpetParam)
	p := json.RawMessage(b)

	response, err := generateConcept(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	var want lsp.WorkspaceEdit
	want.Changes = make(map[string][]lsp.TextEdit)
	uri := util.ConvertPathToURI(filepath.Join(testData, "concept1.cpt"))
	textEdit := lsp.TextEdit{
		NewText: "# foo bar\n* ",
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   lsp.Position{Line: 0, Character: 0},
		},
	}
	want.Changes[string(uri)] = append(want.Changes[string(uri)], textEdit)

	if !reflect.DeepEqual(want, response) {
		t.Errorf("want: `%v`,\n got: `%v`", want, response)
	}
}

func TestGenerateConceptWithParam(t *testing.T) {
	cwd, _ := os.Getwd()
	testData := filepath.Join(cwd, "_testdata")

	extractConcpetParam := concpetInfo{
		ConceptName: "# foo bar <some>\n* ",
		ConceptFile: "New File",
		Dir:         testData,
	}
	b, _ := json.Marshal(extractConcpetParam)
	p := json.RawMessage(b)

	response, err := generateConcept(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	var want lsp.WorkspaceEdit
	want.Changes = make(map[string][]lsp.TextEdit)
	uri := util.ConvertPathToURI(filepath.Join(testData, "concept1.cpt"))
	textEdit := lsp.TextEdit{
		NewText: "# foo bar <some>\n* ",
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   lsp.Position{Line: 0, Character: 0},
		},
	}
	want.Changes[string(uri)] = append(want.Changes[string(uri)], textEdit)

	if !reflect.DeepEqual(want, response) {
		t.Errorf("want: `%v`,\n got: `%v`", want, response)
	}
}

func TestGenerateConceptInExistingFile(t *testing.T) {
	cwd, _ := os.Getwd()
	testData := filepath.Join(cwd, "_testdata")
	cptFile := filepath.Join(testData, "some.cpt")

	extractConcpetParam := concpetInfo{
		ConceptName: "# foo bar <some>\n* ",
		ConceptFile: cptFile,
		Dir:         testData,
	}
	b, _ := json.Marshal(extractConcpetParam)
	p := json.RawMessage(b)

	response, err := generateConcept(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	var want lsp.WorkspaceEdit
	want.Changes = make(map[string][]lsp.TextEdit)
	textEdit := lsp.TextEdit{
		NewText: "# concept heading\n* with a step\n\n# foo bar <some>\n* ",
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   lsp.Position{Line: 2, Character: 0},
		},
	}
	uri := string(util.ConvertPathToURI(cptFile))

	want.Changes[uri] = append(want.Changes[uri], textEdit)

	if !reflect.DeepEqual(want, response) {
		t.Errorf("want: `%v`,\n got: `%v`", want, response)
	}
}

func TestGenerateConceptInNewFileWhenDefaultExisits(t *testing.T) {
	cwd, _ := os.Getwd()
	testData := filepath.Join(cwd, "_testdata")

	cptFile := filepath.Join(testData, "concept1.cpt")
	err := ioutil.WriteFile(cptFile, []byte(""), common.NewFilePermissions)
	if err != nil {
		t.Fatalf("Unable to create Concept %s: %s", cptFile, err.Error())
	}

	defer func() {
		err := common.Remove(cptFile)
		if err != nil {
			t.Fatalf("Unable to delete Concept %s: %s", cptFile, err.Error())
		}
	}()

	extractConcpetParam := concpetInfo{
		ConceptName: "# foo bar <some>\n* ",
		ConceptFile: "New File",
		Dir:         testData,
	}
	b, _ := json.Marshal(extractConcpetParam)
	p := json.RawMessage(b)

	response, err := generateConcept(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	uri := util.ConvertPathToURI(filepath.Join(testData, "concept2.cpt"))

	var want lsp.WorkspaceEdit
	want.Changes = make(map[string][]lsp.TextEdit)

	textEdit := lsp.TextEdit{
		NewText: "# foo bar <some>\n* ",
		Range: lsp.Range{
			Start: lsp.Position{Line: 0, Character: 0},
			End:   lsp.Position{Line: 0, Character: 0},
		},
	}
	want.Changes[string(uri)] = append(want.Changes[string(uri)], textEdit)

	if !reflect.DeepEqual(want, response) {
		t.Errorf("want: `%v`,\n got: `%v`", want, response)
	}
}
