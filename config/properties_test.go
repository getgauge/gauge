/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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

func (f *dummyFormatter) Format(p []Property) (string, error) {
	for _, prop := range p {
		f.p = append(f.p, prop.Value)
	}
	return "", nil
}

func TestPropertiesSetValue(t *testing.T) {
	p := defaults()
	want := "https://gauge.trepo.url"

	err := p.set(gaugeRepositoryURL, want)
	if err != nil {
		t.Error(err)
	}

	got, err := p.get(gaugeRepositoryURL)
	if err != nil {
		t.Errorf("Expected error == nil when setting property, got %s", err.Error())
	}
	if got != want {
		t.Errorf("Setting Property `%s` failed, want: `%s`, got `%s`", gaugeRepositoryURL, want, got)
	}
}

func TestPropertiesFormat(t *testing.T) {
	p := defaults()
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

	p, err := mergedProperties()
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

# Allow Gauge to download template from insecure URLs.
allow_insecure_download = false

# Allow Gauge and its plugin updates to be notified.
check_updates = true

# Url to get plugin versions
gauge_repository_url = https://downloads.gauge.org/plugin

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

	s, err := defaults().String()
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

	s, err := defaults().String()
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
	got, err := GaugeVersionInPropertiesFile(common.GaugePropertiesFile)
	if err != nil {
		t.Error(err)
	}
	if got.String() != want {
		t.Errorf("Expected Gauge Version in gauge.properties %s, got %s", want, got)
	}
	os.Setenv("GAUGE_HOME", oldEnv)
	os.Remove(propFile)
}
