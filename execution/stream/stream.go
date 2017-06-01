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

package stream

import (
	"log"
	"net"

	"fmt"

	"strings"

	"sync"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/order"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/validation"
	"google.golang.org/grpc"
)

func Start() {
	port, err := conn.GetPortFromEnvironmentVariable(common.APIV2PortEnvVariableName)
	if err != nil {
		logger.APILog.Error("Failed to start execution API Service. %s \n", err.Error())
		return
	}
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	gm.RegisterExecutionServer(s, newExecutionServer())
	go s.Serve(listener)
}

type executionServer struct {
	wg *sync.WaitGroup
}

func newExecutionServer() *executionServer {
	return &executionServer{wg: &sync.WaitGroup{}}
}

func (e *executionServer) Execute(req *gm.ExecutionRequest, stream gm.Execution_ExecuteServer) error {
	errs := setFlags(req)
	if len(errs) > 0 {
		stream.Send(getErrorExecutionResponse(errs...))
		return nil
	}
	execute(req.Specs, stream, req.GetDebug(), e.wg)
	defer e.wg.Wait()
	return nil
}

func execute(specDirs []string, stream gm.Execution_ExecuteServer, debug bool, wg *sync.WaitGroup) {
	res := validation.ValidateSpecs(specDirs, debug)
	if len(res.Errs) > 0 {
		stream.Send(getErrorExecutionResponse(res.Errs...))
		return
	}
	event.InitRegistry()

	listenExecutionEvents(stream, res.Runner.Pid(), wg)
	rerun.ListenFailedScenarios(wg)
	execution.Execute(res.SpecCollection, res.Runner, nil, res.ErrMap, execution.InParallel, 0)
}

func listenExecutionEvents(stream gm.Execution_ExecuteServer, pid int, wg *sync.WaitGroup) {
	ch := make(chan event.ExecutionEvent, 0)
	event.Register(ch, event.SuiteStart, event.SpecStart, event.SpecEnd, event.ScenarioStart, event.ScenarioEnd, event.SuiteEnd)
	go func() {
		wg.Add(1)
		defer wg.Done()
		for {
			e := <-ch
			res := getResponse(e, pid)
			if stream.Send(res) != nil || res.Type == gm.ExecutionResponse_SuiteEnd {
				util.SetWorkingDir(config.ProjectRoot)
				return
			}
		}
	}()
}

func getResponse(e event.ExecutionEvent, pid int) *gm.ExecutionResponse {
	switch e.Topic {
	case event.SuiteStart:
		return &gm.ExecutionResponse{Type: gm.ExecutionResponse_SuiteStart, RunnerProcessId: int32(pid)}
	case event.SpecStart:
		return &gm.ExecutionResponse{
			Type:            gm.ExecutionResponse_SpecStart,
			ID:              e.ExecutionInfo.CurrentSpec.FileName,
			RunnerProcessId: int32(pid),
		}
	case event.ScenarioStart:
		return &gm.ExecutionResponse{
			Type:            gm.ExecutionResponse_ScenarioStart,
			ID:              fmt.Sprintf("%s:%d", e.ExecutionInfo.CurrentSpec.GetFileName(), e.Item.(*gauge.Scenario).Heading.LineNo),
			RunnerProcessId: int32(pid),
			Result: &gm.Result{
				TableRowNumber: int64(getDataTableRowNumber(e.Item.(*gauge.Scenario))),
			},
		}
	case event.ScenarioEnd:
		scn := e.Item.(*gauge.Scenario)
		return &gm.ExecutionResponse{
			Type:            gm.ExecutionResponse_ScenarioEnd,
			ID:              fmt.Sprintf("%s:%d", e.ExecutionInfo.CurrentSpec.GetFileName(), scn.Heading.LineNo),
			RunnerProcessId: int32(pid),
			Result: &gm.Result{
				Status:            getStatus(e.Result.(*result.ScenarioResult)),
				ExecutionTime:     e.Result.ExecTime(),
				Errors:            getErrors(e.Result.(*result.ScenarioResult).ProtoScenario.GetScenarioItems()),
				BeforeHookFailure: getHookFailure(e.Result.GetPreHook()),
				AfterHookFailure:  getHookFailure(e.Result.GetPostHook()),
				TableRowNumber:    int64(getDataTableRowNumber(scn)),
			},
		}
	case event.SpecEnd:
		return &gm.ExecutionResponse{
			Type:            gm.ExecutionResponse_SpecEnd,
			ID:              e.ExecutionInfo.CurrentSpec.FileName,
			RunnerProcessId: int32(pid),
			Result: &gm.Result{
				BeforeHookFailure: getHookFailure(e.Result.GetPreHook()),
				AfterHookFailure:  getHookFailure(e.Result.GetPostHook()),
			},
		}
	case event.SuiteEnd:
		return &gm.ExecutionResponse{
			Type:            gm.ExecutionResponse_SuiteEnd,
			RunnerProcessId: int32(pid),
			Result: &gm.Result{
				BeforeHookFailure: getHookFailure(e.Result.GetPreHook()),
				AfterHookFailure:  getHookFailure(e.Result.GetPostHook()),
			},
		}
	}
	return nil
}

