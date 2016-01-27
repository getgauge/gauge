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

	"github.com/dmotylev/goproperties"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
)

const (
	envDefaultDirName = "default"
)

var defaultProperties map[string]string

var currentEnv = "default"

// LoadEnv first loads the default env properties and then the user specified env properties.
// This way user specified env variable can overwrite default if required
func LoadEnv(envName string) {
	currentEnv = envName

	err := loadDefaultProperties()
	if err != nil {
		logger.Fatalf("Failed to load the default property. %s", err.Error())
	}

	err = loadEnvDir(currentEnv)
	if err != nil {
		logger.Fatalf("Failed to load env. %s", err.Error())
	}
}

func loadDefaultProperties() error {
	defaultProperties = make(map[string]string)
	defaultProperties["gauge_reports_dir"] = "reports"
	defaultProperties["overwrite_reports"] = "true"
	defaultProperties["screenshot_on_failure"] = "true"
	defaultProperties["logs_directory"] = "logs"

	for property, value := range defaultProperties {
		if !isPropertySet(property) {
			if err := common.SetEnvVariable(property, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadEnvDir(envDir string) error {
	envDirPath := filepath.Join(config.ProjectRoot, common.EnvDirectoryName, envDir)
	if !common.DirExists(envDirPath) {
		return fmt.Errorf("%s environment does not exist", envDir)
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
		if canOverwriteProperty(property) {
			err := common.SetEnvVariable(property, value)
			if err != nil {
				return fmt.Errorf("%s: %s", path, err.Error())
			}
		}
	}

	return nil
}

func isPropertiesFile(path string) bool {
	return filepath.Ext(path) == ".properties"
}

func canOverwriteProperty(property string) bool {
	if !isPropertySet(property) {
		return true
	}

	defaultVal, ok := defaultProperties[property]
	if !ok {
		return true
	}

	return defaultVal == os.Getenv(property)
}

func isPropertySet(property string) bool {
	return len(os.Getenv(property)) > 0
}

// CurrentEnv returns the value of currentEnv
func CurrentEnv() string {
	return currentEnv
}
