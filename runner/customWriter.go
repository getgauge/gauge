package runner

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/getgauge/gauge/logger"
)

const portPrefix = "Listening on port:"

type customWriter struct {
	file io.Writer
	port chan string
}

func getLine(b []byte) string {
	m := &logger.LogInfo{}
	err := json.Unmarshal(b, m)
	if err != nil {
		return string(b)
	}
	return m.Message
}

func (w customWriter) Write(p []byte) (n int, err error) {
	line := getLine(p)
	if strings.Contains(line, portPrefix) {
		text := strings.Replace(line, "\r\n", "\n", -1)
		w.port <- strings.TrimSuffix(strings.Split(text, portPrefix)[1], "\n")
		return len(p), nil
	}
	return w.file.Write(p)
}

func newCustomWriter(portChan chan string, outFile io.Writer, id string) customWriter {
	return customWriter{
		port: portChan,
		file: logger.Writer{
			File: outFile,
			LoggerID:            id,
			ShouldWriteToStdout: true,
		},
	}
}
