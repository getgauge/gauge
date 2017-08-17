package lang

import (
	"net/url"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func paramCompletion(line, pLine string, params lsp.TextDocumentPositionParams) (interface{}, error) {
	list := completionList{IsIncomplete: false, Items: []completionItem{}}
	argType, suffix, editRange := getParamArgTypeAndEditRange(line, pLine, params.Position)
	fileUrl, _ := url.Parse(params.TextDocument.URI)
	for _, param := range provider.Params(fileUrl.Path) {
		if !shouldAddParam(param.ArgType, argType) {
			continue
		}
		argValue := param.ArgValue()
		list.Items = append(list.Items, completionItem{
			CompletionItem: lsp.CompletionItem{
				Label:      argValue,
				FilterText: argValue + suffix,
				Detail:     argType,
				Kind:       lsp.CIKVariable,
				TextEdit:   lsp.TextEdit{Range: editRange, NewText: argValue + suffix},
			},
			InsertTextFormat: text,
		})
	}
	return list, nil
}

func shouldAddParam(argType gauge.ArgType, wantArgType string) bool {
	if wantArgType == "static" {
		return argType == gauge.Static
	} else {
		return argType != gauge.Static && argType != gauge.TableArg
	}
}

func getParamArgTypeAndEditRange(line, pLine string, position lsp.Position) (string, string, lsp.Range) {
	getRange := func(index int, endSeparator string) lsp.Range {
		start := lsp.Position{Line: position.Line, Character: index + 1}
		endIndex := start.Character
		if len(line) >= position.Character {
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
	quoteIndex := strings.LastIndex(pLine, "\"")
	bracIndex := strings.LastIndex(pLine, "<")
	if quoteIndex > bracIndex {
		return "static", "\"", getRange(quoteIndex, "\"")
	} else {
		return "dynamic", ">", getRange(bracIndex, ">")
	}
}
