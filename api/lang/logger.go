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
	"fmt"
	"os"

	"github.com/getgauge/gauge/logger"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type lspWriter struct {
}

func (w lspWriter) Write(p []byte) (n int, err error) {
	logger.Debugf(false, string(p))
	return os.Stderr.Write(p)
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

func (c *lspLogger) Log(level lsp.MessageType, msg string) {
	c.conn.Notify(c.ctx, "window/logMessage", lsp.LogMessageParams{Type: level, Message: msg})
}

var lspLog *lspLogger

func initialize(ctx context.Context, conn *jsonrpc2.Conn) {
	lspLog = &lspLogger{conn: conn, ctx: ctx}
}

func logDebug(req *jsonrpc2.Request, msg string, args ...interface{}) {
	m := fmt.Sprintf(getLogFormatFor(req, msg), args...)
	logger.Debugf(false, m)
	logToLsp(lsp.Log, m)
}

func logInfo(req *jsonrpc2.Request, msg string, args ...interface{}) {
	m := fmt.Sprintf(getLogFormatFor(req, msg), args...)
	logger.Infof(false, m)
	logToLsp(lsp.Info, m)
}

func logWarning(req *jsonrpc2.Request, msg string, args ...interface{}) {
	m := fmt.Sprintf(getLogFormatFor(req, msg), args...)
	logger.Warningf(false, m)
	logToLsp(lsp.MTWarning, m)
}

func logError(req *jsonrpc2.Request, msg string, args ...interface{}) {
	m := fmt.Sprintf(getLogFormatFor(req, msg), args...)
	logger.Errorf(false, m)
	logToLsp(lsp.MTError, m)
}

func logToLsp(level lsp.MessageType, m string) {
	if lspLog != nil {
		lspLog.Log(level, m)
	}
}
func getLogFormatFor(req *jsonrpc2.Request, msg string) string {
	if req == nil {
		return msg
	}
	formattedMsg := fmt.Sprintf("#%d: %s: ", req.ID, req.Method) + msg
	if req.Notif {
		return "notif " + formattedMsg
	}
	return "request " + formattedMsg
}
