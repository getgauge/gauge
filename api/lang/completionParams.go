/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func paramCompletion(line, pLine string, params lsp.TextDocumentPositionParams) (interface{}, error) {
	list := completionList{IsIncomplete: false, Items: []completionItem{}}
	argType, suffix, editRange := getParamArgTypeAndEditRange(line, pLine, params.Position)
	file := util.ConvertURItoFilePath(params.TextDocument.URI)
	for _, param := range provider.Params(file, argType) {
		if !shouldAddParam(param.ArgType) {
			continue
		}
		argValue := param.ArgValue()
		list.Items = append(list.Items, completionItem{
			CompletionItem: lsp.CompletionItem{
				Label:      argValue,
				FilterText: argValue + suffix,
				Detail:     string(argType),
				Kind:       lsp.CIKVariable,
				TextEdit:   &lsp.TextEdit{Range: editRange, NewText: argValue + suffix},
			},
			InsertTextFormat: text,
		})
	}
	return list, nil
}

func shouldAddParam(argType gauge.ArgType) bool {
	return argType != gauge.TableArg
}

func getParamArgTypeAndEditRange(line, pLine string, position lsp.Position) (gauge.ArgType, string, lsp.Range) {
	quoteIndex := strings.LastIndex(pLine, "\"")
	bracIndex := strings.LastIndex(pLine, "<")
	if quoteIndex > bracIndex {
		return gauge.Static, "\"", getEditRange(quoteIndex, position, pLine, line, "\"")
	} else {
		return gauge.Dynamic, ">", getEditRange(bracIndex, position, pLine, line, ">")
	}
}
