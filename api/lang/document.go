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
	"encoding/json"
	"fmt"

	gm "github.com/getgauge/gauge/gauge_messages"
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
	var err error
	if !util.IsGaugeFile(string(uri)) {
		if lRunner.runner != nil {
			err = cacheFileOnRunner(uri, "", false, gm.CacheFileRequest_DELETED)
		}
	} else {
		publishDiagnostic(uri, []lsp.Diagnostic{}, conn, ctx)
	}
	return err
}
