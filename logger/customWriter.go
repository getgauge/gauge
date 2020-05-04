/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package logger

import (
	"io"
	"regexp"
	"strings"
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
		return len(p), nil
	}
	return w.file.Write(p)
}

func NewCustomWriter(portChan chan string, outFile io.Writer, id string) customWriter {
	return customWriter{
		port: portChan,
		file: Writer{
			File:                outFile,
			LoggerID:            id,
			ShouldWriteToStdout: true,
		},
	}
}
