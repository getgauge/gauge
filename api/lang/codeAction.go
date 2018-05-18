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

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
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
	return getSpecCodeAction(params)
}

func getSpecCodeAction(params lsp.CodeActionParams) ([]lsp.Command, error) {
	var actions []lsp.Command
	for _, d := range params.Context.Diagnostics {
		if d.Code != "" {
			actions = append(actions, createCodeAction(generateStepCommand, generateStubTitle, []interface{}{d.Code}))
			cptInfo, err := createConceptInfo(params.TextDocument.URI, params.Range.Start.Line)
			if err != nil {
				return nil, err
			}
			actions = append(actions, createCodeAction(generateConceptCommand, generateConceptTitle, []interface{}{cptInfo}))
		}
	}
	return actions, nil
}

func createConceptInfo(uri lsp.DocumentURI, line int) (interface{}, error) {
	file := util.ConvertURItoFilePath(uri)
	linetext := getLine(uri, line)
	var stepValue *gauge.StepValue
	if util.IsConcept(file) {
		steps, _ := new(parser.ConceptParser).Parse(getContent(uri), file)
		for _, conStep := range steps {
			for _, step := range conStep.ConceptSteps {
				if step.LineNo-1 == line {
					stepValue = extractStepValueAndParams(step, linetext)
				}
			}
		}
	}
	if util.IsSpec(file) {
		spec, res, err := new(parser.SpecParser).Parse(getContent(uri), &gauge.ConceptDictionary{}, file)
		if err != nil {
			return nil, err
		}
		if !res.Ok {
			return nil, fmt.Errorf("parsing failed for %s. %s", uri, res.Errors())
		}
		for _, step := range spec.Steps() {
			if step.LineNo-1 == line {
				stepValue = extractStepValueAndParams(step, linetext)
			}
		}
	}
	cptName := strings.Replace(stepValue.ParameterizedStepValue, "*", "", -1)
	return concpetInfo{
		ConceptName: fmt.Sprintf("# %s\n* ", strings.TrimSpace(cptName)),
	}, nil
}

func extractStepValueAndParams(step *gauge.Step, linetext string) *gauge.StepValue {
	var stepValue *gauge.StepValue
	if step.HasInlineTable {
		stepValue, _ = parser.ExtractStepValueAndParams(step.LineText, true)
	} else {
		stepValue, _ = parser.ExtractStepValueAndParams(linetext, false)
	}
	return stepValue
}

func createCodeAction(command, titlle string, params []interface{}) lsp.Command {
	return lsp.Command{
		Command:   command,
		Title:     titlle,
		Arguments: params,
	}
}
