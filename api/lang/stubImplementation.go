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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/getgauge/common"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type stubImpl struct {
	ImplementationFilePath string   `json:"implementationFilePath"`
	Codes                  []string `json:"codes"`
}

type concpetInfo struct {
	ConceptName string `json:"conceptName"`
	ConceptFile string `json:"conceptFile"`
	Dir         string `json:"dir"`
}

type editInfo struct {
	fileName  string
	endLineNo int
	newText   string
}

func getImplFiles(req *jsonrpc2.Request) (interface{}, error) {
	fileList := []string{}
	var info = struct {
		Concept bool `json:"concept"`
	}{}
	if err := json.Unmarshal(*req.Params, &info); err != nil {
		return nil, fmt.Errorf("failed to parse request %s", err.Error())
	}
	if info.Concept {
		return append(fileList, util.GetConceptFiles()...), nil
	}
	response, err := getImplementationFileList()
	if err != nil {
		return nil, err
	}
	return append(fileList, response.GetImplementationFilePaths()...), nil
}

func putStubImpl(req *jsonrpc2.Request) (interface{}, error) {
	var stubImplParams stubImpl
	if err := json.Unmarshal(*req.Params, &stubImplParams); err != nil {
		return nil, fmt.Errorf("failed to parse request %s", err.Error())
	}
	fileDiff, err := putStubImplementation(stubImplParams.ImplementationFilePath, stubImplParams.Codes)
	if err != nil {
		return nil, err
	}

	return getWorkspaceEditForStubImpl(fileDiff), nil
}

func getWorkspaceEditForStubImpl(fileDiff *gm.FileDiff) lsp.WorkspaceEdit {
	var result lsp.WorkspaceEdit
	result.Changes = make(map[string][]lsp.TextEdit, 0)
	uri := util.ConvertPathToURI(fileDiff.FilePath)

	var textDiffs = fileDiff.TextDiffs
	for _, textDiff := range textDiffs {
		span := textDiff.Span
		textEdit := createTextEdit(textDiff.Content, int(span.Start), int(span.StartChar), int(span.End), int(span.EndChar))
		result.Changes[string(uri)] = append(result.Changes[string(uri)], textEdit)
	}
	return result
}

func generateConcept(req *jsonrpc2.Request) (interface{}, error) {
	var params concpetInfo
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("Failed to parse request %s", err.Error())
	}
	conceptFile := string(params.ConceptFile)
	edit := editInfo{
		fileName:  conceptFile,
		endLineNo: 0,
		newText:   params.ConceptName,
	}
	content, err := common.ReadFileContents(conceptFile)
	if err != nil {
		edit.fileName = getFileName(params.Dir, 1)
	} else if content != "" {
		content = strings.Join(util.GetLinesFromText(content), "\n")
		edit.newText = fmt.Sprintf("%s\n\n%s", strings.TrimSpace(content), params.ConceptName)
		edit.endLineNo = len(strings.Split(content, "\n"))
	}
	return createWorkSpaceEdits(edit), nil
}

func createWorkSpaceEdits(edit editInfo) lsp.WorkspaceEdit {
	var result = lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{}}
	textEdiit := createTextEdit(edit.newText, 0, 0, edit.endLineNo, 0)
	result.Changes[edit.fileName] = []lsp.TextEdit{textEdiit}
	return result
}

func createTextEdit(text string, start, startChar, end, endChar int) lsp.TextEdit {
	return lsp.TextEdit{
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      start,
				Character: startChar,
			},
			End: lsp.Position{
				Line:      end,
				Character: endChar,
			},
		},
		NewText: text,
	}
}

func getFileName(dir string, count int) string {
	var fileName = filepath.Join(dir, fmt.Sprintf("concept%d.cpt", count))
	if !common.FileExists(fileName) {
		return fileName
	}
	return getFileName(dir, count+1)
}
