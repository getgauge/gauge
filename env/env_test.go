/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestLoadDefaultEnv(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv(common.DefaultEnvDir, nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("gauge_environment"), Equals, common.DefaultEnvDir)
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "specs")
	c.Assert(os.Getenv("csv_delimiter"), Equals, ",")
	defaultScreenshotDir := filepath.Join(config.ProjectRoot, common.DotGauge, "screenshots")
	c.Assert(os.Getenv("gauge_screenshots_dir"), Equals, defaultScreenshotDir)
	c.Assert(os.Getenv("gauge_spec_file_extensions"), Equals, ".spec, .md")
	c.Assert(os.Getenv("allow_case_sensitive_tags"), Equals, "false")
}

// If default env dir is present, the values present in there should overwrite
// the default values (present in the code), even when env flag is passed
func (s *MySuite) TestLoadDefaultEnvFromDirIfPresent(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj2"

	e := LoadEnv("foo", nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports_dir")
	c.Assert(os.Getenv("gauge_environment"), Equals, "foo")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "false")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "false")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "anotherSpecDir")
}

func (s *MySuite) TestLoadDefaultEnvFromCustomDir(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj9"
	os.Setenv("gauge_env_dir", "customEnv")

	e := LoadEnv(common.DefaultEnvDir, nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("gauge_environment"), Equals, common.DefaultEnvDir)
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "specs")
	c.Assert(os.Getenv("csv_delimiter"), Equals, ",")
	defaultScreenshotDir := filepath.Join(config.ProjectRoot, common.DotGauge, "screenshots")
	c.Assert(os.Getenv("gauge_screenshots_dir"), Equals, defaultScreenshotDir)
	c.Assert(os.Getenv("gauge_spec_file_extensions"), Equals, ".spec, .md")
	c.Assert(os.Getenv("allow_case_sensitive_tags"), Equals, "false")
}

func (s *MySuite) TestLoadDefaultEnvFromDirIfPresentFromCustomDir(c *C) {
	os.Clearenv()
	os.Setenv("gauge_env_dir", "customEnv")
	config.ProjectRoot = "_testdata/proj10"

	e := LoadEnv("foo", nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports_dir")
	c.Assert(os.Getenv("gauge_environment"), Equals, "foo")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "false")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "false")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "anotherSpecDir")
}

func (s *MySuite) TestLoadDefaultEnvFromManifestDir(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj7"

	e := LoadEnv(common.DefaultEnvDir, nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("gauge_environment"), Equals, common.DefaultEnvDir)
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "specs")
	c.Assert(os.Getenv("csv_delimiter"), Equals, ",")
	defaultScreenshotDir := filepath.Join(config.ProjectRoot, common.DotGauge, "screenshots")
	c.Assert(os.Getenv("gauge_screenshots_dir"), Equals, defaultScreenshotDir)
	c.Assert(os.Getenv("gauge_spec_file_extensions"), Equals, ".spec, .md")
	c.Assert(os.Getenv("allow_case_sensitive_tags"), Equals, "false")
}

func (s *MySuite) TestLoadDefaultEnvFromDirIfPresentFromManifestDir(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj8"

	e := LoadEnv("foo", nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports_dir")
	c.Assert(os.Getenv("gauge_environment"), Equals, "foo")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "false")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "false")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "anotherSpecDir")
}

// If default env dir is present, the values present in there should overwrite
// the default values (present in the code), even when env flag is passed.
// If the passed env also has the same values, that should take precedence.
func (s *MySuite) TestLoadDefaultEnvFromDirAndOverwritePassedEnv(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj2"

	e := LoadEnv("bar", nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports_dir")
	c.Assert(os.Getenv("gauge_environment"), Equals, "bar")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "false")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "bar/logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "bar/specs")
}

func (s *MySuite) TestLoadDefaultEnvEvenIfDefaultEnvNotPresent(c *C) {
	os.Clearenv()
	config.ProjectRoot = ""

	e := LoadEnv(common.DefaultEnvDir, nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("gauge_environment"), Equals, common.DefaultEnvDir)
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "specs")
}

func (s *MySuite) TestLoadDefaultEnvWithOtherPropertiesSetInShell(c *C) {
	os.Clearenv()
	os.Setenv("foo", "bar")
	os.Setenv("logs_directory", "custom_logs_dir")
	os.Setenv("gauge_specs_dir", "custom_specs_dir")
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv(common.DefaultEnvDir, nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("foo"), Equals, "bar")
	c.Assert(os.Getenv("property1"), Equals, "value1")
	c.Assert(os.Getenv("logs_directory"), Equals, "custom_logs_dir")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "custom_specs_dir")
}

