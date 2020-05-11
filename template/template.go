/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package template

import (
	"fmt"
	"github.com/getgauge/gauge/config"
	"net/url"
	"sort"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/version"
)

const comment = `This file contains Gauge template configurations. Do not delete`
const templateProperties = "template.properties"

type templates struct {
	t map[string]*config.Property
}

func (t *templates) String() (string, error) {
	var buffer strings.Builder
	_, err := buffer.WriteString(fmt.Sprintf("# Version %s\n# %s\n", version.FullVersion(), comment))
	if err != nil {
		return "", err
	}
	var keys []string
	for k := range t.t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := t.t[k]
		_, err := buffer.WriteString(fmt.Sprintf("\n# %s\n%s = %s\n", v.Description, v.Key, v.Value))
		if err != nil {
			return "", err
		}
	}
	return buffer.String(), nil
}

func (t *templates) update(k, v string) error {
	if _, err := url.Parse(v); err != nil {
		return err
	}
	if _, ok := t.t[k]; ok {
		t.t[k].Value = v
	} else {
		t.t[k] = config.NewProperty(k, v, fmt.Sprintf("Template download information for gauge %s projects", k))
	}
	return nil
}

func (t *templates) get(k string) (string, error) {
	if _, ok := t.t[k]; ok {
		return t.t[k].Value, nil
	}
	return "", fmt.Errorf("config '%s' doesn't exist", k)
}

func (t *templates) write() error {
	s, err := t.String()
	if err != nil {
		return err
	}
	return config.Write(s, templateProperties)
}

func Update(name, value string) error {
	t, err := mergeTemplates()
	if err != nil {
		return err
	}
	if err := t.update(name, value); err != nil {
		return err
	}

	return t.write()
}

func Merge() error {
	v, err := config.GaugeVersionInPropertiesFile(templateProperties)
	if err != nil || version.CompareVersions(v, version.CurrentGaugeVersion, version.LesserThanFunc) {
		t, err := mergeTemplates()
		if err != nil {
			return err
		}
		return t.write()
	}
	return nil
}

func Get(name string) (string, error) {
	mp, err := mergeTemplates()
	if err != nil {
		return "", err
	}
	return mp.get(name)
}

func All() (string, error) {
	t, err := mergeTemplates()
	if err != nil {
		return "", err
	}
	var all []string
	for k := range t.t {
		all = append(all, k)
	}
	sort.Strings(all)
	return strings.Join(all, "\n"), nil
}

func List(machineReadable bool) (string, error) {
	f := config.TextFormatter{}
	t, err := mergeTemplates()
	if err != nil {
		return "", err
	}
	var all []config.Property
	for _, v := range t.t {
		all = append(all, *v)
	}
	return f.Format(all)
}

func defaults() *templates {
	return &templates{t: map[string]*config.Property{
		"dotnet":              getProperty("template-dotnet", "dontet"),
		"java":                getProperty("template-java", "java"),
		"java_gradle":         getProperty("template-java-gradle", "java_gradle"),
		"java_maven":          getProperty("template-java-maven", "java_maven"),
		"java_maven_selenium": getProperty("template-java-maven-selenium", "java_maven_selenium"),
		"js":                  getProperty("template-js", "js"),
		"js_simple":           getProperty("template-js-simple", "js_simple"),
		"python":              getProperty("template-python", "python"),
		"python_selenium":     getProperty("template-python-selenium", "python_selenium"),
		"ruby":                getProperty("template-ruby", "ruby"),
		"ruby_selenium":       getProperty("template-ruby-selenium", "ruby_selenium"),
		"ts":                  getProperty("template-ts", "ts"),
	}}
}

func mergeTemplates() (*templates, error) {
	t := defaults()
	config, err := common.GetGaugeConfigurationFor(templateProperties)
	if err != nil {
		return t, nil
	}
	for k, v := range config {
		if err := t.update(k, v); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func getProperty(repoName, templateName string) *config.Property {
	f := "https://github.com/getgauge/%s/releases/latest/download/%s.zip"
	url := fmt.Sprintf(f, repoName, templateName)
	desc := fmt.Sprintf("Template download information for gauge %s projects", templateName)
	return config.NewProperty(templateName, url, desc)
}
