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
	"os"
	"path/filepath"
	"strings"

	"runtime"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/op/go-logging"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logsDirectory    = "logs_directory"
	logs             = "logs"
	gaugeLogFileName = "gauge.log"
	apiLogFileName   = "api.log"
)

var level logging.Level
var isWindows bool

// Info logs INFO messages
func Info(msg string, args ...interface{}) {
	GaugeLog.Info(msg, args...)
	fmt.Println(fmt.Sprintf(msg, args...))
}

// Error logs ERROR messages
func Error(msg string, args ...interface{}) {
	GaugeLog.Error(msg, args...)
	fmt.Println(fmt.Sprintf(msg, args...))
}

// Warning logs WARNING messages
func Warning(msg string, args ...interface{}) {
	GaugeLog.Warning(msg, args...)
	fmt.Println(fmt.Sprintf(msg, args...))
}

// Fatal logs CRITICAL messages and exits
func Fatal(msg string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, args...))
	GaugeLog.Fatalf(msg, args...)
}

// Debug logs DEBUG messages
func Debug(msg string, args ...interface{}) {
	GaugeLog.Debug(msg, args...)
	if level == logging.DEBUG {
		fmt.Println(fmt.Sprintf(msg, args...))
	}
}

// GaugeLog is for logging messages related to spec execution lifecycle
var GaugeLog = logging.MustGetLogger("gauge")

// APILog is for logging API related messages
var APILog = logging.MustGetLogger("gauge-api")

var fileLogFormat = logging.MustStringFormatter("%{time:15:04:05.000} %{message}")
var gaugeLogFile = filepath.Join(logs, gaugeLogFileName)
var apiLogFile = filepath.Join(logs, apiLogFileName)

// Initialize initializes the logger object
func Initialize(logLevel string) {
	level = loggingLevel(logLevel)
	initGaugeFileLogger()
	initAPIFileLogger()
	if runtime.GOOS == "windows" {
		isWindows = true
	}
}

func initGaugeFileLogger() {
	logsDir, err := filepath.Abs(os.Getenv(logsDirectory))
	var gaugeFileLogger logging.Backend
	if logsDir == "" || err != nil {
		gaugeFileLogger = createFileLogger(gaugeLogFile, 20)
	} else {
		gaugeFileLogger = createFileLogger(filepath.Join(logsDir, gaugeLogFileName), 20)
	}
	fileFormatter := logging.NewBackendFormatter(gaugeFileLogger, fileLogFormat)
	fileLoggerLeveled := logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(logging.DEBUG, "")

	GaugeLog.SetBackend(fileLoggerLeveled)
}

func initAPIFileLogger() {
	logsDir, err := filepath.Abs(os.Getenv(logsDirectory))
	var apiFileLogger logging.Backend
	if logsDir == "" || err != nil {
		apiFileLogger = createFileLogger(apiLogFile, 10)
	} else {
		apiFileLogger = createFileLogger(filepath.Join(logsDir, apiLogFileName), 10)
	}
	fileFormatter := logging.NewBackendFormatter(apiFileLogger, fileLogFormat)
	fileLoggerLeveled := logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(logging.DEBUG, "")

	APILog.SetBackend(fileLoggerLeveled)
}

func createFileLogger(name string, size int) logging.Backend {
	if !filepath.IsAbs(name) {
		name = getLogFile(name)
	}
	return logging.NewLogBackend(&lumberjack.Logger{
		Filename:   name,
		MaxSize:    size, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}, "", 0)
}

func getLogFile(fileName string) string {
	if config.ProjectRoot != "" {
		return filepath.Join(config.ProjectRoot, fileName)
	}
	gaugeHome, err := common.GetGaugeHomeDirectory()
	if err != nil {
		return fileName
	}
	return filepath.Join(gaugeHome, fileName)
}

func loggingLevel(logLevel string) logging.Level {
	if logLevel != "" {
		switch strings.ToLower(logLevel) {
		case "debug":
			return logging.DEBUG
		case "info":
			return logging.INFO
		case "warning":
			return logging.WARNING
		case "error":
			return logging.ERROR
		case "critical":
			return logging.CRITICAL
		case "notice":
			return logging.NOTICE
		}
	}
	return logging.INFO
}

// HandleWarningMessages logs multiple messages in WARNING mode
func HandleWarningMessages(warnings []string) {
	for _, warning := range warnings {
		GaugeLog.Warning(warning)
	}
}
