/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package plugin

import (
	"strings"

	"github.com/getgauge/gauge/version"
)

type pluginCapability string

const (
	gRPCSupportCapability pluginCapability = "grpc_support"
)

type PluginDescriptor struct {
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

func (pd *PluginDescriptor) hasScope(scope pluginScope) bool {
	for _, s := range pd.Scope {
		if strings.ToLower(s) == string(scope) {
			return true
		}
	}
	return false
}

func (pd *PluginDescriptor) hasAnyScope() bool {
	return len(pd.Scope) > 0
}

func (pd *PluginDescriptor) hasCapability(capability pluginCapability) bool {
	for _, c := range pd.Capabilities {
		if strings.ToLower(c) == string(capability) {
			return true
		}
	}
	return false
}
