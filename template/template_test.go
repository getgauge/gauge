/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package template

import (
	"fmt"
	"strings"
	"testing"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/version"
)

var templatesContent = "# Version " + version.CurrentGaugeVersion.String() + `
# This file contains Gauge template configurations. Do not delete

# Template download information for gauge dotnet projects
dotnet = https://github.com/getgauge/template-dotnet/releases/latest/download/dotnet.zip

# Template download information for gauge java projects
java = https://github.com/getgauge/template-java/releases/latest/download/java.zip

# Template download information for gauge js projects
js = https://github.com/getgauge/template-js/releases/latest/download/js.zip

# Template download information for gauge python projects
python = https://github.com/getgauge/template-python/releases/latest/download/python.zip

# Template download information for gauge ruby projects
ruby = https://github.com/getgauge/template-ruby/releases/latest/download/ruby.zip

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
	expected := "cannot find a Gauge template 'hello'"
	if e.Error() != expected {
		t.Errorf("expected error to be \n'%s'\nGot:\n'%s'", expected, e.Error())
	}
}

func TestGetShouldGetSimilarTemplatesIfDoesNotExists(t *testing.T) {
	temp := defaults()
	_, e := temp.get("jaba")
	if e == nil {
		t.Errorf("expected error. got nil")
	}
	expected := `cannot find a Gauge template 'jaba'.
The most similar template names are

	java`

	if e.Error() != expected {
		t.Errorf("expected error to be \n'%s'\nGot:\n'%s'", expected, e.Error())
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

func TestTemplateAll(t *testing.T) {
	want := defaults().names
	s, err := All()
	if err != nil {
		t.Error(err)
	}
	got := strings.Split(s, "\n")

	for i, x := range want {
		if got[i] != x {
			fmt.Printf("'%s'\n", got[i])
			fmt.Printf("'%s'\n", x)
			t.Errorf("Expected property no %d = %s, got %s", i, x, got[i])
		}
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

func TestTemplateList(t *testing.T) {
	want := []string{
		"Template Name                 \tLocation                           ",
		"--------------------------------------------------------------------------------------------------------------",
		"dotnet                        \thttps://github.com/getgauge/template-dotnet/releases/latest/download/dotnet.zip",
		"java                          \thttps://github.com/getgauge/template-java/releases/latest/download/java.zip",
		"js                            \thttps://github.com/getgauge/template-js/releases/latest/download/js.zip",
		"python                        \thttps://github.com/getgauge/template-python/releases/latest/download/python.zip",
		"ruby                          \thttps://github.com/getgauge/template-ruby/releases/latest/download/ruby.zip",
		"ts                            \thttps://github.com/getgauge/template-ts/releases/latest/download/ts.zip",
	}
	s, err := List(false)
	if err != nil {
		t.Error(err)
	}
	got := strings.Split(s, "\n")
	if len(got) != len(want) {
		t.Errorf("Expected %d entries, got %d", len(want), len(got))
	}
	for i, x := range want {
		if got[i] != x {
			t.Errorf("Properties text Format failed\nwant:`%s`\ngot: `%s`", x, got[i])
		}
	}
}
