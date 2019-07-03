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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/plugin/pluginInfo"
	"github.com/getgauge/gauge/version"
	"github.com/natefinch/lumberjack"
	"github.com/op/go-logging"
)

const (
	gauge            = "Gauge"
	logsDirectory    = "logs_directory"
	logs             = "logs"
	gaugeLogFileName = "gauge.log"
	apiLogFileName   = "api.log"
	lspLogFileName   = "lsp.log"
	// CLI indicates gauge is used as a CLI.
	CLI = iota
	// API indicates gauge is in daemon mode. Used in IDEs.
	API
	// LSP indicates that gauge is acting as an LSP server.
	LSP
)

var level logging.Level
var initialized bool
var loggersMap map[string]*logging.Logger
var fatalErrors []string
var fileLogFormat = logging.MustStringFormatter("%{time:02-01-2006 15:04:05.000} [%{module}] [%{level}] %{message}")

// ActiveLogFile log file represents the file which will be used for the backend logging
var ActiveLogFile string
var machineReadable bool
var isLSP bool

// Initialize logger with given level
func Initialize(mr bool, logLevel string, c int) {
	loggersMap = make(map[string]*logging.Logger)
	machineReadable = mr
	level = loggingLevel(logLevel)
	switch c {
	case CLI:
		ActiveLogFile = getLogFile(gaugeLogFileName)
	case API:
		ActiveLogFile = getLogFile(apiLogFileName)
	case LSP:
		isLSP = true
		ActiveLogFile = getLogFile(lspLogFileName)
	}
	addLogger(gauge)
	initialized = true
}

// GetLogger gets logger for given modules. It creates a new logger for the module if not exists
func GetLogger(module string) *logging.Logger {
	if module == "" {
		return loggersMap[gauge]
	}
	if _, ok := loggersMap[module]; !ok {
		addLogger(module)
	}
	return loggersMap[module]

}

// OutMessage contains information for output log
type OutMessage struct {
	MessageType string `json:"type"`
	Message     string `json:"message"`
}

// ToJSON converts OutMessage into JSON
func (out *OutMessage) ToJSON() (string, error) {
	json, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

// Info logs INFO messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Info(stdout bool, msg string) {
	Infof(stdout, msg)
}

// Infof logs INFO messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Infof(stdout bool, msg string, args ...interface{}) {
	write(stdout, msg, args...)
	if !initialized {
		return
	}
	GetLogger(gauge).Infof(msg, args...)
}

// Error logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Error(stdout bool, msg string) {
	Errorf(stdout, msg)
}

// Errorf logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Errorf(stdout bool, msg string, args ...interface{}) {
	write(stdout, msg, args...)
	if !initialized {
		fmt.Fprintf(os.Stderr, msg, args...)
		return
	}
	GetLogger(gauge).Errorf(msg, args...)
}

// Warning logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Warning(stdout bool, msg string) {
	Warningf(stdout, msg)
}

// Warningf logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Warningf(stdout bool, msg string, args ...interface{}) {
	write(stdout, msg, args...)
	if !initialized {
		return
	}
	GetLogger(gauge).Warningf(msg, args...)
}

// Fatal logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func Fatal(stdout bool, msg string) {
	Fatalf(stdout, msg)
}

// Fatalf logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func Fatalf(stdout bool, msg string, args ...interface{}) {
	var msgs []string
	var builder strings.Builder
	msgs = append(msgs, getPluginErrorText(), fmt.Sprintf("[gauge]\n\t" + msg, args...))
	builder.WriteString(getErrorText(msgs))
	message := builder.String()

	if !initialized {
		fmt.Fprintf(os.Stderr, msg, args...)
		return
	}
	write(stdout, message)
	GetLogger(gauge).Fatalf(msg, args...)
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
	GetLogger(gauge).Debugf(msg, args...)
	if level == logging.DEBUG {
		write(stdout, msg, args...)
	}
}

func write(stdout bool, msg string, args ...interface{}) {
	if !isLSP && stdout {
		if machineReadable {
			strs := strings.Split(fmt.Sprintf(msg, args...), "\n")
			for _, m := range strs {
				outMessage := &OutMessage{MessageType: "out", Message: m}
				m, _ = outMessage.ToJSON()
				fmt.Println(m)
			}
		} else {
			fmt.Println(fmt.Sprintf(msg, args...))
		}
	}
}

func addLogger(module string) {
	l := logging.MustGetLogger(module)
	loggersMap[module] = l
	initFileLogger(ActiveLogFile, module, l)
}

func initFileLogger(logFileName string, module string, fileLogger *logging.Logger) {
	var backend logging.Backend
	backend = createFileLogger(logFileName, 10)
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

func getLogFile(logFileName string) string {
	logDirPath := addLogsDirPath(logFileName)
	if filepath.IsAbs(logDirPath) {
		return logDirPath
	}
	if config.ProjectRoot != "" {
		return filepath.Join(config.ProjectRoot, logDirPath)
	}
	gaugeHome, err := common.GetGaugeHomeDirectory()
	if err != nil {
		return logDirPath
	}
	return filepath.Join(gaugeHome, logDirPath)
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

func getPluginErrorText() string {
	var fatalTextBuilder strings.Builder
	for _, errorText := range fatalErrors {
		fatalTextBuilder.WriteString(errorText)
	}
	return fatalTextBuilder.String()
}

func getErrorText(msg []string) string {
	env := []string{runtime.GOOS, version.FullVersion()}
	if version.GetCommitHash() != "" {
		env = append(env, version.GetCommitHash())
	}
	envText := strings.Join(env, ", ")
	
	var messageBuilder strings.Builder
	for _, errorText := range msg{
		errMsg := fmt.Sprintf(errorText)
		messageBuilder.WriteString(fmt.Sprintf(`Error ----------------------------------

%s
		
`,errMsg))
	}
	return fmt.Sprintf(`%s
Get Support ----------------------------
	Docs:          https://docs.gauge.org
	Bugs:          https://github.com/getgauge/gauge/issues
	Chat:          https://gitter.im/getgauge/chat

Your Environment Information -----------
	%s
	%s`, messageBuilder.String(),
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

// HandleWarningMessages logs multiple messages in WARNING mode
func HandleWarningMessages(stdout bool, warnings []string) {
	for _, warning := range warnings {
		Warning(stdout, warning)
	}
}
