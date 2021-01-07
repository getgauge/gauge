/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/gauge"
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
	return renameStep(req)
}

func renameStep(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.RenameParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		logDebug(req, "failed to parse rename request %s", err.Error())
		return nil, err
	}

	step, err := getStepToRefactor(params)
	if err != nil {
		return nil, err
	}
	if step == nil {
		return nil, fmt.Errorf("refactoring is supported for steps only")
	}
	newName := getNewStepName(params, step)

	refactortingResult := refactor.GetRefactoringChanges(step.GetLineText(), newName, lRunner.runner, util.GetSpecDirs(), false)
	for _, warning := range refactortingResult.Warnings {
		logWarning(req, warning)
	}
	if !refactortingResult.Success {
		return nil, fmt.Errorf("%s", strings.Join(refactortingResult.Errors, "\t"))
	}
	var result lsp.WorkspaceEdit
	result.Changes = make(map[string][]lsp.TextEdit)
	changes := append(refactortingResult.SpecsChanged, append(refactortingResult.ConceptsChanged, refactortingResult.RunnerFilesChanged...)...)
	if err := addWorkspaceEdits(&result, changes); err != nil {
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

func addWorkspaceEdits(result *lsp.WorkspaceEdit, filesChanges []*gm.FileChanges) error {
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
