package util

import (
	"fmt"
	"runtime"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/logger"
)

func NumberOfCores() int {
	return runtime.NumCPU()
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// DownloadAndUnzip downloads the zip file from given download link and unzips it.
// Returns the unzipped file path.
func DownloadAndUnzip(downloadLink string) (string, error) {
	logger.Debug("Downloading => %s", downloadLink)
	tempDir := common.GetTempDir()
	defer common.Remove(tempDir)
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
