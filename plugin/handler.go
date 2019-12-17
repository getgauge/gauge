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

// Handler manages plugins listed in project manifest.
type Handler interface {
	NotifyPlugins(*gauge_messages.Message)
	GracefullyKillPlugins()
}

// GaugePlugins holds a reference to all plugins launched. The plugins are listed in project manifest
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

// NotifyPlugins passes a message to all plugins listed in the manifest
func (gp *GaugePlugins) NotifyPlugins(message *gauge_messages.Message) {
	var handle = func(id string, p *plugin, err error) {
		if err != nil {
			logger.Errorf(true, "Unable to connect to plugin %s %s. %s\n", p.descriptor.Name, p.descriptor.Version, err.Error())
			gp.killPlugin(id)
		}
	}

	for id, plugin := range gp.pluginsMap {
		handle(id, plugin, plugin.sendMessage(message))
	}
}

func (gp *GaugePlugins) killPlugin(pluginID string) {
	plugin := gp.pluginsMap[pluginID]
	logger.Debugf(true, "Killing Plugin %s %s\n", plugin.descriptor.Name, plugin.descriptor.Version)
	err := plugin.pluginCmd.Process.Kill()
	if err != nil {
		logger.Errorf(true, "Failed to kill plugin %s %s. %s\n", plugin.descriptor.Name, plugin.descriptor.Version, err.Error())
	}
	gp.removePlugin(pluginID)
}

// GracefullyKillPlugins tells the plugins to stop, letting them cleanup whatever they need to
func (gp *GaugePlugins) GracefullyKillPlugins() {
	var wg sync.WaitGroup
	for _, pl := range gp.pluginsMap {
		wg.Add(1)
		logger.Debugf(true, "Sending kill message to %s plugin.", pl.descriptor.Name)
		go func(p *plugin) {
			err := p.kill(&wg)
			if err != nil {
				logger.Errorf(false, "Unable to kill plugin %s : %s", p.descriptor.Name, err.Error())
			}
		}(pl)
	}
	wg.Wait()
}
