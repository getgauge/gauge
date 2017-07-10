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
	"path/filepath"
	"runtime"
	"syscall"

	"strconv"
	"strings"

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
	logger.Infof("Downloading %s", filepath.Base(downloadLink))
	logger.Debugf("Download URL %s", downloadLink)
	downloadedFile, err := Download(downloadLink, tempDir, "", false)
	if err != nil {
		return "", err
	}
	logger.Debugf("Downloaded to %s", downloadedFile)

	unzippedPluginDir, err := common.UnzipArchive(downloadedFile, tempDir)
	if err != nil {
		return "", fmt.Errorf("Failed to Unzip file %s: %s", downloadedFile, err.Error())
	}
	logger.Debugf("Unzipped to => %s\n", unzippedPluginDir)

	return unzippedPluginDir, nil
}

// IsProcessRunning checks if the process with the given process id is still running
func IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	if !IsWindows() {
		return process.Signal(syscall.Signal(0)) == nil
	}

	processState, err := process.Wait()
	if err != nil {
		return false
	}
	if processState.Exited() {
		return false
	}

	return true
}

// SetWorkingDir sets the current working directory to specified location
func SetWorkingDir(workingDir string) {
	targetDir, err := filepath.Abs(workingDir)
	if err != nil {
		logger.Fatalf("Unable to set working directory : %s", err.Error())
	}

	if !common.DirExists(targetDir) {
		err = os.Mkdir(targetDir, 0777)
		if err != nil {
			logger.Fatalf("Unable to set working directory : %s", err.Error())
		}
	}

	err = os.Chdir(targetDir)
	if err != nil {
		logger.Fatalf("Unable to set working directory : %s", err.Error())
	}

	_, err = os.Getwd()
	if err != nil {
		logger.Fatalf("Unable to set working directory : %s", err.Error())
	}
}

func ConvertToBool(value, property string, defaultValue bool) bool {
	boolValue, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		logger.Warningf("Incorrect value for %s in property file. Cannot convert %s to boolean.", property, value)
		logger.Warningf("Using default value %v for property %s.", defaultValue, property)
		return defaultValue
	}
	return boolValue
}
