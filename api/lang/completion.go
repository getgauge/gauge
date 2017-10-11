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
	return stepCompletion(line, pLine, params)
}

func isInTagsContext(line int, uri string) bool {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(getLine(uri, line))), tagIdentifier) {
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
