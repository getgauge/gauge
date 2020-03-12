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
	"strconv"

	"regexp"
	"strings"

	properties "github.com/dmotylev/goproperties"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
)

const (
	// SpecsDir holds the location of spec files
	SpecsDir = "gauge_specs_dir"
	// GaugeReportsDir holds the location of reports
	GaugeReportsDir = "gauge_reports_dir"
	// GaugeEnvironment holds the name of the current environment
	GaugeEnvironment = "gauge_environment"
	// LogsDirectory holds the location of log files
	LogsDirectory = "logs_directory"
	// OverwriteReports = false will create a new directory for reports
	// for every run.
	OverwriteReports = "overwrite_reports"
	// ScreenshotOnFailure indicates if failure should invoke screenshot
	ScreenshotOnFailure = "screenshot_on_failure"
	saveExecutionResult = "save_execution_result"
	// CsvDelimiter holds delimiter used to parse csv files
	CsvDelimiter                   = "csv_delimiter"
	allowMultilineStep             = "allow_multiline_step"
	allowScenarioDatatable         = "allow_scenario_datatable"
	allowFilteredParallelExecution = "allow_filtered_parallel_execution"
	enableMultithreading           = "enable_multithreading"
	// GaugeScreenshotsDir holds the location of screenshots dir
	GaugeScreenshotsDir     = "gauge_screenshots_dir"
	gaugeSpecFileExtensions = "gauge_spec_file_extensions"
)

var envVars map[string]string

var currentEnvironments = []string{}

// LoadEnv first generates the map of the env vars that needs to be set.
// It starts by populating the map with the env passed by the user in --env flag.
// It then adds the default values of the env vars which are required by Gauge,
// but are not present in the map.
//
// Finally, all the env vars present in the map are actually set in the shell.
func LoadEnv(envName string) error {
	allEnvs := strings.Split(envName, ",")

	envVars = make(map[string]string)

	defaultEnvLoaded := false
	for _, env := range allEnvs {
		env = strings.TrimSpace(env)

		err := loadEnvDir(env)
		if err != nil {
			return fmt.Errorf("Failed to load env. %s", err.Error())
		}

		if env == common.DefaultEnvDir {
			defaultEnvLoaded = true
		} else {
			currentEnvironments = append(currentEnvironments, env)
		}
	}

	if !defaultEnvLoaded {
		err := loadEnvDir(common.DefaultEnvDir)
		if err != nil {
			return fmt.Errorf("Failed to load env. %s", err.Error())
		}
	}

	loadDefaultEnvVars()

	err := substituteEnvVars()
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
	addEnvVar(SpecsDir, "specs")
	addEnvVar(GaugeReportsDir, "reports")
	addEnvVar(GaugeEnvironment, common.DefaultEnvDir)
	addEnvVar(LogsDirectory, "logs")
	addEnvVar(OverwriteReports, "true")
	addEnvVar(ScreenshotOnFailure, "true")
	addEnvVar(saveExecutionResult, "false")
	addEnvVar(CsvDelimiter, ",")
	addEnvVar(allowMultilineStep, "false")
	addEnvVar(allowScenarioDatatable, "false")
	addEnvVar(allowFilteredParallelExecution, "false")
	defaultScreenshotDir := filepath.Join(config.ProjectRoot, common.DotGauge, "screenshots")
	addEnvVar(GaugeScreenshotsDir, defaultScreenshotDir)
	addEnvVar(gaugeSpecFileExtensions, ".spec, .md")
	err := os.MkdirAll(defaultScreenshotDir, 0750)
	if err != nil {
		logger.Warningf(true, "Could not create screenshot dir at %s", err.Error())
	}
}

func loadEnvDir(envName string) error {
	envDirPath := filepath.Join(config.ProjectRoot, common.EnvDirectoryName, envName)
	if !common.DirExists(envDirPath) {
		if envName != common.DefaultEnvDir {
			return fmt.Errorf("%s environment does not exist", envName)
		}
		return nil
	}
	addEnvVar(GaugeEnvironment, envName)
	logger.Debugf(true, "'%s' set to '%s'", GaugeEnvironment, envName)
	return filepath.Walk(envDirPath, loadEnvFile)
}

func loadEnvFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !isPropertiesFile(path) {
		return nil
	}

	gaugeProperties, err1 := properties.Load(path)
	if err1 != nil {
		return fmt.Errorf("Failed to parse: %s. %s", path, err1.Error())
	}

	for property, value := range gaugeProperties {
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
				// error if env property is not found
				if !isPropertySet(property) {
					return fmt.Errorf("'%s' env variable was not set", property)
				}
				// get env var from system
				propertyValue := os.Getenv(property)
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
	rStr := `\$\{(\w+)\}`
	r, err := regexp.Compile(rStr)
	if err != nil {
		logger.Errorf(false, "Unable to compile regex %s: %s", rStr, err.Error())
	}
	contains = r.MatchString(value)
	if contains {
		matches = r.FindAllStringSubmatch(value, -1)
	}
	return
}

// comma-separated value of environments
func CurrentEnvironments() string {
	if len(currentEnvironments) == 0 {
		currentEnvironments = append(currentEnvironments, common.DefaultEnvDir)
	}
	return strings.Join(currentEnvironments, ",")
}

func convertToBool(property string, defaultValue bool) bool {
	v := os.Getenv(property)
	boolValue, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		logger.Warningf(true, "Incorrect value for %s in property file. Cannot convert %s to boolean.", property, v)
		logger.Warningf(true, "Using default value %v for property %s.", defaultValue, property)
		return defaultValue
	}
	return boolValue
}

// AllowFilteredParallelExecution - feature toggle for filtered parallel execution
var AllowFilteredParallelExecution = func() bool {
	return convertToBool(allowFilteredParallelExecution, false)
}

// AllowScenarioDatatable -feature toggle for datatables in scenario
var AllowScenarioDatatable = func() bool {
	return convertToBool(allowScenarioDatatable, false)
}

// AllowMultiLineStep - feature toggle for newline in step text
var AllowMultiLineStep = func() bool {
	return convertToBool(allowMultilineStep, false)
}

// SaveExecutionResult determines if last run result should be saved
var SaveExecutionResult = func() bool {
	return convertToBool(saveExecutionResult, false)
}

// EnableMultiThreadedExecution determines if threads should be used instead of process
// for each parallel stream
var EnableMultiThreadedExecution = func() bool {
	return convertToBool(enableMultithreading, false)
}

var GaugeSpecFileExtensions = func() []string {
	e := os.Getenv(gaugeSpecFileExtensions)
	if e == "" {
		e = ".spec, .md" //this was earlier hardcoded, this is a failsafe if env isn't set
	}
	exts := strings.Split(strings.TrimSpace(e), ",")
	var allowedExts = []string{}
	for _, ext := range exts {
		e := strings.TrimSpace(ext)
		if e != "" {
			allowedExts = append(allowedExts, e)
		}
	}
	return allowedExts
}
