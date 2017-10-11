// Copyright 2015 ThoughtWorks, Inc.

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
				TextEdit:   lsp.TextEdit{Range: editRange, NewText: argValue + suffix},
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
		return gauge.Static, "\"", getRange(quoteIndex, position.Character, position, pLine, line, "\"")
	} else {
		return gauge.Dynamic, ">", getRange(bracIndex, position.Character, position, pLine, line, ">")
	}
}

func getRange(index, cursorPosition int, position lsp.Position, pLine, line string, endSeparator string) lsp.Range {
	start := lsp.Position{Line: position.Line, Character: index + 1}
	endIndex := start.Character
	if len(line) >= cursorPosition {
		lineAfterCursor := line[position.Character:]
		endIndex = strings.Index(lineAfterCursor, endSeparator)
	}
	if endIndex == -1 {
		endIndex = len(line)
	} else {
		endIndex = len(pLine) + endIndex + 1
	}
	end := lsp.Position{Line: position.Line, Character: endIndex}
	return lsp.Range{Start: start, End: end}
}
