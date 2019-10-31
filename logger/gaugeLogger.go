// Copyright 2019 ThoughtWorks, Inc.

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
	write(stdout, getFatalErrorMsg())
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
