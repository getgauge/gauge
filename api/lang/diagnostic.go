package lang

import (
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func createDiagnostics(uri string) []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, 0)
	file := util.ConvertURItoFilePath(uri)
	_, res := new(parser.SpecParser).Parse(getContent(uri), gauge.NewConceptDictionary(), file)
	if len(res.ParseErrors) > 0 {
		for _, err := range res.ParseErrors {
			diagnostics = append(diagnostics, createDiagnostic(err, uri))
		}
	}
	return diagnostics
}

func createDiagnostic(err parser.ParseError, uri string) lsp.Diagnostic {
	line := err.LineNo - 1
	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: line, Character: 0},
			End:   lsp.Position{Line: line, Character: len(getLine(uri, line))},
		},
		Message:  err.Message,
		Severity: 1,
	}
}
