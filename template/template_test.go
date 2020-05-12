/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package template

import (
	"fmt"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/version"
	"strings"
	"testing"
)

var templatesContent = "# Version " + version.CurrentGaugeVersion.String() + `
# This file contains Gauge template configurations. Do not delete

# Template download information for gauge dotnet projects
dotnet = https://github.com/getgauge/template-dotnet/releases/latest/download/dotnet.zip

# Template download information for gauge java projects
java = https://github.com/getgauge/template-java/releases/latest/download/java.zip

# Template download information for gauge java_gradle projects
java_gradle = https://github.com/getgauge/template-java-gradle/releases/latest/download/java_gradle.zip

# Template download information for gauge java_maven projects
java_maven = https://github.com/getgauge/template-java-maven/releases/latest/download/java_maven.zip

# Template download information for gauge java_maven_selenium projects
java_maven_selenium = https://github.com/getgauge/template-java-maven-selenium/releases/latest/download/java_maven_selenium.zip

# Template download information for gauge js projects
js = https://github.com/getgauge/template-js/releases/latest/download/js.zip

# Template download information for gauge js_simple projects
js_simple = https://github.com/getgauge/template-js-simple/releases/latest/download/js_simple.zip

# Template download information for gauge python projects
python = https://github.com/getgauge/template-python/releases/latest/download/python.zip

# Template download information for gauge python_selenium projects
python_selenium = https://github.com/getgauge/template-python-selenium/releases/latest/download/python_selenium.zip

# Template download information for gauge ruby projects
ruby = https://github.com/getgauge/template-ruby/releases/latest/download/ruby.zip

# Template download information for gauge ruby_selenium projects
ruby_selenium = https://github.com/getgauge/template-ruby-selenium/releases/latest/download/ruby_selenium.zip

# Template download information for gauge ts projects
ts = https://github.com/getgauge/template-ts/releases/latest/download/ts.zip
`

func TestUpdateShouldAddTemplateIfDoesNotExistss(t *testing.T) {
	temp := &templates{
		t: map[string]*config.Property{},
	}
	e := temp.update("hello", "https://templates.org/foo.zip", false)
	if e != nil {
		t.Errorf("expected error to be nil. got '%s'", e.Error())
	}
}

func TestUpdateShouldNotAddInValidURL(t *testing.T) {
	temp := &templates{
		t: map[string]*config.Property{},
	}
	e := temp.update("hello", "foo/bar", true)
	if e == nil {
		t.Errorf("expected error to not be nil. got '%s'", e.Error())
	}
	expected := "Failed to add template 'hello'. The template location must be a valid (https) URI"
	if e.Error() != expected {
		t.Errorf("expected error to be '%s'. got '%s'", expected, e.Error())
	}
}

func TestGetShouldGetTemplateIfDoesNotExists(t *testing.T) {
	temp := &templates{
		t: map[string]*config.Property{},
	}
	_, e := temp.get("hello")
	if e == nil {
		t.Errorf("expected error. got nil")
	}
}

func TestGetShouldGetTemplateIfExists(t *testing.T) {
	temp := &templates{
		t: map[string]*config.Property{
			"hello": config.NewProperty("hello", "/foo/bar", ""),
		},
	}

	v, e := temp.get("hello")
	if e != nil {
		t.Errorf("expected error to be nil. got '%s'", e.Error())
	}

	if v != "/foo/bar" {
		t.Errorf("Expected: '/foo/bar'\nGot: '%s'", v)
	}
}

func TestTemplateString(t *testing.T) {
	want := strings.Split(templatesContent, "\n\n")

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
			fmt.Printf("'%s'\n", got[i])
			fmt.Printf("'%s'\n", x)
			t.Errorf("Expected property no %d = %s, got %s", i, x, got[i])
		}
	}
}
