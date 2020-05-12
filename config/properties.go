/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
	logging "github.com/op/go-logging"
)

var Log = logging.MustGetLogger("gauge")

const comment = `This file contains Gauge specific internal configurations. Do not delete`

type Property struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	Description  string
	defaultValue string
}

type properties struct {
	p map[string]*Property
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

func (p *properties) Format(f Formatter) (string, error) {
	var all []Property
	for _, v := range p.p {
		all = append(all, *v)
	}
	return f.Format(all)
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
		_, err := buffer.WriteString(fmt.Sprintf("\n# %s\n%s = %s\n", v.Description, v.Key, v.Value))
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

func defaults() *properties {
	return &properties{p: map[string]*Property{
		allowInsecureDownload:   NewProperty(allowInsecureDownload, "false", "Allow Gauge to download template from insecure URLs."),
		gaugeRepositoryURL:      NewProperty(gaugeRepositoryURL, "https://downloads.gauge.org/plugin", "Url to get plugin versions"),
		runnerConnectionTimeout: NewProperty(runnerConnectionTimeout, "30000", "Timeout in milliseconds for making a connection to the language runner."),
		pluginConnectionTimeout: NewProperty(pluginConnectionTimeout, "10000", "Timeout in milliseconds for making a connection to plugins."),
		pluginKillTimeOut:       NewProperty(pluginKillTimeOut, "4000", "Timeout in milliseconds for a plugin to stop after a kill message has been sent."),
		runnerRequestTimeout:    NewProperty(runnerRequestTimeout, "30000", "Timeout in milliseconds for requests from the language runner."),
		ideRequestTimeout:       NewProperty(ideRequestTimeout, "30000", "Timeout in milliseconds for requests from runner when invoked for ide."),
		checkUpdates:            NewProperty(checkUpdates, "true", "Allow Gauge and its plugin updates to be notified."),
	}}
}

func mergedProperties() (*properties, error) {
	p := defaults()
	config, err := common.GetGaugeConfigurationFor(common.GaugePropertiesFile)
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
	p, err := mergedProperties()
	if err != nil {
		return err
	}
	err = p.set(name, value)
	if err != nil {
		return err
	}
	s, err := p.String()
	if err != nil {
		return err
	}
	return Write(s, common.GaugePropertiesFile)
}

func Merge() error {
	v, err := GaugeVersionInPropertiesFile(common.GaugePropertiesFile)
	if err != nil || version.CompareVersions(v, version.CurrentGaugeVersion, version.LesserThanFunc) {
		mp, err := mergedProperties()
		if err != nil {
			return err
		}
		s, err := mp.String()
		if err != nil {
			return err
		}
		return Write(s, common.GaugePropertiesFile)
	}
	return nil
}

func GetProperty(name string) (string, error) {
	mp, err := mergedProperties()
	if err != nil {
		return "", err
	}
	return mp.get(name)
}

func List(machineReadable bool) (string, error) {
	var f Formatter
	f = TextFormatter{Headers: []string{"Key", "Value"}}
	if machineReadable {
		f = &JsonFormatter{}
	}
	mp, err := mergedProperties()
	if err != nil {
		return "", err
	}
	return mp.Format(f)
}

func NewProperty(key, defaultValue, description string) *Property {
	return &Property{
		Key:          key,
		defaultValue: defaultValue,
		Description:  description,
		Value:        defaultValue,
	}
}

func Write(text, file string) error {
	file, err := FilePath(file)
	if err != nil {
		return err
	}
	var f *os.File
	if _, err = os.Stat(file); err != nil {
		f, err = os.Create(file)
		if err != nil {
			return err
		}
	} else {
		f, err = os.OpenFile(file, os.O_WRONLY, os.ModeExclusive)
		if err != nil {
			return err
		}
	}
	defer f.Close()
	_, err = f.Write([]byte(text))
	return err
}

func FilePath(name string) (string, error) {
	dir, err := common.GetConfigurationDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), err
}

func GaugeVersionInPropertiesFile(name string) (*version.Version, error) {
	var v *version.Version
	pf, err := FilePath(name)
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

var GetPropertyFromConfig = func(propertyName string) string {
	config, err := common.GetGaugeConfiguration()
	if err != nil {
		APILog.Warningf("Failed to get configuration from Gauge properties file. Error: %s", err.Error())
		return ""
	}
	return config[propertyName]
}
