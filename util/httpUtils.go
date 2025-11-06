/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/getgauge/gauge/logger"

	"github.com/getgauge/common"
)

// progressReader is for indicating the download / upload progress on the console
type progressReader struct {
	io.Reader
	bytesTransfered   int64
	totalBytes        int64
	progress          float64
	progressDisplayed bool
}

// Read overrides the underlying io.Reader's Read method.
// io.Copy() will be calling this method.
func (w *progressReader) Read(p []byte) (int, error) {
	n, err := w.Reader.Read(p)
	if n > 0 {
		w.bytesTransfered += int64(n)
		percent := float64(w.bytesTransfered) * float64(100) / float64(w.totalBytes)
		if percent-w.progress > 4 {
			fmt.Print(".")
			w.progress = percent
			w.progressDisplayed = true
		}
	}
	return n, err
}

// Download fires a HTTP GET request to download a resource to target directory
func Download(url, targetDir, fileName string, silent bool) (string, error) {
	if !common.DirExists(targetDir) {
		return "", fmt.Errorf("Error downloading file: %s\nTarget dir %s doesn't exists.", url, targetDir)
	}

	if fileName == "" {
		fileName = filepath.Base(url)
	}
	targetFile := filepath.Join(targetDir, fileName)

	logger.Debugf(true, "Downloading %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error downloading file: %s.\n%s", url, resp.Status)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	out, err := os.Create(targetFile)
	if err != nil {
		return "", err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)
	if silent {
		_, err = io.Copy(out, resp.Body)
	} else {
		progressReader := &progressReader{Reader: resp.Body, totalBytes: resp.ContentLength}
		_, err = io.Copy(out, progressReader)
		if progressReader.progressDisplayed {
			fmt.Println()
		}
	}
	return targetFile, err
}
