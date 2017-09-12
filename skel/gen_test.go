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

package skel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/common"
)

func TestCreateSkelFilesIfRequired(t *testing.T) {
	config := "config"
	setupPlugins = func() {}
	origGaugeHome := os.Getenv("GAUGE_HOME")
	gaugeHomeDir := filepath.Join("_testdata", "GaugeHome")
	err := os.Mkdir(gaugeHomeDir, common.NewDirectoryPermissions)
	if err != nil {
		t.Fatalf("Unable to create Gauge Root Dir, %s", err)
	}
	os.Setenv("GAUGE_HOME", gaugeHomeDir)
	expectedFiles := []string{
		"notice.md",
		"gauge.properties",
		filepath.Join("skel", "example.spec"),
		filepath.Join("skel", ".gitignore"),
		filepath.Join("skel", "env", "default.properties"),
	}

	CreateSkelFilesIfRequired()

	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(filepath.Join(gaugeHomeDir, config, expectedFile)); err != nil {
			t.Errorf("Expected %s to exist. %s", expectedFile, err)
		}
	}
	err = os.RemoveAll(gaugeHomeDir)
	if err != nil {
		t.Fatalf("Unable to clean up Gauge Root Dir, %s", err)
	}
	os.Setenv("GAUGE_ROOT", origGaugeHome)
}
