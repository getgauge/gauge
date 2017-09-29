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
	"log"

	"os"

	"errors"

	"encoding/json"

	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type server struct{}

type completionProvider interface {
	Init()
	Steps() []*gauge.StepValue
	Concepts() []*gm.ConceptInfo
	Params(file string, argType gauge.ArgType) []gauge.StepArg
}

var provider completionProvider

func Server(p completionProvider) *server {
	provider = p
	provider.Init()
	return &server{}
}

type lspHandler struct {
	jsonrpc2.Handler
}

type LangHandler struct {
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
		kind := lsp.TDSKFull
		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync:           lsp.TextDocumentSyncOptionsOrKind{Kind: &kind},
				CompletionProvider:         &lsp.CompletionOptions{ResolveProvider: true, TriggerCharacters: []string{"*", "* ", "\"", "<"}},
				DocumentFormattingProvider: true,
				CodeLensProvider:           &lsp.CodeLensOptions{ResolveProvider: true},
			},
		}, nil
	case "initialized":
		return nil, nil
	case "shutdown":
		return nil, nil
	case "exit":
		if c, ok := conn.(*jsonrpc2.Conn); ok {
			c.Close()
		}
		return nil, nil

	case "$/cancelRequest":
		return nil, nil
	case "textDocument/didOpen":
		openFile(req)
		publishDiagnostics(ctx, conn, req)
		return nil, nil
	case "textDocument/didClose":
		closeFile(req)
		return nil, nil
	case "textDocument/didSave":
		return nil, errors.New("Unknown request")
	case "textDocument/didChange":
		changeFile(req)
		publishDiagnostics(ctx, conn, req)
		return nil, nil
	case "textDocument/completion":
		return completion(req)
	case "completionItem/resolve":
		return resolveCompletion(req)
	case "textDocument/hover":
		return nil, errors.New("Unknown request")
	case "textDocument/definition":
		return nil, errors.New("Unknown request")
	case "textDocument/xdefinition":
		return nil, errors.New("Unknown request")
	case "textDocument/references":
		return nil, errors.New("Unknown request")
	case "textDocument/documentSymbol":
		return nil, errors.New("Unknown request")
	case "textDocument/signatureHelp":
		return nil, errors.New("Unknown request")
	case "textDocument/formatting":
		data, err := format(req)
		if err != nil {
			conn.Notify(ctx, "window/showMessage", lsp.ShowMessageParams{Type: 1, Message: err.Error()})
		}
		return data, err
	case "textDocument/codeLens":
		return getCodeLenses(req)
	case "codeLens/resolve":
		return resolveCodeLens(req)
	case "workspace/symbol":
		return nil, errors.New("Unknown request")
	case "workspace/xreferences":
		return nil, errors.New("Unknown request")
	default:
		return nil, errors.New("Unknown request")
	}
}

func publishDiagnostics(ctx context.Context, conn jsonrpc2.JSONRPC2, request *jsonrpc2.Request) {
	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(*request.Params, &params); err != nil {
		logger.APILog.Debugf("failed to parse request %s", err.Error())
	}
	diagnostics := createDiagnostics(params.TextDocument.URI)
	conn.Notify(ctx, "textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{URI: params.TextDocument.URI, Diagnostics: diagnostics})
}

func (s *server) Start() {
	logger.APILog.Info("LangServer: reading on stdin, writing on stdout")
	var connOpt []jsonrpc2.ConnOpt
	connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(os.Stderr, "", 0)))
	<-jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(stdRWC{}, jsonrpc2.VSCodeObjectCodec{}), newHandler(), connOpt...).DisconnectNotify()
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
