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
	"github.com/golang/protobuf/proto"
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
	execute(req.GetSpecs(), stream)
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
			e := <-ch
			switch e.Topic {
			case event.SuiteStart:
				stream.Send(&gm.ExecutionResponse{Type: gm.ExecutionResponse_SuiteStart.Enum()})
			case event.SpecStart:
				stream.Send(&gm.ExecutionResponse{
					Type: gm.ExecutionResponse_SpecStart.Enum(),
					ID:   proto.String(fmt.Sprintf(e.ExecutionInfo.CurrentSpec.GetFileName())),
				})
			case event.ScenarioStart:
				stream.Send(&gm.ExecutionResponse{
					Type: gm.ExecutionResponse_ScenarioStart.Enum(),
					ID:   proto.String(fmt.Sprintf("%s:%d", e.ExecutionInfo.CurrentSpec.GetFileName(), e.Item.(*gauge.Scenario).Heading.LineNo)),
				})
			case event.ScenarioEnd:
				stream.Send(&gm.ExecutionResponse{
					Type:              gm.ExecutionResponse_ScenarioEnd.Enum(),
					ID:                proto.String(fmt.Sprintf("%s:%d", e.ExecutionInfo.CurrentSpec.GetFileName(), e.Item.(*gauge.Scenario).Heading.LineNo)),
					Status:            getStatus(e.Result.(*result.ScenarioResult)),
					ExecutionTime:     proto.Int64(e.Result.ExecTime()),
					Errors:            getErrors(e.Result.(*result.ScenarioResult).ProtoScenario.GetScenarioItems()),
					BeforeHookFailure: getHookFailure(e.Result.GetPreHook()),
					AfterHookFailure:  getHookFailure(e.Result.GetPostHook()),
				})
			case event.SpecEnd:
				stream.Send(&gm.ExecutionResponse{
					Type:              gm.ExecutionResponse_SpecEnd.Enum(),
					ID:                proto.String(fmt.Sprintf(e.ExecutionInfo.CurrentSpec.GetFileName())),
					BeforeHookFailure: getHookFailure(e.Result.GetPreHook()),
					AfterHookFailure:  getHookFailure(e.Result.GetPostHook()),
				})
			case event.SuiteEnd:
				stream.Send(&gm.ExecutionResponse{
					Type:              gm.ExecutionResponse_SuiteEnd.Enum(),
					BeforeHookFailure: getHookFailure(e.Result.GetPreHook()),
					AfterHookFailure:  getHookFailure(e.Result.GetPostHook()),
				})
				return
			}
		}
	}()
}

func getErrorExecutionResponse(errs ...error) *gm.ExecutionResponse {
	var e []*gm.ExecutionResponse_ExecutionError
	for _, err := range errs {
		e = append(e, &gm.ExecutionResponse_ExecutionError{ErrorMessage: proto.String(err.Error())})
	}
	return &gm.ExecutionResponse{Type: gm.ExecutionResponse_ErrorResult.Enum(), Errors: e}
}

func getHookFailure(hookFailure **gm.ProtoHookFailure) *gm.ExecutionResponse_ExecutionError {
	if hookFailure != nil && *hookFailure != nil {
		return &gm.ExecutionResponse_ExecutionError{
			Screenshot:   (**hookFailure).ScreenShot,
			ErrorMessage: (**hookFailure).ErrorMessage,
			StackTrace:   (**hookFailure).StackTrace,
		}
	}
	return nil
}

func getErrors(items []*gm.ProtoItem) []*gm.ExecutionResponse_ExecutionError {
	var errors []*gm.ExecutionResponse_ExecutionError
	for _, item := range items {
		executionResult := item.GetStep().GetStepExecutionResult()
		res := executionResult.GetExecutionResult()
		switch item.GetItemType() {
		case gm.ProtoItem_Step:
			if executionResult.GetSkipped() {
				errors = append(errors, &gm.ExecutionResponse_ExecutionError{ErrorMessage: executionResult.SkippedReason})
			} else if res.GetFailed() {
				errors = append(errors, &gm.ExecutionResponse_ExecutionError{
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

func getStatus(result *result.ScenarioResult) *gm.ExecutionResponse_Status {
	if result.GetFailed() {
		return gm.ExecutionResponse_FAILED.Enum()
	}
	if result.ProtoScenario.GetSkipped() {
		return gm.ExecutionResponse_SKIPPED.Enum()
	}
	return gm.ExecutionResponse_PASSED.Enum()
}
