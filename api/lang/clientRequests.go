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
