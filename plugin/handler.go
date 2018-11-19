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
	var handle = func(id string, p *plugin, err error) {
		if err != nil {
			logger.Errorf(true, "Unable to connect to plugin %s %s. %s\n", p.descriptor.Name, p.descriptor.Version, err.Error())
			gp.killPlugin(id)
		}
	}

	for id, plugin := range gp.pluginsMap {
		if !plugin.descriptor.hasCapability(streamResultCapability) {
			handle(id, plugin, plugin.sendMessage(message))
			return
		}
		items := []*gauge_messages.ProtoItem{}
		for _, sr := range message.SuiteExecutionResult.GetSuiteResult().GetSpecResults() {
			for _, i := range sr.ProtoSpec.Items {
				i.FileName = sr.ProtoSpec.FileName
				items = append(items, i)
			}
			sr.ProtoSpec.ItemCount = int64(len(sr.ProtoSpec.Items))
			sr.ProtoSpec.Items = nil
		}
		if message.MessageType == gauge_messages.Message_SuiteExecutionResult {
			message.SuiteExecutionResult.SuiteResult.Chunked = true
			message.SuiteExecutionResult.SuiteResult.ChunkSize = int64(len(items))
			handle(id, plugin, plugin.sendMessage(message))
			for _, i := range items {
				m := &gauge_messages.Message{MessageType: gauge_messages.Message_SuiteExecutionResultItem, SuiteExecutionResultItem: &gauge_messages.SuiteExecutionResultItem{ResultItem: i}}
				handle(id, plugin, plugin.sendMessage(m))
			}
		} else {
			handle(id, plugin, plugin.sendMessage(message))
		}
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

func (gp *GaugePlugins) GracefullyKillPlugins() {
	var wg sync.WaitGroup
	for _, plugin := range gp.pluginsMap {
		wg.Add(1)
		go plugin.kill(&wg)
	}
	wg.Wait()
}
