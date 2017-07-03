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

package config

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestJSONFormatter(t *testing.T) {
	want := []string{
		"----------------------------------------------------------------------",
		"Key                           	Value                              ",
		"check_updates                 	true                               ",
		"gauge_repository_url          	https://downloads.getgauge.io/plugin",
		"gauge_telemetry_enabled       	false                              ",
		"gauge_telemetry_log_enabled   	false                              ",
		"gauge_templates_url           	https://downloads.getgauge.io/templates",
		"gauge_update_url              	https://downloads.getgauge.io/gauge",
		"plugin_connection_timeout     	10000                              ",
		"plugin_kill_timeout           	4000                               ",
		"runner_connection_timeout     	30000                              ",
		"runner_request_timeout        	30000                              ",
	}
	p := Properties()
	var properties []property
	for _, p := range p.p {
		properties = append(properties, *p)
	}

	f := &textFormatter{}
	text, err := f.format(properties)

	if err != nil {
		t.Errorf("Expected error == nil when using text formatter for properties, got %s", err.Error())
	}
	got := strings.Split(text, "\n")
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Properties text Format failed\nwant: `%s`\ngot: `%s`", want, got)
	}
}
