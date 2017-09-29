package lang

import (
	"encoding/json"
	"fmt"

	"github.com/bugdiver/go-langserver/pkg/lsp"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/sourcegraph/jsonrpc2"
)

const (
	command = "gauge.execute"
)

func getCodeLenses(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.CodeLensParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	uri := string(params.TextDocument.URI)
	file := convertURItoFilePath(string(uri))
	spec, res := new(parser.SpecParser).Parse(getContent(uri), gauge.NewConceptDictionary(), file)
	if !res.Ok {
		err := fmt.Errorf("failed to parse specification %s", file)
		logger.APILog.Debugf(err.Error())
		return nil, err
	}
	codeLenses := createCodeLens(spec.Heading.LineNo-1, "Run Specification", command, getExecutionArgs(spec.FileName))
	return append(getScenariosCodeLenses(spec), codeLenses), nil

}

func resolveCodeLens(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.CodeLens
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	return params, nil
}

func getScenariosCodeLenses(spec *gauge.Specification) []lsp.CodeLens {
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
