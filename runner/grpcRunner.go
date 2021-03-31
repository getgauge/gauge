/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package runner

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	gm "github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	errdetails "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	host  = "127.0.0.1"
	oneGB = 1024 * 1024 * 1024
)

// GrpcRunner handles grpc messages.
type GrpcRunner struct {
	cmd          *exec.Cmd
	conn         *grpc.ClientConn
	LegacyClient gm.LspServiceClient
	RunnerClient gm.RunnerClient
	Timeout      time.Duration
	info         *RunnerInfo
	IsExecuting  bool
}

func (r *GrpcRunner) invokeLegacyLSPService(message *gm.Message) (*gm.Message, error) {
	switch message.MessageType {
	case gm.Message_CacheFileRequest:
		_, err := r.LegacyClient.CacheFile(context.Background(), message.CacheFileRequest)
		return &gm.Message{}, err
	case gm.Message_StepNamesRequest:
		response, err := r.LegacyClient.GetStepNames(context.Background(), message.StepNamesRequest)
		return &gm.Message{StepNamesResponse: response}, err
	case gm.Message_StepPositionsRequest:
		response, err := r.LegacyClient.GetStepPositions(context.Background(), message.StepPositionsRequest)
		return &gm.Message{StepPositionsResponse: response}, err
	case gm.Message_ImplementationFileListRequest:
		response, err := r.LegacyClient.GetImplementationFiles(context.Background(), &gm.Empty{})
		return &gm.Message{ImplementationFileListResponse: response}, err
	case gm.Message_StubImplementationCodeRequest:
		response, err := r.LegacyClient.ImplementStub(context.Background(), message.StubImplementationCodeRequest)
		return &gm.Message{FileDiff: response}, err
	case gm.Message_StepValidateRequest:
		response, err := r.LegacyClient.ValidateStep(context.Background(), message.StepValidateRequest)
		return &gm.Message{MessageType: gm.Message_StepValidateResponse, StepValidateResponse: response}, err
	case gm.Message_RefactorRequest:
		response, err := r.LegacyClient.Refactor(context.Background(), message.RefactorRequest)
		return &gm.Message{MessageType: gm.Message_RefactorResponse, RefactorResponse: response}, err
	case gm.Message_StepNameRequest:
		response, err := r.LegacyClient.GetStepName(context.Background(), message.StepNameRequest)
		return &gm.Message{MessageType: gm.Message_StepNameResponse, StepNameResponse: response}, err
	case gm.Message_ImplementationFileGlobPatternRequest:
		response, err := r.LegacyClient.GetGlobPatterns(context.Background(), &gm.Empty{})
		return &gm.Message{MessageType: gm.Message_ImplementationFileGlobPatternRequest, ImplementationFileGlobPatternResponse: response}, err
	case gm.Message_KillProcessRequest:
		_, err := r.LegacyClient.KillProcess(context.Background(), message.KillProcessRequest)
		return &gm.Message{}, err
	default:
		return nil, nil
	}
}

