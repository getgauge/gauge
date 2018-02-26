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
	"github.com/getgauge/gauge/logger"
	"github.com/op/go-logging"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"os"
)

type lspWriter struct {
}

func (w lspWriter) Write(p []byte) (n int, err error) {
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

func (c lspLogger) Log(logLevel logging.Level, msg string) {
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

func logDebug(req *jsonrpc2.Request, msg string, args ...interface{}) {
	logger.Debugf(false, getLogFormatFor(req, msg), args...)
}

func logWarning(req *jsonrpc2.Request, msg string, args ...interface{}) {
	logger.Warningf(false, getLogFormatFor(req, msg), args...)
}

func logError(req *jsonrpc2.Request, msg string, args ...interface{}) {
	logger.Errorf(false, getLogFormatFor(req, msg), args...)
}

func getLogFormatFor(req *jsonrpc2.Request, msg string) string {
	formattedMsg := fmt.Sprintf("#%d: %s: ", req.ID, req.Method) + msg
	if req.Notif {
		return "notif " + formattedMsg
	}
	return "request " + formattedMsg
}
