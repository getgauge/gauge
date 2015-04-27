package util

import (
	"github.com/getgauge/common"
	"io/ioutil"
	"os"
	"path/filepath"
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

// Finds the concept files in specified directory
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
	nestedDirs := make([]string, 0)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !(path == dir) && err == nil {
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
