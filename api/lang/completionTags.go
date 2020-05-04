/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
				SortText:   "a" + t,
				Label:      t,
				FilterText: t + suffix,
				Detail:     tag,
				Kind:       lsp.CIKVariable,
				TextEdit:   &lsp.TextEdit{Range: editRange, NewText: " " + t + suffix},
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
		editRange = getEditRange(commaIndex, position, pLine, line, comma)
	} else {
		editRange = getEditRange(colonIndex, position, pLine, line, comma)
	}
	if len(line) >= position.Character+1 && len(line) != editRange.End.Character {
		suffix = comma
	}
	return suffix, editRange
}
