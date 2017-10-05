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

	"github.com/getgauge/gauge/gauge"
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
	if util.IsConcept(util.ConvertPathToURI(params.TextDocument.URI)) {
		concepts, _ := new(parser.ConceptParser).Parse(fileContent, "")
		for _, concept := range concepts {
			for _, step := range concept.ConceptSteps {
				if loc := searchConcept(step, params.Position.Line); loc != nil {
					return loc, nil
				}
			}
		}
	} else {
		spec, _ := new(parser.SpecParser).ParseSpecText(fileContent, "")
		for _, item := range spec.AllItems() {
			if item.Kind() == gauge.StepKind {
				step := item.(*gauge.Step)
				if loc := searchConcept(step, params.Position.Line); loc != nil {
					return loc, nil
				}
			}
		}
	}
	return nil, nil
}

func searchConcept(step *gauge.Step, lineNo int) interface{} {
	if (step.LineNo - 1) == lineNo {
		if concept := provider.SearchConceptDictionary(step.Value); concept != nil {
			return lsp.Location{
				URI: util.ConvertPathToURI(concept.FileName),
				Range: lsp.Range{
					Start: lsp.Position{Line: concept.ConceptStep.LineNo - 1, Character: 0},
					End:   lsp.Position{Line: concept.ConceptStep.LineNo - 1, Character: 0},
				},
			}
		}
	}
	return nil
}
