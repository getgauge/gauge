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
	"github.com/op/go-logging"
)

func TestLoggerInitWithInfoLevel(t *testing.T) {
	Initialize(false, "info", CLI)

	if !activeLogger.IsEnabledFor(logging.INFO) {
		t.Error("Expected gaugeLog to be enabled for INFO")
	}
}

func TestLoggerInitWithDefaultLevel(t *testing.T) {
	Initialize(false, "", CLI)

	if !activeLogger.IsEnabledFor(logging.INFO) {
		t.Error("Expected gaugeLog to be enabled for default log level")
	}
}

func TestLoggerInitWithDebugLevel(t *testing.T) {
	Initialize(false, "debug", CLI)

	if !activeLogger.IsEnabledFor(logging.DEBUG) {
		t.Error("Expected gaugeLog to be enabled for DEBUG")
	}
}

func TestLoggerInitWithWarningLevel(t *testing.T) {
	Initialize(false, "warning", CLI)

	if !activeLogger.IsEnabledFor(logging.WARNING) {
		t.Error("Expected gaugeLog to be enabled for WARNING")
	}
}

func TestLoggerInitWithErrorLevel(t *testing.T) {
	Initialize(false, "error", CLI)

	if !activeLogger.IsEnabledFor(logging.ERROR) {
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

func TestGetLogFileInGaugeProjectGivenAbsPath(t *testing.T) {
	config.ProjectRoot, _ = filepath.Abs("_testdata")
	want := filepath.Join(config.ProjectRoot, apiLogFileName)

	got := getLogFile(filepath.Join(config.ProjectRoot, apiLogFileName))
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetLogFileInGaugeProjectCustomPath(t *testing.T) {
	config.ProjectRoot, _ = filepath.Abs("_testdata")
	customLogsDir := filepath.Join(config.ProjectRoot, "myLogsDir")
	want := filepath.Join(customLogsDir, apiLogFileName)
	got := getLogFile(filepath.Join(customLogsDir, apiLogFileName))

	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestGetLogFileInGaugeProjectWhenCustomLogsDirIsSet(t *testing.T) {
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

func TestGetErrorText(t *testing.T) {
	tests := []struct {
		gaugeVersion *version.Version
		commitHash   string
		pluginInfos  []pluginInfo.PluginInfo
		expectedText string
	}{
		{
			gaugeVersion: &version.Version{0, 9, 8},
			commitHash:   "",
			pluginInfos:  []pluginInfo.PluginInfo{{Name: "java", Version: &version.Version{0, 6, 6}}},
			expectedText: fmt.Sprintf(`Error ----------------------------------

An Error has Occurred: some error

Get Support ----------------------------
	Docs:          https://docs.gauge.org
	Bugs:          https://github.com/getgauge/gauge/issues
	Chat:          https://gitter.im/getgauge/chat

Your Environment Information -----------
	%s, 0.9.8
	java (0.6.6)`, runtime.GOOS),
		},
		{
			gaugeVersion: &version.Version{0, 9, 8},
			commitHash:   "",
			pluginInfos: []pluginInfo.PluginInfo{
				{Name: "java", Version: &version.Version{0, 6, 6}},
				{Name: "html-report", Version: &version.Version{0, 4, 0}},
			},
			expectedText: fmt.Sprintf(`Error ----------------------------------

An Error has Occurred: some error

Get Support ----------------------------
	Docs:          https://docs.gauge.org
	Bugs:          https://github.com/getgauge/gauge/issues
	Chat:          https://gitter.im/getgauge/chat

Your Environment Information -----------
	%s, 0.9.8
	java (0.6.6), html-report (0.4.0)`, runtime.GOOS),
		},
		{
			gaugeVersion: &version.Version{0, 9, 8},
			commitHash:   "59effa",
			pluginInfos: []pluginInfo.PluginInfo{
				{Name: "java", Version: &version.Version{0, 6, 6}},
				{Name: "html-report", Version: &version.Version{0, 4, 0}},
			},
			expectedText: fmt.Sprintf(`Error ----------------------------------

An Error has Occurred: some error

Get Support ----------------------------
	Docs:          https://docs.gauge.org
	Bugs:          https://github.com/getgauge/gauge/issues
	Chat:          https://gitter.im/getgauge/chat

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
		got := getErrorText("An Error has Occurred: %s", "some error")
		want := test.expectedText

		if got != want {
			t.Errorf("Got %s, want %s", got, want)
		}
	}
}
