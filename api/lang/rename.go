package lang

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getgauge/common"
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
		logger.APILog.Debugf("failed to parse rename request %s", err.Error())
		return nil, err
	}

	spec, pResult := new(parser.SpecParser).ParseSpecText(getContent(params.TextDocument.URI), util.ConvertURItoFilePath(params.TextDocument.URI))
	if !pResult.Ok {
		return nil, fmt.Errorf("refactoring failed due to parse errors")
	}
	var step *gauge.Step
	for _, item := range spec.AllItems() {
		if item.Kind() == gauge.StepKind && item.(*gauge.Step).LineNo-1 == params.Position.Line {
			step = item.(*gauge.Step)
			break
		}
	}
	if step == nil {
		return nil, fmt.Errorf("refactoring is supported for steps only")
	}
	newName := getNewStepName(params, step)

	refactortingResult := refactor.GetRefactoringChanges(step.GetLineText(), newName, lRunner.runner, []string{common.SpecsDirectoryName})
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
	if err := addWorkspaceEdits(&result, refactortingResult.RunnerFilesChanged); err != nil {
		return nil, err
	}
	return result, nil
}

func getNewStepName(params lsp.RenameParams, step *gauge.Step) string {
	newName := strings.TrimSpace(strings.TrimPrefix(params.NewName, "*"))
	if step.HasInlineTable {
		newName = fmt.Sprintf("%s <%s>", newName, gauge.TableArg)
	}
	return newName
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
