package lang

import (
	"io/ioutil"
	"github.com/getgauge/common"
	"encoding/json"
	"fmt"

	"github.com/getgauge/gauge/gauge"
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

func getSpecs(req *jsonrpc2.Request) (interface{}, error) {
	specFiles := util.FindSpecFilesIn(common.SpecsDirectoryName)
	parser := new(parser.SpecParser)
	specs := make([]specInfo, 0)
	for _, f := range specFiles {
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}
		spec, _ := parser.ParseSpecText(string(content), f)
		specs = append(specs, specInfo{Heading: spec.Heading.Value, ExecutionIdentifier: f})
	}
	return specs, nil
}

func getScenarios(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	file := util.ConvertURItoFilePath(params.TextDocument.URI)
	spec, parseResult := new(parser.SpecParser).Parse(getContent(params.TextDocument.URI), gauge.NewConceptDictionary(), file)
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