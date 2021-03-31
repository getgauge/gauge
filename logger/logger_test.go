/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"os"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/plugin/pluginInfo"
	"github.com/getgauge/gauge/version"
	logging "github.com/op/go-logging"
)

func TestGetLoggerShouldGetTheLoggerForGivenModule(t *testing.T) {
	Initialize(false, "info", CLI)

	l := loggersMap.getLogger("gauge-js")
	if l == nil {
		t.Error("Expected a logger to be initilized for gauge-js")
	}
}

func TestLoggerInitWithInfoLevel(t *testing.T) {
	Initialize(false, "info", CLI)

	if !loggersMap.getLogger(gaugeModuleID).IsEnabledFor(logging.INFO) {
		t.Error("Expected gaugeLog to be enabled for INFO")
	}
}

func TestLoggerInitWithDefaultLevel(t *testing.T) {
	Initialize(false, "", CLI)

	if !loggersMap.getLogger(gaugeModuleID).IsEnabledFor(logging.INFO) {
		t.Error("Expected gaugeLog to be enabled for default log level")
	}
}

func TestLoggerInitWithDebugLevel(t *testing.T) {
	Initialize(false, "debug", CLI)

	if !loggersMap.getLogger(gaugeModuleID).IsEnabledFor(logging.DEBUG) {
		t.Error("Expected gaugeLog to be enabled for DEBUG")
	}
}

func TestLoggerInitWithWarningLevel(t *testing.T) {
	Initialize(false, "warning", CLI)

	if !loggersMap.getLogger(gaugeModuleID).IsEnabledFor(logging.WARNING) {
		t.Error("Expected gaugeLog to be enabled for WARNING")
	}
}

func TestLoggerInitWithErrorLevel(t *testing.T) {
	Initialize(false, "error", CLI)

	if !loggersMap.getLogger(gaugeModuleID).IsEnabledFor(logging.ERROR) {
		t.Error("Expected gaugeLog to be enabled for ERROR")
	}
}

func TestGetLogFileGivenRelativePathInGaugeProject(t *testing.T) {
	config.ProjectRoot, _ = filepath.Abs("_testdata")
	want := filepath.Join(config.ProjectRoot, logs, apiLogFileName)

	got := getLogFile(apiLogFileName)
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetLogFileWhenLogsDirNotSet(t *testing.T) {
	config.ProjectRoot, _ = filepath.Abs("_testdata")
	want := filepath.Join(config.ProjectRoot, logs, apiLogFileName)

	got := getLogFile(apiLogFileName)
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetLogFileInGaugeProjectWhenRelativeCustomLogsDirIsSet(t *testing.T) {
	myLogsDir := "my_logs"
	os.Setenv(logsDirectory, myLogsDir)
	defer os.Unsetenv(logsDirectory)

	config.ProjectRoot, _ = filepath.Abs("_testdata")
	want := filepath.Join(config.ProjectRoot, myLogsDir, apiLogFileName)

	got := getLogFile(apiLogFileName)

	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetLogFileInGaugeProjectWhenAbsoluteCustomLogsDirIsSet(t *testing.T) {
	myLogsDir, err := filepath.Abs("my_logs")
	if err != nil {
		t.Errorf("Unable to convert to absolute path, %s", err.Error())
	}

	os.Setenv(logsDirectory, myLogsDir)
	defer os.Unsetenv(logsDirectory)

	want := filepath.Join(myLogsDir, apiLogFileName)

	got := getLogFile(apiLogFileName)

	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetErrorText(t *testing.T) {
	tests := []struct {
		gaugeVersion *version.Version
		commitHash   string
		pluginInfos  []pluginInfo.PluginInfo
		expectedText string
	}{
		{
			gaugeVersion: &version.Version{Major: 0, Minor: 9, Patch: 8},
			commitHash:   "",
			pluginInfos:  []pluginInfo.PluginInfo{{Name: "java", Version: &version.Version{Major: 0, Minor: 6, Patch: 6}}},
			expectedText: fmt.Sprintf(`Error ----------------------------------

An Error has Occurred: some error

Get Support ----------------------------
	Docs:          https://docs.gauge.org
	Bugs:          https://github.com/getgauge/gauge/issues
	Chat:          https://github.com/getgauge/gauge/discussions

Your Environment Information -----------
	%s, 0.9.8
	java (0.6.6)`, runtime.GOOS),
		},
		{
			gaugeVersion: &version.Version{Major: 0, Minor: 9, Patch: 8},
			commitHash:   "",
			pluginInfos: []pluginInfo.PluginInfo{
				{Name: "java", Version: &version.Version{Major: 0, Minor: 6, Patch: 6}},
				{Name: "html-report", Version: &version.Version{Major: 0, Minor: 4, Patch: 0}},
			},
			expectedText: fmt.Sprintf(`Error ----------------------------------

An Error has Occurred: some error

Get Support ----------------------------
	Docs:          https://docs.gauge.org
	Bugs:          https://github.com/getgauge/gauge/issues
	Chat:          https://github.com/getgauge/gauge/discussions

Your Environment Information -----------
	%s, 0.9.8
	java (0.6.6), html-report (0.4.0)`, runtime.GOOS),
		},
		{
			gaugeVersion: &version.Version{Major: 0, Minor: 9, Patch: 8},
			commitHash:   "59effa",
			pluginInfos: []pluginInfo.PluginInfo{
				{Name: "java", Version: &version.Version{Major: 0, Minor: 6, Patch: 6}},
				{Name: "html-report", Version: &version.Version{Major: 0, Minor: 4, Patch: 0}},
			},
			expectedText: fmt.Sprintf(`Error ----------------------------------

An Error has Occurred: some error

Get Support ----------------------------
	Docs:          https://docs.gauge.org
	Bugs:          https://github.com/getgauge/gauge/issues
	Chat:          https://github.com/getgauge/gauge/discussions

Your Environment Information -----------
	%s, 0.9.8, 59effa
	java (0.6.6), html-report (0.4.0)`, runtime.GOOS),
		},
	}

	for _, test := range tests {
		version.CurrentGaugeVersion = test.gaugeVersion
		version.CommitHash = test.commitHash
		pluginInfo.GetAllInstalledPluginsWithVersion = func() ([]pluginInfo.PluginInfo, error) {
			return test.pluginInfos, nil
		}
		fatalErrors = append(fatalErrors, fmt.Sprintf("An Error has Occurred: %s", "some error"))
		got := getFatalErrorMsg()
		want := test.expectedText

		if got != want {
			t.Errorf("Got %s, want %s", got, want)
		}
		fatalErrors = []string{}
	}
}

func TestToJSONWithPlainText(t *testing.T) {
	outMessage := &OutMessage{MessageType: "out", Message: "plain text"}
	want := "{\"type\":\"out\",\"message\":\"plain text\"}"

	got, _ := outMessage.ToJSON()
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestToJSONWithInvalidJSONCharacters(t *testing.T) {
	outMessage := &OutMessage{MessageType: "out", Message: "\n, \t, and \\ needs to be escaped to create a valid JSON"}
	want := "{\"type\":\"out\",\"message\":\"\\n, \\t, and \\\\ needs to be escaped to create a valid JSON\"}"

	got, _ := outMessage.ToJSON()
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}
