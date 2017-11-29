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
	"github.com/natefinch/lumberjack"
	"github.com/op/go-logging"
)

const (
	logsDirectory    = "logs_directory"
	logs             = "logs"
	GaugeLogFileName = "gauge.log"
	apiLogFileName   = "api.log"
	lspLogFileName   = "lsp.log"
)

var level logging.Level
var isWindows bool
var customLogger CustomLogger

type CustomLogger interface {
	Log(logLevel logging.Level, msg string)
}

func SetCustomLogger(l CustomLogger) {
	customLogger = l
}

// Infof logs INFO messages
func Infof(msg string, args ...interface{}) {
	GaugeLog.Infof(msg, args...)
	write(logging.INFO, msg, args...)
}

// Errorf logs ERROR messages
func Errorf(msg string, args ...interface{}) {
	GaugeLog.Errorf(msg, args...)
	write(logging.ERROR, msg, args...)
}

// Warningf logs WARNING messages
func Warningf(msg string, args ...interface{}) {
	GaugeLog.Warningf(msg, args...)
	write(logging.WARNING, msg, args...)
}

// Fatalf logs CRITICAL messages and exits
func Fatalf(msg string, args ...interface{}) {
	write(logging.CRITICAL, msg, args...)
	GaugeLog.Fatalf(msg, args...)
}

// Debugf logs DEBUG messages
func Debugf(msg string, args ...interface{}) {
	GaugeLog.Debugf(msg, args...)
	if level == logging.DEBUG {
		write(logging.DEBUG, msg, args...)
	}
}

func write(logLevel logging.Level, msg string, args ...interface{}) {
	if customLogger != nil {
		customLogger.Log(logLevel, fmt.Sprintf(msg, args...))
	} else {
		fmt.Println(fmt.Sprintf(msg, args...))
	}
}

// GaugeLog is for logging messages related to spec execution lifecycle
var GaugeLog = logging.MustGetLogger("gauge")

// APILog is for logging API related messages
var APILog = logging.MustGetLogger("gauge-api")

var LspLog = logging.MustGetLogger("gauge-lsp")

var fileLogFormat = logging.MustStringFormatter("%{time:15:04:05.000} %{message}")

// Initialize initializes the logger object
func Initialize(logLevel string) {
	level = loggingLevel(logLevel)
	initFileLogger(GaugeLogFileName, GaugeLog)
	initFileLogger(apiLogFileName, APILog)
	initFileLogger(lspLogFileName, LspLog)
	if runtime.GOOS == "windows" {
		isWindows = true
	}
}

func GetLogFilePath(logFileName string) string {
	customLogsDir := os.Getenv(logsDirectory)
	if customLogsDir == "" {
		return filepath.Join(logs, logFileName)
	} else {
		return filepath.Join(customLogsDir, logFileName)
	}
}

func initFileLogger(logFileName string, fileLogger *logging.Logger) {
	var backend logging.Backend
	backend = createFileLogger(GetLogFilePath(logFileName), 10)
	fileFormatter := logging.NewBackendFormatter(backend, fileLogFormat)
	fileLoggerLeveled := logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(logging.DEBUG, "")

	fileLogger.SetBackend(fileLoggerLeveled)
}

func createFileLogger(name string, size int) logging.Backend {
	name = getLogFile(name)
	return logging.NewLogBackend(&lumberjack.Logger{
		Filename:   name,
		MaxSize:    size, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}, "", 0)
}

func getLogFile(fileName string) string {
	if filepath.IsAbs(fileName) {
		return fileName
	}
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
		}
	}
	return logging.INFO
}

// HandleWarningMessages logs multiple messages in WARNING mode
func HandleWarningMessages(warnings []string) {
	for _, warning := range warnings {
		Warningf(warning)
	}
}
