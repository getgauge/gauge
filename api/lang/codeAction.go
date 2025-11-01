/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

const (
	generateStepCommand    = "gauge.generate.step"
	generateStubTitle      = "Create step implementation"
	generateConceptCommand = "gauge.generate.concept"
	generateConceptTitle   = "Create concept"
)

func codeActions(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.CodeActionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse request %v", err)
	}
	return getSpecCodeAction(params)
}

func getSpecCodeAction(params lsp.CodeActionParams) ([]lsp.Command, error) {
	var actions []lsp.Command
	line := params.Range.Start.Line
	for _, d := range params.Context.Diagnostics {
		if d.Code != "" {
			actions = append(actions, createCodeAction(generateStepCommand, generateStubTitle, []interface{}{d.Code}))
			cptInfo, err := createConceptInfo(params.TextDocument.URI, line)
			if err != nil {
				return nil, err
			}
			if cptInfo != nil {
				actions = append(actions, createCodeAction(generateConceptCommand, generateConceptTitle, []interface{}{cptInfo}))
			}
		}
	}
	return actions, nil
}

func createConceptInfo(uri lsp.DocumentURI, line int) (interface{}, error) {
	file := util.ConvertURItoFilePath(uri)
	linetext := getLine(uri, line)
	var stepValue *gauge.StepValue
	if util.IsConcept(file) {
		steps, _ := new(parser.ConceptParser).Parse(getContent(uri), file)
		for _, conStep := range steps {
			for _, step := range conStep.ConceptSteps {
				if step.LineNo-1 == line {
					stepValue = extractStepValueAndParams(step, linetext)
				}
			}
		}
	}
	if util.IsSpec(file) {
		spec, res, err := new(parser.SpecParser).Parse(getContent(uri), &gauge.ConceptDictionary{}, file)
		if err != nil {
			return nil, err
		}
		if !res.Ok {
			return nil, fmt.Errorf("parsing failed for %s. %s", uri, res.Errors())
		}
		for _, step := range spec.Steps() {
			if step.LineNo-1 == line {
				stepValue = extractStepValueAndParams(step, linetext)
			}
		}
	}
	if stepValue != nil {
		count := strings.Count(stepValue.StepValue, "{}")
		for i := 0; i < count; i++ {
			stepValue.StepValue = strings.Replace(stepValue.StepValue, "{}", fmt.Sprintf("<arg%d>", i), 1)
		}
		cptName := strings.ReplaceAll(stepValue.StepValue, "*", "")
		return concpetInfo{
			ConceptName: fmt.Sprintf("# %s\n* ", strings.TrimSpace(cptName)),
		}, nil
	}
	return nil, nil
}

func extractStepValueAndParams(step *gauge.Step, linetext string) *gauge.StepValue {
	var stepValue *gauge.StepValue
	if step.HasInlineTable {
		stepValue, _ = parser.ExtractStepValueAndParams(step.LineText, true)
	} else {
		stepValue, _ = parser.ExtractStepValueAndParams(linetext, false)
	}
	return stepValue
}

func createCodeAction(command, titlle string, params []interface{}) lsp.Command {
	return lsp.Command{
		Command:   command,
		Title:     titlle,
		Arguments: params,
	}
}
