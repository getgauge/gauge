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
			return nil, fmt.Errorf("failed to format %s. Fix all the problems first", file)
		}
		newString := formatter.FormatSpecification(spec)
		oldString := getContent(params.TextDocument.URI)
		textEdit := createTextEdit(newString, 0, 0, len(strings.Split(oldString, "\n")), len(oldString))
		return []lsp.TextEdit{textEdit}, nil
	} else if util.IsValidConceptExtension(file) {
		conceptsDictionary := gauge.NewConceptDictionary()
		conceptSteps, parseResult := new(parser.ConceptParser).Parse(getContent(params.TextDocument.URI), file)
		if !parseResult.Ok {
			return nil, fmt.Errorf("failed to format %s. Fix all the problems first", file)
		}
		parseErrs, err := parser.AddConcept(conceptSteps, file, conceptsDictionary)
		if err != nil {
			return nil, err
		}
		if len(parseErrs) > 0 {
			return nil, fmt.Errorf("failed to format %s. Fix all the problems first", file)
		}
		conceptMap := formatter.FormatConcepts(conceptsDictionary)
		newString := conceptMap[file]
		oldString := getContent(params.TextDocument.URI)
		textEdit := createTextEdit(newString, 0, 0, len(strings.Split(oldString, "\n")), len(oldString))
		return []lsp.TextEdit{textEdit}, nil
	}
	return nil, fmt.Errorf("failed to format %s. Not a valid spec/cpt file", file)
}
