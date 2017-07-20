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
	"fmt"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type insertTextFormat int

const (
	text     insertTextFormat = 1
	snippet  insertTextFormat = 2
	asterisk byte             = 42
	concept                   = "Concept"
	step                      = "Step"
)

type completionItem struct {
	lsp.CompletionItem
	InsertTextFormat insertTextFormat `json:"insertTextFormat,omitempty"`
}

type completionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []completionItem `json:"items"`
}

func isStepCompletion(params lsp.TextDocumentPositionParams) bool {
	line := getLine(params.TextDocument.URI, params.Position.Line)
	if params.Position.Character == 0 {
		return false
	}
	if !strings.HasPrefix(strings.TrimSpace(line), "*") {
		return false
	}
	return !inParameterContext(line, params.Position.Character)
}

func inParameterContext(line string, charPos int) bool {
	lineAfterCharPos := strings.SplitAfter(line[:charPos], "*")
	if len(lineAfterCharPos) == 1 {
		return false
	}
	l := strings.TrimPrefix(strings.SplitAfter(line[:charPos], "*")[1], " ")
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

func completion(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	if !isStepCompletion(params) {
		return nil, nil
	}
	list := completionList{IsIncomplete: false, Items: []completionItem{}}
	prefix := getPrefix(params)
	for _, c := range provider.Concepts() {
		cText := addPlaceHolders(c.StepValue.StepValue, c.StepValue.Parameters)
		list.Items = append(list.Items, newCompletionItem(c.StepValue.ParameterizedStepValue, cText, prefix, concept, params.Position))
	}
	for _, s := range provider.Steps() {
		cText := addPlaceHolders(s.StepValue, s.Args)
		list.Items = append(list.Items, newCompletionItem(s.ParameterizedStepValue, cText, prefix, step, params.Position))
	}
	return list, nil
}

func getPrefix(p lsp.TextDocumentPositionParams) string {
	if p.Position.Character > 0 && getChar(p.TextDocument.URI, p.Position.Line, p.Position.Character-1) == asterisk {
		return " "
	}
	return ""
}

func resolveCompletion(req *jsonrpc2.Request) (interface{}, error) {
	var params completionItem
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	return params, nil
}

func newCompletionItem(stepText, text, prefix, kind string, p lsp.Position) completionItem {
	return completionItem{
		CompletionItem: lsp.CompletionItem{
			Label:      stepText,
			Detail:     kind,
			Kind:       lsp.CIKFunction,
			TextEdit:   lsp.TextEdit{Range: lsp.Range{Start: p, End: p}, NewText: prefix + text},
			FilterText: prefix + stepText,
		},
		InsertTextFormat: snippet,
	}
}

func addPlaceHolders(text string, args []string) string {
	text = strings.Replace(text, "{}", "\"{}\"", -1)
	for i, v := range args {
		value := i + 1
		if value == len(args) {
			value = 0
		}
		text = strings.Replace(text, "{}", fmt.Sprintf("${%d:%s}", value, v), 1)
	}
	return text
}
