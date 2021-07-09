/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	properties "github.com/magiconair/properties"
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
	allowCaseSensitiveTags         = "allow_case_sensitive_tags"
	allowMultilineStep             = "allow_multiline_step"
	allowScenarioDatatable         = "allow_scenario_datatable"
	allowFilteredParallelExecution = "allow_filtered_parallel_execution"
	enableMultithreading           = "enable_multithreading"
	// GaugeScreenshotsDir holds the location of screenshots dir
	GaugeScreenshotsDir     = "gauge_screenshots_dir"
	gaugeSpecFileExtensions = "gauge_spec_file_extensions"
)

var envVars map[string]string
var expansionVars map[string]string

var currentEnvironments = []string{}

// LoadEnv first generates the map of the env vars that needs to be set.
// It starts by populating the map with the env passed by the user in --env flag.
// It then adds the default values of the env vars which are required by Gauge,
// but are not present in the map.
//
// Finally, all the env vars present in the map are actually set in the shell.
func LoadEnv(envName string, errorHandler properties.ErrorHandlerFunc) error {
	properties.ErrorHandler = errorHandler
	allEnvs := strings.Split(envName, ",")

	envVars = make(map[string]string)
	expansionVars = make(map[string]string)

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
	err := checkEnvVarsExpanded()
	if err != nil {
		return fmt.Errorf("Failed to load env. %s", err.Error())
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
	addEnvVar(allowCaseSensitiveTags, "false")
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
	files := common.FindFilesInDir(envDirPath,
		isPropertiesFile,
		func(p string, f os.FileInfo) bool { return false },
	)
	gaugeProperties := properties.MustLoadFiles(files, properties.UTF8, false)
	processedProperties, err := GetProcessedPropertiesMap(gaugeProperties)
	if err != nil {
		return fmt.Errorf("Failed to parse properties in %s. %s", envDirPath, err.Error())
	}
	LoadEnvProperties(processedProperties)
	return nil
}

func GetProcessedPropertiesMap(propertiesMap *properties.Properties) (*properties.Properties, error) {
	for propertyKey := range propertiesMap.Map() {
		// Update properties if an env var is set.
		if envVarValue, present := os.LookupEnv(propertyKey); present && len(envVarValue) > 0 {
			if _, _, err := propertiesMap.Set(propertyKey, envVarValue); err != nil {
				return propertiesMap, fmt.Errorf("%s", err.Error())
			}
		}
		// Update the properties if it has already been added to envVars map.
		if _, ok := envVars[propertyKey]; ok {
			if _, _, err := propertiesMap.Set(propertyKey, envVars[propertyKey]); err != nil {
				return propertiesMap, fmt.Errorf("%s", err.Error())
			}
		}
	}
	return propertiesMap, nil
}

func LoadEnvProperties(propertiesMap *properties.Properties) {
	for propertyKey, propertyValue := range propertiesMap.Map() {
		if contains, matches := containsEnvVar(propertyValue); contains {
			for _, match := range matches {
				key, defaultValue := match[1], match[0]
				// Dont need to add to expansions if it's already set by env var
				if !isPropertySet(key) {
					expansionVars[key] = propertiesMap.GetString(key, defaultValue)
				}
			}
		}
		addEnvVar(propertyKey, propertiesMap.GetString(propertyKey, propertyValue))
	}
}

func checkEnvVarsExpanded() error {
	for key, value := range expansionVars {
		if _, ok := envVars[key]; ok {
			delete(expansionVars, key)
		}
		if err := isCircular(key, value); err != nil {
			return err
		}
	}
	if len(expansionVars) > 0 {
		keys := make([]string, 0, len(expansionVars))
		for key := range expansionVars {
			keys = append(keys, key)
		}
		return fmt.Errorf("[%s] env variable(s) are not set", strings.Join(keys, ", "))
	}
	return nil
}

func isCircular(key, value string) error {
	if keyValue, exists := envVars[key]; exists {
		if len(keyValue) > 0 {
			value = keyValue
		}
		_, err := properties.LoadString(fmt.Sprintf("%s=%s", key, value))
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
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

func addEnvVar(name, value string) {
	if _, ok := envVars[name]; !ok {
		envVars[name] = value
	}
}

func isPropertiesFile(path string) bool {
	return filepath.Ext(path) == ".properties"
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

// AllowCaseSensitiveTags determines if the casing is ignored in tags filtering
var AllowCaseSensitiveTags = func() bool {
	return convertToBool(allowCaseSensitiveTags, false)
}
