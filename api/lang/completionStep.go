/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func stepCompletion(line, pLine string, params lsp.TextDocumentPositionParams) (interface{}, error) {
	list := completionList{IsIncomplete: false, Items: make([]completionItem, 0)}
	editRange := getStepEditRange(line, params.Position)
	prefix := getPrefix(pLine)
	givenArgs, err := getStepArgs(strings.TrimSpace(pLine))
	if err != nil {
		return nil, err
	}
	for _, c := range provider.Concepts() {
		fText := prefix + getStepFilterText(c.StepValue.StepValue, c.StepValue.Parameters, givenArgs)
		cText := prefix + addPlaceHolders(c.StepValue.StepValue, c.StepValue.Parameters)
		list.Items = append(list.Items, newStepCompletionItem(c.StepValue.ParameterizedStepValue, cText, concept, fText, editRange))
	}
	s, err := allImplementedStepValues()
	allSteps := append(allUsedStepValues(), s...)
	for _, sv := range removeDuplicates(allSteps) {
		fText := prefix + getStepFilterText(sv.StepValue, sv.Args, givenArgs)
		cText := prefix + addPlaceHolders(sv.StepValue, sv.Args)
		list.Items = append(list.Items, newStepCompletionItem(sv.ParameterizedStepValue, cText, step, fText, editRange))
	}
	return list, err
}

func removeDuplicates(steps []gauge.StepValue) []gauge.StepValue {
	encountered := map[string]bool{}
	result := []gauge.StepValue{}
	for _, v := range steps {
		if !encountered[v.StepValue] {
			encountered[v.StepValue] = true
			result = append(result, v)
		}
	}
	return result
}

func allUsedStepValues() []gauge.StepValue {
	var stepValues []gauge.StepValue
	for _, s := range provider.Steps(true) {
		stepValues = append(stepValues, parser.CreateStepValue(s))
	}
	return stepValues
}
func allImplementedStepValues() ([]gauge.StepValue, error) {
	var stepValues []gauge.StepValue
	res, err := getAllStepsResponse()
	if err != nil {
		return stepValues, fmt.Errorf("failed to get steps from runner. %s", err.Error())
	}
	for _, stepText := range res.GetSteps() {
		stepValue, err := parser.ExtractStepValueAndParams(stepText, false)
		if err != nil {
			logError(nil, "Unable to extract StepValueAndParams for step '%s', error : %s", stepText, err.Error())
		} else {
			stepValues = append(stepValues, *stepValue)
		}
	}
	return stepValues, nil
}

func getStepArgs(line string) ([]gauge.StepArg, error) {
	givenArgs := make([]gauge.StepArg, 0)
	if line != "" && strings.TrimSpace(line) != "*" {
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

func getStepFilterText(text string, params []string, givenArgs []gauge.StepArg) string {
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

func getStepEditRange(line string, cursorPos lsp.Position) lsp.Range {
	start := 1
	loc := regexp.MustCompile(`^\s*\*(\s*)`).FindIndex([]byte(line))
	if loc != nil {
		start = loc[1]
	}
	if start > cursorPos.Character {
		start = cursorPos.Character
	}
	end := len(line)
	if end < 2 {
		end = 1
	}
	if end < cursorPos.Character {
		end = cursorPos.Character
	}
	startPos := lsp.Position{Line: cursorPos.Line, Character: start}
	endPos := lsp.Position{Line: cursorPos.Line, Character: end}
	return lsp.Range{Start: startPos, End: endPos}
}

func getPrefix(line string) string {
	if strings.HasPrefix(strings.TrimPrefix(line, " "), "* ") {
		return ""
	}
	return " "
}

func addPlaceHolders(text string, args []string) string {
	text = strings.ReplaceAll(text, "{}", "\"{}\"")
	for i, v := range args {
		value := i + 1
		if value == len(args) {
			value = 0
		}
		text = strings.Replace(text, "{}", fmt.Sprintf("${%d:%s}", value, v), 1)
	}
	return text
}

func newStepCompletionItem(stepText, text, kind, fText string, editRange lsp.Range) completionItem {
	return completionItem{
		CompletionItem: lsp.CompletionItem{
			Label:         stepText,
			Detail:        kind,
			Kind:          lsp.CIKFunction,
			TextEdit:      &lsp.TextEdit{Range: editRange, NewText: text},
			FilterText:    fText,
			Documentation: stepText,
		},
		InsertTextFormat: snippet,
	}
}
