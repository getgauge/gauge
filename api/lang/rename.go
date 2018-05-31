// Copyright 2018 ThoughtWorks, Inc.

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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/refactor"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func rename(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request) (interface{}, error) {
	if err := sendSaveFilesRequest(ctx, conn); err != nil {
		return nil, err
	}
	var params lsp.RenameParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		logDebug(req, "failed to parse rename request %s", err.Error())
		return nil, err
	}

	step, err := getStepToRefactor(params)

	if step == nil {
		return nil, fmt.Errorf("refactoring is supported for steps only")
	}
	newName := getNewStepName(params, step)

	refactortingResult := refactor.GetRefactoringChanges(step.GetLineText(), newName, lRunner.runner, util.GetSpecDirs())
	for _, warning := range refactortingResult.Warnings {
		logWarning(req, warning)
	}
	if !refactortingResult.Success {
		return nil, fmt.Errorf("%s", strings.Join(refactortingResult.Errors, "\t"))
	}
	var result lsp.WorkspaceEdit
	result.Changes = make(map[string][]lsp.TextEdit, 0)
	if err := addSrcWorkspaceEdits(&result, refactortingResult.SpecsChanged); err != nil {
		return nil, err
	}
	if err := addWorkspaceEdits(&result, refactortingResult.ConceptsChanged); err != nil {
		return nil, err
	}
	if err := addSrcWorkspaceEdits(&result, refactortingResult.RunnerFilesChanged); err != nil {
		return nil, err
	}
	return result, nil
}

func getStepToRefactor(params lsp.RenameParams) (*gauge.Step, error) {
	file := util.ConvertURItoFilePath(params.TextDocument.URI)
	if util.IsSpec(file) {
		spec, pResult := new(parser.SpecParser).ParseSpecText(getContent(params.TextDocument.URI), util.ConvertURItoFilePath(params.TextDocument.URI))
		if !pResult.Ok {
			return nil, fmt.Errorf("refactoring failed due to parse errors: \n%s", strings.Join(pResult.Errors(), "\n"))
		}
		for _, item := range spec.AllItems() {
			if item.Kind() == gauge.StepKind && item.(*gauge.Step).LineNo-1 == params.Position.Line {
				return item.(*gauge.Step), nil

			}
		}
	}
	if util.IsConcept(file) {
		steps, _ := new(parser.ConceptParser).Parse(getContent(params.TextDocument.URI), file)
		for _, conStep := range steps {
			for _, step := range conStep.ConceptSteps {
				if step.LineNo-1 == params.Position.Line {
					return step, nil
				}
			}
		}
	}
	return nil, nil
}

func getNewStepName(params lsp.RenameParams, step *gauge.Step) string {
	newName := strings.TrimSpace(strings.TrimPrefix(params.NewName, "*"))
	if step.HasInlineTable {
		newName = fmt.Sprintf("%s <%s>", newName, gauge.TableArg)
	}
	return newName
}

func addSrcWorkspaceEdits(result *lsp.WorkspaceEdit, filesChanges []*gm.FileChanges) error {
	for _, fileChange := range filesChanges {
		uri := util.ConvertPathToURI(fileChange.FileName)
		for _, diff := range fileChange.Diffs {
			textEdit := lsp.TextEdit{
				NewText: diff.Content,
				Range: lsp.Range{
					Start: lsp.Position{Line: int(diff.Span.Start - 1), Character: int(diff.Span.StartChar)},
					End:   lsp.Position{Line: int(diff.Span.End - 1), Character: int(diff.Span.EndChar)},
				},
			}
			result.Changes[string(uri)] = append(result.Changes[string(uri)], textEdit)
		}
	}
	return nil
}

func addWorkspaceEdits(result *lsp.WorkspaceEdit, filesChanged map[string]string) error {
	diskFileCache := &files{cache: make(map[lsp.DocumentURI][]string)}
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
		result.Changes[string(uri)] = append(result.Changes[string(uri)], textEdit)
	}
	return nil
}
