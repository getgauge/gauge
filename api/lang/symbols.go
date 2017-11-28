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
	spec, parseResult := new(parser.SpecParser).Parse(content, gauge.NewConceptDictionary(), file)
	if !parseResult.Ok {
		return nil, fmt.Errorf("parsing failed")
	}
	var symbols = make([]lsp.SymbolInformation, 0)
	symbols = append(symbols, lsp.SymbolInformation{
		ContainerName: file,
		Name:          spec.Heading.Value,
		Kind:          lsp.SKClass,
		Location: lsp.Location{
			URI: params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{Line: spec.Heading.LineNo, Character: 0},
				End:   lsp.Position{Line: spec.Heading.LineNo, Character: len(spec.Heading.Value)},
			},
		},
	})
	for _, scn := range spec.Scenarios {
		symbols = append(symbols, lsp.SymbolInformation{
			ContainerName: file,
			Name:          scn.Heading.Value,
			Kind:          lsp.SKFunction,
			Location: lsp.Location{
				URI: params.TextDocument.URI,
				Range: lsp.Range{
					Start: lsp.Position{Line: scn.Heading.LineNo, Character: 0},
					End:   lsp.Position{Line: scn.Heading.LineNo, Character: len(scn.Heading.Value)},
				},
			},
		})
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
	var symbols = make([]lsp.SymbolInformation, 0)
	specDetails := provider.GetAvailableSpecDetails([]string{})
	for _, specDetail := range specDetails {
		if !specDetail.HasSpec() {
			continue
		}
		spec := specDetail.Spec
		if strings.HasPrefix(strings.ToLower(specDetail.Spec.Heading.Value), strings.ToLower(params.Query)) {
			symbols = append(symbols, lsp.SymbolInformation{
				ContainerName: spec.FileName,
				Name:          spec.Heading.Value,
				Kind:          lsp.SKClass,
				Location: lsp.Location{
					URI: util.ConvertPathToURI(spec.FileName),
					Range: lsp.Range{
						Start: lsp.Position{Line: spec.Heading.LineNo, Character: 0},
						End:   lsp.Position{Line: spec.Heading.LineNo, Character: len(spec.Heading.Value)},
					},
				},
			})
		}
		for _, scn := range spec.Scenarios {
			if strings.HasPrefix(strings.ToLower(scn.Heading.Value), strings.ToLower(params.Query)) {
				symbols = append(symbols, lsp.SymbolInformation{
					ContainerName: spec.FileName,
					Name:          scn.Heading.Value,
					Kind:          lsp.SKFunction,
					Location: lsp.Location{
						URI: util.ConvertPathToURI(spec.FileName),
						Range: lsp.Range{
							Start: lsp.Position{Line: scn.Heading.LineNo, Character: 0},
							End:   lsp.Position{Line: scn.Heading.LineNo, Character: len(scn.Heading.Value)},
						},
					},
				})
			}
		}
	}
	return symbols, nil
}
