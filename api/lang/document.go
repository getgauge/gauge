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
		cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{
			Content:  params.TextDocument.Text,
			FilePath: string(util.ConvertURItoFilePath(params.TextDocument.URI)),
			IsClosed: false,
			Status:   gm.CacheFileRequest_OPENED,
		}}
		err = sendMessageToRunner(cacheFileRequest)
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
		cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{
			Content:  params.ContentChanges[0].Text,
			FilePath: string(util.ConvertURItoFilePath(file)),
			IsClosed: false,
			Status:   gm.CacheFileRequest_CHANGED,
		}}
		err = sendMessageToRunner(cacheFileRequest)
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
	} else if lRunner.runner != nil {
		cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{
			FilePath: string(util.ConvertURItoFilePath(params.TextDocument.URI)),
			IsClosed: true,
			Status:   gm.CacheFileRequest_CLOSED,
		}}
		err = sendMessageToRunner(cacheFileRequest)
	}
	go publishDiagnostics(ctx, conn)
	return err
}

func documentCreate(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.TextDocumentIdentifier
	var err error
	if err = json.Unmarshal(*req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse request %v", err)
	}
	if !util.IsGaugeFile(string(params.URI)) {
		if lRunner.runner != nil {
			cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{
				FilePath: string(util.ConvertURItoFilePath(params.URI)),
				IsClosed: false,
				Status:   gm.CacheFileRequest_CREATED,
			}}
			err = sendMessageToRunner(cacheFileRequest)

		}
	}
	go publishDiagnostics(ctx, conn)
	return err
}

func documentDelete(req *jsonrpc2.Request, ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	var params lsp.TextDocumentIdentifier
	var err error
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return fmt.Errorf("failed to parse request. %v", err)
	}
	if !util.IsGaugeFile(string(params.URI)) {
		if lRunner.runner != nil {
			cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{
				FilePath: string(util.ConvertURItoFilePath(params.URI)),
				IsClosed: true,
				Status:   gm.CacheFileRequest_DELETED,
			}}
			err = sendMessageToRunner(cacheFileRequest)
		}
	}
	publishDiagnostic(params.URI, []lsp.Diagnostic{}, conn, ctx)
	go publishDiagnostics(ctx, conn)
	return err
}
