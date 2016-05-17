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
	"strconv"
	"time"

	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/validation"
)

var NumberOfExecutionStreams int
var InParallel bool

type execution interface {
	run() *result.SuiteResult
}

type executionInfo struct {
	manifest        *manifest.Manifest
	specs           *gauge.SpecCollection
	runner          runner.Runner
	pluginHandler   *plugin.Handler
	consoleReporter reporter.Reporter
	errMaps         *validation.ValidationErrMaps
	inParallel      bool
	numberOfStreams int
	stream          int
}

func newExecutionInfo(s *gauge.SpecCollection, r runner.Runner, ph *plugin.Handler, e *validation.ValidationErrMaps, p bool, stream int) *executionInfo {
	m, err := manifest.ProjectManifest()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	return &executionInfo{
		manifest:        m,
		specs:           s,
		runner:          r,
		pluginHandler:   ph,
		errMaps:         e,
		inParallel:      p,
		numberOfStreams: NumberOfExecutionStreams,
		stream:          stream,
	}
}

func ExecuteSpecs(specDirs []string) int {
	validateFlags()
	if config.CheckUpdates() {
		i := &install.UpdateFacade{}
		i.BufferUpdateDetails()
		defer i.PrintUpdateBuffer()
	}

	runner := startAPI()
	specs, errMap := validation.ValidateSpecs(specDirs, runner)
	event.InitRegistry()
	reporter.ListenExecutionEvents()
	ei := newExecutionInfo(specs, runner, nil, errMap, InParallel, 0)
	e := newExecution(ei)
	return printExecutionStatus(e.run(), errMap)
}

func newExecution(executionInfo *executionInfo) execution {
	if executionInfo.inParallel {
		return newParallelExecution(executionInfo)
	}
	return newSimpleExecution(executionInfo)
}

func startAPI() runner.Runner {
	sc := api.StartAPI()
	select {
	case runner := <-sc.RunnerChan:
		return runner
	case err := <-sc.ErrorChan:
		logger.Fatalf("Failed to start gauge API: %s", err.Error())
	}
	return nil
}

func printExecutionStatus(suiteResult *result.SuiteResult, errMap *validation.ValidationErrMaps) int {
	nSkippedScenarios := len(errMap.ScenarioErrs)
	nSkippedSpecs := len(errMap.SpecErrs)
	nExecutedSpecs := len(suiteResult.SpecResults) - nSkippedSpecs
	nFailedSpecs := suiteResult.SpecsFailedCount
	nPassedSpecs := nExecutedSpecs - nFailedSpecs

	nExecutedScenarios := 0
	nFailedScenarios := 0
	nPassedScenarios := 0
	for _, specResult := range suiteResult.SpecResults {
		nExecutedScenarios += specResult.ScenarioCount
		nFailedScenarios += specResult.ScenarioFailedCount
	}
	nExecutedScenarios -= nSkippedScenarios
	nPassedScenarios = nExecutedScenarios - nFailedScenarios

	if nExecutedScenarios < 0 {
		nExecutedScenarios = 0
	}

	if nPassedScenarios < 0 {
		nPassedScenarios = 0
	}

	logger.Info("Specifications:\t%d executed\t%d passed\t%d failed\t%d skipped", nExecutedSpecs, nPassedSpecs, nFailedSpecs, nSkippedSpecs)
	logger.Info("Scenarios:\t%d executed\t%d passed\t%d failed\t%d skipped", nExecutedScenarios, nPassedScenarios, nFailedScenarios, nSkippedScenarios)
	logger.Info("\nTotal time taken: %s", time.Millisecond*time.Duration(suiteResult.ExecutionTime))

	if suiteResult.IsFailed || (nSkippedSpecs+nSkippedScenarios) > 0 {
		return 1
	}
	return 0
}

func validateFlags() {
	if !InParallel {
		return
	}
	if NumberOfExecutionStreams < 1 {
		logger.Fatalf("Invalid input(%s) to --n flag.", strconv.Itoa(NumberOfExecutionStreams))
	}
	if !isValidStrategy(Strategy) {
		logger.Fatalf("Invalid input(%s) to --strategy flag.", Strategy)
	}
}
