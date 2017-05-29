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

	"io/ioutil"

	"sync"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/golang/protobuf/proto"
)

const (
	dotGauge      = ".gauge"
	lastRunResult = "last_run_result"
)

// ListenSuiteEndAndSaveResult listens to execution events and writes the failed scenarios to JSON file
func ListenSuiteEndAndSaveResult(wg *sync.WaitGroup) {
	ch := make(chan event.ExecutionEvent, 0)
	event.Register(ch, event.SuiteEnd)
	wg.Add(1)

	go func() {
		for {
			e := <-ch
			if e.Topic == event.SuiteEnd {
				writeResult(e.Result.(*result.SuiteResult))
				wg.Done()
			}
		}
	}()
}

func writeResult(res *result.SuiteResult) {
	dotGaugeDir := filepath.Join(config.ProjectRoot, dotGauge)
	resultFile := filepath.Join(config.ProjectRoot, dotGauge, lastRunResult)
	if err := os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Errorf("Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	r, err := proto.Marshal(gauge.ConvertToProtoSuiteResult(res))
	if err != nil {
		logger.Errorf("Unable to marshal suite execution result, skipping save. %s", err.Error())
	}
	err = ioutil.WriteFile(resultFile, r, common.NewFilePermissions)
	if err != nil {
		logger.Errorf("Failed to write to %s. Reason: %s", resultFile, err.Error())
	} else {
		logger.Debug("Last run result saved to %s", resultFile)
	}
}
