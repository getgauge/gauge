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
	"context"
	"fmt"
	"sync"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/validation"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// Diagnostics lock ensures only one goroutine publishes diagnostics at a time.
var diagnosticsLock sync.Mutex

// isInQueue ensures that only one other goroutine waits for the diagnostic lock.
// Since diagnostics are published for all files, multiple threads need not wait to publish diagnostics.
var isInQueue = false

func publishDiagnostics(ctx context.Context, conn jsonrpc2.JSONRPC2) {
	defer recoverPanic(nil)
	if !isInQueue {
		isInQueue = true

		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()

		isInQueue = false

		diagnosticsMap, err := getDiagnostics()
		if err != nil {
			logError(nil, "Unable to publish diagnostics, error : %s", err.Error())
			return
		}
		for uri, diagnostics := range diagnosticsMap {
			err := publishDiagnostic(uri, diagnostics, conn, ctx)
			if err != nil {
				logError(nil, "Unable to publish diagnostics for %s, error : %s", uri, err.Error())
			}
		}
	}
}

func publishDiagnostic(uri lsp.DocumentURI, diagnostics []lsp.Diagnostic, conn jsonrpc2.JSONRPC2, ctx context.Context) error {
	params := lsp.PublishDiagnosticsParams{URI: uri, Diagnostics: diagnostics}
	return conn.Notify(ctx, "textDocument/publishDiagnostics", params)
}

func getDiagnostics() (map[lsp.DocumentURI][]lsp.Diagnostic, error) {
	diagnostics := make(map[lsp.DocumentURI][]lsp.Diagnostic, 0)
	conceptDictionary, err := validateConcepts(diagnostics)
	if err != nil {
		return nil, err
	}
	if err = validateSpecs(conceptDictionary, diagnostics); err != nil {
		return nil, err
	}
	return diagnostics, nil
}

func createValidationDiagnostics(errors []error, diagnostics map[lsp.DocumentURI][]lsp.Diagnostic) {
	for _, err := range errors {
		uri := util.ConvertPathToURI(err.(validation.StepValidationError).FileName())
		d := createDiagnostic(uri, err.(validation.StepValidationError).Message(), err.(validation.StepValidationError).Step().LineNo-1, 1)
		if err.(validation.StepValidationError).ErrorType() == gm.StepValidateResponse_STEP_IMPLEMENTATION_NOT_FOUND {
			d.Code = err.(validation.StepValidationError).Suggestion()
		}
		diagnostics[uri] = append(diagnostics[uri], d)
	}
	return
}

func validateSpecifications(specs []*gauge.Specification, conceptDictionary *gauge.ConceptDictionary) []error {
	if lRunner.runner == nil {
		return []error{}
	}
	vErrs := validation.NewValidator(specs, lRunner.runner, conceptDictionary).Validate()
	return validation.FilterDuplicates(vErrs)
}

func validateSpecs(conceptDictionary *gauge.ConceptDictionary, diagnostics map[lsp.DocumentURI][]lsp.Diagnostic) error {
	specFiles := util.GetSpecFiles(util.GetSpecDirs())
	specs := make([]*gauge.Specification, 0)
	for _, specFile := range specFiles {
		uri := util.ConvertPathToURI(specFile)
		if _, ok := diagnostics[uri]; !ok {
			diagnostics[uri] = make([]lsp.Diagnostic, 0)
		}
		content, err := getContentFromFileOrDisk(specFile)
		if err != nil {
			return fmt.Errorf("Unable to read file %s", err)
		}
		spec, res, err := new(parser.SpecParser).Parse(content, conceptDictionary, specFile)
		if err != nil {
			return err
		}
		createDiagnostics(res, diagnostics)
		if res.Ok {
			specs = append(specs, spec)
		}
	}
	createValidationDiagnostics(validateSpecifications(specs, conceptDictionary), diagnostics)
	return nil
}

func validateConcepts(diagnostics map[lsp.DocumentURI][]lsp.Diagnostic) (*gauge.ConceptDictionary, error) {
	conceptFiles := util.GetConceptFiles()
	conceptDictionary := gauge.NewConceptDictionary()
	for _, conceptFile := range conceptFiles {
		uri := util.ConvertPathToURI(conceptFile)
		if _, ok := diagnostics[uri]; !ok {
			diagnostics[uri] = make([]lsp.Diagnostic, 0)
		}
		content, err := getContentFromFileOrDisk(conceptFile)
		if err != nil {
			return nil, fmt.Errorf("Unable to read file %s", err)
		}
		cpts, pRes := new(parser.ConceptParser).Parse(content, conceptFile)
		pErrs, err := parser.AddConcept(cpts, conceptFile, conceptDictionary)
		if err != nil {
			return nil, err
		}
		pRes.ParseErrors = append(pRes.ParseErrors, pErrs...)
		createDiagnostics(pRes, diagnostics)
	}
	createDiagnostics(parser.ValidateConcepts(conceptDictionary), diagnostics)
	return conceptDictionary, nil
}

func createDiagnostics(res *parser.ParseResult, diagnostics map[lsp.DocumentURI][]lsp.Diagnostic) {
	for _, err := range res.ParseErrors {
		uri := util.ConvertPathToURI(err.FileName)
		diagnostics[uri] = append(diagnostics[uri], createDiagnostic(uri, err.Message, err.LineNo-1, 1))
	}
	for _, warning := range res.Warnings {
		uri := util.ConvertPathToURI(warning.FileName)
		diagnostics[uri] = append(diagnostics[uri], createDiagnostic(uri, warning.Message, warning.LineNo-1, 2))
	}
}

func createDiagnostic(uri lsp.DocumentURI, message string, line int, severity lsp.DiagnosticSeverity) lsp.Diagnostic {
	endChar := 10000
	if isOpen(uri) {
		endChar = len(getLine(uri, line))
	}
	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: line, Character: 0},
			End:   lsp.Position{Line: line, Character: endChar},
		},
		Message:  message,
		Severity: severity,
	}
}

func getContentFromFileOrDisk(fileName string) (string, error) {
	uri := util.ConvertPathToURI(fileName)
	if isOpen(uri) {
		return getContent(uri), nil
	} else {
		return common.ReadFileContents(fileName)
	}
}
