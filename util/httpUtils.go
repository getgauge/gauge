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

package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/getgauge/common"
)

// progressReader is for indicating the download / upload progress on the console
type progressReader struct {
	io.Reader
	bytesTransfered int64
	totalBytes      int64
	progress        float64
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
		}
	}
	return n, err
}

// Download fires a HTTP GET request to download a resource to target directory
func Download(url, targetDir string, silent bool) (string, error) {
	if !common.DirExists(targetDir) {
		return "", fmt.Errorf("Error downloading file: %s\nTarget dir %s doesn't exists.", url, targetDir)
	}
	targetFile := filepath.Join(targetDir, filepath.Base(url))

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Error downloading file: %s.\n%s", url, resp.Status)
	}
	defer resp.Body.Close()

	out, err := os.Create(targetFile)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if silent {
		_, err = io.Copy(out, resp.Body)
	} else {
		progressReader := &progressReader{Reader: resp.Body, totalBytes: resp.ContentLength}
		_, err = io.Copy(out, progressReader)
		fmt.Println()
	}
	return targetFile, err
}
