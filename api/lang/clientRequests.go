/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"context"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func showErrorMessageOnClient(ctx context.Context, conn jsonrpc2.JSONRPC2, err error) error {
	return conn.Notify(ctx, "window/showMessage", lsp.ShowMessageParams{Type: lsp.MTError, Message: err.Error()})
}

func sendSaveFilesRequest(ctx context.Context, conn jsonrpc2.JSONRPC2) error {
	if clientCapabilities.SaveFiles {
		var result interface{}
		return conn.Call(ctx, "workspace/saveFiles", nil, &result)
	}
	return nil
}
