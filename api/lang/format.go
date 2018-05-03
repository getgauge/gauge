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

	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func format(request *jsonrpc2.Request) (interface{}, error) {
	var params lsp.DocumentFormattingParams
	if err := json.Unmarshal(*request.Params, &params); err != nil {
		return nil, err
	}
	logDebug(request, "LangServer: request received : Type: Format Document URI: %s", params.TextDocument.URI)
	file := util.ConvertURItoFilePath(params.TextDocument.URI)
	if util.IsValidSpecExtension(file) {
		spec, parseResult, err := new(parser.SpecParser).Parse(getContent(params.TextDocument.URI), gauge.NewConceptDictionary(), file)
		if err != nil {
			return nil, err
		}
		if !parseResult.Ok {
			return nil, fmt.Errorf("failed to format document. Fix all the problems first")
		}
		newString := formatter.FormatSpecification(spec)
		oldString := getContent(params.TextDocument.URI)
		textEdit := createTextEdit(newString, 0, 0, len(strings.Split(oldString, "\n")), len(oldString))
		return []lsp.TextEdit{textEdit}, nil
	}
	return nil, fmt.Errorf("failed to format document. %s is not a valid spec file", file)
}
