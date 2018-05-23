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

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func stepReferences(req *jsonrpc2.Request) (interface{}, error) {
	var params string
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse request %v", err)
	}
	return getLocationFor(params)
}

func stepValueAt(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse request %v", err)
	}
	stepPositionsResponse, err := getStepPositionResponse(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	for _, step := range stepPositionsResponse.StepPositions {
		if (int(step.GetSpan().GetStart()) <= params.Position.Line+1) && (int(step.GetSpan().GetEnd()) >= params.Position.Line+1) {
			return step.GetStepValue(), nil
		}
	}
	return nil, nil
}

func getLocationFor(stepValue string) (interface{}, error) {
	allSteps := provider.AllSteps(false)
	var locations []lsp.Location
	diskFileCache := &files{cache: make(map[lsp.DocumentURI][]string)}
	for _, step := range allSteps {
		if stepValue == step.Value {
			uri := util.ConvertPathToURI(step.FileName)
			var endPos int
			lineNo := step.LineNo - 1
			if isOpen(uri) {
				endPos = len(getLine(uri, lineNo))
			} else {
				if !diskFileCache.exists(uri) {
					contents, err := common.ReadFileContents(step.FileName)
					if err != nil {
						return nil, err
					}
					diskFileCache.add(uri, contents)
				}
				endPos = len(diskFileCache.line(uri, lineNo))
			}
			locations = append(locations, lsp.Location{
				URI: uri,
				Range: lsp.Range{
					Start: lsp.Position{Line: lineNo, Character: 0},
					End:   lsp.Position{Line: lineNo, Character: endPos},
				},
			})
		}
	}
	return locations, nil
}
