/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package main

import (
	"os"
	"runtime/debug"

	"github.com/getgauge/gauge/cmd"
	"github.com/getgauge/gauge/logger"
)

func main() {
	defer recoverPanic()
	if err := cmd.Parse(); err != nil {
		logger.Info(true, err.Error())
		os.Exit(1)
	}
}

func recoverPanic() {
	if r := recover(); r != nil {
		logger.Fatalf(true, "Error: %v\n%s", r, string(debug.Stack()))
	}
}