func (s *MySuite) TestLoadDefaultEnvWithOtherPropertiesNotSetInShell(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv(common.DefaultEnvDir, nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("property1"), Equals, "value1")
}

func (s *MySuite) TestLoadCustomEnvAlongWithDefaultEnv(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("foo", nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "false")
	c.Assert(os.Getenv("logs_directory"), Equals, "foo/logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "foo/specs")
}

func (s *MySuite) TestLoadCustomEnvAlongWithOtherPropertiesSetInShell(c *C) {
	os.Clearenv()
	os.Setenv("gauge_reports_dir", "custom_reports_dir")
	os.Setenv("gauge_specs_dir", "custom_specs_dir")
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("foo", nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "custom_reports_dir")
	c.Assert(os.Getenv("overwrite_reports"), Equals, "true")
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "false")
	c.Assert(os.Getenv("logs_directory"), Equals, "foo/logs")
	c.Assert(os.Getenv("gauge_specs_dir"), Equals, "custom_specs_dir")
}

func (s *MySuite) TestLoadCustomEnvWithCommentsInPropertiesSet(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("test", nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("test_url"), Equals, "http://testurl")
}

func (s *MySuite) TestLoadMultipleEnv(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj2"

	e := LoadEnv("bar, foo", nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "bar/logs")
}

func (s *MySuite) TestLoadMultipleEnvEnsureFirstOneDecides(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj2"

	e := LoadEnv("bar, default", nil)

	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("screenshot_on_failure"), Equals, "true")
	c.Assert(os.Getenv("logs_directory"), Equals, "bar/logs")
	c.Assert(os.Getenv("gauge_reports_dir"), Equals, "reports_dir")
}

func (s *MySuite) TestEnvPropertyIsSet(c *C) {
	os.Clearenv()
	os.Setenv("foo", "bar")

	actual := isPropertySet("foo")

	c.Assert(actual, Equals, true)
}

func (s *MySuite) TestEnvPropertyIsNotSet(c *C) {
	os.Clearenv()

	actual := isPropertySet("foo")

	c.Assert(actual, Equals, false)
}

func (s *MySuite) TestFatalErrorIsThrownIfEnvNotFound(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("bar", nil)
	c.Assert(e.Error(), Equals, "Failed to load env. bar environment does not exist")
}

func (s *MySuite) TestLoadDefaultEnvWithSubstitutedVariables(c *C) {
	os.Clearenv()
	os.Setenv("foo", "bar")
	os.Setenv("property1", "value1")

	config.ProjectRoot = "_testdata/proj3"

	e := LoadEnv(common.DefaultEnvDir, nil)

	c.Assert(e, Equals, nil)

	c.Assert(os.Getenv("property1"), Equals, "value1")
	c.Assert(os.Getenv("property3"), Equals, "bar/value1")
	c.Assert(os.Getenv("property2"), Equals, "value1/value2")
}

func (s *MySuite) TestLoadDefaultEmptyEnvWithSubstitutedVariables(c *C) {
	os.Clearenv()
	os.Setenv("property1", "")

	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv(common.DefaultEnvDir, nil)

	c.Assert(e, Equals, nil)

	c.Assert(os.Getenv("property1"), Equals, "value1")
}

func (s *MySuite) TestLoadDefaultEnvWithMissingSubstitutedVariable(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj5"
	e := LoadEnv(common.DefaultEnvDir, nil)
	c.Assert(e.Error(), Equals, "Failed to load env. [missingProperty] env variable(s) are not set")
}

func (s *MySuite) TestLoadDefaultEnvWithMissingSubstitutedVariableWhenAssignedToProperty(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj4"
	e := LoadEnv("missing", nil)
	c.Assert(e.Error(), Equals, "Failed to load env. [c] env variable(s) are not set")
}

func (s *MySuite) TestLoadDefaultEnvWithMissingSubstitutedVariableAcrossMultipleFiles(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj4"
	e := LoadEnv("missing-multi", nil)
	c.Assert(e.Error(), Equals, "Failed to load env. [d] env variable(s) are not set")
}

func (s *MySuite) TestCurrentEnvironmentIsPopulated(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj1"

	e := LoadEnv("foo", nil)

	c.Assert(e, Equals, nil)
	c.Assert(CurrentEnvironments(), Equals, "foo")
}

func (s *MySuite) TestGetDefaultSpecFileExtensions(c *C) {
	os.Clearenv()
	var contains = func(c []string, what string) bool {
		for _, x := range c {
			if x == what {
				return true
			}
		}
		return false
	}

	exts := GaugeSpecFileExtensions()
	for _, expected := range []string{".spec", ".md"} {
		c.Assert(contains(exts, expected), Equals, true)
	}
}

