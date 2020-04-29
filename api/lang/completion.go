/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"encoding/json"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type insertTextFormat int

const (
	text          insertTextFormat = 1
	snippet       insertTextFormat = 2
	concept                        = "Concept"
	step                           = "Step"
	tag                            = "Tag"
	tagIdentifier                  = "tags:"
	emptyString                    = ""
	colon                          = ":"
	comma                          = ","
)

type completionItem struct {
	lsp.CompletionItem
	InsertTextFormat insertTextFormat `json:"insertTextFormat,omitempty"`
}

type completionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []completionItem `json:"items"`
}

func completion(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	line := getLine(params.TextDocument.URI, params.Position.Line)
	pLine := line
	if len(line) > params.Position.Character {
		pLine = line[:params.Position.Character]
	}
	if isInTagsContext(params.Position.Line, params.TextDocument.URI) {
		return tagsCompletion(line, pLine, params)
	}
	if !isStepCompletion(pLine, params.Position.Character) {
		return completionList{IsIncomplete: false, Items: []completionItem{}}, nil
	}
	if inParameterContext(line, params.Position.Character) {
		return paramCompletion(line, pLine, params)
	}
	v, err := stepCompletion(line, pLine, params)
	if err != nil && v != nil {
		// there were errors, but gauge will return completions on a best effort promise.
		logError(req, err.Error())
		return v, nil
	}
	return v, err
}

func isInTagsContext(line int, uri lsp.DocumentURI) bool {
	if strings.HasPrefix(strings.ToLower(strings.Join(strings.Fields(getLine(uri, line)), "")), tagIdentifier) {
		return true
	} else if line != 0 && (endsWithComma(getLine(uri, line-1)) && isInTagsContext(line-1, uri)) {
		return true
	}
	return false
}

func endsWithComma(line string) bool {
	return strings.HasSuffix(strings.TrimSpace(line), comma)
}

func isStepCompletion(line string, character int) bool {
	if character == 0 {
		return false
	}
	if !strings.HasPrefix(strings.TrimSpace(line), "*") {
		return false
	}
	return true
}

func inParameterContext(line string, charPos int) bool {
	pl := line
	if len(line) > charPos {
		pl = line[:charPos]
	}
	lineAfterCharPos := strings.SplitAfter(pl, "*")
	if len(lineAfterCharPos) == 1 {
		return false
	}
	l := strings.TrimPrefix(strings.SplitAfter(pl, "*")[1], " ")
	var stack string
	for _, value := range l {
		if string(value) == "<" {
			stack = stack + string(value)
		}
		if string(value) == ">" && len(stack) != 0 && stack[len(stack)-1:] == "<" {
			stack = stack[:len(stack)-1]
		}
		if string(value) == "\"" {
			if len(stack) != 0 && stack[len(stack)-1:] == "\"" {
				stack = stack[:len(stack)-1]
			} else {
				stack = stack + string(value)
			}
		}
	}
	return len(stack) != 0
}

func resolveCompletion(req *jsonrpc2.Request) (interface{}, error) {
	var params completionItem
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	return params, nil
}

func getEditRange(index int, position lsp.Position, pLine, line string, endSeparator string) lsp.Range {
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
