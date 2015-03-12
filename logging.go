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

package main

import (
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/op/go-logging"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
)

const LOGS_DIRECTORY = "logs_directory"

var log = logging.MustGetLogger("gauge")
var apiLog = logging.MustGetLogger("gauge-api")

var gaugeLogFile = filepath.Join("logs", "gauge.log")
var apiLogFile = filepath.Join("logs", "api.log")

var format = logging.MustStringFormatter(
	"%{time:15:04:05.000} [%{level:.4s}] %{message}",
)

func initLoggers() {
	initGaugeLogger()
	initApiLogger()
}

func initGaugeLogger() {
	stdOutLogger := logging.NewLogBackend(os.Stdout, "", 0)
	logsDir := os.Getenv(LOGS_DIRECTORY)
	var gaugeFileLogger logging.Backend
	if logsDir == "" {
		gaugeFileLogger = createFileLogger(gaugeLogFile, 20)
	} else {
		gaugeFileLogger = createFileLogger(logsDir+fmt.Sprintf("%c", filepath.Separator)+"gauge.log", 20)
	}
	stdOutFormatter := logging.NewBackendFormatter(stdOutLogger, format)
	fileFormatter := logging.NewBackendFormatter(gaugeFileLogger, format)

	stdOutLoggerLeveled := logging.AddModuleLevel(stdOutFormatter)
	stdOutLoggerLeveled.SetLevel(loggingLevel(), "")

	fileLoggerLeveled := logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(logging.DEBUG, "")

	logging.SetBackend(fileLoggerLeveled, stdOutLoggerLeveled)
}

func initApiLogger() {
	logsDir, err := filepath.Abs(os.Getenv(LOGS_DIRECTORY))
	var apiFileLogger logging.Backend
	if logsDir == "" || err != nil {
		apiFileLogger = createFileLogger(apiLogFile, 10)
	} else {
		apiFileLogger = createFileLogger(logsDir+fmt.Sprintf("%c", filepath.Separator)+"api.log", 10)
	}

	fileFormatter := logging.NewBackendFormatter(apiFileLogger, format)
	fileLoggerLeveled := logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(loggingLevel(), "")
	apiLog.SetBackend(fileLoggerLeveled)
}

func createFileLogger(name string, size int) logging.Backend {
	return logging.NewLogBackend(&lumberjack.Logger{
		Filename:   getLogFile(name),
		MaxSize:    size, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}, "", 0)
}

func getLogFile(fileName string) string {
	if config.ProjectRoot != "" {
		return filepath.Join(config.ProjectRoot, fileName)
	} else {
		gaugeHome, err := common.GetGaugeHomeDirectory()
		if err != nil {
			return fileName
		}
		return filepath.Join(gaugeHome, fileName)
	}
}

func loggingLevel() logging.Level {
	if *verbosity {
		return logging.DEBUG
	}
	if *logLevel != "" {
		switch strings.ToLower(*logLevel) {
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
