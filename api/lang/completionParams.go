package lang

import (
	"net/url"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func paramCompletion(line, pLine string, params lsp.TextDocumentPositionParams) (interface{}, error) {
	list := completionList{IsIncomplete: false, Items: []completionItem{}}
	editRange := getParamEditRange(line, pLine, params.Position)
	fileUrl, _ := url.Parse(params.TextDocument.URI)
	for _, param := range provider.Params(fileUrl.Path) {
		list.Items = append(list.Items, completionItem{
			CompletionItem: lsp.CompletionItem{
				Label:    param,
				Detail:   "Param",
				Kind:     lsp.CIKVariable,
				TextEdit: lsp.TextEdit{Range: editRange, NewText: param},
			},
			InsertTextFormat: text,
		})
	}
	return list, nil
}

func getParamEditRange(line, pLine string, position lsp.Position) lsp.Range {
	getRange := func(index int, endSeparator string) lsp.Range {
		start := lsp.Position{Line: position.Line, Character: index + 1}
		endIndex := start.Character
		if len(line) >= position.Character  {
			lineAfterCursor := line[position.Character:]
			endIndex = strings.Index(lineAfterCursor, endSeparator)
		}
		if endIndex == -1 {
			endIndex = len(line)
		} else {
			endIndex = len(pLine) + endIndex
		}
		end := lsp.Position{Line: position.Line, Character: endIndex}
		return lsp.Range{Start:start,End: end}

	}
	quoteIndex := strings.LastIndex(pLine, "\"")
	bracIndex := strings.LastIndex(pLine, "<")
	if quoteIndex > bracIndex {
		return getRange(quoteIndex, "\"")
	} else {
		return getRange(bracIndex, ">")
	}
}
