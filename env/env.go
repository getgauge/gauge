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

package env

import (
	"fmt"
	"os"
	"path/filepath"

	"regexp"
	"strings"

	"github.com/dmotylev/goproperties"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
)

const (
	GaugeReportsDir     = "gauge_reports_dir"
	LogsDirectory       = "logs_directory"
	OverwriteReports    = "overwrite_reports"
	ScreenshotOnFailure = "screenshot_on_failure"
	SaveExecutionResult = "save_execution_result" // determines if last run result should be saved
)

var envVars map[string]string

var currentEnv = "default"

// LoadEnv first generates the map of the env vars that needs to be set.
// It starts by populating the map with the env passed by the user in --env flag.
// It then adds the default values of the env vars which are required by Gauge,
// but are not present in the map.
//
// Finally, all the env vars present in the map are actually set in the shell.
func LoadEnv(envName string) error {
	envVars = make(map[string]string)
	currentEnv = envName

	err := loadEnvDir(currentEnv)
	if err != nil {
		return fmt.Errorf("Failed to load env. %s", err.Error())
	}

	if currentEnv != "default" {
		err := loadEnvDir("default")
		if err != nil {
			return fmt.Errorf("Failed to load env. %s", err.Error())
		}
	}

	loadDefaultEnvVars()

	err = substituteEnvVars()
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}

	err = setEnvVars()
	if err != nil {
		return fmt.Errorf("Failed to load env. %s", err.Error())
	}
	return nil
}

func loadDefaultEnvVars() {
	addEnvVar(GaugeReportsDir, "reports")
	addEnvVar(LogsDirectory, "logs")
	addEnvVar(OverwriteReports, "true")
	addEnvVar(ScreenshotOnFailure, "true")
	addEnvVar(SaveExecutionResult, "false")
}

func loadEnvDir(envName string) error {
	envDirPath := filepath.Join(config.ProjectRoot, common.EnvDirectoryName, envName)
	if !common.DirExists(envDirPath) {
		if envName != "default" {
			return fmt.Errorf("%s environment does not exist", envName)
		}
		return nil
	}

	return filepath.Walk(envDirPath, loadEnvFile)
}

func loadEnvFile(path string, info os.FileInfo, err error) error {
	if !isPropertiesFile(path) {
		return nil
	}

	properties, err := properties.Load(path)
	if err != nil {
		return fmt.Errorf("Failed to parse: %s. %s", path, err.Error())
	}

	for property, value := range properties {
		addEnvVar(property, value)
	}

	return nil
}

func addEnvVar(name, value string) {
	if _, ok := envVars[name]; !ok {
		envVars[name] = value
	}
}

func isPropertiesFile(path string) bool {
	return filepath.Ext(path) == ".properties"
}

func substituteEnvVars() error {
	for name, value := range envVars {
		contains, matches := containsEnvVar(value)
		// if value contains an env var E.g. ${foo}
		if contains {
			for _, match := range matches {
				// check if match is from properties file
				// if not, get from system env
				envKey, property := match[0], match[1]
				propertyValue := envVars[property]
				if _, ok := envVars[property]; !ok {
					// error if env property is not found
					if !isPropertySet(property) {
						return fmt.Errorf("'%s' env property was not set.", property)
					}
					// get env var from system
					propertyValue = os.Getenv(property)
				}
				// replace env key with property value
				value = strings.Replace(value, envKey, propertyValue, -1)
			}
			// overwrite the envVar value
			envVars[name] = value
		}
	}
	return nil
}

func setEnvVars() error {
	for name, value := range envVars {
		if !isPropertySet(name) {
			err := common.SetEnvVariable(name, value)
			if err != nil {
				return fmt.Errorf("%s", err.Error())
			}
		}
	}
	return nil
}

func isPropertySet(property string) bool {
	return len(os.Getenv(property)) > 0
}

func containsEnvVar(value string) (contains bool, matches [][]string) {
	// match for any ${foo}
	r, _ := regexp.Compile("\\$\\{(\\w+)\\}")
	contains = r.MatchString(value)
	if contains {
		matches = r.FindAllStringSubmatch(value, -1)
	}
	return
}

// CurrentEnv returns the value of currentEnv
func CurrentEnv() string {
	return currentEnv
}
