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

	"regexp"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type insertTextFormat int

const (
	text    insertTextFormat = 1
	snippet insertTextFormat = 2
	concept                  = "Concept"
	step                     = "Step"
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
	if !isStepCompletion(line, params.Position.Character) {
		return nil, nil
	}
	list := completionList{IsIncomplete: false, Items: []completionItem{}}
	startPos, endPos := getEditPosition(line, params.Position)
	prefix := getPrefix(line)
	var givenArgs []gauge.StepArg
	if startPos.Character != endPos.Character {
		var err error
		givenArgs, err = getStepArgs(line[:params.Position.Character])
		if err != nil {
			return nil, err
		}
	}
	for _, c := range provider.Concepts() {
		fText := getFilterText(c.StepValue.StepValue, c.StepValue.Parameters, givenArgs)
		cText := prefix + addPlaceHolders(c.StepValue.StepValue, c.StepValue.Parameters)
		list.Items = append(list.Items, newCompletionItem(c.StepValue.ParameterizedStepValue, cText, concept, fText, startPos, endPos))
	}
	for _, s := range provider.Steps() {
		fText := getFilterText(s.StepValue, s.Args, givenArgs)
		cText := prefix + addPlaceHolders(s.StepValue, s.Args)
		list.Items = append(list.Items, newCompletionItem(s.ParameterizedStepValue, cText, step, fText, startPos, endPos))
	}
	return list, nil
}

func getStepArgs(line string) ([]gauge.StepArg, error) {
	var givenArgs []gauge.StepArg
	if line != "" {
		specParser := new(parser.SpecParser)
		tokens, errs := specParser.GenerateTokens(line, "")
		if len(errs) > 0 {
			return nil, fmt.Errorf("Unable to parse text entered")
		}
		var err error
		givenArgs, err = parser.ExtractStepArgsFromToken(tokens[0])
		if err != nil {
			return nil, fmt.Errorf("Unable to parse text entered")
		}
	}
	return givenArgs, nil

}
func isStepCompletion(line string, character int) bool {
	if character == 0 {
		return false
	}
	if !strings.HasPrefix(strings.TrimSpace(line), "*") {
		return false
	}
	return !inParameterContext(line, character)
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

func getFilterText(text string, params []string, givenArgs []gauge.StepArg) string {
	if len(params) > 0 {
		for i, p := range params {
			if len(givenArgs) > i {
				if givenArgs[i].ArgType == gauge.Static {
					text = strings.Replace(text, "{}", fmt.Sprintf("\"%s\"", givenArgs[i].ArgValue()), 1)
				} else {
					text = strings.Replace(text, "{}", fmt.Sprintf("<%s>", givenArgs[i].ArgValue()), 1)
				}
			} else {
				text = strings.Replace(text, "{}", fmt.Sprintf("<%s>", p), 1)
			}
		}
	}
	return text
}

func getEditPosition(line string, cursorPos lsp.Position) (lsp.Position, lsp.Position) {
	start := 1
	loc := regexp.MustCompile(`^\*(\s*)`).FindIndex([]byte(line))
	if loc != nil {
		start = loc[1]
	}
	end := len(line)
	if end < 2 {
		end = 1
	}
	if end < cursorPos.Character {
		end = cursorPos.Character
	}
	return lsp.Position{Line: cursorPos.Line, Character: start}, lsp.Position{Line: cursorPos.Line, Character: end}
}

func getPrefix(line string) string {
	if strings.HasPrefix(line, "* ") {
		return ""
	}
	return " "
}

func resolveCompletion(req *jsonrpc2.Request) (interface{}, error) {
	var params completionItem
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}
	return params, nil
}

func newCompletionItem(stepText, text, kind, fText string, startPos, endPos lsp.Position) completionItem {
	return completionItem{
		CompletionItem: lsp.CompletionItem{
			Label:      stepText,
			Detail:     kind,
			Kind:       lsp.CIKFunction,
			TextEdit:   lsp.TextEdit{Range: lsp.Range{Start: startPos, End: endPos}, NewText: text},
			FilterText: fText,
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
