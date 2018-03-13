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
	"strings"

	"github.com/getgauge/gauge/parser"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

const (
	generateStepCommand    = "gauge.generate.step"
	generateStubTitle      = "Create step implementation"
	generateConceptCommand = "gauge.generate.concept"
	generateConceptTitle   = "Create concept"
)

func codeActions(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.CodeActionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse request %v", err)
	}
	return getSpecCodeAction(params), nil
}

func getSpecCodeAction(params lsp.CodeActionParams) []lsp.Command {
	var actions []lsp.Command
	for _, d := range params.Context.Diagnostics {
		if d.Code != "" {
			actions = append(actions, createCodeAction(generateStepCommand, generateStubTitle, []interface{}{d.Code}))
			cptInfo := createConceptInfo(params.TextDocument.URI, params.Range.Start.Line)
			actions = append(actions, createCodeAction(generateConceptCommand, generateConceptTitle, []interface{}{cptInfo}))
		}
	}
	return actions
}

func createConceptInfo(uri lsp.DocumentURI, line int) concpetInfo {
	lineText := getLine(uri, line)
	stepValue, _ := parser.ExtractStepValueAndParams(lineText, false)
	cptName := strings.Replace(stepValue.ParameterizedStepValue, "*", "", -1)
	return concpetInfo{
		ConceptName: fmt.Sprintf("# %s\n* ", strings.TrimSpace(cptName)),
	}
}

func createCodeAction(command, titlle string, params []interface{}) lsp.Command {
	return lsp.Command{
		Command:   command,
		Title:     titlle,
		Arguments: params,
	}
}
