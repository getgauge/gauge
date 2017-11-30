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
	"sort"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func documentSymbols(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DocumentSymbolParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}
	file := util.ConvertURItoFilePath(params.TextDocument.URI)
	content := getContent(params.TextDocument.URI)
	if util.IsConcept(file) {
		return getConceptSymbols(content, file), nil
	}
	spec, parseResult := new(parser.SpecParser).Parse(content, gauge.NewConceptDictionary(), file)
	if !parseResult.Ok {
		return nil, fmt.Errorf("parsing failed for %s. %s", file, parseResult.Errors())
	}
	var symbols = make([]*lsp.SymbolInformation, 0)
	symbols = append(symbols, getSpecSymbol(spec))
	for _, scn := range spec.Scenarios {
		symbols = append(symbols, getScenarioSymbol(scn, file))
	}
	return symbols, nil
}

func workspaceSymbols(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.WorkspaceSymbolParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return nil, err
	}

	if len(params.Query) < 2 {
		return nil, nil
	}

	var specSymbols = make([]*lsp.SymbolInformation, 0)
	var scnSymbols = make([]*lsp.SymbolInformation, 0)
	specDetails := provider.GetAvailableSpecDetails([]string{})
	for _, specDetail := range specDetails {
		if !specDetail.HasSpec() {
			continue
		}
		spec := specDetail.Spec
		if strings.Contains(strings.ToLower(specDetail.Spec.Heading.Value), strings.ToLower(params.Query)) {
			specSymbols = append(specSymbols, getSpecSymbol(spec))
		}
		for _, scn := range spec.Scenarios {
			if strings.Contains(strings.ToLower(scn.Heading.Value), strings.ToLower(params.Query)) {
				scnSymbols = append(scnSymbols, getScenarioSymbol(scn, spec.FileName))
			}
		}
	}
	sort.Sort(byName(specSymbols))
	sort.Sort(byName(scnSymbols))
	return append(specSymbols, scnSymbols...), nil
}

func getSpecSymbol(s *gauge.Specification) *lsp.SymbolInformation {
	return &lsp.SymbolInformation{
		Name: fmt.Sprintf("# %s", s.Heading.Value),
		Kind: lsp.SKNamespace,
		Location: lsp.Location{
			URI: util.ConvertPathToURI(s.FileName),
			Range: lsp.Range{
				Start: lsp.Position{Line: s.Heading.LineNo - 1, Character: 0},
				End:   lsp.Position{Line: s.Heading.LineNo - 1, Character: len(s.Heading.Value)},
			},
		},
	}
}

func getScenarioSymbol(s *gauge.Scenario, path string) *lsp.SymbolInformation {
	return &lsp.SymbolInformation{
		Name: fmt.Sprintf("## %s", s.Heading.Value),
		Kind: lsp.SKNamespace,
		Location: lsp.Location{
			URI: util.ConvertPathToURI(path),
			Range: lsp.Range{
				Start: lsp.Position{Line: s.Heading.LineNo - 1, Character: 0},
				End:   lsp.Position{Line: s.Heading.LineNo - 1, Character: len(s.Heading.Value)},
			},
		},
	}
}

func getConceptSymbols(content, file string) []*lsp.SymbolInformation {
	concepts, _ := new(parser.ConceptParser).Parse(content, file)
	var symbols = make([]*lsp.SymbolInformation, 0)
	for _, cpt := range concepts {
		symbols = append(symbols, &lsp.SymbolInformation{
			Name: fmt.Sprintf("# %s", cpt.LineText),
			Kind: lsp.SKNamespace,
			Location: lsp.Location{
				URI: util.ConvertPathToURI(file),
				Range: lsp.Range{
					Start: lsp.Position{Line: cpt.LineNo - 1, Character: 0},
					End:   lsp.Position{Line: cpt.LineNo - 1, Character: len(cpt.LineText)},
				},
			},
		})
	}
	return symbols
}

type byName []*lsp.SymbolInformation

func (s byName) Len() int {
	return len(s)
}

func (s byName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byName) Less(i, j int) bool {
	return strings.Compare(s[i].Name, s[j].Name) < 0
}
