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

	"fmt"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type insertTextFormat int

const (
	text    insertTextFormat = 1
	snippet insertTextFormat = 2
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
	list := completionList{IsIncomplete: false, Items: []completionItem{}}
	for _, c := range provider.Concepts() {
		list.Items = append(list.Items, completionItem{
			CompletionItem: lsp.CompletionItem{
				Label:    c.StepValue.ParameterizedStepValue,
				Detail:   "Concept",
				Kind:     lsp.CIKFunction,
				TextEdit: lsp.TextEdit{Range: lsp.Range{Start: params.Position, End: params.Position}, NewText: addPlaceHolders(c.StepValue.StepValue, c.StepValue.Parameters)},
			},
			InsertTextFormat: snippet,
		})
	}
	for _, s := range provider.Steps() {
		list.Items = append(list.Items, completionItem{
			CompletionItem: lsp.CompletionItem{
				Label:    s.ParameterizedStepValue,
				Detail:   "Step",
				Kind:     lsp.CIKFunction,
				TextEdit: lsp.TextEdit{Range: lsp.Range{Start: params.Position, End: params.Position}, NewText: addPlaceHolders(s.StepValue, s.Args)},
			},
			InsertTextFormat: snippet,
		})
	}
	return list, nil
}

func resolveCompletion(req *jsonrpc2.Request) (interface{}, error) {
	var params completionItem
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	return params, nil
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
