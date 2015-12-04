package util

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"fmt"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
)

func init() {
	AcceptedExtensions[".spec"] = true
	AcceptedExtensions[".md"] = true
}

var AcceptedExtensions = make(map[string]bool)

// Finds all the files in the directory of a given extension
func findFilesIn(dirRoot string, isValidFile func(path string) bool) []string {
	absRoot, _ := filepath.Abs(dirRoot)
	files := common.FindFilesInDir(absRoot, isValidFile)
	return files
}

// Finds spec files in the given directory
func FindSpecFilesIn(dir string) []string {
	return findFilesIn(dir, IsValidSpecExtension)
}

// Checks if the path has a spec file extension
func IsValidSpecExtension(path string) bool {
	return AcceptedExtensions[filepath.Ext(path)]
}

// FindConceptFilesIn Finds the concept files in specified directory
func FindConceptFilesIn(dir string) []string {
	return findFilesIn(dir, IsValidConceptExtension)
}

// Checks if the path has a concept file extension
func IsValidConceptExtension(path string) bool {
	return filepath.Ext(path) == ".cpt"
}

// Returns true if concept file
func IsConcept(path string) bool {
	return IsValidConceptExtension(path)
}

// Returns true if spec file file
func IsSpec(path string) bool {
	return IsValidSpecExtension(path)
}

func FindAllNestedDirs(dir string) []string {
	var nestedDirs []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && !(path == dir) {
			nestedDirs = append(nestedDirs, path)
		}
		return nil
	})
	return nestedDirs
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func CreateFileIn(dir string, fileName string, data []byte) (string, error) {
	os.MkdirAll(dir, 0755)
	err := ioutil.WriteFile(filepath.Join(dir, fileName), data, 0644)
	return filepath.Join(dir, fileName), err
}

func CreateDirIn(dir string, dirName string) (string, error) {
	tempDir, err := ioutil.TempDir(dir, dirName)
	fullDirName := filepath.Join(dir, dirName)
	err = os.Rename(tempDir, fullDirName)
	return fullDirName, err
}

func GetSpecFiles(specSource string) []string {
	var specFiles []string
	if common.DirExists(specSource) {
		specFiles = append(specFiles, FindSpecFilesIn(specSource)...)
	} else if common.FileExists(specSource) && IsValidSpecExtension(specSource) {
		specFile, _ := filepath.Abs(specSource)
		specFiles = append(specFiles, specFile)
	}
	return specFiles
}

func SaveFile(fileName string, content string, backup bool) {
	err := common.SaveFile(fileName, content, backup)
	if err != nil {
		logger.Error("Failed to refactor '%s': %s\n", fileName, err)
	}
}

func GetPathToFile(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(config.ProjectRoot, path)
}

func Remove(dir string) {
	err := common.Remove(dir)
	if err != nil {
		logger.Warning("Failed to remove directory %s. Remove it manually. %s", dir, err.Error())
	}
}

func RemoveTempDir() {
	Remove(common.GetTempDir())
}

// DownloadAndUnzip downloads the zip file from given download link and unzips it.
// Returns the unzipped file path.
func DownloadAndUnzip(downloadLink string) (string, error) {
	logger.Debug("Downloading => %s", downloadLink)
	downloadedFile, err := common.DownloadToTempDir(downloadLink)
	if err != nil {
		return "", fmt.Errorf("Could not download file %s: %s", downloadLink, err.Error())
	}
	logger.Info("Downloaded to %s", downloadedFile)
	defer Remove(downloadedFile)

	unzippedPluginDir, err := common.UnzipArchive(downloadedFile)
	if err != nil {
		return "", fmt.Errorf("Failed to Unzip file %s: %s", downloadedFile, err.Error())
	}
	logger.Debug("Unzipped to => %s\n", unzippedPluginDir)

	return unzippedPluginDir, nil
}
