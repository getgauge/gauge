/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
	"google.golang.org/protobuf/proto"
)

const (
	dotGauge      = ".gauge"
	lastRunResult = "last_run_result"
)

// ListenSuiteEndAndSaveResult listens to execution events and writes the failed scenarios to JSON file
func ListenSuiteEndAndSaveResult(wg *sync.WaitGroup) {
	ch := make(chan event.ExecutionEvent)
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
		logger.Errorf(true, "Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	r, err := proto.Marshal(gauge.ConvertToProtoSuiteResult(res))
	if err != nil {
		logger.Errorf(true, "Unable to marshal suite execution result, skipping save. %s", err.Error())
	}
	err = ioutil.WriteFile(resultFile, r, common.NewFilePermissions)
	if err != nil {
		logger.Errorf(true, "Failed to write to %s. Reason: %s", resultFile, err.Error())
	} else {
		logger.Debugf(true, "Last run result saved to %s", resultFile)
	}
}
