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

package run_failed

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/logger"
)

const (
	dotGauge   = ".gauge"
	failedFile = "failed.json"
)

var Environment string
var Tags string
var TableRows string
var SimpleConsole bool
var Verbose bool

type FailedMetadata struct {
	Env             string
	Tags            string
	TableRows       string
	Verbose         bool
	SimpleConsole   bool
	FailedScenarios []string
}

func (m *FailedMetadata) AddFailedScenario(sce string) {
	m.FailedScenarios = append(m.FailedScenarios, sce)
}

func ListenFailedScenarios() {
	ch := make(chan event.ExecutionEvent, 0)
	event.Register(ch, event.SuiteEnd)

	go func() {
		for {
			e := <-ch
			switch e.Topic {
			case event.SuiteEnd:
				failedMeta := getFailedMetadata(e.Result.(*result.SuiteResult).SpecResults)
				writeFailedMeta(getJSON(failedMeta))
			}
		}
	}()
}

func getFailedMetadata(specResults []*result.SpecResult) *FailedMetadata {
	failedMeta := &FailedMetadata{Env: Environment, Tags: Tags, TableRows: TableRows, Verbose: Verbose, SimpleConsole: SimpleConsole, FailedScenarios: []string{}}
	for _, specRes := range specResults {
		if specRes.GetFailed() {
			specPath := *specRes.ProtoSpec.FileName
			failedScenario := strings.TrimPrefix(specPath, config.ProjectRoot+string(filepath.Separator))
			for _, i := range specRes.FailedScenarioIndices {
				failedMeta.AddFailedScenario(fmt.Sprintf("%s:%v", failedScenario, i))
			}
		}
	}
	return failedMeta
}

func writeFailedMeta(contents string) {
	failedPath := filepath.Join(config.ProjectRoot, dotGauge, failedFile)
	dotGaugeDir := filepath.Join(config.ProjectRoot, dotGauge)
	if err := os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf("Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err := ioutil.WriteFile(failedPath, []byte(contents), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf("Failed to write to %s. Reason: %s", failedPath, err.Error())
	}
}

func getJSON(failedMeta *FailedMetadata) string {
	json, err := json.MarshalIndent(failedMeta, "", "\t")
	if err != nil {
		logger.Fatalf("Failed to read last run information. Reason: %s", err.Error())
	}
	return string(json)
}
