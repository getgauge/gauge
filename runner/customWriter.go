package runner

import (
	"io"
	"regexp"
	"strings"

	"github.com/getgauge/gauge/logger"
)

const portPrefix = "Listening on port:"

type customWriter struct {
	file io.Writer
	port chan string
}

func (w customWriter) Write(p []byte) (n int, err error) {
	line := string(p)
	if strings.Contains(line, portPrefix) {
		text := strings.Replace(line, "\r\n", "\n", -1)
		re := regexp.MustCompile(portPrefix + "([0-9]+)")
		f := re.FindStringSubmatch(text)
		w.port <- f[1]
		// return len(p), nil
	}
	return w.file.Write(p)
}

func newCustomWriter(portChan chan string, outFile io.Writer, id string) customWriter {
	return customWriter{
		port: portChan,
		file: logger.Writer{
			File:                outFile,
			LoggerID:            id,
			ShouldWriteToStdout: true,
		},
	}
}
