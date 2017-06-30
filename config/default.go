package config

import (
	"bytes"
	"io"

	"os"
	"path/filepath"

	"github.com/getgauge/common"
)

type property struct {
	name         string
	key          string
	defaultValue string
	description  string
	value        string
}

type properties struct {
	p map[string]*property
}

func (p *properties) set(k, v string) {
	if _, ok := p.p[k]; ok {
		p.p[k].value = v
	}
}

func (p *properties) update(name, value string) {
	for _, property := range p.p {
		if property.name == name {
			property.value = value
		}
	}
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
		buffer.WriteString(v.key)
		buffer.WriteString(" = ")
		buffer.WriteString(v.value)
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func (p *properties) Write(w io.Writer) (int, error) {
	return w.Write([]byte(p.String()))
}

func NewProperties() *properties {
	return &properties{p: gaugeProperties}
}

func newProperty(name, key, defaultValue, description string) *property {
	return &property{
		name:         name,
		key:          key,
		defaultValue: defaultValue,
		description:  description,
		value:        defaultValue,
	}
}

var gaugeProperties = map[string]*property{
	"gauge_repository_url":        newProperty("repository_url", "gauge_repository_url", "https://downloads.getgauge.io/plugin", "Url to get plugin versions"),
	"gauge_update_url":            newProperty("update_url", "gauge_update_url", "https://downloads.getgauge.io/gauge", "Url for latest gauge version"),
	"gauge_templates_url":         newProperty("templates_url", "gauge_templates_url", "https://downloads.getgauge.io/templates", "Url to get templates list"),
	"runner_connection_timeout":   newProperty("runner_connection_timeout", "runner_connection_timeout", "30000", "Timeout in milliseconds for making a connection to the language runner."),
	"plugin_connection_timeout":   newProperty("plugin_connection_timeout", "plugin_connection_timeout", "10000", "Timeout in milliseconds for making a connection to plugins."),
	"plugin_kill_timeout":         newProperty("plugin_kill_timeout", "plugin_kill_timeout", "4000", "Timeout in milliseconds for a plugin to stop after a kill message has been sent."),
	"runner_request_timeout":      newProperty("runner_request_timeout", "runner_request_timeout", "30000", "Timeout in milliseconds for requests from the language runner."),
	"check_updates":               newProperty("updates", "check_updates", "true", "Allow Gauge and its plugin updates to be notified."),
	"gauge_analytics_enabled":     newProperty("analytics", "gauge_analytics_enabled", "false", "Allow Gauge to collect anonymous usage statistics"),
	"gauge_analytics_log_enabled": newProperty("log_analytics", "gauge_analytics_log_enabled", "false", "Log request sent to Gauge analytics engine"),
}

const comment = `This file contains Gauge specific internal configurations. Do not delete`

func getProperties() *properties {
	p := NewProperties()
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
	p := getProperties()
	p.update(name, value)
	dir, err := common.GetConfigurationDir()
	f, err := os.OpenFile(filepath.Join(dir, common.GaugePropertiesFile), os.O_WRONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = p.Write(f)
	return err
}
