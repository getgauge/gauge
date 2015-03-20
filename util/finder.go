package util

import (
	"path/filepath"
	"github.com/getgauge/common"
	"io/ioutil"
	"fmt"
	"os"
)

func init() {
	AcceptedExtensions[".spec"] = true
	AcceptedExtensions[".md"] = true
}

var AcceptedExtensions = make(map[string]bool)

func findFilesIn(dirRoot string, isValidFile func(path string) bool) []string {
	absRoot, _ := filepath.Abs(dirRoot)
	files := common.FindFilesInDir(absRoot, isValidFile)
	return files
}

func FindSpecFilesIn(dir string) []string {
	return findFilesIn(dir, IsValidSpecExtension)
}

func IsValidSpecExtension(path string) bool {
	return AcceptedExtensions[filepath.Ext(path)]
}

func FindConceptFilesIn(dir string) []string {
	return findFilesIn(dir, IsValidConceptExtension)
}

func IsValidConceptExtension(path string) bool {
	return filepath.Ext(path) == ".cpt"
}

func CreateFileIn(dir string, fileName string, data []byte) (string, error) {
	tempFile, err := ioutil.TempFile(dir, "gauge1")
	fullFileName := dir + fmt.Sprintf("%c",filepath.Separator)+fileName
	err = os.Rename(tempFile.Name(), fullFileName)
	err = ioutil.WriteFile(fullFileName, data, 0644)
	return fullFileName, err
}

func CreateDirIn(dir string, dirName string) (string, error) {
	tempDir , err := ioutil.TempDir(dir, dirName)
	fullDirName := dir + fmt.Sprintf("%c",filepath.Separator)+ dirName
	err = os.Rename(tempDir, fullDirName)
	return fullDirName, err
}
