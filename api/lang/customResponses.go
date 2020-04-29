/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"encoding/json"
	"fmt"

	"github.com/getgauge/gauge/gauge"
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

func specs() (interface{}, error) {
	specDetails := provider.GetAvailableSpecDetails([]string{})
	specs := make([]specInfo, 0)
	for _, d := range specDetails {
		specs = append(specs, specInfo{Heading: d.Spec.Heading.Value, ExecutionIdentifier: d.Spec.FileName})
	}
	return specs, nil
}

func scenarios(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse request %s", err)
	}
	file := util.ConvertURItoFilePath(params.TextDocument.URI)
	content := ""
	if !isOpen(params.TextDocument.URI) {
		specDetails := provider.GetAvailableSpecDetails([]string{file})
		return getScenarioAt(specDetails[0].Spec.Scenarios, file, params.Position.Line), nil
	}
	content = getContent(params.TextDocument.URI)
	spec, parseResult, err := new(parser.SpecParser).Parse(content, gauge.NewConceptDictionary(), string(file))
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
