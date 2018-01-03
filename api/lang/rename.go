package lang

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/refactor"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func rename(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.RenameParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	line := strings.TrimSpace(getLine(params.TextDocument.URI, params.Position.Line))
	if !parser.IsStep(line) {
		return nil, fmt.Errorf("rename is supported only for steps")
	}
	oldName := strings.TrimSpace(strings.TrimPrefix(line, "*"))
	newName := strings.TrimSpace(strings.TrimPrefix(params.NewName, "*"))
	refactortingResult := refactor.GetRefactoredSteps(oldName, newName, nil, []string{common.SpecsDirectoryName})
	for _, warning := range refactortingResult.Warnings {
		logger.Warningf(warning)
	}
	if !refactortingResult.Success {
		return nil, fmt.Errorf("Refactoring failed due to errors:\n%s", strings.Join(refactortingResult.Errors, "\n"))
	}
	return getRefactoringWorkspaceEdits(refactortingResult.StepsChanged), nil
}

func getRefactoringWorkspaceEdits(stepsChanged []*gauge.Step) (result lsp.WorkspaceEdit) {
	result.Changes = make(map[string][]lsp.TextEdit, 0)
	for _, step := range stepsChanged {
		textEdit := lsp.TextEdit{
			NewText: getNewText(step),
			Range: lsp.Range{
				Start: lsp.Position{Line: step.LineNo - 1, Character: 0},
				End:   lsp.Position{Line: step.LineNo - 1, Character: 10000},
			},
		}
		uri := util.ConvertPathToURI(step.FileName)
		result.Changes[uri] = append(result.Changes[uri], textEdit)
	}
	return
}

func getNewText(step *gauge.Step) (newText string) {
	newText = strings.TrimSuffix(formatter.FormatStep(step), "\n")
	if step.IsConcept {
		newText = strings.TrimSpace(strings.Replace(newText, "*", "#", 1))
	}
	return
}
