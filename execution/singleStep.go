package execution

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/skel"
)

type ExecutionStatus struct {
	Args        []string `json:"Args"`
	FailedItems []string `json:"FailedItems"`
}

type StepLocation struct {
	SpecFile   string
	LineNumber int
}

type singleStepExecutor interface {
	run() *result.StepResult
}

const (
	failedFile = "failures.json"
)

func readFailedSpecLocations() ([]string, error) {
	projectRoot, _ := common.GetProjectRoot()
	failedStatusFile := filepath.Join(projectRoot, common.DotGauge, failedFile)

	content, err := os.ReadFile(failedStatusFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("failed tests file not found at %s", err)
		}
		return nil, fmt.Errorf("failed to read %s: %w", failedStatusFile, err)
	}

	var status ExecutionStatus
	if err := json.Unmarshal(content, &status); err != nil {
		return nil, fmt.Errorf("failed to parse failed status file: %w", err)
	}

	if len(status.FailedItems) == 0 {
		return nil, fmt.Errorf("no failed items found in failures.json")
	}

	return status.FailedItems, nil
}

func parseSpecLocation(location string) (string, int, error) {
	parts := strings.Split(location, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid format. Expected format: filepath:line")
	}

	specFile := parts[0]
	lineNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid line number: %s", parts[1])
	}

	return specFile, lineNum, nil
}

func GetFailedStepLocations() ([]StepLocation, error) {
	failedItems, err := readFailedSpecLocations()
	if err != nil {
		return nil, fmt.Errorf("failed to read failed items: %w", err)
	}

	var locations []StepLocation
	for _, item := range failedItems {
		specFile, lineNum, err := parseSpecLocation(item)
		if err != nil {
			return nil, fmt.Errorf("failed to parse location %s: %w", item, err)
		}
		locations = append(locations, StepLocation{
			SpecFile:   specFile,
			LineNumber: lineNum,
		})
	}

	return locations, nil
}

func startAPI(debug bool) runner.Runner {
	sc := api.StartAPI(debug)
	select {
	case runner := <-sc.RunnerChan:
		return runner
	case err := <-sc.ErrorChan:
		logger.Fatalf(true, "Failed to start gauge API: %s", err.Error())
	}
	return nil
}

var ExecuteStep = func(step *gauge.Step) int {

	startTime := time.Now()
	failed := 0
	skipped := 0

	if err := validateFlags(); err != nil {
		logger.Fatal(true, err.Error())
	}

	if config.CheckUpdates() {
		i := &install.UpdateFacade{}
		i.BufferUpdateDetails()
		defer i.PrintUpdateBuffer()
	}

	skel.SetupPlugins(MachineReadable)
	if err := os.Setenv(gaugeParallelStreamCountEnv, strconv.Itoa(NumberOfExecutionStreams)); err != nil {
		logger.Fatalf(true, "failed to set env %s. %s", gaugeParallelStreamCountEnv, err.Error())
	}

	r := startAPI(false)
	if r == nil {
		return ExecutionFailed
	}
	defer func() {
		err := r.Kill()
		if err != nil {
			logger.Errorf(false, "unable to kill runner: %s", err.Error())
		}
	}()

	_, res, err := parser.ParseConcepts()
	if err != nil {
		logger.Fatalf(true, "Unable to parse concepts: %s", err.Error())
		return ExecutionFailed
	}

	if !res.Ok {
		return ExecutionFailed
	}

	event.InitRegistry()
	wg := &sync.WaitGroup{}
	reporter.ListenExecutionEvents(wg)
	if env.SaveExecutionResult() {
		ListenSuiteEndAndSaveResult(wg)
	}
	defer wg.Wait()

	e := &stepExecutor{
		runner:               r,
		stream:               0,
		currentExecutionInfo: &gauge_messages.ExecutionInfo{},
	}

	protoStep := &gauge_messages.ProtoStep{}
	stepResult := ExecuteSingleStep(e, step, protoStep)
	if stepResult.GetFailed() {
		failed++
	} else if stepResult.GetSkippedScenario() {
		skipped++
	}

	// attach an executor
	executionTime := time.Since(startTime).Milliseconds()

	logger.Infof(true, "\nTotal time taken: %s", time.Millisecond*time.Duration(executionTime))

	if failed > 0 {
		return ExecutionFailed
	}
	return Success
}

func ExecuteSingleStep(e *stepExecutor, step *gauge.Step, protoStep *gauge_messages.ProtoStep) *result.StepResult {
	stepRequest := e.createStepRequest(protoStep)
	e.currentExecutionInfo.CurrentStep = &gauge_messages.StepInfo{Step: stepRequest, IsFailed: false}
	stepResult := result.NewStepResult(protoStep)
	for i := range step.GetFragments() {
		stepFragmet := step.GetFragments()[i]
		protoStepFragmet := protoStep.GetFragments()[i]
		if stepFragmet.FragmentType == gauge_messages.Fragment_Parameter && stepFragmet.Parameter.ParameterType == gauge_messages.Parameter_Dynamic {
			stepFragmet.GetParameter().Value = protoStepFragmet.GetParameter().Value
		}
	}
	event.Notify(event.NewExecutionEvent(event.StepStart, step, nil, e.stream, e.currentExecutionInfo))

	e.notifyBeforeStepHook(stepResult)
	if !stepResult.GetFailed() {
		executeStepMessage := &gauge_messages.Message{MessageType: gauge_messages.Message_ExecuteStep, ExecuteStepRequest: stepRequest}
		stepExecutionStatus := e.runner.ExecuteAndGetStatus(executeStepMessage)
		stepExecutionStatus.Message = append(stepResult.ProtoStepExecResult().GetExecutionResult().Message, stepExecutionStatus.Message...)
		if stepExecutionStatus.GetFailed() {
			e.currentExecutionInfo.CurrentStep.ErrorMessage = stepExecutionStatus.GetErrorMessage()
			e.currentExecutionInfo.CurrentStep.StackTrace = stepExecutionStatus.GetStackTrace()
			setStepFailure(e.currentExecutionInfo)
			stepResult.SetStepFailure()
		} else if stepResult.GetSkippedScenario() {
			e.currentExecutionInfo.CurrentStep.ErrorMessage = stepExecutionStatus.GetErrorMessage()
			e.currentExecutionInfo.CurrentStep.StackTrace = stepExecutionStatus.GetStackTrace()
		}
		stepResult.SetProtoExecResult(stepExecutionStatus)
	}
	e.notifyAfterStepHook(stepResult)

	event.Notify(event.NewExecutionEvent(event.StepEnd, *step, stepResult, e.stream, e.currentExecutionInfo))
	defer e.currentExecutionInfo.CurrentStep.Reset()
	return stepResult
}
