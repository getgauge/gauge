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
		logger.Infof(true, "", "Installing required plugins.")
	}
	for _, p := range plugins {
		installPlugin(p, silent)
	}
}

func installPlugin(name string, silent bool) {
	logger.Debugf(true, "", "Installing plugin '%s'", name)
	res := install.Plugin(name, "", silent)
	if res.Error != nil {
		logger.Debugf(true, "", res.Error.Error())
	} else if res.Version != "" {
		logger.Infof(true, "", "Successfully installed plugin '%s' version %s", name, res.Version)
	} else {
		logger.Infof(true, "", "Successfully installed plugin '%s'", name)
	}
}
