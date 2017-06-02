// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/common"
)

func stubGetFromConfig(propertyName string) string {
	return ""
}

func stub2GetFromConfig(propertyName string) string {
	return "10000"
}

func stub3GetFromConfig(propertyName string) string {
	return "false"
}

func stub4GetFromConfig(propertyName string) string {
	return "true	"
}

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

	os.Setenv(runnerRequestTimeout, "1000")
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

func TestReadUniqueID(t *testing.T) {
	expected := "foo"
	idFile := filepath.Join("_testData", ".gauge_id")
	ioutil.WriteFile(idFile, []byte(expected), common.NewFilePermissions)

	s, err := filepath.Abs("_testData")
	if err != nil {
		t.Error(err)
	}

	os.Setenv("GAUGE_ROOT", s)
	got := UniqueID()

	if got != expected {
		t.Errorf("Expected UniqueID=%s, got %s", expected, got)
	}
	os.Setenv("GAUGE_ROOT", "")
	err = os.Remove(idFile)
	if err != nil {
		t.Error(err)
	}
}
