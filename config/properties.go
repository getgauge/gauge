package config

import (
	"bytes"
	"io"

	"os"
	"path/filepath"

	"errors"

	"github.com/getgauge/common"
)

var gaugeProperties = map[string]*property{
	"gauge_repository_url":        newProperty("gauge_repository_url", "https://downloads.getgauge.io/plugin", "Url to get plugin versions"),
	"gauge_update_url":            newProperty("gauge_update_url", "https://downloads.getgauge.io/gauge", "Url for latest gauge version"),
	"gauge_templates_url":         newProperty("gauge_templates_url", "https://downloads.getgauge.io/templates", "Url to get templates list"),
	"runner_connection_timeout":   newProperty("runner_connection_timeout", "30000", "Timeout in milliseconds for making a connection to the language runner."),
	"plugin_connection_timeout":   newProperty("plugin_connection_timeout", "10000", "Timeout in milliseconds for making a connection to plugins."),
	"plugin_kill_timeout":         newProperty("plugin_kill_timeout", "4000", "Timeout in milliseconds for a plugin to stop after a kill message has been sent."),
	"runner_request_timeout":      newProperty("runner_request_timeout", "30000", "Timeout in milliseconds for requests from the language runner."),
	"check_updates":               newProperty("check_updates", "true", "Allow Gauge and its plugin updates to be notified."),
	"gauge_analytics_enabled":     newProperty("gauge_analytics_enabled", "false", "Allow Gauge to collect anonymous usage statistics"),
	"gauge_analytics_log_enabled": newProperty("gauge_analytics_log_enabled", "false", "Log request sent to Gauge analytics engine"),
}

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
	}
	return errors.New("Config property doesn't exist.")
}

func (p *properties) get(k string) (*property, error) {
	if _, ok := p.p[k]; ok {
		return p.p[k], nil
	}
	return nil, errors.New("Config property doesn't exist.")
}

func (p *properties) Format(f Formatter) (string, error) {
	var all []property
	for _, v := range p.p {
		all = append(all, *v)
	}
	return f.Format(all)
}

func (p *properties) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("# ")
	buffer.WriteString(comment)
	buffer.WriteString("\n")
	for _, v := range p.p {
		buffer.WriteString("\n")
		buffer.WriteString("# ")
		buffer.WriteString(v.description)
		buffer.WriteString("\n")
		buffer.WriteString(v.Key)
		buffer.WriteString(" = ")
		buffer.WriteString(v.Value)
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func (p *properties) Write(w io.Writer) (int, error) {
	return w.Write([]byte(p.String()))
}

func GaugeProperties() *properties {
	return &properties{p: gaugeProperties}
}

func GetProperties() *properties {
	p := GaugeProperties()
	config, err := common.GetGaugeConfiguration()
	if err != nil {
		return p
	}
	for k, v := range config {
		p.set(k, v)
	}
	return p
}

func Update(name, value string) error {
	p := GetProperties()
	err := p.set(name, value)
	if err != nil {
		return err
	}
	return writeConfig(p)
}

func GetProperty(name string) (*property, error) {
	p := GetProperties()
	return p.get(name)
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
	dir, err := common.GetConfigurationDir()
	f, err := os.OpenFile(filepath.Join(dir, common.GaugePropertiesFile), os.O_WRONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = p.Write(f)
	return err
}
