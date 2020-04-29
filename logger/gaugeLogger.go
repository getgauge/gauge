// Copyright 2019 ThoughtWorks, Inc.

/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package logger

import (
	"fmt"
	"os"
)

// Info logs INFO messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Info(stdout bool, msg string) {
	logInfo(loggersMap.getLogger(gaugeModuleID), stdout, msg)
}

// Infof logs INFO messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Infof(stdout bool, msg string, args ...interface{}) {
	Info(stdout, fmt.Sprintf(msg, args...))
}

// Error logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Error(stdout bool, msg string) {
	logError(loggersMap.getLogger(gaugeModuleID), stdout, msg)
}

// Errorf logs ERROR messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Errorf(stdout bool, msg string, args ...interface{}) {
	Error(stdout, fmt.Sprintf(msg, args...))
}

// Warning logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Warning(stdout bool, msg string) {
	logWarning(loggersMap.getLogger(gaugeModuleID), stdout, msg)
}

// Warningf logs WARNING messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Warningf(stdout bool, msg string, args ...interface{}) {
	Warning(stdout, fmt.Sprintf(msg, args...))
}

// Fatal logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func Fatal(stdout bool, msg string) {
	logCritical(loggersMap.getLogger(gaugeModuleID), msg)
	addFatalError(gaugeModuleID, msg)
	write(stdout, getFatalErrorMsg(), os.Stdout)
	os.Exit(1)
}

// Fatalf logs CRITICAL messages and exits. stdout flag indicates if message is to be written to stdout in addition to log.
func Fatalf(stdout bool, msg string, args ...interface{}) {
	Fatal(stdout, fmt.Sprintf(msg, args...))
}

// Debug logs DEBUG messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Debug(stdout bool, msg string) {
	logDebug(loggersMap.getLogger(gaugeModuleID), stdout, msg)
}

// Debugf logs DEBUG messages. stdout flag indicates if message is to be written to stdout in addition to log.
func Debugf(stdout bool, msg string, args ...interface{}) {
	Debug(stdout, fmt.Sprintf(msg, args...))
}

// HandleWarningMessages logs multiple messages in WARNING mode
func HandleWarningMessages(stdout bool, warnings []string) {
	for _, warning := range warnings {
		Warning(stdout, warning)
	}
}
