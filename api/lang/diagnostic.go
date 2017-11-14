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
	"strconv"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/validation"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
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
	if lRunner.runner == nil {
		return
	}
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
		if strings.Contains(err.Message, "Duplicate concept definition found") {
			diagnostics = createDiagnosticsForDuplicateConcepts(err, diagnostics, uri)
		} else {
			diagnostics = append(diagnostics, createDiagnostic(err.Message, err.LineNo-1, 1, uri))
		}
	}
	for _, warning := range res.Warnings {
		diagnostics = append(diagnostics, createDiagnostic(warning.Message, warning.LineNo-1, 2, uri))
	}
	return diagnostics
}

func createDiagnosticsForDuplicateConcepts(err parser.ParseError, diagnostics []lsp.Diagnostic, uri string) []lsp.Diagnostic {
	values := strings.Split(err.Message, "\t")[1:]
	for _, val := range values {
		arr := strings.Split(val, ":")
		l, error := strconv.Atoi(strings.TrimSuffix(arr[len(arr)-1], "\n"))
		if error != nil {
			logger.Errorf("Error while getting line number for duplicate concepts: ", error)
		}
		diagnostics = append(diagnostics, createDiagnostic(err.Message, l-1, 1, uri))
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
