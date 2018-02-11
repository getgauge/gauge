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

	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type ScenarioInfo struct {
	Heading             string `json:"heading"`
	LineNo              int    `json:"lineNo"`
	ExecutionIdentifier string `json:"executionIdentifier"`
}

type specInfo struct {
	Heading             string `json:"heading"`
	ExecutionIdentifier string `json:"executionIdentifier"`
}

type StubImpl struct {
	ImplementationFilePath string   `json:"implementationFilePath"`
	StepTexts              []string `json:"stepTexts"`
}

func specs() (interface{}, error) {
	specDetails := provider.GetAvailableSpecDetails([]string{})
	specs := make([]specInfo, 0)
	for _, d := range specDetails {
		specs = append(specs, specInfo{Heading: d.Spec.Heading.Value, ExecutionIdentifier: d.Spec.FileName})
	}
	return specs, nil
}

func getImplFiles() (interface{}, error) {
	if lRunner.runner == nil {
		return nil, nil
	}
	implementationFileListResponse, err := getImplementationFileList()
	if err != nil {
		return nil, err
	}
	return implementationFileListResponse.ImplementationFilePaths, nil
}

func putStubImpl(req *jsonrpc2.Request) (interface{}, error) {
	var stubImplParams StubImpl
	if err := json.Unmarshal(*req.Params, &stubImplParams); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	if lRunner.runner == nil {
		return nil, nil
	}
	var stepVals []*gm.ProtoStepValue
	for _, stepText := range stubImplParams.StepTexts {
		stepValue, err := parser.ExtractStepValueAndParams(stepText, false)
		if err != nil {
			return nil, err
		}
		protoStepValue := gauge.ConvertToProtoStepValue(stepValue)

		stepVals = append(stepVals, protoStepValue)
	}
	putStubImplementationResponse, err := putStubImplementation(stubImplParams.ImplementationFilePath, stepVals)
	if err != nil {
		return nil, err
	}
	return putStubImplementationResponse, nil
}

func scenarios(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	file := util.ConvertURItoFilePath(params.TextDocument.URI)
	content := ""
	if !isOpen(params.TextDocument.URI) {
		specDetails := provider.GetAvailableSpecDetails([]string{file})
		return getScenarioAt(specDetails[0].Spec.Scenarios, file, params.Position.Line), nil
	}
	content = getContent(params.TextDocument.URI)
	spec, parseResult, err := new(parser.SpecParser).Parse(content, gauge.NewConceptDictionary(), file)
	if err != nil {
		return nil, err
	}
	if !parseResult.Ok {
		return nil, fmt.Errorf("parsing failed")
	}
	return getScenarioAt(spec.Scenarios, file, params.Position.Line), nil
}

func getScenarioAt(scenarios []*gauge.Scenario, file string, line int) interface{} {
	var ifs []ScenarioInfo
	for _, sce := range scenarios {
		info := getScenarioInfo(sce, file)
		if sce.InSpan(line + 1) {
			return info
		}
		ifs = append(ifs, info)
	}
	return ifs
}
func getScenarioInfo(sce *gauge.Scenario, file string) ScenarioInfo {
	return ScenarioInfo{
		Heading:             sce.Heading.Value,
		LineNo:              sce.Heading.LineNo,
		ExecutionIdentifier: fmt.Sprintf("%s:%d", file, sce.Heading.LineNo),
	}
}
