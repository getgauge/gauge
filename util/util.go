/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/env"
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
func DownloadAndUnzip(downloadLink, tempDir string) (string, error) {
	logger.Debugf(true, "Download URL %s", downloadLink)
	downloadedFile, err := Download(downloadLink, tempDir, "", false)
	if err != nil {
		return "", err
	}
	logger.Debugf(true, "Downloaded to %s", downloadedFile)

	unzippedPluginDir, err := common.UnzipArchive(downloadedFile, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to unzip file %s: %s", downloadedFile, err.Error())
	}
	logger.Debugf(true, "Unzipped to => %s", unzippedPluginDir)
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
		logger.Fatalf(true, "Unable to set working directory : %s", err.Error())
	}

	if !common.DirExists(targetDir) {
		err = os.Mkdir(targetDir, 0750)
		if err != nil {
			logger.Fatalf(true, "Unable to set working directory : %s", err.Error())
		}
	}

	err = os.Chdir(targetDir)
	if err != nil {
		logger.Fatalf(true, "Unable to set working directory : %s", err.Error())
	}

	_, err = os.Getwd()
	if err != nil {
		logger.Fatalf(true, "Unable to set working directory : %s", err.Error())
	}
}

// GetSpecDirs returns the specification directory.
// It checks whether the environment variable for gauge_specs_dir is set.
// It returns 'specs' otherwise
func GetSpecDirs() []string {
	var specFromProperties = os.Getenv(env.SpecsDir)
	if specFromProperties != "" {
		var specDirectories = strings.Split(specFromProperties, ",")
		for index, ele := range specDirectories {
			specDirectories[index] = strings.TrimSpace(ele)
		}
		return specDirectories
	}
	return []string{common.SpecsDirectoryName}
}

func ListContains(list []string, val string) bool {
	for _, s := range list {
		if s == val {
			return true
		}
	}
	return false
}

func GetFileContents(filepath string) (string, error) {
	return common.ReadFileContents(GetPathToFile(filepath))
}
