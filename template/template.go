/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package template

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"

	"github.com/schollz/closestmatch"
)

const templateProperties = "template.properties"

type templates struct {
	t     map[string]*config.Property
	names []string
}

func (t *templates) String() (string, error) {
	var buffer strings.Builder
	for _, k := range t.names {
		v := t.t[k]
		_, err := buffer.WriteString(fmt.Sprintf("\n# %s\n%s = %s\n", v.Description, v.Key, v.Value))
		if err != nil {
			return "", err
		}
	}
	return buffer.String(), nil
}

func (t *templates) update(k, v string, validate bool) error {
	if validate {
		if _, err := url.ParseRequestURI(v); err != nil {
			return fmt.Errorf("Failed to add template '%s'. The template location must be a valid (https) URI", k)
		}
	}
	if _, ok := t.t[k]; ok {
		t.t[k].Value = v
	} else {
		t.t[k] = config.NewProperty(k, v, fmt.Sprintf("Template download information for gauge %s projects", k))
		t.names = append(t.names, k)
	}
	sort.Strings(t.names)
	return nil
}

func (t *templates) get(k string) (string, error) {
	if _, ok := t.t[k]; ok {
		return t.t[k].Value, nil
	}
	matches := t.closestMatch(k)
	if len(matches) > 0 {
		return "", fmt.Errorf("Cannot find Gauge template '%s'.\nThe most similar template names are\n\n\t%s", k, strings.Join(matches, "\n\t"))
	}
	return "", fmt.Errorf("Cannot find Gauge template '%s'", k)
}

func (t *templates) closestMatch(k string) []string {
	matches := []string{}
	cm := closestmatch.New(t.names, []int{2})
	for _, m := range cm.ClosestN(k, 5) {
		if m != "" {
			matches = append(matches, m)
		}
	}
	sort.Strings(matches)
	return matches
}

func (t *templates) write() error {
	s, err := t.String()
	if err != nil {
		return err
	}
	return config.Write(s, templateProperties)
}

func Update(name, value string) error {
	t, err := getTemplates()
	if err != nil {
		return err
	}
	if err := t.update(name, value, true); err != nil {
		return err
	}
	return t.write()
}

func Generate() error {
	cd, err := common.GetConfigurationDir()
	if err != nil {
		return err
	}
	if !common.FileExists(filepath.Join(cd, templateProperties)) {
		return defaults().write()
	}
	return nil
}

func Get(name string) (string, error) {
	mp, err := getTemplates()
	if err != nil {
		return "", err
	}
	return mp.get(name)
}

func All() (string, error) {
	t, err := getTemplates()
	if err != nil {
		return "", err
	}
	return strings.Join(t.names, "\n"), nil
}

func List(machineReadable bool) (string, error) {
	var f config.Formatter
	f = config.TextFormatter{Headers: []string{"Template Name", "Location"}}
	if machineReadable {
		f = config.JsonFormatter{}
	}
	t, err := getTemplates()
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
	prop := map[string]*config.Property{
		"dotnet":              getProperty("template-dotnet", "dotnet"),
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
	}
	return &templates{t: prop, names: getKeys(prop)}
}

func getKeys(prop map[string]*config.Property) []string {
	var keys []string
	for k := range prop {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func getTemplates() (*templates, error) {
	prop, err := common.GetGaugeConfigurationFor(templateProperties)
	if err != nil {
		return nil, err
	}
	t := &templates{t: make(map[string]*config.Property), names: []string{}}
	for k, v := range prop {
		if err := t.update(k, v, false); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func getProperty(repoName, templateName string) *config.Property {
	f := "https://github.com/getgauge/%s/releases/latest/download/%s.zip"
	templateURL := fmt.Sprintf(f, repoName, templateName)
	desc := fmt.Sprintf("Template for gauge %s projects", templateName)
	return config.NewProperty(templateName, templateURL, desc)
}
