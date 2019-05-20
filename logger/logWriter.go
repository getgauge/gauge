package logger

import (
	"encoding/json"
	"fmt"
	"io"
)

type Writer struct {
	ShouldWriteToStdout bool
	stream              int
}

type LogInfo struct {
	ID       string `josn:"id"`
	LogLevel string `json:"logLevel`
	Message  string `json:"message"`
}

func (w Writer) Write(p []byte) (int, error) {
	m := &LogInfo{}
	err := json.Unmarshal(p, m)
	if err != nil {
		m.LogLevel = "info"
		m.Message = string(p)
	}
	if w.stream > 0 {
		m.Message = fmt.Sprintf("[runner: %d] %s", w.stream, m.Message)
	}
	switch m.LogLevel {
	case "debug":
		Debug(w.ShouldWriteToStdout, m.Message)
	case "info":
		Info(w.ShouldWriteToStdout, m.Message)
	case "error":
		Error(w.ShouldWriteToStdout, m.Message)
	case "fatal":
		Fatal(w.ShouldWriteToStdout, m.Message)
	}
	return len(p), nil
}

type LogWriter struct {
	Stderr io.Writer
	Stdout io.Writer
}

func NewLogWriter(loggerID string, stdout bool, stream int) *LogWriter {
	return &LogWriter{
		Stderr: Writer{ShouldWriteToStdout: stdout, stream: stream},
		Stdout: Writer{ShouldWriteToStdout: stdout, stream: stream},
	}
}
