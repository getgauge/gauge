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
	"log"

	"os"

	"encoding/json"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	"github.com/op/go-logging"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type infoProvider interface {
	Init()
	Steps() []*gauge.Step
	AllSteps() []*gauge.Step
	Concepts() []*gm.ConceptInfo
	Params(file string, argType gauge.ArgType) []gauge.StepArg
	Tags() []string
	SearchConceptDictionary(string) *gauge.Concept
	GetAvailableSpecDetails(specs []string) []*infoGatherer.SpecDetail
}

var provider infoProvider
var clientCapabilities ClientCapabilities

type lspHandler struct {
	jsonrpc2.Handler
}

type LangHandler struct {
}

type registrationParams struct {
	Registrations []registration `json:"registrations"`
}

type registration struct {
	Id              string      `json:"id"`
	Method          string      `json:"method"`
	RegisterOptions interface{} `json:"registerOptions"`
}

type textDocumentRegistrationOptions struct {
	DocumentSelector documentSelector `json:"documentSelector"`
}

type textDocumentChangeRegistrationOptions struct {
	textDocumentRegistrationOptions
	SyncKind lsp.TextDocumentSyncKind `json:"syncKind,omitempty"`
}

type codeLensRegistrationOptions struct {
	textDocumentRegistrationOptions
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

type documentSelector struct {
	Scheme   string `json:"scheme"`
	Language string `json:"language"`
	Pattern  string `json:"pattern"`
}

type InitializeParams struct {
	RootPath     string             `json:"rootPath,omitempty"`
	Capabilities ClientCapabilities `json:"capabilities,omitempty"`
}

type ClientCapabilities struct {
	SaveFiles bool `json:"saveFiles,omitempty"`
}

func newHandler() jsonrpc2.Handler {
	return lspHandler{jsonrpc2.HandlerWithError((&LangHandler{}).handle)}
}

func (h lspHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	go h.Handler.Handle(ctx, conn, req)
}

func (h *LangHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	return h.Handle(ctx, conn, req)
}

func (h *LangHandler) Handle(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request) (interface{}, error) {
	switch req.Method {
	case "initialize":
		if err := cacheInitializeParams(req); err != nil {
			return nil, err
		}
		return gaugeLSPCapabilities(), nil
	case "initialized":
		err := registerRunnerCapabilities(conn, ctx)
		go publishDiagnostics(ctx, conn)
		return nil, err
	case "shutdown":
		killRunner()
		return nil, nil
	case "exit":
		if c, ok := conn.(*jsonrpc2.Conn); ok {
			c.Close()
			os.Exit(0)
		}
		return nil, nil
	case "$/cancelRequest":
		return nil, nil
	case "textDocument/didOpen":
		return nil, documentOpened(req, ctx, conn)
	case "textDocument/didClose":
		return nil, documentClosed(req, ctx, conn)
	case "textDocument/didChange":
		return nil, documentChange(req, ctx, conn)
	case "textDocument/completion":
		return completion(req)
	case "completionItem/resolve":
		return resolveCompletion(req)
	case "textDocument/definition":
		return definition(req)
	case "textDocument/formatting":
		data, err := format(req)
		if err != nil {
			conn.Notify(ctx, "window/showMessage", lsp.ShowMessageParams{Type: 1, Message: err.Error()})
		}
		return data, err
	case "textDocument/codeLens":
		return codeLenses(req)
	case "textDocument/codeAction":
		return codeActions(req)
	case "textDocument/rename":
		result, err := rename(ctx, conn, req)
		if err != nil {
			showErrorMessageOnClient(ctx, conn, err)
			return nil, err
		}
		return result, nil
	case "textDocument/documentSymbol":
		return documentSymbols(req)
	case "workspace/symbol":
		return workspaceSymbols(req)
	case "gauge/stepReferences":
		return stepReferences(req)
	case "gauge/stepValueAt":
		return stepValueAt(req)
	case "gauge/scenarios":
		return scenarios(req)
	case "gauge/getImplFiles":
		return getImplFiles()
	case "gauge/putStubImpl":
		return putStubImpl(req)
	case "gauge/specs":
		return specs()
	case "gauge/executionStatus":
		return execution.ReadExecutionStatus()
	default:
		return nil, nil
	}
}

func cacheInitializeParams(req *jsonrpc2.Request) error {
	var params InitializeParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		return err
	}
	clientCapabilities = params.Capabilities
	return nil
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

func documentOpened(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.DidOpenTextDocumentParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return err
	}
	if util.IsGaugeFile(params.TextDocument.URI) {
		openFile(params)
	} else if lRunner.runner != nil {
		err = cacheFileOnRunner(params.TextDocument.URI, params.TextDocument.Text)
	}
	go publishDiagnostics(ctx, conn)
	return err
}

