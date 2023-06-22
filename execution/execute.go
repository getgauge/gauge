/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

/*
Package execution handles gauge's execution of spec/scenario/steps

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

	"encoding/json"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/validation"
)

const (
	executionStatusFile = "executionStatus.json"
)

// Count of iterations
var MaxRetriesCount int

// Tags to filter specs/scenarios to retry
var RetryOnlyTags string

// NumberOfExecutionStreams shows the number of execution streams, in parallel execution.
var NumberOfExecutionStreams int

// InParallel if true executes the specs in parallel else in serial.
var InParallel bool

var TagsToFilterForParallelRun string

// Verbose if true prints additional details about the execution
var Verbose bool

// MachineReadable indicates that the output is in json format
var MachineReadable bool

var ExecutionArgs []*gauge.ExecutionArg

type suiteExecutor interface {
	run() *result.SuiteResult
}

type executor interface {
	execute(i gauge.Item, r result.Result)
}

// ExecuteSpecs : Check for updates, validates the specs (by invoking the respective language runners), initiates the registry which is needed for console reporting, execution API and Rerunning of specs
// and finally saves the execution result as binary in .gauge folder.
var ExecuteSpecs = func(specDirs []string) int {
	err := validateFlags()
	if err != nil {
		logger.Fatalf(true, err.Error())
	}
	if config.CheckUpdates() {
		i := &install.UpdateFacade{}
		i.BufferUpdateDetails()
		defer i.PrintUpdateBuffer()
	}
	skel.SetupPlugins(MachineReadable)
	err = os.Setenv(gaugeParallelStreamCountEnv, strconv.Itoa(NumberOfExecutionStreams))
	if err != nil {
		logger.Fatalf(true, "failed to set env %s. %s", gaugeParallelStreamCountEnv, err.Error())
	}

	res := validation.ValidateSpecs(specDirs, false)
	if len(res.Errs) > 0 {
		if res.ParseOk {
			return ParseFailed
		}
		return ValidationFailed
	}
	if res.SpecCollection.Size() < 1 {
		logger.Infof(true, "No specifications found in %s.", strings.Join(specDirs, ", "))
		err := res.Runner.Kill()
		if err != nil {
			logger.Errorf(false, "unable to kill runner: %s", err.Error())
		}
		if res.ParseOk {
			return Success
		}
		return ExecutionFailed
	}
	event.InitRegistry()
	wg := &sync.WaitGroup{}
	reporter.ListenExecutionEvents(wg)
	rerun.ListenFailedScenarios(wg, specDirs)
	if env.SaveExecutionResult() {
		ListenSuiteEndAndSaveResult(wg)
	}
	defer wg.Wait()
	ei := newExecutionInfo(res.SpecCollection, res.Runner, nil, res.ErrMap, InParallel, 0)

	e := ei.getExecutor()
	logger.Debug(true, "Run started")
	return printExecutionResult(e.run(), res.ParseOk)
}

func writeExecutionResult(content string) {
	executionStatusFile := filepath.Join(config.ProjectRoot, common.DotGauge, executionStatusFile)
	dotGaugeDir := filepath.Join(config.ProjectRoot, common.DotGauge)
	if err := os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf(true, "Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err := os.WriteFile(executionStatusFile, []byte(content), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf(true, "Failed to write to %s. Reason: %s", executionStatusFile, err.Error())
	}
}

// ReadLastExecutionResult returns the result of previous execution in JSON format
// This is stored in $GAUGE_PROJECT_ROOT/.gauge/executionStatus.json file after every execution
func ReadLastExecutionResult() (interface{}, error) {
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, common.DotGauge, executionStatusFile))
	if err != nil {
		logger.Fatalf(true, "Failed to read execution status information. Reason: %s", err.Error())
	}
	meta := &executionStatus{}
	if err = json.Unmarshal([]byte(contents), meta); err != nil {
		logger.Fatalf(true, "Invalid execution status information. Reason: %s", err.Error())
		return meta, err
	}
	return meta, nil
}

func printExecutionResult(suiteResult *result.SuiteResult, isParsingOk bool) int {
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

	s := statusJSON(nExecutedSpecs, nPassedSpecs, nFailedSpecs, nSkippedSpecs, nExecutedScenarios, nPassedScenarios, nFailedScenarios, nSkippedScenarios)
	logger.Infof(true, "Specifications:\t%d executed\t%d passed\t%d failed\t%d skipped", nExecutedSpecs, nPassedSpecs, nFailedSpecs, nSkippedSpecs)
	logger.Infof(true, "Scenarios:\t%d executed\t%d passed\t%d failed\t%d skipped", nExecutedScenarios, nPassedScenarios, nFailedScenarios, nSkippedScenarios)
	logger.Infof(true, "\nTotal time taken: %s", time.Millisecond*time.Duration(suiteResult.ExecutionTime))
	writeExecutionResult(s)

	if !isParsingOk {
		return ParseFailed
	}
	if suiteResult.IsFailed {
		return ExecutionFailed
	}
	return Success
}

func validateFlags() error {
	if MaxRetriesCount < 1 {
		return fmt.Errorf("invalid input(%s) to --max-retries-count flag", strconv.Itoa(MaxRetriesCount))
	}
	if !InParallel {
		return nil
	}
	if NumberOfExecutionStreams < 1 {
		return fmt.Errorf("invalid input(%s) to --n flag", strconv.Itoa(NumberOfExecutionStreams))
	}
	if !isValidStrategy(Strategy) {
		return fmt.Errorf("invalid input(%s) to --strategy flag", Strategy)
	}
	return nil
}
