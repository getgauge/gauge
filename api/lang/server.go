/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/api/infoGatherer"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/gauge"
	"github.com/sourcegraph/jsonrpc2"
)

type infoProvider interface {
	Init()
	Steps(filterConcepts bool) []*gauge.Step
	AllSteps(filterConcepts bool) []*gauge.Step
	Concepts() []*gm.ConceptInfo
	Params(file string, argType gauge.ArgType) []gauge.StepArg
	Tags() []string
	SearchConceptDictionary(string) *gauge.Concept
	GetAvailableSpecDetails(specs []string) []*infoGatherer.SpecDetail
	GetSpecDirs() []string
}

var provider infoProvider

type lspHandler struct {
	jsonrpc2.Handler
}

type LangHandler struct {
}

type InitializeParams struct {
	RootPath     string             `json:"rootPath,omitempty"`
	Capabilities ClientCapabilities `json:"capabilities,omitempty"`
}

func newHandler() jsonrpc2.Handler {
	return lspHandler{jsonrpc2.HandlerWithError((&LangHandler{}).handle)}
}

func (h lspHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	go h.Handler.Handle(ctx, conn, req)
}

func (h *LangHandler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
	defer recoverPanic(req)
	return h.Handle(ctx, conn, req)
}

func (h *LangHandler) Handle(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request) (interface{}, error) {
	switch req.Method {
	case "initialize":
		if err := cacheInitializeParams(req); err != nil {
			logError(req, err.Error())
			return nil, err
		}
		return gaugeLSPCapabilities(), nil
	case "initialized":
		err := registerFileWatcher(conn, ctx)
		if err != nil {
			logError(req, err.Error())
		}
		err = registerRunnerCapabilities(conn, ctx)
		if err != nil {
			logError(req, err.Error())
		}
		go publishDiagnostics(ctx, conn)
		return nil, nil
	case "shutdown":
		return nil, killRunner()
	case "exit":
		if c, ok := conn.(*jsonrpc2.Conn); ok {
			err := c.Close()
			if err != nil {
				logError(req, err.Error())
			}
			os.Exit(0)
		}
		return nil, nil
	case "$/cancelRequest":
		return nil, nil
	case "textDocument/didOpen":
		err := documentOpened(req, ctx, conn)
		if err != nil {
			logDebug(req, err.Error())
		}
		return nil, err
	case "textDocument/didClose":
		err := documentClosed(req, ctx, conn)
		if err != nil {
			logDebug(req, err.Error())
		}
		return nil, err
	case "textDocument/didChange":
		err := documentChange(req, ctx, conn)
		if err != nil {
			logDebug(req, err.Error())
		}
		return nil, err
	case "workspace/didChangeWatchedFiles":
		err := documentChangeWatchedFiles(req, ctx, conn)
		if err != nil {
			logDebug(req, err.Error())
		}
		return nil, err
	case "textDocument/completion":
		val, err := completion(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "completionItem/resolve":
		val, err := resolveCompletion(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "textDocument/definition":
		val, err := definition(req)
		if err != nil {
			logDebug(req, err.Error())
			if e := showErrorMessageOnClient(ctx, conn, err); e != nil {
				return nil, fmt.Errorf("unable to send '%s' error to LSP server. %s", err.Error(), e.Error())
			}

		}
		return val, err
	case "textDocument/formatting":
		data, err := format(req)
		if err != nil {
			logDebug(req, err.Error())
			e := showErrorMessageOnClient(ctx, conn, err)
			if e != nil {
				return nil, fmt.Errorf("unable to send '%s' error to LSP server. %s", err.Error(), e.Error())
			}
		}
		return data, err
	case "textDocument/codeLens":
		val, err := codeLenses(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "textDocument/codeAction":
		val, err := codeActions(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "textDocument/rename":
		result, err := rename(ctx, conn, req)
		if err != nil {
			logDebug(req, err.Error())
			e := showErrorMessageOnClient(ctx, conn, err)
			if e != nil {
				return nil, fmt.Errorf("unable to send '%s' error to LSP server. %s", err.Error(), e.Error())
			}
			return nil, err
		}
		return result, nil
	case "textDocument/documentSymbol":
		val, err := documentSymbols(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "workspace/symbol":
		val, err := workspaceSymbols(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "gauge/stepReferences":
		val, err := stepReferences(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "gauge/stepValueAt":
		val, err := stepValueAt(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "gauge/scenarios":
		val, err := scenarios(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "gauge/getImplFiles":
		val, err := getImplFiles(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "gauge/putStubImpl":
		if err := sendSaveFilesRequest(ctx, conn); err != nil {
			logDebug(req, err.Error())
			e := showErrorMessageOnClient(ctx, conn, err)
			if e != nil {
				return nil, fmt.Errorf("unable to send '%s' error to LSP server. %s", err.Error(), e.Error())
			}
			return nil, err
		}
		val, err := putStubImpl(req)
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "gauge/specs":
		val, err := specs()
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "gauge/executionStatus":
		val, err := execution.ReadLastExecutionResult()
		if err != nil {
			logDebug(req, err.Error())
		}
		return val, err
	case "gauge/generateConcept":
		if err := sendSaveFilesRequest(ctx, conn); err != nil {
			e := showErrorMessageOnClient(ctx, conn, err)
			if e != nil {
				return nil, fmt.Errorf("unable to send '%s' error to LSP server. %s", err.Error(), e.Error())
			}
			return nil, err
		}
		return generateConcept(req)
	case "gauge/getRunnerLanguage":
		return lRunner.lspID, nil
	case "gauge/specDirs":
		return provider.GetSpecDirs(), nil
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

func startLsp(logLevel string) (context.Context, *jsonrpc2.Conn) {
	logInfo(nil, "LangServer: reading on stdin, writing on stdout")
	var connOpt []jsonrpc2.ConnOpt
	if logLevel == "debug" {
		connOpt = append(connOpt, jsonrpc2.LogMessages(log.New(lspWriter{}, "", 0)))
	}
	ctx := context.Background()
	return ctx, jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(stdRWC{}, jsonrpc2.VSCodeObjectCodec{}), newHandler(), connOpt...)
}

func initializeRunner() error {
	id, err := getLanguageIdentifier()
	if err != nil || id == "" {
		e := fmt.Errorf("There are version incompatibilities between Gauge and it's plugins in this project. Some features will not work as expected.")
		logDebug(nil, "%s", e.Error())
		return e
	}
	err = startRunner()
	if err != nil {
		logDebug(nil, "%s\nSome of the gauge lsp feature will not work as expected.", err.Error())
		return err
	}
	lRunner.lspID = id
	return nil
}

func Start(p infoProvider, logLevel string) {
	provider = p
	provider.Init()
	err := initializeRunner()
	ctx, conn := startLsp(logLevel)
	if err != nil {
		_ = showErrorMessageOnClient(ctx, conn, err)
	}
	initialize(ctx, conn)
	<-conn.DisconnectNotify()
	if killRunner() != nil {
		logInfo(nil, "failed to kill runner with pid %d", lRunner.runner.Pid())
	}
	logInfo(nil, "Connection closed")
}

func recoverPanic(req *jsonrpc2.Request) {
	if r := recover(); r != nil {
		logFatal(req, "%v\n%s", r, string(debug.Stack()))
	}
}
