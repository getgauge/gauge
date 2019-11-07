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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/version"
)

const comment = `This file contains Gauge specific internal configurations. Do not delete`

type property struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	description  string
	defaultValue string
}

type properties struct {
	p map[string]*property
}

func (p *properties) set(k, v string) error {
	if _, ok := p.p[k]; ok {
		p.p[k].Value = v
		return nil
	}
	return fmt.Errorf("config '%s' doesn't exist", k)
}

func (p *properties) get(k string) (string, error) {
	if _, ok := p.p[k]; ok {
		return p.p[k].Value, nil
	}
	return "", fmt.Errorf("config '%s' doesn't exist", k)
}

func (p *properties) Format(f formatter) (string, error) {
	var all []property
	for _, v := range p.p {
		all = append(all, *v)
	}
	return f.format(all)
}

func (p *properties) String() (string, error) {
	var buffer strings.Builder
	_, err := buffer.WriteString(fmt.Sprintf("# Version %s\n# %s\n", version.FullVersion(), comment))
	if err != nil {
		return "", err
	}
	var keys []string
	for k := range p.p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := p.p[k]
		_, err := buffer.WriteString(fmt.Sprintf("\n# %s\n%s = %s\n", v.description, v.Key, v.Value))
		if err != nil {
			return "", err
		}
	}
	return buffer.String(), nil
}

func (p *properties) Write(w io.Writer) (int, error) {
	s, err := p.String()
	if err != nil {
		return 0, err
	}
	return w.Write([]byte(s))
}

func Properties() *properties {
	return &properties{p: map[string]*property{
		gaugeRepositoryURL:      newProperty(gaugeRepositoryURL, "https://downloads.gauge.org/plugin", "Url to get plugin versions"),
		gaugeTemplatesURL:       newProperty(gaugeTemplatesURL, "https://templates.gauge.org", "Url to get templates list"),
		runnerConnectionTimeout: newProperty(runnerConnectionTimeout, "30000", "Timeout in milliseconds for making a connection to the language runner."),
		pluginConnectionTimeout: newProperty(pluginConnectionTimeout, "10000", "Timeout in milliseconds for making a connection to plugins."),
		pluginKillTimeOut:       newProperty(pluginKillTimeOut, "4000", "Timeout in milliseconds for a plugin to stop after a kill message has been sent."),
		runnerRequestTimeout:    newProperty(runnerRequestTimeout, "30000", "Timeout in milliseconds for requests from the language runner."),
		ideRequestTimeout:       newProperty(ideRequestTimeout, "30000", "Timeout in milliseconds for requests from runner when invoked for ide."),
		checkUpdates:            newProperty(checkUpdates, "true", "Allow Gauge and its plugin updates to be notified."),
	}}
}

func MergedProperties() (*properties, error) {
	p := Properties()
	config, err := common.GetGaugeConfiguration()
	if err != nil {
		// if unable to get from gauge.properties, just return defaults.
		return p, nil
	}
	for k, v := range config {
		if _, ok := p.p[k]; ok {
			err := p.set(k, v)
			if err != nil {
				return nil, err
			}
		}
	}
	return p, nil
}

func Update(name, value string) error {
	p, err := MergedProperties()
	if err != nil {
		return err
	}
	err = p.set(name, value)
	if err != nil {
		return err
	}
	return writeConfig(p)
}

func Merge() error {
	v, err := gaugeVersionInProperties()
	if err != nil || version.CompareVersions(v, version.CurrentGaugeVersion, version.LesserThanFunc) {
		mp, err := MergedProperties()
		if err != nil {
			return err
		}
		return writeConfig(mp)
	}
	return nil
}

func GetProperty(name string) (string, error) {
	mp, err := MergedProperties()
	if err != nil {
		return "", err
	}
	return mp.get(name)
}

func List(machineReadable bool) (string, error) {
	var f formatter
	f = textFormatter{}
	if machineReadable {
		f = &jsonFormatter{}
	}
	mp, err := MergedProperties()
	if err != nil {
		return "", err
	}
	return mp.Format(f)
}

func newProperty(key, defaultValue, description string) *property {
	return &property{
		Key:          key,
		defaultValue: defaultValue,
		description:  description,
		Value:        defaultValue,
	}
}

func writeConfig(p *properties) error {
	gaugePropertiesFile, err := gaugePropertiesFile()
	if err != nil {
		return err
	}
	var f *os.File
	if _, err = os.Stat(gaugePropertiesFile); err != nil {
		f, err = os.Create(gaugePropertiesFile)
		if err != nil {
			return err
		}
	} else {
		f, err = os.OpenFile(gaugePropertiesFile, os.O_WRONLY, os.ModeExclusive)
		if err != nil {
			return err
		}
	}
	defer f.Close()
	_, err = p.Write(f)
	return err
}

func gaugePropertiesFile() (string, error) {
	dir, err := common.GetConfigurationDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, common.GaugePropertiesFile), err
}

func gaugeVersionInProperties() (*version.Version, error) {
	var v *version.Version
	pf, err := gaugePropertiesFile()
	if err != nil {
		return v, err
	}
	f, err := os.Open(pf)
	if err != nil {
		return v, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	l, _, err := r.ReadLine()
	if err != nil {
		return v, err
	}
	return version.ParseVersion(strings.TrimPrefix(string(l), "# Version "))
}
