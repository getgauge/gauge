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
	"strings"

	"github.com/getgauge/gauge/version"
)

type pluginCapability string

const (
	gRPCSupportCapability pluginCapability = "grpc_support"
)

type pluginDescriptor struct {
	ID          string
	Version     string
	Name        string
	Description string
	Command     struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	Scope               []string
	GaugeVersionSupport version.VersionSupport
	pluginPath          string
	Capabilities        []string
}

func (pd *pluginDescriptor) hasScope(scope pluginScope) bool {
	for _, s := range pd.Scope {
		if strings.ToLower(s) == string(scope) {
			return true
		}
	}
	return false
}

func (pd *pluginDescriptor) hasAnyScope() bool {
	return len(pd.Scope) > 0
}

func (pd *pluginDescriptor) hasCapability(cap pluginCapability) bool {
	for _, c := range pd.Capabilities {
		if strings.ToLower(c) == string(cap) {
			return true
		}
	}
	return false
}
