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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/getgauge/common"
)

type dummyFormatter struct {
	p []string
}

func (f *dummyFormatter) Format(p []property) (string, error) {
	for _, prop := range p {
		f.p = append(f.p, prop.Value)
	}
	return "", nil
}

func TestPropertiesSetValue(t *testing.T) {
	p := Properties()
	want := "https://gauge.templates.url"

	p.set(gaugeTemplatesURL, want)

	got, err := p.get(gaugeTemplatesURL)
	if err != nil {
		t.Errorf("Expected error == nil when setting property, got %s", err.Error())
	}
	if got.Value != want {
		t.Errorf("Setting Property `%s` failed, want: `%s`, got `%s`", gaugeTemplatesURL, want, got.Value)
	}
}

func TestPropertiesFormat(t *testing.T) {
	p := Properties()
	var want []string
	for _, p := range p.p {
		want = append(want, (*p).Value)
	}
	sort.Strings(want)
	f := &dummyFormatter{}

	p.Format(f)

	sort.Strings(f.p)
	if !reflect.DeepEqual(f.p, want) {
		t.Errorf("Properties Format failed, want: `%s`, got `%s`", want, f.p)
	}
}

func TestMergedProperties(t *testing.T) {
	want := "false"
	idFile := filepath.Join("_testData", "gauge.properties")
	ioutil.WriteFile(idFile, []byte("check_updates=false"), common.NewFilePermissions)
	s, err := filepath.Abs("_testData")
	if err != nil {
		t.Error(err)
	}
	os.Setenv("GAUGE_ROOT", s)

	p := MergedProperties()
	got, err := p.get(checkUpdates)

	if err != nil {
		t.Errorf("Expected error == nil when getting property after merge, got %s", err.Error())
	}
	if got.Value != want {
		t.Errorf("Properties Merge failed, want: %s == `%s`, got `%s`", checkUpdates, want, got.Value)
	}
	os.Setenv("GAUGE_ROOT", "")
	err = os.Remove(idFile)
	if err != nil {
		t.Error(err)
	}
}

func TestPropertiesString(t *testing.T) {
	propertiesContent := `# This file contains Gauge specific internal configurations. Do not delete

# Url to get templates list
gauge_templates_url = https://downloads.getgauge.io/templates

# Timeout in milliseconds for making a connection to plugins.
plugin_connection_timeout = 10000

# Timeout in milliseconds for a plugin to stop after a kill message has been sent.
plugin_kill_timeout = 4000

# Timeout in milliseconds for requests from the language runner.
runner_request_timeout = 30000

# Log request sent to Gauge analytics engine
gauge_analytics_log_enabled = false

# Url to get plugin versions
gauge_repository_url = https://downloads.getgauge.io/plugin

# Url for latest gauge version
gauge_update_url = https://downloads.getgauge.io/gauge

# Timeout in milliseconds for making a connection to the language runner.
runner_connection_timeout = 30000

# Allow Gauge and its plugin updates to be notified.
check_updates = true

# Allow Gauge to collect anonymous usage statistics
gauge_analytics_enabled = false
`
	want := strings.Split(propertiesContent, "\n")
	sort.Strings(want)

	got := strings.Split(Properties().String(), "\n")

	sort.Strings(got)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Properties String failed\ngot: `%s`\nwant:`%s`", got, want)
	}
}