func getDataTableRowNumber(scn *gauge.Scenario) int {
	index := scn.DataTableRowIndex
	if scn.DataTableRow.IsInitialized() {
		index++
	}
	return index
}

func getErrorExecutionResponse(errs ...error) *gm.ExecutionResponse {
	var e []*gm.Result_ExecutionError
	for _, err := range errs {
		e = append(e, &gm.Result_ExecutionError{ErrorMessage: err.Error()})
	}
	return &gm.ExecutionResponse{
		Type: gm.ExecutionResponse_ErrorResult,
		Result: &gm.Result{
			Errors: e,
		},
	}
}

func getHookFailure(hookFailure []*gm.ProtoHookFailure) *gm.Result_ExecutionError {
	if len(hookFailure) > 0 {
		return &gm.Result_ExecutionError{
			Screenshot:   hookFailure[0].ScreenShot,
			ErrorMessage: hookFailure[0].ErrorMessage,
			StackTrace:   hookFailure[0].StackTrace,
		}
	}
	return nil
}

func getErrors(items []*gm.ProtoItem) []*gm.Result_ExecutionError {
	var errors []*gm.Result_ExecutionError
	for _, item := range items {
		executionResult := item.GetStep().GetStepExecutionResult()
		res := executionResult.GetExecutionResult()
		switch item.GetItemType() {
		case gm.ProtoItem_Step:
			if executionResult.GetSkipped() {
				errors = append(errors, &gm.Result_ExecutionError{ErrorMessage: executionResult.SkippedReason})
			} else if res.GetFailed() {
				errors = append(errors, &gm.Result_ExecutionError{
					StackTrace:   res.StackTrace,
					ErrorMessage: res.ErrorMessage,
					Screenshot:   res.ScreenShot,
				})
			}
		case gm.ProtoItem_Concept:
			errors = append(errors, getErrors(item.GetConcept().GetSteps())...)
		}
	}
	return errors
}

func getStatus(result *result.ScenarioResult) gm.Result_Status {
	if result.ProtoScenario.GetExecutionStatus() == gm.ExecutionStatus_FAILED {
		return gm.Result_FAILED
	}
	if result.ProtoScenario.GetExecutionStatus() == gm.ExecutionStatus_SKIPPED {
		return gm.Result_SKIPPED
	}
	return gm.Result_PASSED
}

func setFlags(req *gm.ExecutionRequest) []error {
	resetFlags()
	reporter.IsParallel = req.GetIsParallel()
	execution.InParallel = req.GetIsParallel()
	filter.ExecuteTags = req.GetTags()
	execution.SetTableRows(req.GetTableRows())
	streams := int(req.GetParallelStreams())
	if streams < 1 {
		streams = util.NumberOfCores()
	}
	execution.NumberOfExecutionStreams = streams
	reporter.NumberOfExecutionStreams = streams
	filter.NumberOfExecutionStreams = streams
	execution.Strategy = strings.ToLower(req.GetStrategy().String())
	order.Sorted = req.GetSort()
	reporter.Verbose = true
	logger.Initialize(strings.ToLower(req.GetLogLevel().String()))
	if req.GetWorkingDir() != "" {
		util.SetWorkingDir(req.GetWorkingDir())
	}
	var errs []error
	if req.GetEnv() != "" {
		if err := env.LoadEnv(req.GetEnv()); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func resetFlags() {
	cores := util.NumberOfCores()
	reporter.IsParallel = false
	execution.InParallel = false
	reporter.Verbose = false
	filter.ExecuteTags = ""
	execution.SetTableRows("")
	execution.NumberOfExecutionStreams = cores
	reporter.NumberOfExecutionStreams = cores
	filter.NumberOfExecutionStreams = cores
	execution.Strategy = "lazy"
	order.Sorted = false
	util.SetWorkingDir(config.ProjectRoot)
}
