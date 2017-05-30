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

package plugin

import (
	"sync"

	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
)

type Handler interface {
	NotifyPlugins(*gauge_messages.Message)
	GracefullyKillPlugins()
}

type GaugePlugins struct {
	pluginsMap map[string]*plugin
}

func (gp *GaugePlugins) addPlugin(pluginID string, pluginToAdd *plugin) {
	if gp.pluginsMap == nil {
		gp.pluginsMap = make(map[string]*plugin)
	}
	gp.pluginsMap[pluginID] = pluginToAdd
}

func (gp *GaugePlugins) removePlugin(pluginID string) {
	delete(gp.pluginsMap, pluginID)
}

func (gp *GaugePlugins) NotifyPlugins(message *gauge_messages.Message) {
	for id, plugin := range gp.pluginsMap {
		err := plugin.sendMessage(message)
		if err != nil {
			logger.Errorf("Unable to connect to plugin %s %s. %s\n", plugin.descriptor.Name, plugin.descriptor.Version, err.Error())
			gp.killPlugin(id)
		}
	}
}

func (gp *GaugePlugins) killPlugin(pluginID string) {
	plugin := gp.pluginsMap[pluginID]
	logger.Debug("Killing Plugin %s %s\n", plugin.descriptor.Name, plugin.descriptor.Version)
	err := plugin.pluginCmd.Process.Kill()
	if err != nil {
		logger.Errorf("Failed to kill plugin %s %s. %s\n", plugin.descriptor.Name, plugin.descriptor.Version, err.Error())
	}
	gp.removePlugin(pluginID)
}

func (gp *GaugePlugins) GracefullyKillPlugins() {
	var wg sync.WaitGroup
	for _, plugin := range gp.pluginsMap {
		wg.Add(1)
		go plugin.kill(&wg)
	}
	wg.Wait()
}