func (r *GrpcRunner) invokeServiceFor(message *gm.Message) (*gm.Message, error) {
	switch message.MessageType {
	case gm.Message_SuiteDataStoreInit:
		response, err := r.RunnerClient.InitializeSuiteDataStore(context.Background(), message.SuiteDataStoreInitRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_SpecDataStoreInit:
		response, err := r.RunnerClient.InitializeSpecDataStore(context.Background(), message.SpecDataStoreInitRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ScenarioDataStoreInit:
		response, err := r.RunnerClient.InitializeScenarioDataStore(context.Background(), message.ScenarioDataStoreInitRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ExecutionStarting:
		response, err := r.RunnerClient.StartExecution(context.Background(), message.ExecutionStartingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_SpecExecutionStarting:
		response, err := r.RunnerClient.StartSpecExecution(context.Background(), message.SpecExecutionStartingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ScenarioExecutionStarting:
		response, err := r.RunnerClient.StartScenarioExecution(context.Background(), message.ScenarioExecutionStartingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_StepExecutionStarting:
		response, err := r.RunnerClient.StartStepExecution(context.Background(), message.StepExecutionStartingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ExecuteStep:
		response, err := r.RunnerClient.ExecuteStep(context.Background(), message.ExecuteStepRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_StepExecutionEnding:
		response, err := r.RunnerClient.FinishStepExecution(context.Background(), message.StepExecutionEndingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ScenarioExecutionEnding:
		response, err := r.RunnerClient.FinishScenarioExecution(context.Background(), message.ScenarioExecutionEndingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_SpecExecutionEnding:
		response, err := r.RunnerClient.FinishSpecExecution(context.Background(), message.SpecExecutionEndingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ExecutionEnding:
		response, err := r.RunnerClient.FinishExecution(context.Background(), message.ExecutionEndingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err

	case gm.Message_CacheFileRequest:
		_, err := r.RunnerClient.CacheFile(context.Background(), message.CacheFileRequest)
		return &gm.Message{}, err
	case gm.Message_StepNamesRequest:
		response, err := r.RunnerClient.GetStepNames(context.Background(), message.StepNamesRequest)
		return &gm.Message{StepNamesResponse: response}, err
	case gm.Message_StepPositionsRequest:
		response, err := r.RunnerClient.GetStepPositions(context.Background(), message.StepPositionsRequest)
		return &gm.Message{StepPositionsResponse: response}, err
	case gm.Message_ImplementationFileListRequest:
		response, err := r.RunnerClient.GetImplementationFiles(context.Background(), &gm.Empty{})
		return &gm.Message{ImplementationFileListResponse: response}, err
	case gm.Message_StubImplementationCodeRequest:
		response, err := r.RunnerClient.ImplementStub(context.Background(), message.StubImplementationCodeRequest)
		return &gm.Message{FileDiff: response}, err
	case gm.Message_RefactorRequest:
		response, err := r.RunnerClient.Refactor(context.Background(), message.RefactorRequest)
		return &gm.Message{MessageType: gm.Message_RefactorResponse, RefactorResponse: response}, err
	case gm.Message_StepNameRequest:
		response, err := r.RunnerClient.GetStepName(context.Background(), message.StepNameRequest)
		return &gm.Message{MessageType: gm.Message_StepNameResponse, StepNameResponse: response}, err
	case gm.Message_ImplementationFileGlobPatternRequest:
		response, err := r.RunnerClient.GetGlobPatterns(context.Background(), &gm.Empty{})
		return &gm.Message{MessageType: gm.Message_ImplementationFileGlobPatternRequest, ImplementationFileGlobPatternResponse: response}, err
	case gm.Message_StepValidateRequest:
		response, err := r.RunnerClient.ValidateStep(context.Background(), message.StepValidateRequest)
		return &gm.Message{MessageType: gm.Message_StepValidateResponse, StepValidateResponse: response}, err
	case gm.Message_KillProcessRequest:
		_, err := r.RunnerClient.Kill(context.Background(), message.KillProcessRequest)
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.Canceled {
			// Ref https://www.grpc.io/docs/guides/error/#general-errors
			// GRPC_STATUS_UNAVAILABLE is thrown when Server is shutting down. Ignore it here.
			return &gm.Message{}, nil
		}
		return &gm.Message{}, err
	default:
		return nil, nil
	}
}

func (r *GrpcRunner) invokeRPC(message *gm.Message, resChan chan *gm.Message, errChan chan error) {
	var res *gm.Message
	var err error
	if r.LegacyClient != nil {
		res, err = r.invokeLegacyLSPService(message)
	} else {
		res, err = r.invokeServiceFor(message)
	}
	if err != nil {
		errChan <- err
	} else {
		resChan <- res
	}
}

func (r *GrpcRunner) executeMessage(message *gm.Message, timeout time.Duration) (*gm.Message, error) {
	resChan := make(chan *gm.Message)
	errChan := make(chan error)
	go r.invokeRPC(message, resChan, errChan)

	timer := setupTimer(timeout, errChan, message.GetMessageType().String())
	defer stopTimer(timer)

	select {
	case response := <-resChan:
		return response, nil
	case err := <-errChan:
		return nil, err
	}
}

// ExecuteMessageWithTimeout process request and give back the response
func (r *GrpcRunner) ExecuteMessageWithTimeout(message *gm.Message) (*gm.Message, error) {
	return r.executeMessage(message, r.Timeout)
}

// ExecuteAndGetStatus executes a given message and response without timeout.
func (r *GrpcRunner) ExecuteAndGetStatus(m *gm.Message) *gm.ProtoExecutionResult {
	if r.Info().Killed {
		return &gauge_messages.ProtoExecutionResult{Failed: true,  ErrorMessage:"Runner is not Alive"}
	}
	res, err := r.executeMessage(m, 0)
	if err != nil {
		e, ok := status.FromError(err)
		if ok {
			var stackTrace = ""
			for _, detail := range e.Details() {
				if t, ok := detail.(*errdetails.DebugInfo); ok {
					for _, entry := range t.GetStackEntries() {
						stackTrace = fmt.Sprintf("%s%s\n", stackTrace, entry)
					}
				}
			}
			var data = strings.Split(e.Message(), "||")
			var message = data[0]
			if len(data) > 1 && stackTrace == "" {
				stackTrace = data[1]
			}
			if e.Code() == codes.Unavailable {
				r.Info().Killed = true
				return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: message, StackTrace: stackTrace}
			}
			return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: message, StackTrace: stackTrace}
		}
		return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: err.Error()}
	}
	return res.ExecutionStatusResponse.ExecutionResult
}

// Alive check if the runner process is still alive
func (r *GrpcRunner) Alive() bool {
	ps := r.cmd.ProcessState
	return ps == nil || !ps.Exited()
}

// Kill closes the grpc connection and kills the process
func (r *GrpcRunner) Kill() error {
	if r.IsExecuting {
		return nil
	}
	m := &gm.Message{
		MessageType:        gm.Message_KillProcessRequest,
		KillProcessRequest: &gm.KillProcessRequest{},
	}
	m, err := r.executeMessage(m, r.Timeout)
	if m == nil || err != nil {
		return err
	}
	if r.conn == nil && r.cmd == nil {
		return nil
	}
	defer r.conn.Close()
	if r.Alive() {
		exited := make(chan bool, 1)
		go func() {
			for {
				if r.Alive() {
					time.Sleep(100 * time.Millisecond)
				} else {
					exited <- true
					return
				}
			}
		}()

		select {
		case done := <-exited:
			if done {
				logger.Debugf(true, "Runner with PID:%d has exited", r.cmd.Process.Pid)
				return nil
			}
		case <-time.After(config.PluginKillTimeout()):
			logger.Warningf(true, "Killing runner with PID:%d forcefully", r.cmd.Process.Pid)
			return r.cmd.Process.Kill()
		}
	}
	return nil
}

// Connection return the client connection
func (r *GrpcRunner) Connection() net.Conn {
	return nil
}

// IsMultithreaded tells if the runner has multithreaded capability
func (r *GrpcRunner) IsMultithreaded() bool {
	return r.info.Multithreaded
}

// Info gives the information about runner
func (r *GrpcRunner) Info() *RunnerInfo {
	return r.info
}

// Pid return the runner's command pid
func (r *GrpcRunner) Pid() int {
	return r.cmd.Process.Pid
}

// StartGrpcRunner makes a connection with grpc server
func StartGrpcRunner(m *manifest.Manifest, stdout, stderr io.Writer, timeout time.Duration, shouldWriteToStdout bool) (*GrpcRunner, error) {
	portChan := make(chan string)
	errChan := make(chan error)
	logWriter := &logger.LogWriter{
		Stderr: logger.NewCustomWriter(portChan, stderr, m.Language, true),
		Stdout: logger.NewCustomWriter(portChan, stdout, m.Language, false),
	}
	cmd, info, err := runRunnerCommand(m, "0", false, logWriter)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while starting runner process.\nError : %w", err)
	}

	go func() {
		err = cmd.Wait()
		if err != nil {
			e := fmt.Errorf("Error occurred while waiting for runner process to finish.\nError : %w", err)
			logger.Errorf(true, e.Error())
			errChan <- e
		}
		errChan <- nil
	}()

	var port string
	select {
	case port = <-portChan:
		close(portChan)
	case err = <-errChan:
		return nil, err
	case <-time.After(config.RunnerConnectionTimeout()):
		return nil, fmt.Errorf("Timed out connecting to %s", m.Language)
	}
	logger.Debugf(true, "Attempting to connect to grpc server at port: %s", port)
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port),
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(oneGB), grpc.MaxCallSendMsgSize(oneGB)),
		grpc.WithBlock())
	logger.Debugf(true, "Successfully made the connection with runner with port: %s", port)
	if err != nil {
		return nil, err
	}
	r := &GrpcRunner{cmd: cmd, conn: conn, Timeout: timeout, info: info}

	if info.GRPCSupport {
		r.RunnerClient = gm.NewRunnerClient(conn)
	} else {
		r.LegacyClient = gm.NewLspServiceClient(conn)
	}
	return r, nil
}

func setupTimer(timeout time.Duration, errChan chan error, messageType string) *time.Timer {
	if timeout > 0 {
		return time.AfterFunc(timeout, func() {
			errChan <- fmt.Errorf("request timed out for message %s", messageType)
		})
	}
	return nil
}

func stopTimer(timer *time.Timer) {
	if timer != nil && !timer.Stop() {
		<-timer.C
	}
}
