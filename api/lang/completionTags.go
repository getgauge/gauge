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

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func tagsCompletion(line string, pLine string, params lsp.TextDocumentPositionParams) (interface{}, error) {
	list := completionList{IsIncomplete: false, Items: []completionItem{}}
	suffix, editRange := getTagsEditRange(line, pLine, params.Position)
	for _, t := range provider.Tags() {
		item := completionItem{
			InsertTextFormat: text,
			CompletionItem: lsp.CompletionItem{
				Label:      t,
				FilterText: t + suffix,
				Detail:     tag,
				Kind:       lsp.CIKVariable,
				TextEdit:   lsp.TextEdit{Range: editRange, NewText: " " + t + suffix},
			},
		}
		list.Items = append(list.Items, item)
	}
	return list, nil
}

func getTagsEditRange(line, pLine string, position lsp.Position) (string, lsp.Range) {
	var editRange lsp.Range
	suffix := emptyString
	commaIndex := strings.LastIndex(pLine, comma)
	colonIndex := strings.LastIndex(pLine, colon)
	if commaIndex > colonIndex {
		editRange = getRange(commaIndex, position.Character+1, position, pLine, line, comma)
	} else {
		editRange = getRange(colonIndex, position.Character+1, position, pLine, line, colon)
	}
	if len(line) >= position.Character+1 && len(line) != editRange.End.Character {
		suffix = comma
	}
	return suffix, editRange
}
