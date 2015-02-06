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
	pluginKillTimeOut       = "plugin_kill_timeout"

	defaultApiRefreshInterval      = time.Second * 3
	defaultRunnerConnectionTimeout = time.Second * 25
	defaultPluginConnectionTimeout = time.Second * 10
	defaultPluginKillTimeout       = time.Second * 4
	defaultRefactorTimeout         = time.Second * 10
	defaultRunnerRequestTimeout    = time.Second * 3
)

var ProjectRoot string

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

func PluginKillTimeout() time.Duration {
	intervalString := getFromConfig(pluginKillTimeOut)
	return convertToTime(intervalString, defaultPluginKillTimeout)
}

func RefactorTimeout() time.Duration {
	return defaultRefactorTimeout
}

func RunnerRequestTimeout() time.Duration {
	return defaultRunnerRequestTimeout
}

func GaugeRepositoryUrl() string {
	return getFromConfig(gaugeRepositoryUrl)
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
