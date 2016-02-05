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
	"os"
	"runtime"
	"syscall"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/logger"
)

// NumberOfCores returns the number of CPU cores on the system
func NumberOfCores() int {
	return runtime.NumCPU()
}

// IsWindows returns if Gauge is running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// DownloadAndUnzip downloads the zip file from given download link and unzips it.
// Returns the unzipped file path.
func DownloadAndUnzip(downloadLink string, tempDir string) (string, error) {
	logger.Debug("Downloading => %s", downloadLink)
	downloadedFile, err := common.Download(downloadLink, tempDir)
	if err != nil {
		return "", fmt.Errorf("Could not download file %s: %s", downloadLink, err.Error())
	}
	logger.Debug("Downloaded to %s", downloadedFile)

	unzippedPluginDir, err := common.UnzipArchive(downloadedFile, tempDir)
	if err != nil {
		return "", fmt.Errorf("Failed to Unzip file %s: %s", downloadedFile, err.Error())
	}
	logger.Debug("Unzipped to => %s\n", unzippedPluginDir)

	return unzippedPluginDir, nil
}

// IsProcessRunning checks if the process with the given process id is still running
func IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}
