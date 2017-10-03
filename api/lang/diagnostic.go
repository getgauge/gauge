package lang

import (
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func createDiagnostics(uri string) []lsp.Diagnostic {
	file := util.ConvertURItoFilePath(uri)
	var res *parser.ParseResult
	if util.IsConcept(file) {
		res = validateConcept(uri, file)
	} else {
		_, res = new(parser.SpecParser).Parse(getContent(uri), provider.GetConceptDictionary(), file)
	}
	return createDiagnosticsFrom(res, uri)
}

func validateConcept(uri string, file string) *parser.ParseResult {
	dictionary := provider.GetConceptDictionary()
	for _, cpt := range dictionary.ConceptsMap {
		if cpt.FileName == file {
			delete(dictionary.ConceptsMap, cpt.ConceptStep.Value)
		}
	}
	cpts, res := new(parser.ConceptParser).Parse(getContent(uri), file)
	if errs := parser.AddConcept(cpts, file, dictionary); len(errs) > 0 {
		res.ParseErrors = append(res.ParseErrors, errs...)
	}
	vRes := parser.ValidateConcepts(dictionary)
	res.ParseErrors = append(res.ParseErrors, vRes.ParseErrors...)
	res.Warnings = append(res.Warnings, vRes.Warnings...)
	return res
}

func createDiagnosticsFrom(res *parser.ParseResult, uri string) []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, 0)
	for _, err := range res.ParseErrors {
		diagnostics = append(diagnostics, createDiagnostic(err.Message, err.LineNo-1, 1, uri))
	}
	for _, warning := range res.Warnings {
		diagnostics = append(diagnostics, createDiagnostic(warning.Message, warning.LineNo-1, 2, uri))
	}
	return diagnostics
}

func createDiagnostic(message string, line int, severity lsp.DiagnosticSeverity, uri string) lsp.Diagnostic {
	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: line, Character: 0},
			End:   lsp.Position{Line: line, Character: len(getLine(uri, line))},
		},
		Message:  message,
		Severity: severity,
	}
}
