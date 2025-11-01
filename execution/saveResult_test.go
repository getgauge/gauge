/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package execution

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/result"
)

func TestIfResultFileIsCreated(t *testing.T) {
	msg := &result.SuiteResult{}

	writeResult(msg)

	file := filepath.Join(config.ProjectRoot, dotGauge, lastRunResult)

	if !common.FileExists(file) {
		t.Errorf("Expected file %s to exist", file)
	}
	_ = os.RemoveAll(filepath.Join(config.ProjectRoot, dotGauge))
}
