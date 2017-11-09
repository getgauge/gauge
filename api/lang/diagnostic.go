package lang

import (
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/validation"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/getgauge/gauge/gauge"
)

func createDiagnostics(uri string) []lsp.Diagnostic {
	file := util.ConvertURItoFilePath(uri)
	if util.IsConcept(file) {
		return createDiagnosticsFrom(validateConcept(uri, file), uri)
	} else {
		spec, res := new(parser.SpecParser).Parse(getContent(uri), provider.GetConceptDictionary(), file)
		vRes := validateSpec(spec)
		return append(createDiagnosticsFrom(res, uri), createValidationDiagnostics(vRes, uri)...)
	}
}

func createValidationDiagnostics(errors []validation.StepValidationError, uri string) (diagnostics []lsp.Diagnostic) {
	for _, err := range errors {
		diagnostics = append(diagnostics, createDiagnostic(err.Message(), err.Step().LineNo-1, 1, uri))
	}
	return
}

func validateSpec(spec *gauge.Specification) (vErrors []validation.StepValidationError) {
	v := validation.NewSpecValidator(spec, lRunner.runner, provider.GetConceptDictionary(), []error{}, map[string]error{})
	for _, e := range v.Validate() {
		vErrors = append(vErrors, e.(validation.StepValidationError))
	}
	return
}

func validateConcept(uri string, file string) *parser.ParseResult {
	res := provider.UpdateConceptCache(file, getContent(uri))
	vRes := parser.ValidateConcepts(provider.GetConceptDictionary())
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
