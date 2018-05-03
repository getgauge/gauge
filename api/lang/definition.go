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

	"github.com/getgauge/gauge/util"

	"fmt"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func definition(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.TextDocumentPositionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	fileContent := getContent(params.TextDocument.URI)
	if util.IsConcept(util.ConvertURItoFilePath(params.TextDocument.URI)) {
		concepts, _ := new(parser.ConceptParser).Parse(fileContent, "")
		for _, concept := range concepts {
			for _, step := range concept.ConceptSteps {
				if (step.LineNo - 1) == params.Position.Line {
					return search(step)
				}
			}
		}
	} else {
		spec, _ := new(parser.SpecParser).ParseSpecText(fileContent, "")
		for _, item := range spec.AllItems() {
			if item.Kind() == gauge.StepKind {
				step := item.(*gauge.Step)
				if (step.LineNo - 1) == params.Position.Line {
					return search(step)
				}
			}
		}
	}
	return nil, nil
}

func search(step *gauge.Step) (interface{}, error) {
	if loc, _ := searchConcept(step); loc != nil {
		return loc, nil
	}
	return searchStep(step)

}

func searchStep(step *gauge.Step) (interface{}, error) {
	if lRunner.runner == nil {
		return nil, nil
	}
	responseMessage, err := getStepNameResponse(step.Value)
	if err != nil {
		return nil, err
	}
	if responseMessage == nil || !(responseMessage.GetIsStepPresent()) {
		return nil, fmt.Errorf("Step implementation not found for step : %s", step.Value)
	}
	return getLspLocationForStep(responseMessage.GetFileName(), responseMessage.GetSpan()), nil
}

func searchConcept(step *gauge.Step) (interface{}, error) {
	if concept := provider.SearchConceptDictionary(step.Value); concept != nil {
		return getLspLocationForConcept(concept.FileName, concept.ConceptStep.LineNo)
	}
	return nil, nil
}

func getLspLocationForStep(fileName string, span *gauge_messages.Span) lsp.Location {
	return lsp.Location{
		URI: util.ConvertPathToURI(fileName),
		Range: lsp.Range{
			Start: lsp.Position{Line: int(span.Start - 1), Character: int(span.StartChar)},
			End:   lsp.Position{Line: int(span.End - 1), Character: int(span.EndChar)},
		},
	}
}

func getLspLocationForConcept(fileName string, lineNumber int) (interface{}, error) {
	uri := util.ConvertPathToURI(fileName)
	var endPos int
	lineNo := lineNumber - 1
	if isOpen(uri) {
		endPos = len(getLine(uri, lineNo))
	} else {
		contents, err := common.ReadFileContents(fileName)
		if err != nil {
			return nil, err
		}
		lines := util.GetLinesFromText(contents)
		if len(lines) < lineNo {
			return nil, fmt.Errorf("unable to read line %d from disk", lineNo+1)
		}
		endPos = len(lines[lineNo])
	}
	return lsp.Location{
		URI: util.ConvertPathToURI(fileName),
		Range: lsp.Range{
			Start: lsp.Position{Line: lineNo, Character: 0},
			End:   lsp.Position{Line: lineNo, Character: endPos},
		},
	}, nil
}
