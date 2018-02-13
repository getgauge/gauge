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
	"path/filepath"
	"testing"

	"os"

	"github.com/getgauge/gauge/config"
	"github.com/op/go-logging"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestLoggerInitWithInfoLevel(c *C) {
	Initialize("info")

	c.Assert(GaugeLog.IsEnabledFor(logging.INFO), Equals, true)
	c.Assert(APILog.IsEnabledFor(logging.INFO), Equals, true)
}

func (s *MySuite) TestLoggerInitWithDefaultLevel(c *C) {
	Initialize("")

	c.Assert(GaugeLog.IsEnabledFor(logging.INFO), Equals, true)
	c.Assert(APILog.IsEnabledFor(logging.INFO), Equals, true)
}

func (s *MySuite) TestLoggerInitWithDebugLevel(c *C) {
	Initialize("debug")

	c.Assert(GaugeLog.IsEnabledFor(logging.DEBUG), Equals, true)
	c.Assert(APILog.IsEnabledFor(logging.DEBUG), Equals, true)
}

func (s *MySuite) TestLoggerInitWithWarningLevel(c *C) {
	Initialize("warning")

	c.Assert(GaugeLog.IsEnabledFor(logging.WARNING), Equals, true)
	c.Assert(APILog.IsEnabledFor(logging.WARNING), Equals, true)
}

func (s *MySuite) TestLoggerInitWithErrorLevel(c *C) {
	Initialize("error")

	c.Assert(GaugeLog.IsEnabledFor(logging.ERROR), Equals, true)
	c.Assert(APILog.IsEnabledFor(logging.ERROR), Equals, true)
}

func (s *MySuite) TestGetLogFileGivenRelativePathInGaugeProject(c *C) {
	config.ProjectRoot, _ = filepath.Abs("_testdata")
	expected := filepath.Join(config.ProjectRoot, logs, apiLogFileName)

	c.Assert(GetLogFile(apiLogFileName), Equals, expected)
}

func (s *MySuite) TestGetLogFileInGaugeProjectGivenAbsPath(c *C) {
	config.ProjectRoot, _ = filepath.Abs("_testdata")
	expected := filepath.Join(config.ProjectRoot, apiLogFileName)

	c.Assert(GetLogFile(filepath.Join(config.ProjectRoot, apiLogFileName)), Equals, expected)
}

func (s *MySuite) TestGetLogFileInGaugeProjectCustomPath(c *C) {
	config.ProjectRoot, _ = filepath.Abs("_testdata")
	customLogsDir := filepath.Join(config.ProjectRoot, "myLogsDir")

	logFile := GetLogFile(filepath.Join(customLogsDir, apiLogFileName))

	c.Assert(logFile, Equals, filepath.Join(customLogsDir, apiLogFileName))
}

func (s *MySuite) TestGetLogFileInGaugeProjectWhenCustomLogsDirIsSet(c *C) {
	myLogsDir := "my_logs"
	os.Setenv(logsDirectory, myLogsDir)
	defer os.Unsetenv(logsDirectory)

	config.ProjectRoot, _ = filepath.Abs("_testdata")
	expected := filepath.Join(config.ProjectRoot, myLogsDir, apiLogFileName)

	logFile := GetLogFile(apiLogFileName)

	c.Assert(logFile, Equals, expected)
}
