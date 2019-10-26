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
	"sync"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/version"
)

type dummyFormatter struct {
	p []string
}

func (f *dummyFormatter) format(p []property) (string, error) {
	for _, prop := range p {
		f.p = append(f.p, prop.Value)
	}
	return "", nil
}

func TestPropertiesSetValue(t *testing.T) {
	p := Properties()
	want := "https://gauge.templates.url"

	err := p.set(gaugeTemplatesURL, want)
	if err != nil {
		t.Error(err)
	}

	got, err := p.get(gaugeTemplatesURL)
	if err != nil {
		t.Errorf("Expected error == nil when setting property, got %s", err.Error())
	}
	if got != want {
		t.Errorf("Setting Property `%s` failed, want: `%s`, got `%s`", gaugeTemplatesURL, want, got)
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

	_, err := p.Format(f)
	if err != nil {
		t.Error(err)
	}
	sort.Strings(f.p)
	if !reflect.DeepEqual(f.p, want) {
		t.Errorf("Properties Format failed, want: `%s`, got `%s`", want, f.p)
	}
}

func TestMergedProperties(t *testing.T) {
	want := "false"
	idFile := filepath.Join("_testData", "config", "gauge.properties")
	err := ioutil.WriteFile(idFile, []byte("check_updates=false"), common.NewFilePermissions)
	if err != nil {
		t.Error(err)
	}

	s, err := filepath.Abs("_testData")
	if err != nil {
		t.Error(err)
	}
	os.Setenv("GAUGE_HOME", s)

	p, err := MergedProperties()
	if err != nil {
		t.Errorf("Unable to read MergedProperties: %s", err.Error())
	}

	got, err := p.get(checkUpdates)

	if err != nil {
		t.Errorf("Expected error == nil when getting property after merge, got %s", err.Error())
	}
	if got != want {
		t.Errorf("Properties Merge failed, want: %s == `%s`, got `%s`", checkUpdates, want, got)
	}
	os.Setenv("GAUGE_HOME", "")
	err = os.Remove(idFile)
	if err != nil {
		t.Error(err)
	}
}

var propertiesContent = "# Version " + version.CurrentGaugeVersion.String() + `
# This file contains Gauge specific internal configurations. Do not delete

# Allow Gauge and its plugin updates to be notified.
check_updates = true

# Url to get plugin versions
gauge_repository_url = https://downloads.gauge.org/plugin

# Record user opt in/out for telemetry
gauge_telemetry_action_recorded = false

# Allow Gauge to collect anonymous usage statistics
gauge_telemetry_enabled = true

# Log request sent to Gauge telemetry engine
gauge_telemetry_log_enabled = false

# Url to get templates list
gauge_templates_url = https://templates.gauge.org

# Timeout in milliseconds for requests from runner when invoked for ide.
ide_request_timeout = 30000

# Timeout in milliseconds for making a connection to plugins.
plugin_connection_timeout = 10000

# Timeout in milliseconds for a plugin to stop after a kill message has been sent.
plugin_kill_timeout = 4000

# Timeout in milliseconds for making a connection to the language runner.
runner_connection_timeout = 30000

# Timeout in milliseconds for requests from the language runner.
runner_request_timeout = 30000
`

func TestPropertiesString(t *testing.T) {
	want := strings.Split(propertiesContent, "\n\n")

	s, err := Properties().String()
	if err != nil {
		t.Error(err)
	}
	got := strings.Split(s, "\n\n")

	if len(got) != len(want) {
		t.Errorf("Expected %d properties, got %d", len(want), len(got))
	}

	for i, x := range want {
		if got[i] != x {
			t.Errorf("Expected property no %d = %s, got %s", i, x, got[i])
		}
	}
}

func TestPropertiesStringConcurrent(t *testing.T) {
	want := strings.Split(propertiesContent, "\n\n")

	writeFunc := func(wg *sync.WaitGroup) {
		err := Merge()
		if err != nil {
			t.Log(err)
		}
		wg.Done()
	}

	wg := &sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go writeFunc(wg)
	}

	wg.Wait()

	s, err := Properties().String()
	if err != nil {
		t.Error(err)
	}
	got := strings.Split(s, "\n\n")

	if len(got) != len(want) {
		t.Errorf("Expected %d properties, got %d", len(want), len(got))
	}

	for i, x := range want {
		if got[i] != x {
			t.Errorf("Expected property no %d = %s, got %s", i, x, got[i])
		}
	}
}

func TestWriteGaugePropertiesOnlyForNewVersion(t *testing.T) {
	oldEnv := os.Getenv("GAUGE_HOME")
	os.Setenv("GAUGE_HOME", filepath.Join(".", "_testData"))
	propFile := filepath.Join("_testData", "config", "gauge.properties")
	err := ioutil.WriteFile(propFile, []byte("# Version 0.8.0"), common.NewFilePermissions)
	if err != nil {
		t.Error(err)
	}

	err = Merge()
	if err != nil {
		t.Error(err)
	}

	want := version.FullVersion()
	got, err := gaugeVersionInProperties()
	if err != nil {
		t.Error(err)
	}
	if got.String() != want {
		t.Errorf("Expected Gauge Version in gauge.properties %s, got %s", want, got)
	}
	os.Setenv("GAUGE_HOME", oldEnv)
	os.Remove(propFile)
}
