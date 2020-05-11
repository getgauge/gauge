/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func documentSymbols(req *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DocumentSymbolParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to parse request %v", err)
	}
	file := util.ConvertURItoFilePath(params.TextDocument.URI)
	content := getContent(params.TextDocument.URI)
	if util.IsConcept(file) {
		return getConceptSymbols(content, file), nil
	}
	spec, parseResult, err := new(parser.SpecParser).Parse(content, gauge.NewConceptDictionary(), file)
	if err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("failed to parse request %v", err)
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
