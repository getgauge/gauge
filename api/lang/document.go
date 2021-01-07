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

	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func documentOpened(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.DidOpenTextDocumentParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse request %s", err.Error())
	}
	if util.IsGaugeFile(string(params.TextDocument.URI)) {
		openFile(params)
	} else if lRunner.runner != nil {
		err = cacheFileOnRunner(params.TextDocument.URI, params.TextDocument.Text, false, gm.CacheFileRequest_OPENED)
	}
	go publishDiagnostics(ctx, conn)
	return err
}

func documentChange(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.DidChangeTextDocumentParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse request %s", err.Error())
	}
	file := params.TextDocument.URI
	if util.IsGaugeFile(string(file)) {
		changeFile(params)
	} else if lRunner.runner != nil {
		err = cacheFileOnRunner(params.TextDocument.URI, params.ContentChanges[0].Text, false, gm.CacheFileRequest_CHANGED)
	}
	go publishDiagnostics(ctx, conn)
	return err
}

func documentClosed(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.DidCloseTextDocumentParams
	var err error
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse request. %s", err.Error())
	}
	if util.IsGaugeFile(string(params.TextDocument.URI)) {
		closeFile(params)
	} else if lRunner.runner != nil {
		err = cacheFileOnRunner(params.TextDocument.URI, "", true, gm.CacheFileRequest_CLOSED)
	}
	go publishDiagnostics(ctx, conn)
	return err
}

func documentChangeWatchedFiles(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.DidChangeWatchedFilesParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse request. %s", err.Error())
	}
	for _, fileEvent := range params.Changes {
		if fileEvent.Type == int(lsp.Created) {
			if err := documentCreate(fileEvent.URI, ctx, conn); err != nil {
				return err
			}
		} else if fileEvent.Type == int(lsp.Deleted) {
			if err := documentDelete(fileEvent.URI, ctx, conn); err != nil {
				return err
			}
		} else {
			if err := documentCreate(fileEvent.URI, ctx, conn); err != nil {
				return err
			}
		}
	}
	go publishDiagnostics(ctx, conn)
	return nil
}

func documentCreate(uri lsp.DocumentURI, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var err error
	if !util.IsGaugeFile(string(uri)) {
		if lRunner.runner != nil {
			err = cacheFileOnRunner(uri, "", false, gm.CacheFileRequest_CREATED)
		}
	}
	return err
}

func documentDelete(uri lsp.DocumentURI, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	if !util.IsGaugeFile(string(uri)) {
		if lRunner.runner != nil {
			return cacheFileOnRunner(uri, "", false, gm.CacheFileRequest_DELETED)
		}
		return fmt.Errorf("Language runner is not instantiated.")
	}
	return publishDiagnostic(uri, []lsp.Diagnostic{}, conn, ctx)
}
