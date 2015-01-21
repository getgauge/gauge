package config

import (
	"github.com/getgauge/common"
	"strconv"
	"time"
)

const (
	gaugeRepositoryUrl      = "gauge_repository_url"
	apiRefreshInterval      = "api_refresh_interval"
	runnerConnectionTimeout = "runner_connection_timeout"
	pluginConnectionTimeout = "plugin_connection_timeout"
	runnerKillTimeOut       = "runner_kill_timeout"

	defaultApiRefreshInterval      = time.Second * 2
	defaultRunnerConnectionTimeout = time.Second * 25
	defaultPluginConnectionTimeout = time.Second * 10
	defaultRunnerKillTimeOut       = time.Second * 2
	defaultRefactorTimeout         = time.Second * 10
	defaultRunnerAPIRequestTimeout = time.Second * 2
)

func ApiRefreshInterval() time.Duration {
	intervalString := getFromConfig(apiRefreshInterval)
	return convertToTime(intervalString, defaultApiRefreshInterval)
}

func RunnerConnectionTimeout() time.Duration {
	intervalString := getFromConfig(runnerConnectionTimeout)
	return convertToTime(intervalString, defaultRunnerConnectionTimeout)
}

func PluginConnectionTimeout() time.Duration {
	intervalString := getFromConfig(pluginConnectionTimeout)
	return convertToTime(intervalString, defaultPluginConnectionTimeout)
}

func RunnerKillTimeout() time.Duration {
	intervalString := getFromConfig(runnerKillTimeOut)
	return convertToTime(intervalString, defaultRunnerKillTimeOut)
}

func RefactorTimeout() time.Duration {
	return defaultRefactorTimeout

}

func RunnerAPIRequestTimeout() time.Duration {
	return defaultRunnerAPIRequestTimeout
}

func GaugeRepositoryUrl() string {
	return getFromConfig(gaugeRepositoryUrl)
}

func convertToTime(value string, defaultValue time.Duration) time.Duration {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return time.Millisecond * time.Duration(intValue)
}

func getFromConfig(propertyName string) string {
	config, err := common.GetGaugeConfiguration()
	if err != nil {
		return ""
	}
	return config[propertyName]
}
