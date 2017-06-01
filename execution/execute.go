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

	"fmt"

	"strings"

	"os"

	"sync"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/validation"
)

var NumberOfExecutionStreams int
var InParallel bool

type suiteExecutor interface {
	run() *result.SuiteResult
}

type executor interface {
	execute(i gauge.Item, r result.Result)
}

type executionInfo struct {
	manifest        *manifest.Manifest
	specs           *gauge.SpecCollection
	runner          runner.Runner
	pluginHandler   plugin.Handler
	errMaps         *gauge.BuildErrors
	inParallel      bool
	numberOfStreams int
	stream          int
}

func newExecutionInfo(s *gauge.SpecCollection, r runner.Runner, ph plugin.Handler, e *gauge.BuildErrors, p bool, stream int) *executionInfo {
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
	err := validateFlags()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	if config.CheckUpdates() {
		i := &install.UpdateFacade{}
		i.BufferUpdateDetails()
		defer i.PrintUpdateBuffer()
	}

	res := validation.ValidateSpecs(specDirs, false)
	if len(res.Errs) > 0 {
		return 1
	}
	if res.SpecCollection.Size() < 1 {
		logger.Info("No specifications found in %s.", strings.Join(specDirs, ", "))
		res.Runner.Kill()
		if res.ParseOk {
			return 0
		}
		return 1
	}
	event.InitRegistry()
	wg := &sync.WaitGroup{}
	reporter.ListenExecutionEvents(wg)
	rerun.ListenFailedScenarios(wg)
	if util.ConvertToBool(os.Getenv(env.SaveExecutionResult), env.SaveExecutionResult, false) {
		ListenSuiteEndAndSaveResult(wg)
	}
	defer wg.Wait()
	ei := newExecutionInfo(res.SpecCollection, res.Runner, nil, res.ErrMap, InParallel, 0)
	e := newExecution(ei)
	return printExecutionStatus(e.run(), res.ParseOk)
}

func Execute(s *gauge.SpecCollection, r runner.Runner, ph plugin.Handler, e *gauge.BuildErrors, p bool, n int) {
	newExecution(newExecutionInfo(s, r, ph, e, p, n)).run()
}

func newExecution(executionInfo *executionInfo) suiteExecutor {
	if executionInfo.inParallel {
		return newParallelExecution(executionInfo)
	}
	return newSimpleExecution(executionInfo, true)
}

func printExecutionStatus(suiteResult *result.SuiteResult, isParsingOk bool) int {
	nSkippedSpecs := suiteResult.SpecsSkippedCount
	var nExecutedSpecs int
	if len(suiteResult.SpecResults) != 0 {
		nExecutedSpecs = len(suiteResult.SpecResults) - nSkippedSpecs
	}
	nFailedSpecs := suiteResult.SpecsFailedCount
	nPassedSpecs := nExecutedSpecs - nFailedSpecs

	nExecutedScenarios := 0
	nFailedScenarios := 0
	nPassedScenarios := 0
	nSkippedScenarios := 0
	for _, specResult := range suiteResult.SpecResults {
		nExecutedScenarios += specResult.ScenarioCount
		nFailedScenarios += specResult.ScenarioFailedCount
		nSkippedScenarios += specResult.ScenarioSkippedCount
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

	if suiteResult.IsFailed || !isParsingOk {
		return 1
	}
	return 0
}

func validateFlags() error {
	if !InParallel {
		return nil
	}
	if NumberOfExecutionStreams < 1 {
		return fmt.Errorf("Invalid input(%s) to --n flag.", strconv.Itoa(NumberOfExecutionStreams))
	}
	if !isValidStrategy(Strategy) {
		return fmt.Errorf("Invalid input(%s) to --strategy flag.", Strategy)
	}
	return nil
}
