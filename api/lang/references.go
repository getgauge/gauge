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

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func getStepReferences(req *jsonrpc2.Request) (interface{}, error) {
	var params string
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	allSteps := provider.AllSteps()
	var locations []lsp.Location
	diskFileCache := &files{cache: make(map[string][]string)}
	for _, step := range allSteps {
		if params == step.Value {
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
