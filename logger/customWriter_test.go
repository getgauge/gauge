/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package logger

import (
	"bytes"
	"testing"
	"time"
)

func TestCustomWriterShouldExtractPortNumberFromStdout(t *testing.T) {
	portChan := make(chan string)
	var b bytes.Buffer
	w := CustomWriter{
		file: &b,
		port: portChan,
	}

	go func() {
		n, err := w.Write([]byte("Listening on port:23454"))
		if n <= 0 || err != nil {
			t.Errorf("failed to write port information")
		}
	}()

	select {
	case port := <-portChan:
		close(portChan)
		if port != "23454" {
			t.Errorf("Expected:%s\nGot     :%s", "23454", port)
		}
	case <-time.After(3 * time.Second):
		t.Errorf("Timed out!! Failed to get port info.")
	}
}

func TestCustomWriterShouldExtractPortNumberFromStdoutWithMultipleLines(t *testing.T) {
	portChan := make(chan string)
	var b bytes.Buffer
	w := CustomWriter{
		file: &b,
		port: portChan,
	}

	go func() {
		s := `{"logLevel": "debug", "message": "Loading step implementations from spec0_3389569547211323752\step_impl dirs."}
{"logLevel": "debug", "message": "Starting grpc server.."}
{"logLevel": "info", "message": "Listening on port:50042"}`
		n, err := w.Write([]byte(s))
		if n <= 0 || err != nil {
			t.Errorf("failed to write port information")
		}
	}()

	select {
	case port := <-portChan:
		close(portChan)
		if port != "50042" {
			t.Errorf("Expected:%s\nGot     :%s", "50042", port)
		}
	case <-time.After(3 * time.Second):
		t.Errorf("Timed out!! Failed to get port info.")
	}
}
