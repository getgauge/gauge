/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package skel

import (
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/plugin/install"
	"os"
	"strconv"
	"strings"
)

const (
	screenshot = "screenshot"
	html       = "html-report"
)

var SetupPlugins = func(silent bool) {
	installPlugins(getPluginsToInstall(), silent)
}

func getPluginsToInstall() (plugins []string) {
	requiredPlugins := []string{html}
	if screenshotEnabled, err := strconv.ParseBool(strings.TrimSpace(os.Getenv(env.ScreenshotOnFailure))); err == nil && screenshotEnabled {
		requiredPlugins = append(requiredPlugins, screenshot)
	}
	for _, p := range requiredPlugins {
		if !plugin.IsPluginInstalled(p, "") {
			plugins = append(plugins, p)
		}
	}
	return
}

func installPlugins(plugins []string, silent bool) {
	if len(plugins) > 0 {
		logger.Info(true, "Installing required plugins.")
	}
	for _, p := range plugins {
		installPlugin(p, silent)
	}
}

func installPlugin(name string, silent bool) {
	logger.Debugf(true, "Installing plugin '%s'", name)
	res := install.Plugin(name, "", silent)
	if res.Error != nil {
		logger.Debug(true, res.Error.Error())
	} else if res.Version != "" {
		logger.Infof(true, "Successfully installed plugin '%s' version %s", name, res.Version)
	} else {
		logger.Infof(true, "Successfully installed plugin '%s'", name)
	}
}
