/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package config

import (
	"os"
	"testing"
)

func stubGetFromConfig(propertyName string) string { return "" }
func stub2GetFromConfig(propertyName string) string { return "10000" }
func stub3GetFromConfig(propertyName string) string { return "false" }
func stub4GetFromConfig(propertyName string) string { return "true	" }

func stubLogMaxBackupsCountDefault(propertyName string) string { return "" }
func stubLogMaxBackupsCountCustom(propertyName string) string { return "10" }

func stubDataTableModeFilter(propertyName string) string { return "filter" }
func stubDataTableModeMatrix(propertyName string) string { return "matrix" }
func stubDataTableModeInvalid(propertyName string) string { return "banana" }


func TestRunnerRequestTimeout(t *testing.T) {
	getFromConfig = stubGetFromConfig
	expected := defaultRunnerRequestTimeout
	got := RunnerRequestTimeout()
	if got != expected {
		t.Errorf("Expected RunnerRequestTimeout == defaultRunnerRequestTimeout(%s), got %s", expected, got)
	}

	getFromConfig = stub2GetFromConfig
	got1 := RunnerRequestTimeout().Seconds()
	expected1 := float64(10)
	if got1 != expected1 {
		t.Errorf("Expected RunnerRequestTimeout == defaultRunnerRequestTimeout(%f), got %f", expected1, got1)
	}

	_ = os.Setenv(runnerRequestTimeout, "1000")
	got1 = RunnerRequestTimeout().Seconds()
	expected1 = float64(1)
	if got != expected {
		t.Errorf("Expected RunnerRequestTimeout == defaultRunnerRequestTimeout(%f), got %f", expected1, got1)
	}
}

func TestAllowUpdates(t *testing.T) {
	getFromConfig = stubGetFromConfig
	if !CheckUpdates() {
		t.Error("Expected CheckUpdates=true, got false")
	}

	getFromConfig = stub2GetFromConfig
	if !CheckUpdates() {
		t.Error("Expected CheckUpdates=true, got false")
	}

	getFromConfig = stub3GetFromConfig
	if CheckUpdates() {
		t.Error("Expected CheckUpdates=true, got true")
	}

	getFromConfig = stub4GetFromConfig
	if !CheckUpdates() {
		t.Error("Expected CheckUpdates=true, got false")
	}
}

func TestLogMaxBackups(t *testing.T) {
	getFromConfig = stubLogMaxBackupsCountDefault
	got := LogMaxBackupsCount()
	if got != 3 {
		t.Errorf("expected default LogMaxBackupsCount to be 3, got %d", got)
	}

	getFromConfig = stubLogMaxBackupsCountCustom
	got = LogMaxBackupsCount()
	if got != 10 {
		t.Errorf("expected LogMaxBackupsCount to be 10, got %d", got)
	}
}

func TestDataTableMode(t *testing.T) {
	getFromConfig = stubGetFromConfig
	if DataTableMode() != "filter" {
		t.Errorf("expected default to be 'filter'")
	}

	getFromConfig = stubDataTableModeFilter
	if DataTableMode() != "filter" {
		t.Errorf("expected 'filter'")
	}

	getFromConfig = stubDataTableModeMatrix
	if DataTableMode() != "matrix" {
		t.Errorf("expected 'matrix'")
	}

	getFromConfig = stubDataTableModeInvalid
	if DataTableMode() != "filter" {
		t.Errorf("expected fallback to 'filter' on invalid value")
	}
}
