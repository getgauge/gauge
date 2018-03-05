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

	"github.com/getgauge/common"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type textDocumentRegistrationOptions struct {
	DocumentSelector documentSelector `json:"documentSelector"`
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

func documentOpened(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.DidOpenTextDocumentParams
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse request %v", err)
	}
	if util.IsGaugeFile(string(params.TextDocument.URI)) {
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
		return fmt.Errorf("failed to parse request %v", err)
	}
	file := params.TextDocument.URI
	if util.IsGaugeFile(string(file)) {
		changeFile(params)
	} else if lRunner.runner != nil {
		err = cacheFileOnRunner(file, params.ContentChanges[0].Text)
	}
	go publishDiagnostics(ctx, conn)
	return err
}

func documentClosed(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.DidCloseTextDocumentParams
	var err error
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse request. %v", err)
	}
	if util.IsGaugeFile(string(params.TextDocument.URI)) {
		closeFile(params)
		if !common.FileExists(string(util.ConvertPathToURI(params.TextDocument.URI))) {
			publishDiagnostic(params.TextDocument.URI, []lsp.Diagnostic{}, conn, ctx)
		}
	} else if lRunner.runner != nil {
		cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{FilePath: string(util.ConvertURItoFilePath(params.TextDocument.URI)), IsClosed: true}}
		err = sendMessageToRunner(cacheFileRequest)
	}
	go publishDiagnostics(ctx, conn)
	return err
}
