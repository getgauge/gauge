package lang

import (
	"encoding/json"
	"github.com/getgauge/gauge/logger"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

const (
	copyStubCommand = "gauge.copy.unimplemented.stub"
	copyStubTitle   = "Copy function stub for this step to clipboard"
)

func getCodeActions(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.CodeActionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	return getSpecCodeAction(params), nil
}

func getSpecCodeAction(params lsp.CodeActionParams) interface{} {
	var actions []lsp.Command
	for _, d := range params.Context.Diagnostics {
		if d.Code != "" {
			actions = append(actions, lsp.Command{
				Command:   copyStubCommand,
				Title:     copyStubTitle,
				Arguments: []interface{}{d.Code},
			})
		}
	}
	return actions
}
