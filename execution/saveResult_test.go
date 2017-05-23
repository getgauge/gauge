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
	os.RemoveAll(filepath.Join(config.ProjectRoot, dotGauge))
}
