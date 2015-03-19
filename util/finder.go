package util

import (
	"path/filepath"
	"github.com/getgauge/common"
)

func init() {
	acceptedExtensions[".spec"] = true
	acceptedExtensions[".md"] = true
}

var acceptedExtensions = make(map[string]bool)

func findFilesIn(dirRoot string, isValidFile func(path string) bool) []string {
	absRoot, _ := filepath.Abs(dirRoot)
	files := common.FindFilesInDir(absRoot, isValidFile)
	return files
}

func findSpecFilesIn(dir string) []string {
	return findFilesIn(dir, isValidSpecExtension)
}

func isValidSpecExtension(path string) bool {
	return acceptedExtensions[filepath.Ext(path)]
}

func findConceptFilesIn(dir string) []string {
	return findFilesIn(dir, isValidConceptExtension)
}

func isValidConceptExtension(path string) bool {
	return filepath.Ext(path) == ".cpt"
}
