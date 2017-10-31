package lang

import (
	"fmt"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"encoding/json"

	"strconv"

	"github.com/sourcegraph/jsonrpc2"
)

const (
	command           = "gauge.execute"
	inParallelCommand = "gauge.execute.inParallel"
)

func getCodeLenses(req *jsonrpc2.Request) (interface{}, error) {
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
	spec, res := new(parser.SpecParser).Parse(getContent(uri), gauge.NewConceptDictionary(), file)

	if !res.Ok {
		err := fmt.Errorf("failed to parse specification %s", file)
		logger.APILog.Debugf(err.Error())
		return nil, err
	}
	var codeLenses []lsp.CodeLens
	specLenses := createCodeLens(spec.Heading.LineNo-1, "Run Spec", command, getExecutionArgs(spec.FileName))
	codeLenses = append(codeLenses, specLenses)
	if spec.DataTable.IsInitialized() {
		codeLenses = append(codeLenses, getDataTableLenses(spec)...)
	}
	return append(getScenarioCodeLenses(spec), codeLenses...), nil

}

func getReferenceCodeLenses(params lsp.CodeLensParams) (interface{}, error) {
	cacheFileRequest := &gm.Message{MessageType: gm.Message_StepPositionsRequest, StepPositionsRequest: &gm.StepPositionsRequest{FilePath: util.ConvertURItoFilePath(params.TextDocument.URI)}}
	response, err := conn.GetResponseForMessageWithTimeout(cacheFileRequest, lRunner.runner.Connection(), config.RunnerConnectionTimeout())
	if err != nil {
		logger.APILog.Infof("Error while connecting to runner : %s", err.Error())
		return nil, err
	}
	stepPositionsResponse := response.GetStepPositionsResponse()
	if stepPositionsResponse.GetError() != "" {
		logger.APILog.Infof("Error while connecting to runner : %s", stepPositionsResponse.GetError())
	}
	allSteps := provider.AllSteps()
	var lenses []lsp.CodeLens
	for _, stepPosition := range stepPositionsResponse.GetStepPositions() {
		stepValue, err := parser.ExtractStepValueAndParams(stepPosition.GetStepValue(), false)
		var count int
		var locations []lsp.Location
		if err != nil {
			continue
		}
		for _, step := range allSteps {
			if stepValue.StepValue == step.Value {
				count++
				locations = append(locations, lsp.Location{URI: util.ConvertPathToURI(step.FileName), Range: lsp.Range{
					Start: lsp.Position{Line: step.LineNo - 1, Character: 0},
					End:   lsp.Position{Line: step.LineNo - 1, Character: 0},
				}})
			}
		}

		lens := createCodeLens(int(stepPosition.GetLineNumber())-1, strconv.Itoa(count)+" references", "gauge.showReferences", []interface{}{params.TextDocument.URI, lsp.Position{Line: int(stepPosition.GetLineNumber()) - 1, Character: 0}, locations})
		lenses = append(lenses, lens)
	}
	return lenses, nil
}

func getDataTableLenses(spec *gauge.Specification) []lsp.CodeLens {
	var lenses []lsp.CodeLens
	lenses = append(lenses, createCodeLens(spec.Heading.LineNo-1, "Run in parallel", inParallelCommand, getExecutionArgs(spec.FileName)))
	return lenses
}

func getScenarioCodeLenses(spec *gauge.Specification) []lsp.CodeLens {
	var lenses []lsp.CodeLens
	for _, sce := range spec.Scenarios {
		args := getExecutionArgs(fmt.Sprintf("%s:%d", spec.FileName, sce.Heading.LineNo))
		lens := createCodeLens(sce.Heading.LineNo-1, "Run Scenario", command, args)
		lenses = append(lenses, lens)
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
