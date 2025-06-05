// Copyright 2019 ThoughtWorks, Inc.

/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Writer represents to a custom writer.
// It intercepts the log messages and redirects them to logger according the log level given in info
type Writer struct {
	LoggerID            string
	ShouldWriteToStdout bool
	stream              int
	File                io.Writer
	isErrorStream       bool
}

// LogInfo represents the log message structure for plugins
type LogInfo struct {
	LogLevel string `json:"logLevel"`
	Message  string `json:"message"`
}

func (w Writer) Write(p []byte) (int, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(p)))
	for scanner.Scan() {
		_logEntry := strings.Trim(scanner.Text(), " ")
		if _logEntry == "" {
			continue
		}
		_p := []byte(_logEntry)
		m := &LogInfo{}
		err := json.Unmarshal(_p, m)
		if err != nil {
			if w.isErrorStream {
				logError(loggersMap.getLogger(w.LoggerID), w.ShouldWriteToStdout, string(_p))
			} else {
				logInfo(loggersMap.getLogger(w.LoggerID), w.ShouldWriteToStdout, string(_p))
			}
		}
		if w.stream > 0 {
			m.Message = fmt.Sprintf("[runner: %2d] %s", w.stream, m.Message)
		}
		switch m.LogLevel {
		case "debug":
			logDebug(loggersMap.getLogger(w.LoggerID), w.ShouldWriteToStdout, m.Message)
		case "info":
			logInfo(loggersMap.getLogger(w.LoggerID), w.ShouldWriteToStdout, m.Message)
		case "error":
			logError(loggersMap.getLogger(w.LoggerID), w.ShouldWriteToStdout, m.Message)
		case "warning":
			logWarning(loggersMap.getLogger(w.LoggerID), w.ShouldWriteToStdout, m.Message)
		case "fatal":
			logCritical(loggersMap.getLogger(w.LoggerID), m.Message)
			addFatalError(w.LoggerID, m.Message)
		}
	}
	return len(p), nil
}

// LogWriter represents the type which consists of two custom writers
type LogWriter struct {
	Stderr io.Writer
	Stdout io.Writer
}

// NewLogWriter creates a new logWriter for given id
func NewLogWriter(LoggerID string, stdout bool, stream int) *LogWriter {
	return &LogWriter{
		Stderr: Writer{ShouldWriteToStdout: stdout, stream: stream, LoggerID: LoggerID, File: os.Stderr},
		Stdout: Writer{ShouldWriteToStdout: stdout, stream: stream, LoggerID: LoggerID, File: os.Stdout},
	}
}