func documentChange(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.DidChangeTextDocumentParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return err
	}
	if util.IsGaugeFile(params.TextDocument.URI) {
		changeFile(params)
	} else if lRunner.runner != nil {
		err = cacheFileOnRunner(params.TextDocument.URI, params.ContentChanges[0].Text)
	}
	go publishDiagnostics(ctx, conn)
	return err
}

func documentClosed(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.DidCloseTextDocumentParams
	var err error
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
		return err
	}
	if util.IsGaugeFile(params.TextDocument.URI) {
		closeFile(params)
		if !common.FileExists(util.ConvertPathToURI(params.TextDocument.URI)) {
			publishDiagnostic(params.TextDocument.URI, []lsp.Diagnostic{}, conn, ctx)
		}
	} else if lRunner.runner != nil {
		cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{FilePath: util.ConvertURItoFilePath(params.TextDocument.URI), IsClosed: true}}
		err = sendMessageToRunner(cacheFileRequest)
	}
	go publishDiagnostics(ctx, conn)
	return err
}

func registerRunnerCapabilities(conn jsonrpc2.JSONRPC2, ctx context.Context) error {
	if lRunner.lspID == "" {
		return fmt.Errorf("current runner is not compatible with gauge LSP")
	}
	var result interface{}
	ds := documentSelector{"file", lRunner.lspID, fmt.Sprintf("%s/**/*", config.ProjectRoot)}
	conn.Call(ctx, "client/registerCapability", registrationParams{[]registration{
		{Id: "gauge-runner-didOpen", Method: "textDocument/didOpen", RegisterOptions: textDocumentRegistrationOptions{DocumentSelector: ds}},
		{Id: "gauge-runner-didClose", Method: "textDocument/didClose", RegisterOptions: textDocumentRegistrationOptions{DocumentSelector: ds}},
		{Id: "gauge-runner-didChange", Method: "textDocument/didChange", RegisterOptions: textDocumentChangeRegistrationOptions{textDocumentRegistrationOptions: textDocumentRegistrationOptions{DocumentSelector: ds}, SyncKind: lsp.TDSKFull}},
		{Id: "gauge-runner-codelens", Method: "textDocument/codeLens", RegisterOptions: codeLensRegistrationOptions{textDocumentRegistrationOptions: textDocumentRegistrationOptions{DocumentSelector: ds}, ResolveProvider: false}},
	}}, &result)
	return nil
}

type lspWriter struct {
}

func (w lspWriter) Write(p []byte) (n int, err error) {
	logger.LspLog.Debug(string(p))
	return os.Stderr.Write(p)
}

func startLsp(logLevel string) (context.Context, *jsonrpc2.Conn) {
	logger.APILog.Info("LangServer: reading on stdin, writing on stdout")
	var connOpt []jsonrpc2.ConnOpt
	if logLevel == "debug" {
		connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(lspWriter{}, "", 0)))
	}
	ctx := context.Background()
	return ctx, jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(stdRWC{}, jsonrpc2.VSCodeObjectCodec{}), newHandler(), connOpt...)
}

func initializeRunner() {
	id, err := getLanguageIdentifier()
	if err != nil || id == "" {
		logger.APILog.Debug("Current runner is not compatible with gauge LSP.")
	}
	err = startRunner()
	if err != nil {
		logger.APILog.Debugf("%s\nSome of the gauge lsp feature will not work as expected.", err.Error())
	}
	lRunner.lspID = id
}

func Start(p infoProvider, logLevel string) {
	provider = p
	provider.Init()
	initializeRunner()
	ctx, conn := startLsp(logLevel)
	logger.SetCustomLogger(lspLogger{conn, ctx})
	<-conn.DisconnectNotify()
	logger.APILog.Info("Connection closed")
}

type stdRWC struct{}

func (stdRWC) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdRWC) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdRWC) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}

type lspLogger struct {
	conn *jsonrpc2.Conn
	ctx  context.Context
}

func (c lspLogger) Log(logLevel logging.Level, msg string) {
	logger.APILog.Info(logLevel)
	var level lsp.MessageType
	switch logLevel {
	case logging.DEBUG:
		level = lsp.Log
	case logging.INFO:
		level = lsp.Info
	case logging.WARNING:
		level = lsp.MTWarning
	case logging.ERROR:
		level = lsp.MTError
	case logging.CRITICAL:
		level = lsp.MTError
	default:
		level = lsp.Info
	}
	c.conn.Notify(c.ctx, "window/logMessage", lsp.LogMessageParams{Type: level, Message: msg})
}
