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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
)

func init() {
	AcceptedExtensions[".spec"] = true
	AcceptedExtensions[".md"] = true
}

var AcceptedExtensions = make(map[string]bool)

// findFilesIn Finds all the files in the directory of a given extension
func findFilesIn(dirRoot string, isValidFile func(path string) bool) []string {
	absRoot, _ := filepath.Abs(dirRoot)
	files := common.FindFilesInDir(absRoot, isValidFile)
	return files
}

// FindSpecFilesIn Finds spec files in the given directory
func FindSpecFilesIn(dir string) []string {
	return findFilesIn(dir, IsValidSpecExtension)
}

// IsValidSpecExtension Checks if the path has a spec file extension
func IsValidSpecExtension(path string) bool {
	return AcceptedExtensions[filepath.Ext(path)]
}

// FindConceptFilesIn Finds the concept files in specified directory
func FindConceptFilesIn(dir string) []string {
	return findFilesIn(dir, IsValidConceptExtension)
}

// IsValidConceptExtension Checks if the path has a concept file extension
func IsValidConceptExtension(path string) bool {
	return filepath.Ext(path) == ".cpt"
}

// IsConcept Returns true if concept file
func IsConcept(path string) bool {
	return IsValidConceptExtension(path)
}

// IsSpec Returns true if spec file file
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
		logger.Error("Failed to refactor '%s': %s\n", fileName, err.Error())
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
