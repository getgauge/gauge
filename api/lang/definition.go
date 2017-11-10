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

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
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
	if loc := searchConcept(step); loc != nil {
		return loc, nil
	}
	return searchStep(step)

}

func searchStep(step *gauge.Step) (interface{}, error) {
	if lRunner.runner == nil {
		return nil,nil
	}
	stepNameMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_StepNameRequest, StepNameRequest: &gauge_messages.StepNameRequest{StepValue: step.Value}}
	responseMessage, err := conn.GetResponseForMessageWithTimeout(stepNameMessage, lRunner.runner.Connection(), config.RunnerRequestTimeout())
	if err != nil {
		logger.APILog.Infof("%s", err.Error())
		return nil, err
	}
	if responseMessage == nil || !(responseMessage.GetStepNameResponse().GetIsStepPresent()) {
		logger.APILog.Debugf("Step implementation not found for step : %s", step.Value)
		return nil, fmt.Errorf("Step implementation not found for step : %s", step.Value)
	}
	return getLspLocation(responseMessage.GetStepNameResponse().GetFileName(), int(responseMessage.GetStepNameResponse().GetLineNumber())), nil
}

func searchConcept(step *gauge.Step) interface{} {
	if concept := provider.SearchConceptDictionary(step.Value); concept != nil {
		return getLspLocation(concept.FileName, concept.ConceptStep.LineNo)
	}
	return nil
}

func getLspLocation(fileName string, lineNumber int) lsp.Location {
	return lsp.Location{
		URI: util.ConvertPathToURI(fileName),
		Range: lsp.Range{
			Start: lsp.Position{Line: lineNumber - 1, Character: 0},
			End:   lsp.Position{Line: lineNumber - 1, Character: 0},
		},
	}
}
