// Common functions shared across all files
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	newDirectoryPermissions = 0755
	newFilePermissions      = 0644
	defaultEnvJSONFileName  = "default.json"
	envDirectoryName        = "env"
)

func getSearchPathForSharedFiles() []string {
	return []string{"/usr/local/share/twist2", "/usr/share/twist2"}
}

func getLanguageJSONFilePath(language string) (string, error) {
	searchPaths := getSearchPathForSharedFiles()
	for _, p := range searchPaths {
		languageJson := filepath.Join(p, "languages", fmt.Sprintf("%s.json", language))
		_, err := os.Stat(languageJson)
		if err == nil {
			return languageJson, nil
		}
	}

	return "", errors.New(fmt.Sprintf("Failed to find the implementation for: %s", language))
}

func getSkeletonFilePath(filename string) (string, error) {
	searchPaths := getSearchPathForSharedFiles()
	for _, p := range searchPaths {
		skelFile := filepath.Join(p, "skel", filename)
		if fileExists(skelFile) {
			return skelFile, nil
		}
	}

	return "", errors.New(fmt.Sprintf("Failed to find the skeleton file: %s", filename))
}

func isASupportedLanguage(language string) bool {
	_, err := getLanguageJSONFilePath(language)
	return err == nil
}

func readFileContents(file string) string {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Failed to read: %s. %s\n", file, err.Error())
		os.Exit(1)
	}

	return string(bytes)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func dirExists(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	if err == nil && stat.IsDir() {
		return true
	}

	return false
}

func getUniqueId() int64 {
	return time.Now().UnixNano()
}

func copyFile(src, dest string) error {
	if !fileExists(src) {
		return errors.New(fmt.Sprintf("%s doesn't exist", src))
	}

	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, b, newFilePermissions)
	if err != nil {
		return err
	}

	return nil
}

// A wrapper around os.SetEnv
// This handles duplicate env variable assignments and fails
func setEnvVariable(key, value string) error {
	existingValue := os.Getenv(key)
	if existingValue == "" {
		if strings.TrimSpace(value) == "" {
			return nil
		}
		err := os.Setenv(key, value)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to set: %s = %s. %s", key, value, err.Error()))
		}
	} else {
		return errors.New(fmt.Sprintf("Failed to set: %s = %s. It is already assigned a value '%s'. Multiple assignments to same variable is not allowed", key, value, existingValue))
	}

	return nil
}
