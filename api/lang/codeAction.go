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

	"github.com/getgauge/gauge/logger"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

const (
	generateStubCommand   = "gauge.generate.unimplemented.stub"
	generateStubTitle     = "Generate step implementation stub"
	extractConceptCommand = "gauge.extract.concept"
	extractConceptTitle   = "Extract to concept"
)

func codeActions(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.CodeActionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	return append(getSpecCodeAction(params), getExtractConceptCodeAction(params)...), nil
}

func getExtractConceptCodeAction(params lsp.CodeActionParams) []lsp.Command {
	if len(params.Context.Diagnostics) > 0 || params.Range.Start.Line == params.Range.End.Line {
		return []lsp.Command{}
	}
	return []lsp.Command{{
		Command: extractConceptCommand,
		Title:   extractConceptTitle,
		Arguments: []interface{}{extractConceptInfo{
			Uri:   params.TextDocument.URI,
			Range: params.Range,
		}},
	}}
}

func getSpecCodeAction(params lsp.CodeActionParams) []lsp.Command {
	var actions []lsp.Command
	for _, d := range params.Context.Diagnostics {
		if d.Code != "" {
			actions = append(actions, lsp.Command{
				Command:   generateStubCommand,
				Title:     generateStubTitle,
				Arguments: []interface{}{d.Code},
			})
		}
	}
	return actions
}
