/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
)

var templatesContent = `
# Template for gauge dotnet projects
dotnet = https://github.com/getgauge/template-dotnet/releases/latest/download/dotnet.zip

# Template for gauge java projects
java = https://github.com/getgauge/template-java/releases/latest/download/java.zip

# Template for gauge java_gradle projects
java_gradle = https://github.com/getgauge/template-java-gradle/releases/latest/download/java_gradle.zip

# Template for gauge java_maven projects
java_maven = https://github.com/getgauge/template-java-maven/releases/latest/download/java_maven.zip

# Template for gauge java_maven_selenium projects
java_maven_selenium = https://github.com/getgauge/template-java-maven-selenium/releases/latest/download/java_maven_selenium.zip

# Template for gauge js projects
js = https://github.com/getgauge/template-js/releases/latest/download/js.zip

# Template for gauge js_simple projects
js_simple = https://github.com/getgauge/template-js-simple/releases/latest/download/js_simple.zip

# Template for gauge python projects
python = https://github.com/getgauge/template-python/releases/latest/download/python.zip

# Template for gauge python_selenium projects
python_selenium = https://github.com/getgauge/template-python-selenium/releases/latest/download/python_selenium.zip

# Template for gauge ruby projects
ruby = https://github.com/getgauge/template-ruby/releases/latest/download/ruby.zip

# Template for gauge ruby_selenium projects
ruby_selenium = https://github.com/getgauge/template-ruby-selenium/releases/latest/download/ruby_selenium.zip

# Template for gauge ts projects
ts = https://github.com/getgauge/template-ts/releases/latest/download/ts.zip
`

func TestMain(m *testing.M) {
	oldHome := os.Getenv("GAUGE_HOME")
	var assertNil = func(err error) {
		if err != nil {
			fmt.Println("Failed to setup test data")
			os.Exit(1)
		}
	}
	home, err := filepath.Abs("_testdata")
	assertNil(err)
	assertNil(os.Setenv("GAUGE_HOME", "_testdata"))
	assertNil(os.MkdirAll(filepath.Join(home, "config"), common.NewDirectoryPermissions))
	assertNil(Generate())
	e := m.Run()
	assertNil(os.Setenv("GAUGE_HOME", oldHome))
	assertNil(os.RemoveAll(home))
	os.Exit(e)
}

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
	expected := "Cannot find Gauge template 'hello'"
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
	expected := `Cannot find Gauge template 'jaba'.
The most similar template names are

	java
	java_gradle
	java_maven
	java_maven_selenium`

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
		"----------------------------------------------------------------------------------------------------------------------------------------",
		"dotnet                        \thttps://github.com/getgauge/template-dotnet/releases/latest/download/dotnet.zip",
		"java                          \thttps://github.com/getgauge/template-java/releases/latest/download/java.zip",
		"java_gradle                   \thttps://github.com/getgauge/template-java-gradle/releases/latest/download/java_gradle.zip",
		"java_maven                    \thttps://github.com/getgauge/template-java-maven/releases/latest/download/java_maven.zip",
		"java_maven_selenium           \thttps://github.com/getgauge/template-java-maven-selenium/releases/latest/download/java_maven_selenium.zip",
		"js                            \thttps://github.com/getgauge/template-js/releases/latest/download/js.zip",
		"js_simple                     \thttps://github.com/getgauge/template-js-simple/releases/latest/download/js_simple.zip",
		"python                        \thttps://github.com/getgauge/template-python/releases/latest/download/python.zip",
		"python_selenium               \thttps://github.com/getgauge/template-python-selenium/releases/latest/download/python_selenium.zip",
		"ruby                          \thttps://github.com/getgauge/template-ruby/releases/latest/download/ruby.zip",
		"ruby_selenium                 \thttps://github.com/getgauge/template-ruby-selenium/releases/latest/download/ruby_selenium.zip",
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

func TestTemplateListAfterUpdate(t *testing.T) {

	want := []string{
		"Template Name                 \tLocation                           ",
		"----------------------------------------------------------------------------------------------------------------------------------------",
		"dotnet                        \thttps://github.com/getgauge/template-dotnet/releases/latest/download/dotnet.zip",
		"foo                           \thttps://github.com/getgauge/template-foo/releases/latest/download/foo.zip",
		"java                          \thttps://github.com/getgauge/template-java/releases/latest/download/java.zip",
		"java_gradle                   \thttps://github.com/getgauge/template-java-gradle/releases/latest/download/java_gradle.zip",
		"java_maven                    \thttps://github.com/getgauge/template-java-maven/releases/latest/download/java_maven.zip",
		"java_maven_selenium           \thttps://github.com/getgauge/template-java-maven-selenium/releases/latest/download/java_maven_selenium.zip",
		"js                            \thttps://github.com/getgauge/template-js/releases/latest/download/js.zip",
		"js_simple                     \thttps://github.com/getgauge/template-js-simple/releases/latest/download/js_simple.zip",
		"python                        \thttps://github.com/getgauge/template-python/releases/latest/download/python.zip",
		"python_selenium               \thttps://github.com/getgauge/template-python-selenium/releases/latest/download/python_selenium.zip",
		"ruby                          \thttps://github.com/getgauge/template-ruby/releases/latest/download/ruby.zip",
		"ruby_selenium                 \thttps://github.com/getgauge/template-ruby-selenium/releases/latest/download/ruby_selenium.zip",
		"ts                            \thttps://github.com/getgauge/template-ts/releases/latest/download/ts.zip",
	}
	err := Update("foo", "https://github.com/getgauge/template-foo/releases/latest/download/foo.zip")
	if err != nil {
		t.Error(err)
	}
	s, err := List(false)
	got := strings.Split(s, "\n")
	if err != nil {
		t.Error(err)
	}
	if len(got) != len(want) {
		t.Errorf("Expected %d entries, got %d", len(want), len(got))
	}
	for i, x := range want {
		if got[i] != x {
			t.Errorf("Properties text Format failed\nwant:`%s`\ngot: `%s`", x, got[i])
		}
	}
}
