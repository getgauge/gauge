/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
	logger.Debug(false, string(p))
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
	err := c.conn.Notify(c.ctx, "window/logMessage", lsp.LogMessageParams{Type: level, Message: msg})
	if err != nil {
		logger.Errorf(false, "Unable to log error '%s' to LSP: %s", msg, err.Error())
	}
}

var lspLog *lspLogger

func initialize(ctx context.Context, conn *jsonrpc2.Conn) {
	lspLog = &lspLogger{conn: conn, ctx: ctx}
}

func logDebug(req *jsonrpc2.Request, msg string, args ...interface{}) {
	m := fmt.Sprintf(getLogFormatFor(req, msg), args...)
	logger.Debug(false, m)
	logToLsp(lsp.Log, m)
}

func logInfo(req *jsonrpc2.Request, msg string, args ...interface{}) {
	m := fmt.Sprintf(getLogFormatFor(req, msg), args...)
	logger.Info(false, m)
	logToLsp(lsp.Info, m)
}

func logWarning(req *jsonrpc2.Request, msg string, args ...interface{}) {
	m := fmt.Sprintf(getLogFormatFor(req, msg), args...)
	logger.Warning(false, m)
	logToLsp(lsp.MTWarning, m)
}

func logError(req *jsonrpc2.Request, msg string, args ...interface{}) {
	m := fmt.Sprintf(getLogFormatFor(req, msg), args...)
	logger.Error(false, m)
	logToLsp(lsp.MTError, m)
}

func logFatal(req *jsonrpc2.Request, msg string, args ...interface{}) {
	logToLsp(lsp.MTError, "An error occurred. Refer lsp.log for more details.")
	logger.Fatalf(true, getLogFormatFor(req, msg), args...)
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
	formattedMsg := fmt.Sprintf("#%s: %s: ", req.ID.String(), req.Method) + msg
	if req.Notif {
		return "notif " + formattedMsg
	}
	return "request " + formattedMsg
}
