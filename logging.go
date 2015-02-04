package main

import (
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/op/go-logging"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
)

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
	gaugeFileLogger := createFileLogger(gaugeLogFile, 20)

	stdOutFormatter := logging.NewBackendFormatter(stdOutLogger, format)
	fileFormatter := logging.NewBackendFormatter(gaugeFileLogger, format)

	stdOutLoggerLeveled := logging.AddModuleLevel(stdOutFormatter)
	stdOutLoggerLeveled.SetLevel(loggingLevel(), "")

	fileLoggerLeveled := logging.AddModuleLevel(fileFormatter)
	fileLoggerLeveled.SetLevel(logging.DEBUG, "")

	logging.SetBackend(fileLoggerLeveled, stdOutLoggerLeveled)
}

func initApiLogger() {
	apiFileLogger := createFileLogger(apiLogFile, 10)
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
