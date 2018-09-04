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

	"github.com/getgauge/gauge/plugin/pluginInfo"
	"github.com/getgauge/gauge/version"

	"runtime"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/natefinch/lumberjack"
	"github.com/op/go-logging"
)

// Channel specifies the logging channel. Can be one of CLI, API or LSP
type channel int

const (
	logsDirectory    = "logs_directory"
	logs             = "logs"
	gaugeLogFileName = "gauge.log"
	apiLogFileName   = "api.log"
	LspLogFileName   = "lsp.log"
	// CLI indicates gauge is used as a CLI.
	CLI channel = iota
	// API indicates gauge is in daemon mode. Used in IDEs.
	API
	// LSP indicates that gauge is acting as an LSP server.
	LSP
)

var level logging.Level
var activeLogger *logging.Logger
var fileLogFormat = logging.MustStringFormatter("%{time:15:04:05.000} %{message}")
var isLSP bool
var initialized bool
var ActiveLogFile string
var machineReadable bool

// Info logs INFO messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Info(stdout bool, msg string) {
	Infof(stdout, msg)
}

// Infof logs INFO messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Infof(stdout bool, msg string, args ...interface{}) {
	if !initialized {
		return
	}
	write(stdout, msg, args...)
	activeLogger.Infof(msg, args...)
}

// Error logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Error(stdout bool, msg string) {
	Errorf(stdout, msg)
}

// Errorf logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Errorf(stdout bool, msg string, args ...interface{}) {
	if !initialized {
		fmt.Fprintf(os.Stderr, msg, args)
		return
	}
	write(stdout, msg, args...)
	activeLogger.Errorf(msg, args...)
}

// Warning logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Warning(stdout bool, msg string) {
	Warningf(stdout, msg)
}

// Warningf logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Warningf(stdout bool, msg string, args ...interface{}) {
	if !initialized {
		return
	}
	write(stdout, msg, args...)
	activeLogger.Warningf(msg, args...)
}

// Fatal logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func Fatal(stdout bool, msg string) {
	Fatalf(stdout, msg)
}

// Fatalf logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func Fatalf(stdout bool, msg string, args ...interface{}) {
	message := getErrorText(msg, args...)
	if !initialized {
		fmt.Fprintf(os.Stderr, msg, args)
		return
	}
	write(stdout, message)
	activeLogger.Fatalf(msg, args...)
}

// Debug logs DEBUG messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Debug(stdout bool, msg string) {
	Debugf(stdout, msg)
}

// Debugf logs DEBUG messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Debugf(stdout bool, msg string, args ...interface{}) {
	if !initialized {
		return
	}
	activeLogger.Debugf(msg, args...)
	if level == logging.DEBUG {
		write(stdout, msg, args...)
	}
}

func getErrorText(msg string, args ...interface{}) string {
	env := []string{runtime.GOOS, version.FullVersion()}
	if version.GetCommitHash() != "" {
		env = append(env, version.GetCommitHash())
	}
	envText := strings.Join(env, ", ")
	return fmt.Sprintf(`Error ----------------------------------

%s

Get Support ----------------------------
	Docs:          https://docs.gauge.org
	Bugs:          https://github.com/getgauge/gauge/issues
	Chat:          https://gitter.im/getgauge/chat

Your Environment Information -----------
	%s
	%s`, fmt.Sprintf(msg, args...),
		envText,
		getPluginVersions())
}

func getPluginVersions() string {
	pis, err := pluginInfo.GetAllInstalledPluginsWithVersion()
	if err != nil {
		return fmt.Sprintf("Could not retrieve plugin information.")
	}
	pluginVersions := make([]string, 0, 0)
	for _, pi := range pis {
		pluginVersions = append(pluginVersions, fmt.Sprintf(`%s (%s)`, pi.Name, pi.Version))
	}
	return strings.Join(pluginVersions, ", ")
}

func write(stdout bool, msg string, args ...interface{}) {
	if !isLSP && stdout {
		if machineReadable {
			strs := strings.Split(fmt.Sprintf(msg, args...), "\n")
			for _, m := range strs {
				fmt.Printf("{\"type\": \"out\", \"message\": \"%s\"}\n", strings.Trim(m, "\n "))
			}
		} else {
			fmt.Println(fmt.Sprintf(msg, args...))
		}
	}
}

// Initialize initializes the logger object
func Initialize(isMachineReadable bool, logLevel string, c channel) {
	initialized = true
	level = loggingLevel(logLevel)
	activeLogger = logger(c)
	machineReadable = isMachineReadable
}

func logger(c channel) *logging.Logger {
	var l *logging.Logger
	switch c {
	case LSP:
		l = logging.MustGetLogger("gauge-lsp")
		initFileLogger(LspLogFileName, l)
		isLSP = true
		break
	case API:
		l = logging.MustGetLogger("gauge-api")
		initFileLogger(apiLogFileName, l)
		break
	default:
		l = logging.MustGetLogger("gauge")
		initFileLogger(gaugeLogFileName, l)
	}
	return l
}

func initFileLogger(logFileName string, fileLogger *logging.Logger) {
	var backend logging.Backend
	ActiveLogFile = getLogFile(logFileName)
	backend = createFileLogger(ActiveLogFile, 10)
	fileFormatter := logging.NewBackendFormatter(backend, fileLogFormat)
	fileLoggerLeveled := logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(logging.DEBUG, "")

	fileLogger.SetBackend(fileLoggerLeveled)
}

func createFileLogger(name string, size int) logging.Backend {
	return logging.NewLogBackend(&lumberjack.Logger{
		Filename:   name,
		MaxSize:    size, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}, "", 0)
}

func addLogsDirPath(logFileName string) string {
	customLogsDir := os.Getenv(logsDirectory)
	if customLogsDir == "" {
		return filepath.Join(logs, logFileName)
	}
	return filepath.Join(customLogsDir, logFileName)
}

func getLogFile(fileName string) string {
	if filepath.IsAbs(fileName) {
		return fileName
	}
	fileName = addLogsDirPath(fileName)
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
func HandleWarningMessages(stdout bool, warnings []string) {
	for _, warning := range warnings {
		Warning(stdout, warning)
	}
}
