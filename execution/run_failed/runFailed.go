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
	failedFile = "failed.txt"
)

var Environment string
var Tags string
var TableRows string
var SimpleConsole bool
var Verbose bool

var failedInfo string

func appendFailedInfo(info string) {
	failedInfo += info + "\n"
}

func ListenFailedScenarios() {
	ch := make(chan event.ExecutionEvent, 0)
	event.Register(ch, event.SuiteEnd, event.SpecEnd)

	go func() {
		for {
			e := <-ch
			switch e.Topic {
			case event.SuiteEnd:
				addFailedInfo()
			case event.SpecEnd:
				addSpecToFailedInfo(e.Result)
			}
		}
	}()
}

func addFailedInfo() {
	failedPath := filepath.Join(config.ProjectRoot, dotGauge, failedFile)
	dotGaugeDir := filepath.Join(config.ProjectRoot, dotGauge)
	if err := os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf("Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	contents := prepareCmd() + failedInfo
	err := ioutil.WriteFile(failedPath, []byte(contents), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf("Failed to write to %s. Reason: %s", failedPath, err.Error())
	}
}

func addSpecToFailedInfo(res result.Result) {
	if res.GetFailed() {
		specPath := *res.(*result.SpecResult).ProtoSpec.FileName
		specRelPath := strings.TrimPrefix(specPath, config.ProjectRoot+string(filepath.Separator))
		appendFailedInfo(specRelPath)
	}
}

func prepareCmd() string {
	cmd := []string{"gauge"}

	if Environment != "default" && Environment != "" {
		cmd = append(cmd, "--env="+Environment)
	}
	if Tags != "" {
		cmd = append(cmd, "--tags="+Tags)
	}
	if TableRows != "" {
		cmd = append(cmd, "--tableRows="+TableRows)
	}
	if SimpleConsole {
		cmd = append(cmd, "--simple-console")
	}
	if Verbose {
		cmd = append(cmd, "--verbose")
	}
	return strings.Join(cmd, " ") + "\n"
}
