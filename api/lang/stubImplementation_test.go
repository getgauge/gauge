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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type mockLspClient struct {
	response interface{}
	err      error
}

func (r *mockLspClient) GetStepNames(ctx context.Context, in *gm.StepNamesRequest, opts ...grpc.CallOption) (*gm.StepNamesResponse, error) {
	return r.response.(*gm.StepNamesResponse), r.err
}
func (r *mockLspClient) CacheFile(ctx context.Context, in *gm.CacheFileRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return r.response.(*gm.Empty), r.err
}
func (r *mockLspClient) GetStepPositions(ctx context.Context, in *gm.StepPositionsRequest, opts ...grpc.CallOption) (*gm.StepPositionsResponse, error) {
	return r.response.(*gm.StepPositionsResponse), r.err
}
func (r *mockLspClient) GetImplementationFiles(ctx context.Context, in *gm.Empty, opts ...grpc.CallOption) (*gm.ImplementationFileListResponse, error) {
	return r.response.(*gm.ImplementationFileListResponse), r.err
}
func (r *mockLspClient) ImplementStub(ctx context.Context, in *gm.StubImplementationCodeRequest, opts ...grpc.CallOption) (*gm.FileDiff, error) {
	return r.response.(*gm.FileDiff), r.err
}
func (r *mockLspClient) ValidateStep(ctx context.Context, in *gm.StepValidateRequest, opts ...grpc.CallOption) (*gm.StepValidateResponse, error) {
	return r.response.(*gm.StepValidateResponse), r.err
}
func (r *mockLspClient) Refactor(ctx context.Context, in *gm.RefactorRequest, opts ...grpc.CallOption) (*gm.RefactorResponse, error) {
	return r.response.(*gm.RefactorResponse), r.err
}
func (r *mockLspClient) GetStepName(ctx context.Context, in *gm.StepNameRequest, opts ...grpc.CallOption) (*gm.StepNameResponse, error) {
	return r.response.(*gm.StepNameResponse), r.err
}

func (r *mockLspClient) GetGlobPatterns(ctx context.Context, in *gm.Empty, opts ...grpc.CallOption) (*gm.ImplementationFileGlobPatternResponse, error) {
	return r.response.(*gm.ImplementationFileGlobPatternResponse), r.err
}

func (r *mockLspClient) KillProcess(ctx context.Context, in *gm.KillProcessRequest, opts ...grpc.CallOption) (*gm.Empty, error) {
	return nil, nil
}

func TestGetImplementationFilesShouldReturnFilePaths(t *testing.T) {

	var params = struct {
		Concept bool
	}{}

	b, _ := json.Marshal(params)
	p := json.RawMessage(b)

	response := &gm.ImplementationFileListResponse{
		ImplementationFilePaths: []string{"file"},
	}
	lRunner.runner = &runner.GrpcRunner{Client: &mockLspClient{response: response}, Timeout: config.IdeRequestTimeout()}
	implFiles, err := getImplFiles(&jsonrpc2.Request{Params: &p})

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
	response := &gm.ImplementationFileListResponse{
		ImplementationFilePaths: nil,
	}
	lRunner.runner = &runner.GrpcRunner{Client: &mockLspClient{response: response}, Timeout: config.IdeRequestTimeout()}

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
	response := &gm.FileDiff{
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
	lRunner.runner = &runner.GrpcRunner{Client: &mockLspClient{response: response}, Timeout: config.IdeRequestTimeout()}

	stubImplResponse, err := putStubImpl(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	var want lsp.WorkspaceEdit
	want.Changes = make(map[string][]lsp.TextEdit, 0)
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
	want.Changes = make(map[string][]lsp.TextEdit, 0)
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
	want.Changes = make(map[string][]lsp.TextEdit, 0)
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

func TestGenerateConceptInExisitingFile(t *testing.T) {
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
	want.Changes = make(map[string][]lsp.TextEdit, 0)

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
	ioutil.WriteFile(cptFile, []byte(""), common.NewFilePermissions)
	defer common.Remove(cptFile)

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
	want.Changes = make(map[string][]lsp.TextEdit, 0)

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