func (s *MySuite) TestGetSpecFileExtensionsSetViaEnv(c *C) {
	os.Clearenv()
	os.Setenv(gaugeSpecFileExtensions, ".foo, .bar")
	var contains = func(c []string, what string) bool {
		for _, x := range c {
			if x == what {
				return true
			}
		}
		return false
	}

	exts := GaugeSpecFileExtensions()
	for _, expected := range []string{".foo", ".bar"} {
		c.Assert(contains(exts, expected), Equals, true)
	}
}

func (s *MySuite) TestShouldNotGetDefaultExtensionsWhenEnvIsSet(c *C) {
	os.Clearenv()
	os.Setenv(gaugeSpecFileExtensions, ".foo, .bar")
	var contains = func(c []string, what string) bool {
		for _, x := range c {
			if x == what {
				return true
			}
		}
		return false
	}

	exts := GaugeSpecFileExtensions()
	for _, expected := range []string{".spec", ".md"} {
		c.Assert(contains(exts, expected), Equals, false)
	}
}

func (s *MySuite) TestLoadDefaultEnvWithSubstitutedVariablesFromProperties(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj4"
	e := LoadEnv(common.DefaultEnvDir, nil)
	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("service_url"), Equals, "http://default.service.com")
	c.Assert(os.Getenv("path"), Equals, "api/default/endpoint")
	c.Assert(os.Getenv("api_url"), Equals, "http://default.service.com/api/default/endpoint")
}

func (s *MySuite) TestLoadEnvWithSubstitutedVariablesFromProperties(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj4"
	e := LoadEnv("foo", nil)
	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("service_url"), Equals, "http://foo.service.com")
	c.Assert(os.Getenv("path"), Equals, "api/default/endpoint")
	c.Assert(os.Getenv("api_url"), Equals, "http://foo.service.com/api/default/endpoint")
	c.Assert(os.Getenv("nested"), Equals, "http://foo.service.com/api/default/endpoint")
}

func (s *MySuite) TestLoadEnvWithSubstitutedVariablesFromPropertiesAndSetInShell(c *C) {
	os.Clearenv()
	os.Setenv("service_url", "http://system.service.com")
	config.ProjectRoot = "_testdata/proj4"
	e := LoadEnv("foo", nil)
	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("service_url"), Equals, "http://system.service.com")
	c.Assert(os.Getenv("path"), Equals, "api/default/endpoint")
	c.Assert(os.Getenv("api_url"), Equals, "http://system.service.com/api/default/endpoint")
	c.Assert(os.Getenv("nested"), Equals, "http://system.service.com/api/default/endpoint")
}

func (s *MySuite) TestLoadEnvWithCircularPropertiesAcrossEnvironments(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj6"
	e := LoadEnv("circular", nil)
	c.Assert(e, ErrorMatches, "Failed to load env. circular reference in:\n.*\n")
}

func (s *MySuite) TestLoadEnvWithCircularPropertiesAcrossFiles(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj4"
	e := LoadEnv("circular", func(err error) {
		c.Assert(err, ErrorMatches, "circular reference in:\n.*\n.*\n.*\n")
	})
	c.Assert(e, Equals, nil)
}

func (s *MySuite) TestLoadEnvWithAcyclicProperties(c *C) {
	os.Clearenv()
	config.ProjectRoot = "_testdata/proj4"
	e := LoadEnv("acyclic", nil)
	c.Assert(e, Equals, nil)
	c.Assert(os.Getenv("a"), Equals, "foo/foo/foo")
	c.Assert(os.Getenv("b"), Equals, "foo")
	c.Assert(os.Getenv("c"), Equals, "foo")
	c.Assert(os.Getenv("d"), Equals, "foo")
	c.Assert(os.Getenv("e"), Equals, "foo")
	c.Assert(os.Getenv("f"), Equals, "foo")
}

func (s *MySuite) TestScenarioInitStrategyDefaultsToEager(c *C) {
	os.Clearenv()
	strategy := ScenarioInitStrategy()
	c.Assert(strategy, Equals, "eager")
}

func (s *MySuite) TestScenarioInitStrategyReturnsLazy(c *C) {
	os.Clearenv()
	os.Setenv("scenario_init_strategy", "lazy")
	strategy := ScenarioInitStrategy()
	c.Assert(strategy, Equals, "lazy")
}

func (s *MySuite) TestScenarioInitStrategyWithInvalidValueDefaultsToEager(c *C) {
	os.Clearenv()
	os.Setenv("scenario_init_strategy", "invalid")
	strategy := ScenarioInitStrategy()
	c.Assert(strategy, Equals, "eager")
}
