// Copyright 2018 ThoughtWorks, Inc.

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
	"os"
	"strconv"
	"strings"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type registrationParams struct {
	Registrations []registration `json:"registrations"`
}

type registration struct {
	Id              string      `json:"id"`
	Method          string      `json:"method"`
	RegisterOptions interface{} `json:"registerOptions"`
}

type codeLensRegistrationOptions struct {
	textDocumentRegistrationOptions
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

type didChangeWatchedFilesRegistrationOptions struct {
	Watchers []fileSystemWatcher `json:"watchers"`
}

type fileSystemWatcher struct {
	GlobPattern string `json:"globPattern"`
	Kind        int    `json:"kind"`
}

type watchKind int

const (
	created watchKind = 1
	changed           = 2
	deleted           = 4
)

type textDocumentRegistrationOptions struct {
	DocumentSelector []documentSelector `json:"documentSelector"`
}

type textDocumentChangeRegistrationOptions struct {
	textDocumentRegistrationOptions
	SyncKind lsp.TextDocumentSyncKind `json:"syncKind,omitempty"`
}

type documentSelector struct {
	Scheme   string `json:"scheme"`
	Language string `json:"language"`
	Pattern  string `json:"pattern"`
}

var clientCapabilities ClientCapabilities

type ClientCapabilities struct {
	SaveFiles bool `json:"saveFiles,omitempty"`
}

func gaugeLSPCapabilities() lsp.InitializeResult {
	kind := lsp.TDSKFull
	return lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync:           lsp.TextDocumentSyncOptionsOrKind{Kind: &kind, Options: &lsp.TextDocumentSyncOptions{Save: &lsp.SaveOptions{IncludeText: true}}},
			CompletionProvider:         &lsp.CompletionOptions{ResolveProvider: true, TriggerCharacters: []string{"*", "* ", "\"", "<", ":", ","}},
			DocumentFormattingProvider: true,
			CodeLensProvider:           &lsp.CodeLensOptions{ResolveProvider: false},
			DefinitionProvider:         true,
			CodeActionProvider:         true,
			DocumentSymbolProvider:     true,
			WorkspaceSymbolProvider:    true,
			RenameProvider:             true,
		},
	}
}

func registerFileWatcher(conn jsonrpc2.JSONRPC2, ctx context.Context) {
	fileExtensions := strings.Join(util.GaugeFileExtensions(), ",")
	regParams := didChangeWatchedFilesRegistrationOptions{
		Watchers: []fileSystemWatcher{{
			GlobPattern: strings.Replace(config.ProjectRoot, util.WindowsSep, util.UnixSep, -1) + "/**/*{" + fileExtensions + "}",
			Kind:        int(created) + int(deleted),
		}},
	}
	var result interface{}
	conn.Call(ctx, "client/registerCapability", registrationParams{[]registration{
		{Id: "gauge-fileWatcher", Method: "workspace/didChangeWatchedFiles", RegisterOptions: regParams},
	}}, &result)
}

func registerRunnerCapabilities(conn jsonrpc2.JSONRPC2, ctx context.Context) error {
	if lRunner.lspID == "" {
		return fmt.Errorf("current runner is not compatible with gauge LSP")
	}

	implFileGlobPatternResponse, err := globPatternRequest()
	if err != nil {
		return err
	}
	filePatterns := make([]fileSystemWatcher, 0)
	documentSelectors := make([]documentSelector, 0)
	for _, globPattern := range implFileGlobPatternResponse.GlobPatterns {
		filePatterns = append(filePatterns, fileSystemWatcher{
			GlobPattern: globPattern,
			Kind:        int(created) + int(deleted),
		})
		documentSelectors = append(documentSelectors, documentSelector{
			Scheme:   "file",
			Language: lRunner.lspID,
			Pattern:  globPattern,
		})
	}
	var result interface{}
	var registrations = []registration{
		{Id: "gauge-runner-didOpen", Method: "textDocument/didOpen", RegisterOptions: textDocumentRegistrationOptions{DocumentSelector: documentSelectors}},
		{Id: "gauge-runner-didClose", Method: "textDocument/didClose", RegisterOptions: textDocumentRegistrationOptions{DocumentSelector: documentSelectors}},
		{Id: "gauge-runner-didChange", Method: "textDocument/didChange", RegisterOptions: textDocumentChangeRegistrationOptions{textDocumentRegistrationOptions: textDocumentRegistrationOptions{DocumentSelector: documentSelectors}, SyncKind: lsp.TDSKFull}},
		{Id: "gauge-runner-fileWatcher", Method: "workspace/didChangeWatchedFiles", RegisterOptions: didChangeWatchedFilesRegistrationOptions{Watchers: filePatterns}},
	}
	registrations = addReferenceCodeLensRegistration(registrations, documentSelectors)
	conn.Call(ctx, "client/registerCapability", registrationParams{registrations}, &result)
	return nil
}

func addReferenceCodeLensRegistration(registrations []registration, selectors []documentSelector) []registration {
	if enabled, err := strconv.ParseBool(os.Getenv("gauge_lsp_reference_codelens")); err == nil && !enabled {
		return registrations
	}
	codeLenseRegistration := registration{Id: "gauge-runner-codelens",
		Method: "textDocument/codeLens",
		RegisterOptions: codeLensRegistrationOptions{
			textDocumentRegistrationOptions: textDocumentRegistrationOptions{
				DocumentSelector: selectors,
			},
			ResolveProvider: false},
	}
	return append(registrations, codeLenseRegistration)
}
