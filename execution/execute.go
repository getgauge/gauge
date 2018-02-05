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

/*
   Execution can be of two types
	- Simple execution
	- Paralell execution

   Execution Flow :
   	- Checks for updates
    	- Validation
    	- Init Registry
    	- Saving Execution result

   Strategy
    	- Lazy : Lazy is a parallelization strategy for execution. In this case tests assignment will be dynamic during execution, i.e. assign the next spec in line to the stream that has completed it’s previous execution and is waiting for more work.
    	- Eager : Eager is a parallelization strategy for execution. In this case tests are distributed before execution, thus making them an equal number based distribution.
*/
package execution

import (
	"strconv"
	"time"

	"github.com/getgauge/gauge/skel"

	"fmt"

	"strings"

	"os"

	"sync"

	"runtime/debug"

	"encoding/json"
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
	"io/ioutil"
	"path/filepath"
	"github.com/getgauge/common"
)

const (
	executionStatusFile = "executionStatus.json"
)

// NumberOfExecutionStreams shows the number of execution streams, in parallel execution.
var NumberOfExecutionStreams int

// InParallel if true executes the specs in parallel else in serial.
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

// ExecuteSpecs : Check for updates, validates the specs (by invoking the respective language runners), initiates the registry which is needed for console reporting, execution API and Rerunning of specs
// and finally saves the execution result as binary in .gauge folder.
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
	skel.SetupPlugins()
	res := validation.ValidateSpecs(specDirs, false)
	if len(res.Errs) > 0 {
		return 1
	}
	if res.SpecCollection.Size() < 1 {
		logger.Infof("No specifications found in %s.", strings.Join(specDirs, ", "))
		res.Runner.Kill()
		if res.ParseOk {
			return 0
		}
		return 1
	}
	event.InitRegistry()
	wg := &sync.WaitGroup{}
	reporter.ListenExecutionEvents(wg)
	rerun.ListenFailedScenarios(wg, specDirs)
	if util.ConvertToBool(os.Getenv(env.SaveExecutionResult), env.SaveExecutionResult, false) {
		ListenSuiteEndAndSaveResult(wg)
	}
	defer wg.Wait()
	defer recoverPanic()
	ei := newExecutionInfo(res.SpecCollection, res.Runner, nil, res.ErrMap, InParallel, 0)
	e := newExecution(ei)
	return printExecutionStatus(e.run(), res.ParseOk)
}

func recoverPanic() {
	if r := recover(); r != nil {
		logger.Infof("%v\n%s", r, string(debug.Stack()))
		os.Exit(1)
	}
}

func newExecution(executionInfo *executionInfo) suiteExecutor {
	if executionInfo.inParallel {
		return newParallelExecution(executionInfo)
	}
	return newSimpleExecution(executionInfo, true)
}

type executionStatus struct {
	Executed int
	Passed   int
	Failed   int
	Skipped  int
}

func newExecutionStatus() *executionStatus {
	return &executionStatus{Executed: 0, Passed: 0, Failed: 0, Skipped: 0}
}

func (status *executionStatus) getJSON() (string, error) {
	j, err := json.MarshalIndent(status, "", "\t")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func writeExecutionStatus(executedScenarios int, passedScenarios int, failedScenarios int, skippedScenarios int) {
	executionStatus := newExecutionStatus()
	executionStatus.Executed = executedScenarios
	executionStatus.Passed = passedScenarios
	executionStatus.Failed = failedScenarios
	executionStatus.Skipped = skippedScenarios
	contents, err := executionStatus.getJSON()
	if err != nil {
		logger.Fatalf("Unable to parse execution status information : %v", err.Error())
	}
	executionStatusFile := filepath.Join(config.ProjectRoot, common.DotGauge, executionStatusFile)
	dotGaugeDir := filepath.Join(config.ProjectRoot, common.DotGauge)
	if err = os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf("Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err = ioutil.WriteFile(executionStatusFile, []byte(contents), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf("Failed to write to %s. Reason: %s", executionStatusFile, err.Error())
	}
}

func ReadExecutionStatus() (interface{}, error){
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, common.DotGauge, executionStatusFile))
	if err != nil {
		logger.Fatalf("Failed to read execution status information. Reason: %s", err.Error())
	}
	meta := newExecutionStatus()
	if err = json.Unmarshal([]byte(contents), meta); err != nil {
		logger.Fatalf("Invalid execution status information. Reason: %s", err.Error())
		return meta, err
	}
	return meta, nil
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

	logger.Infof("Specifications:\t%d executed\t%d passed\t%d failed\t%d skipped", nExecutedSpecs, nPassedSpecs, nFailedSpecs, nSkippedSpecs)
	logger.Infof("Scenarios:\t%d executed\t%d passed\t%d failed\t%d skipped", nExecutedScenarios, nPassedScenarios, nFailedScenarios, nSkippedScenarios)
	logger.Infof("\nTotal time taken: %s", time.Millisecond*time.Duration(suiteResult.ExecutionTime))

	writeExecutionStatus(nExecutedScenarios, nPassedScenarios, nFailedScenarios, nSkippedScenarios)

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
