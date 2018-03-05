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

	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type stubImpl struct {
	ImplementationFilePath string   `json:"implementationFilePath"`
	Codes                  []string `json:"codes"`
}

func getImplFiles(req *jsonrpc2.Request) (interface{}, error) {
	var info = struct {
		Concept bool `json:"concept"`
	}{}
	if err := json.Unmarshal(*req.Params, &info); err != nil {
		return nil, fmt.Errorf("failed to parse request %s", err.Error())
	}
	if info.Concept {
		return util.GetConceptFiles(), nil
	}
	implementationFileListResponse, err := getImplementationFileList()
	if err != nil {
		return nil, err
	}
	return implementationFileListResponse.ImplementationFilePaths, nil
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
	uri := util.ConvertPathToURI(lsp.DocumentURI(fileDiff.FilePath))

	var textDiffs = fileDiff.TextDiffs
	for _, textDiff := range textDiffs {
		textEdit := lsp.TextEdit{
			NewText: textDiff.Content,
			Range: lsp.Range{
				Start: lsp.Position{Line: int(textDiff.Span.Start), Character: int(textDiff.Span.StartChar)},
				End:   lsp.Position{Line: int(textDiff.Span.End), Character: int(textDiff.Span.EndChar)},
			},
		}
		result.Changes[string(uri)] = append(result.Changes[string(uri)], textEdit)
	}
	return result
}
