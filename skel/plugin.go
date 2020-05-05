/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package skel

import (
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/plugin/install"
)

const (
	screenshot = "screenshot"
	html       = "html-report"
)

var requiredPlugins = []string{screenshot, html}

var SetupPlugins = func(silent bool) {
	installPlugins(getPluginsToInstall(), silent)
}

func getPluginsToInstall() (plugins []string) {
	for _, p := range requiredPlugins {
		if !plugin.IsPluginInstalled(p, "") {
			plugins = append(plugins, p)
		}
	}
	return
}

func installPlugins(plugins []string, silent bool) {
	if len(plugins) > 0 {
		logger.Infof(true, "Installing required plugins.")
	}
	for _, p := range plugins {
		installPlugin(p, silent)
	}
}

func installPlugin(name string, silent bool) {
	logger.Debugf(true, "Installing plugin '%s'", name)
	res := install.Plugin(name, "", silent)
	if res.Error != nil {
		logger.Debugf(true, res.Error.Error())
	} else if res.Version != "" {
		logger.Infof(true, "Successfully installed plugin '%s' version %s", name, res.Version)
	} else {
		logger.Infof(true, "Successfully installed plugin '%s'", name)
	}
}
