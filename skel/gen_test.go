/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package skel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/common"
)

func TestCreateSkelFilesIfRequired(t *testing.T) {
	config := "config"
	origGaugeHome := os.Getenv("GAUGE_HOME")
	gaugeHomeDir := filepath.Join("_testdata", "GaugeHome")
	os.RemoveAll(gaugeHomeDir)
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
	os.Setenv("GAUGE_HOME", origGaugeHome)
}
