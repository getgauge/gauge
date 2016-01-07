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

package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/getgauge/common"
	"github.com/op/go-logging"
)

const (
	gaugeRepositoryUrl      = "gauge_repository_url"
	gaugeUpdateUrl          = "gauge_update_url"
	runnerConnectionTimeout = "runner_connection_timeout"
	pluginConnectionTimeout = "plugin_connection_timeout"
	pluginKillTimeOut       = "plugin_kill_timeout"
	runnerRequestTimeout    = "runner_request_timeout"
	allowUpdates            = "allow_updates"

	defaultRunnerConnectionTimeout = time.Second * 25
	defaultPluginConnectionTimeout = time.Second * 10
	defaultPluginKillTimeout       = time.Second * 4
	defaultRefactorTimeout         = time.Second * 10
	defaultRunnerRequestTimeout    = time.Second * 3
	LayoutForTimeStamp             = "Jan 2, 2006 at 3:04pm"
)

var APILog = logging.MustGetLogger("gauge-api")
var ProjectRoot string

// Timeout in milliseconds for making a connection to the language runner
func RunnerConnectionTimeout() time.Duration {
	intervalString := getFromConfig(runnerConnectionTimeout)
	return convertToTime(intervalString, defaultRunnerConnectionTimeout, runnerConnectionTimeout)
}

// Timeout in milliseconds for making a connection to plugins
func PluginConnectionTimeout() time.Duration {
	intervalString := getFromConfig(pluginConnectionTimeout)
	return convertToTime(intervalString, defaultPluginConnectionTimeout, pluginConnectionTimeout)
}

// Timeout in milliseconds for a plugin to stop after a kill message has been sent
func PluginKillTimeout() time.Duration {
	intervalString := getFromConfig(pluginKillTimeOut)
	return convertToTime(intervalString, defaultPluginKillTimeout, pluginKillTimeOut)
}

func AllowUpdates() bool {
	allow := getFromConfig(allowUpdates)
	return convertToBool(allow, allowUpdates, true)
}

func convertToBool(value string, property string, defaultValue bool) bool {
	boolValue, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		APILog.Warning("Incorrect value for %s in property file. Cannot convert %s to boolean.", property, value)
		return defaultValue
	}
	return boolValue
}

func RefactorTimeout() time.Duration {
	return defaultRefactorTimeout
}

// Timeout in milliseconds for requests from the language runner.
func RunnerRequestTimeout() time.Duration {
	intervalString := os.Getenv(runnerRequestTimeout)
	if intervalString == "" {
		intervalString = getFromConfig(runnerRequestTimeout)
	}
	return convertToTime(intervalString, defaultRunnerRequestTimeout, runnerRequestTimeout)
}

func GaugeRepositoryUrl() string {
	return getFromConfig(gaugeRepositoryUrl)
}

func GaugeUpdateUrl() string {
	return getFromConfig(gaugeUpdateUrl)
}

func SetProjectRoot(args []string) error {
	if ProjectRoot != "" {
		return setCurrentProjectEnvVariable()
	}
	value := ""
	if len(args) != 0 {
		value = args[0]
	}
	root, err := common.GetProjectRootFromSpecPath(value)
	if err != nil {
		return err
	}
	ProjectRoot = root
	return setCurrentProjectEnvVariable()
}

func setCurrentProjectEnvVariable() error {
	return common.SetEnvVariable(common.GaugeProjectRootEnv, ProjectRoot)
}

func convertToTime(value string, defaultValue time.Duration, name string) time.Duration {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		APILog.Warning("Incorrect value for %s in property file. Cannot convert %s to time", name, value)
		return defaultValue
	}
	return time.Millisecond * time.Duration(intValue)
}

var getFromConfig = func(propertyName string) string {
	config, err := common.GetGaugeConfiguration()
	if err != nil {
		APILog.Warning("Failed to get configuration from Gauge properties file. Error: %s", err.Error())
		return ""
	}
	return config[propertyName]
}
