package lang

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getgauge/common"
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
		return nil, fmt.Errorf("rename is supported only for Steps")
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
	var result lsp.WorkspaceEdit
	result.Changes = make(map[string][]lsp.TextEdit, 0)
	if err := addWorkspaceEdits(&result, refactortingResult.SpecsChanged); err != nil {
		return nil, err
	}
	if err := addWorkspaceEdits(&result, refactortingResult.ConceptsChanged); err != nil {
		return nil, err
	}
	return result, nil
}

func addWorkspaceEdits(result *lsp.WorkspaceEdit, filesChanged map[string]string) error {
	diskFileCache := &files{cache: make(map[string][]string)}
	for fileName, text := range filesChanged {
		uri := util.ConvertPathToURI(fileName)
		var lastLineNo int
		var lastLineLength int
		if isOpen(uri) {
			lastLineNo = getLineCount(uri) - 1
			lastLineLength = len(getLine(uri, lastLineNo))
		} else {
			if !diskFileCache.exists(uri) {
				contents, err := common.ReadFileContents(fileName)
				if err != nil {
					return err
				}
				diskFileCache.add(uri, contents)
			}
			lastLineNo = len(diskFileCache.content(uri)) - 1
			lastLineLength = len(diskFileCache.line(uri, lastLineNo))
		}
		textEdit := lsp.TextEdit{
			NewText: text,
			Range: lsp.Range{
				Start: lsp.Position{Line: 0, Character: 0},
				End:   lsp.Position{Line: lastLineNo, Character: lastLineLength},
			},
		}
		result.Changes[uri] = append(result.Changes[uri], textEdit)
	}
	return nil
}
