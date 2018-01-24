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
	"fmt"

	"encoding/json"
	"strconv"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

const (
	executeCommand    = "gauge.execute"
	debugCommand      = "gauge.debug"
	inParallelCommand = "gauge.execute.inParallel"
	referencesCommand = "gauge.showReferences"

	runSpecCodeLens       = "Run Spec"
	debugSpecCodeLens     = "Debug Spec"
	runInParallelCodeLens = "Run in parallel"
	runScenarioCodeLens   = "Run Scenario"
	debugScenarioCodeLens = "Debug Scenario"
)

func codeLenses(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.CodeLensParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	if util.IsGaugeFile(params.TextDocument.URI) {
		return getExecutionCodeLenses(params)
	} else {
		return getReferenceCodeLenses(params)
	}
}

func getExecutionCodeLenses(params lsp.CodeLensParams) (interface{}, error) {
	uri := string(params.TextDocument.URI)
	file := util.ConvertURItoFilePath(uri)
	if !util.IsSpec(file) {
		return nil, nil
	}
	spec, res, err := new(parser.SpecParser).Parse(getContent(uri), gauge.NewConceptDictionary(), file)
	if err != nil {
		return nil, err
	}

	if !res.Ok {
		err := fmt.Errorf("failed to parse specification %s", file)
		logger.APILog.Debugf(err.Error())
		return nil, err
	}
	var codeLenses []lsp.CodeLens
	runCodeLens := createCodeLens(spec.Heading.LineNo-1, runSpecCodeLens, executeCommand, getExecutionArgs(spec.FileName))
	codeLenses = append(codeLenses, runCodeLens)
	debugCodeLens := createCodeLens(spec.Heading.LineNo-1, debugSpecCodeLens, debugCommand, getExecutionArgs(spec.FileName))
	codeLenses = append(codeLenses, debugCodeLens)
	if spec.DataTable.IsInitialized() {
		codeLenses = append(codeLenses, getDataTableLenses(spec)...)
	}
	return append(getScenarioCodeLenses(spec), codeLenses...), nil

}

func getReferenceCodeLenses(params lsp.CodeLensParams) (interface{}, error) {
	if lRunner.runner == nil {
		return nil, nil
	}
	uri := params.TextDocument.URI
	stepPositionsResponse, err := getStepPositionResponse(uri)
	if err != nil {
		return nil, err
	}
	allSteps := provider.AllSteps()
	var lenses []lsp.CodeLens
	for _, stepPosition := range stepPositionsResponse.GetStepPositions() {
		var count int
		stepValue := stepPosition.GetStepValue()
		for _, step := range allSteps {
			if stepValue == step.Value {
				count++
			}
		}
		lensTitle := strconv.Itoa(count) + " reference(s)"
		lensPosition := lsp.Position{Line: int(stepPosition.GetSpan().GetStart()) - 1, Character: 0}
		lineNo := int(stepPosition.GetSpan().GetStart()) - 1
		args := []interface{}{uri, lensPosition, stepValue}

		lens := createCodeLens(lineNo, lensTitle, referencesCommand, args)
		lenses = append(lenses, lens)
	}
	return lenses, nil
}

func getDataTableLenses(spec *gauge.Specification) []lsp.CodeLens {
	var lenses []lsp.CodeLens
	lenses = append(lenses, createCodeLens(spec.Heading.LineNo-1, runInParallelCodeLens, inParallelCommand, getExecutionArgs(spec.FileName)))
	return lenses
}

func getScenarioCodeLenses(spec *gauge.Specification) []lsp.CodeLens {
	var lenses []lsp.CodeLens
	for _, sce := range spec.Scenarios {
		args := getExecutionArgs(fmt.Sprintf("%s:%d", spec.FileName, sce.Heading.LineNo))
		lens := createCodeLens(sce.Heading.LineNo-1, runScenarioCodeLens, executeCommand, args)
		lenses = append(lenses, lens)
		debugCodeLens := createCodeLens(sce.Heading.LineNo-1, debugScenarioCodeLens, debugCommand, args)
		lenses = append(lenses, debugCodeLens)

	}
	return lenses
}

func createCodeLens(lineNo int, lensTitle, command string, args []interface{}) lsp.CodeLens {
	return lsp.CodeLens{
		Range: lsp.Range{
			Start: lsp.Position{Line: lineNo, Character: 0},
			End:   lsp.Position{Line: lineNo, Character: len(lensTitle)},
		},
		Command: lsp.Command{
			Command:   command,
			Title:     lensTitle,
			Arguments: args,
		},
	}
}

func getExecutionArgs(id string) []interface{} {
	var args []interface{}
	return append(args, id)
}
