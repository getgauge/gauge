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
	"net"

	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/golang/protobuf/proto"
)

type keepAliveHandler struct {
	ph Handler
}

func (h *keepAliveHandler) MessageBytesReceived(b []byte, c net.Conn) {
	m := &gauge_messages.Message{}
	err := proto.Unmarshal(b, m)
	if err != nil {
		logger.Errorf(true, "", "Failed to read proto message: %s\n", err.Error())
	} else {
		if m.GetMessageType() == gauge_messages.Message_KeepAlive {
			id := m.KeepAlive.PluginId
			logger.Debugf(true, "", "KeepAlive request received for pluginID: %s", id)
			h.ph.ExtendTimeout(id)
		}
	}
}
