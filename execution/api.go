package execution

import (
	"log"
	"net"

	"fmt"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
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
	gm.RegisterExecutionServer(s, &executionServer{})
	go s.Serve(listener)
}

type executionServer struct {
}

func (e *executionServer) Execute(req *gm.ExecutionRequest, stream gm.Execution_ExecuteServer) error {
	execute(req.Specs, stream)
	return nil
}

func execute(specDirs []string, stream gm.Execution_ExecuteServer) {
	err := validateFlags()
	if err != nil {
		stream.Send(getErrorExecutionResponse(err))
		return
	}
	res := validation.ValidateSpecs(specDirs)
	if len(res.Errs) > 0 {
		stream.Send(getErrorExecutionResponse(res.Errs...))
		return
	}
	event.InitRegistry()
	listenExecutionEvents(stream)
	rerun.ListenFailedScenarios()
	ei := newExecutionInfo(res.SpecCollection, res.Runner, nil, res.ErrMap, InParallel, 0)
	e := newExecution(ei)
	e.run()
}

func listenExecutionEvents(stream gm.Execution_ExecuteServer) {
	ch := make(chan event.ExecutionEvent, 0)
	event.Register(ch, event.SuiteStart, event.SpecStart, event.SpecEnd, event.ScenarioStart, event.ScenarioEnd, event.SuiteEnd)
	go func() {
		for {
			var err error
			e := <-ch
			switch e.Topic {
			case event.SuiteStart:
				err = stream.Send(&gm.ExecutionResponse{Type: gm.ExecutionResponse_SuiteStart})
			case event.SpecStart:
				err = stream.Send(&gm.ExecutionResponse{
					Type: gm.ExecutionResponse_SpecStart,
					ID:   fmt.Sprintf(e.ExecutionInfo.CurrentSpec.GetFileName()),
				})
			case event.ScenarioStart:
				err = stream.Send(&gm.ExecutionResponse{
					Type: gm.ExecutionResponse_ScenarioStart,
					ID:   fmt.Sprintf("%s:%d", e.ExecutionInfo.CurrentSpec.GetFileName(), e.Item.(*gauge.Scenario).Heading.LineNo),
					Result: &gm.Result{
						TableRowNumber: int64(getDataTableRowNumber(e.Item.(*gauge.Scenario))),
					},
				})
			case event.ScenarioEnd:
				scn := e.Item.(*gauge.Scenario)
				err = stream.Send(&gm.ExecutionResponse{
					Type: gm.ExecutionResponse_ScenarioEnd,
					ID:   fmt.Sprintf("%s:%d", e.ExecutionInfo.CurrentSpec.GetFileName(), scn.Heading.LineNo),
					Result: &gm.Result{
						Status:            getStatus(e.Result.(*result.ScenarioResult)),
						ExecutionTime:     e.Result.ExecTime(),
						Errors:            getErrors(e.Result.(*result.ScenarioResult).ProtoScenario.GetScenarioItems()),
						BeforeHookFailure: getHookFailure(e.Result.GetPreHook()),
						AfterHookFailure:  getHookFailure(e.Result.GetPostHook()),
						TableRowNumber:    int64(getDataTableRowNumber(scn)),
					},
				})
			case event.SpecEnd:
				err = stream.Send(&gm.ExecutionResponse{
					Type: gm.ExecutionResponse_SpecEnd,
					ID:   fmt.Sprintf(e.ExecutionInfo.CurrentSpec.GetFileName()),
					Result: &gm.Result{
						BeforeHookFailure: getHookFailure(e.Result.GetPreHook()),
						AfterHookFailure:  getHookFailure(e.Result.GetPostHook()),
					},
				})
			case event.SuiteEnd:
				err = stream.Send(&gm.ExecutionResponse{
					Type: gm.ExecutionResponse_SuiteEnd,
					Result: &gm.Result{
						BeforeHookFailure: getHookFailure(e.Result.GetPreHook()),
						AfterHookFailure:  getHookFailure(e.Result.GetPostHook()),
					},
				})
				return
			}
			if err != nil {
				return
			}
		}
	}()
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

func getHookFailure(hookFailure **gm.ProtoHookFailure) *gm.Result_ExecutionError {
	if hookFailure != nil && *hookFailure != nil {
		return &gm.Result_ExecutionError{
			Screenshot:   (**hookFailure).GetScreenShot(),
			ErrorMessage: (**hookFailure).GetErrorMessage(),
			StackTrace:   (**hookFailure).GetStackTrace(),
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
				errors = append(errors, &gm.Result_ExecutionError{ErrorMessage: executionResult.GetSkippedReason()})
			} else if res.GetFailed() {
				errors = append(errors, &gm.Result_ExecutionError{
					StackTrace:   res.GetStackTrace(),
					ErrorMessage: res.GetErrorMessage(),
					Screenshot:   res.GetScreenShot(),
				})
			}
		case gm.ProtoItem_Concept:
			errors = append(errors, getErrors(item.GetConcept().GetSteps())...)
		}
	}
	return errors
}

func getStatus(result *result.ScenarioResult) gm.Result_Status {
	if result.GetFailed() {
		return gm.Result_FAILED
	}
	if result.ProtoScenario.GetSkipped() {
		return gm.Result_SKIPPED
	}
	return gm.Result_PASSED
}
