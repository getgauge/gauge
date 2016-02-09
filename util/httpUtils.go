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

// Download fires a HTTP GET request to download a resource to target directory
func Download(url, targetDir string) (string, error) {
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

	_, err = io.Copy(out, resp.Body)
	return "", err
}
