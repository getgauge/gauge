// Copyright 2019 ThoughtWorks, Inc.

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

package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Writer reperesnts to a custom writer.
// It intercents the log messages and redirects them to logger according the log level given in info
type Writer struct {
	loggerID            string
	ShouldWriteToStdout bool
	stream              int
	File                io.Writer
}

// LogInfo repesents the log message structure for plugins
type LogInfo struct {
	LogLevel string `json:"logLevel"`
	Message  string `json:"message"`
}

func (w Writer) Write(p []byte) (int, error) {
	m := &LogInfo{}
	err := json.Unmarshal(p, m)
	if err != nil {
		fmt.Fprint(w.File, string(p))
		return len(p), nil
	}
	if w.stream > 0 {
		m.Message = fmt.Sprintf("[runner: %d] %s", w.stream, m.Message)
	}
	switch m.LogLevel {
	case "debug":
		write(w.ShouldWriteToStdout, m.Message)
		if initialized {
			GetLogger(w.loggerID).Debug(m.Message)
		}
	case "info":
		write(w.ShouldWriteToStdout, m.Message)
		if initialized {
			GetLogger(w.loggerID).Info(m.Message)
		}
	case "error":
		write(w.ShouldWriteToStdout, m.Message)
		fmt.Fprintf(os.Stderr, m.Message)
		if initialized {
			GetLogger(w.loggerID).Error(m.Message)
		}
	case "warning":
		write(w.ShouldWriteToStdout, m.Message)
		if initialized {
			GetLogger(w.loggerID).Warning(m.Message)
		}
	case "fatal":
		if initialized {
			GetLogger(w.loggerID).Warning(m.Message)
		}
		fatalErrors = append(fatalErrors, fmt.Sprintf("[%s]\n\t%s", w.loggerID, m.Message))
		//TODO: Aggregate the fatal erros from the plugins and print it at the end of execution
		// Or print them when Gauge's fataf is used.
	}
	return len(p), nil
}

// LogWriter reperesents the type which consists of two custom writers
type LogWriter struct {
	Stderr io.Writer
	Stdout io.Writer
}

// NewLogWriter creates a new logWriter for given id
func NewLogWriter(loggerID string, stdout bool, stream int) *LogWriter {
	return &LogWriter{
		Stderr: Writer{ShouldWriteToStdout: stdout, stream: stream, loggerID: loggerID, File: os.Stderr},
		Stdout: Writer{ShouldWriteToStdout: stdout, stream: stream, loggerID: loggerID, File: os.Stdout},
	}
}
