/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/logger"
)

const (
	gaugeExcludeDirectories = "gauge_exclude_dirs"
	cptFileExtension        = ".cpt"
)

var ignoredDirectories = make(map[string]bool)

func add(value string) {
	value = strings.TrimSpace(value)
	if !filepath.IsAbs(value) {
		path, err := filepath.Abs(filepath.Join(config.ProjectRoot, value))
		if err != nil {
			logger.Errorf(true, "Error getting absolute path. %v", err)
			return
		}
		value = path
	}
	ignoredDirectories[value] = true
}

func addDirectories(value string) {
	for _, dir := range strings.Split(value, ",") {
		add(dir)
	}
}

func addIgnoredDirectories() {
	ignoredDirectories[filepath.Join(config.ProjectRoot, "gauge_bin")] = true
	ignoredDirectories[filepath.Join(config.ProjectRoot, "reports")] = true
	ignoredDirectories[filepath.Join(config.ProjectRoot, "logs")] = true
	ignoredDirectories[filepath.Join(config.ProjectRoot, common.EnvDirectoryName)] = true
	addDirFromEnv(env.GaugeReportsDir, add)
	addDirFromEnv(env.LogsDirectory, add)
	addDirFromEnv(gaugeExcludeDirectories, addDirectories)
}

func addDirFromEnv(name string, add func(value string)) {
	value := os.Getenv(name)
	if value != "" {
		add(value)
	}
}

// findFilesIn Finds all the files in the directory of a given extension
func findFilesIn(dirRoot string, isValidFile func(path string) bool, shouldSkip func(path string, f os.FileInfo) bool) []string {
	absRoot, _ := filepath.Abs(dirRoot)
	files := common.FindFilesInDir(absRoot, isValidFile, shouldSkip)
	return files
}

// FindSpecFilesIn Finds spec files in the given directory
var FindSpecFilesIn = func(dir string) []string {
	return findFilesIn(dir, IsValidSpecExtension, func(path string, f os.FileInfo) bool {
		return false
	})
}

// IsValidSpecExtension Checks if the path has a spec file extension
func IsValidSpecExtension(path string) bool {
	for _, ext := range env.GaugeSpecFileExtensions() {
		if ext == strings.ToLower(filepath.Ext(path)) {
			return true
		}
	}
	return false
}

// FindConceptFilesIn Finds the concept files in specified directory
func FindConceptFilesIn(dir string) []string {
	addIgnoredDirectories()
	return findFilesIn(dir, IsValidConceptExtension, func(path string, f os.FileInfo) bool {
		if !f.IsDir() {
			return false
		}
		_, ok := ignoredDirectories[path]
		return strings.HasPrefix(f.Name(), ".") || ok
	})
}

// IsValidConceptExtension Checks if the path has a concept file extension
func IsValidConceptExtension(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == cptFileExtension
}

// IsConcept Returns true if concept file
func IsConcept(path string) bool {
	return IsValidConceptExtension(path)
}

// IsSpec Returns true if spec file file
func IsSpec(path string) bool {
	return IsValidSpecExtension(path)
}

// IsGaugeFile Returns true if spec file or concept file
func IsGaugeFile(path string) bool {
	return IsConcept(path) || IsSpec(path)
}

// IsGaugeFile Returns true if spec file or concept file
func GaugeFileExtensions() []string {
	return append(env.GaugeSpecFileExtensions(), cptFileExtension)
}

// FindAllNestedDirs returns list of all nested directories in given path
func FindAllNestedDirs(dir string) []string {
	var nestedDirs []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && !(path == dir) {
			nestedDirs = append(nestedDirs, path)
		}
		return nil
	})
	if err != nil {
		logger.Errorf(false, "Failed to find nested directories for %s: %s", dir, err.Error())
	}
	return nestedDirs
}

// IsDir reports whether path describes a directory.
func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// GetSpecFiles returns the list of spec files present at the given path.
// If the path itself represents a spec file, it returns the same.
var exitWithMessage = func(message string) {
	logger.Errorf(true, message)
	os.Exit(1)
}
var GetSpecFiles = func(paths []string) []string {
	var specFiles []string
	for _, path := range paths {
		if !common.FileExists(path) {
			exitWithMessage(fmt.Sprintf("Specs directory %s does not exists.", path))
		}
		if common.DirExists(path) {
			specFilesInpath := FindSpecFilesIn(path)
			if len(specFilesInpath) < 1 {
				exitWithMessage(fmt.Sprintf("No specifications found in %s.", path))
			}
			specFiles = append(specFiles, specFilesInpath...)
		} else if IsValidSpecExtension(path) {
			f, _ := filepath.Abs(path)
			specFiles = append(specFiles, f)
		}
	}
	return specFiles
}

// GetConceptFiles returns the list of concept files present in the PROJECTROOT
var GetConceptFiles = func() []string {
	projRoot := config.ProjectRoot
	if projRoot == "" {
		logger.Fatalf(true, "Failed to get project root.")
	}
	absPath, err := filepath.Abs(projRoot)
	if err != nil {
		logger.Fatalf(true, "Error getting absolute path. %v", err)
	}
	files := FindConceptFilesIn(absPath)
	var specFromProperties = os.Getenv(env.SpecsDir)
	if specFromProperties == "" {
		return files
	}
	var specDirectories = strings.Split(specFromProperties, ",")
	for _, dir := range specDirectories {
		absSpecPath, err := filepath.Abs(strings.TrimSpace(dir))
		if err != nil {
			logger.Fatalf(true, "Error getting absolute path. %v", err)
		}
		files = append(files, FindConceptFilesIn(absSpecPath)...)
	}
	return removeDuplicateValues(files)
}

func removeDuplicateValues(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// SaveFile saves contents at the given path
func SaveFile(fileName string, content string, backup bool) {
	err := common.SaveFile(fileName, content, backup)
	if err != nil {
		logger.Errorf(true, "Failed to refactor '%s': %s\n", fileName, err.Error())
	}
}

func RelPathToProjectRoot(path string) string {
	return strings.TrimPrefix(path, config.ProjectRoot+string(filepath.Separator))
}

// GetPathToFile returns the path to a given file from the Project root
func GetPathToFile(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(config.ProjectRoot, path)
}

// Remove removes all the files and directories recursively for the given path
func Remove(dir string) {
	err := common.Remove(dir)
	if err != nil {
		logger.Warningf(true, "Failed to remove directory %s. Remove it manually. %s", dir, err.Error())
	}
}

// RemoveTempDir removes the temp dir
func RemoveTempDir() {
	Remove(common.GetTempDir())
}

// GetLinesFromText gets lines of a text in an array
func GetLinesFromText(text string) []string {
	text = strings.Replace(text, "\r\n", "\n", -1)
	return strings.Split(text, "\n")
}

// GetLineCount give no of lines in given text
func GetLineCount(text string) int {
	return len(GetLinesFromText(text))
}

func OpenFile(fileName string) (io.Writer, error) {
	return os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
}
