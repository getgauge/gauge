package lang

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getgauge/gauge/formatter"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
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
	logger.APILog.Debugf("LangServer: request received : Type: Format Document URI: %s", params.TextDocument.URI)
	file := convertURItoFilePath(params.TextDocument.URI)
	if util.IsValidSpecExtension(file) {
		spec, err := parseSpec(params.TextDocument.URI, file)
		if err != nil {
			return nil, err
		}
		newString := formatter.FormatSpecification(spec)
		return createTextEdit(getContent(params.TextDocument.URI), newString), nil
	}
	return nil, fmt.Errorf("file %s is not a valid spec file", file)
}

func createTextEdit(oldContent string, newString string) []lsp.TextEdit {
	return []lsp.TextEdit{
		{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      0,
					Character: 0,
				},
				End: lsp.Position{
					Line:      len(strings.Split(oldContent, "\n")),
					Character: len(oldContent),
				},
			},
			NewText: newString,
		},
	}
}

func parseSpec(uri, specFile string) (*gauge.Specification, error) {
	spec, parseResult := new(parser.SpecParser).Parse(getContent(uri), gauge.NewConceptDictionary(), specFile)
	if !parseResult.Ok {
		err := parseResult.ParseErrors[0]
		return nil, fmt.Errorf("ParseError : %s, Location : %s:%d", err.Message, err.FileName, err.LineNo)
	}
	return spec, nil
}
