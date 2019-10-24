// Copyright 2018 ThoughtWorks, Inc.

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

package runner

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"time"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge_messages"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"google.golang.org/grpc"
)

const (
	host = "127.0.0.1"
)

// GrpcRunner handles grpc messages.
type GrpcRunner struct {
	cmd             *exec.Cmd
	conn            *grpc.ClientConn
	Client          gm.LspServiceClient
	ExecutionClient gm.RunnerClient
	Timeout         time.Duration
}

func (r *GrpcRunner) invokeLSPService(message *gm.Message) (*gm.Message, error) {
	switch message.MessageType {
	case gm.Message_CacheFileRequest:
		r.Client.CacheFile(context.Background(), message.CacheFileRequest)
		return &gm.Message{}, nil
	case gm.Message_StepNamesRequest:
		response, err := r.Client.GetStepNames(context.Background(), message.StepNamesRequest)
		return &gm.Message{StepNamesResponse: response}, err
	case gm.Message_StepPositionsRequest:
		response, err := r.Client.GetStepPositions(context.Background(), message.StepPositionsRequest)
		return &gm.Message{StepPositionsResponse: response}, err
	case gm.Message_ImplementationFileListRequest:
		response, err := r.Client.GetImplementationFiles(context.Background(), &gm.Empty{})
		return &gm.Message{ImplementationFileListResponse: response}, err
	case gm.Message_StubImplementationCodeRequest:
		response, err := r.Client.ImplementStub(context.Background(), message.StubImplementationCodeRequest)
		return &gm.Message{FileDiff: response}, err
	case gm.Message_StepValidateRequest:
		response, err := r.Client.ValidateStep(context.Background(), message.StepValidateRequest)
		return &gm.Message{MessageType: gm.Message_StepValidateResponse, StepValidateResponse: response}, err
	case gm.Message_RefactorRequest:
		response, err := r.Client.Refactor(context.Background(), message.RefactorRequest)
		return &gm.Message{MessageType: gm.Message_RefactorResponse, RefactorResponse: response}, err
	case gm.Message_StepNameRequest:
		response, err := r.Client.GetStepName(context.Background(), message.StepNameRequest)
		return &gm.Message{MessageType: gm.Message_StepNameResponse, StepNameResponse: response}, err
	case gm.Message_ImplementationFileGlobPatternRequest:
		response, err := r.Client.GetGlobPatterns(context.Background(), &gm.Empty{})
		return &gm.Message{MessageType: gm.Message_ImplementationFileGlobPatternRequest, ImplementationFileGlobPatternResponse: response}, err
	case gm.Message_KillProcessRequest:
		_, err := r.Client.KillProcess(context.Background(), message.KillProcessRequest)
		return &gm.Message{}, err
	default:
		return nil, nil
	}
}

func (r *GrpcRunner) invokeRunnerService(message *gm.Message) (*gm.Message, error) {
	switch message.MessageType {
	case gm.Message_SuiteDataStoreInit:
		response, err := r.ExecutionClient.SuiteDataStoreInit(context.Background(), &gm.Empty{})
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_SpecDataStoreInit:
		response, err := r.ExecutionClient.SpecDataStoreInit(context.Background(), &gm.Empty{})
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ScenarioDataStoreInit:
		response, err := r.ExecutionClient.ScenarioDataStoreInit(context.Background(), &gm.Empty{})
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err

	case gm.Message_ExecutionStarting:
		response, err := r.ExecutionClient.ExecutionStarting(context.Background(), message.ExecutionStartingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_SpecExecutionStarting:
		response, err := r.ExecutionClient.SpecExecutionStarting(context.Background(), message.SpecExecutionStartingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ScenarioExecutionStarting:
		response, err := r.ExecutionClient.ScenarioExecutionStarting(context.Background(), message.ScenarioExecutionStartingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_StepExecutionStarting:
		response, err := r.ExecutionClient.StepExecutionStarting(context.Background(), message.StepExecutionStartingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ExecuteStep:
		response, err := r.ExecutionClient.ExecuteStep(context.Background(), message.ExecuteStepRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_StepExecutionEnding:
		response, err := r.ExecutionClient.StepExecutionEnding(context.Background(), message.StepExecutionEndingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ScenarioExecutionEnding:
		response, err := r.ExecutionClient.ScenarioExecutionEnding(context.Background(), message.ScenarioExecutionEndingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_SpecExecutionEnding:
		response, err := r.ExecutionClient.SpecExecutionEnding(context.Background(), message.SpecExecutionEndingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err
	case gm.Message_ExecutionEnding:
		response, err := r.ExecutionClient.ExecutionEnding(context.Background(), message.ExecutionEndingRequest)
		return &gm.Message{MessageType: gm.Message_ExecutionStatusResponse, ExecutionStatusResponse: response}, err

	case gm.Message_CacheFileRequest:
		_, err := r.ExecutionClient.CacheFile(context.Background(), message.CacheFileRequest)
		return &gm.Message{}, err
	case gm.Message_StepNamesRequest:
		response, err := r.ExecutionClient.GetStepNames(context.Background(), message.StepNamesRequest)
		return &gm.Message{StepNamesResponse: response}, err
	case gm.Message_StepPositionsRequest:
		response, err := r.ExecutionClient.GetStepPositions(context.Background(), message.StepPositionsRequest)
		return &gm.Message{StepPositionsResponse: response}, err
	case gm.Message_ImplementationFileListRequest:
		response, err := r.ExecutionClient.GetImplementationFiles(context.Background(), &gm.Empty{})
		return &gm.Message{ImplementationFileListResponse: response}, err
	case gm.Message_StubImplementationCodeRequest:
		response, err := r.ExecutionClient.ImplementStub(context.Background(), message.StubImplementationCodeRequest)
		return &gm.Message{FileDiff: response}, err
	case gm.Message_StepValidateRequest:
		response, err := r.ExecutionClient.ValidateStep(context.Background(), message.StepValidateRequest)
		return &gm.Message{MessageType: gm.Message_StepValidateResponse, StepValidateResponse: response}, err
	case gm.Message_RefactorRequest:
		response, err := r.ExecutionClient.Refactor(context.Background(), message.RefactorRequest)
		return &gm.Message{MessageType: gm.Message_RefactorResponse, RefactorResponse: response}, err
	case gm.Message_StepNameRequest:
		response, err := r.ExecutionClient.GetStepName(context.Background(), message.StepNameRequest)
		return &gm.Message{MessageType: gm.Message_StepNameResponse, StepNameResponse: response}, err
	case gm.Message_ImplementationFileGlobPatternRequest:
		response, err := r.ExecutionClient.GetGlobPatterns(context.Background(), &gm.Empty{})
		return &gm.Message{MessageType: gm.Message_ImplementationFileGlobPatternRequest, ImplementationFileGlobPatternResponse: response}, err
	case gm.Message_KillProcessRequest:
		_, err := r.ExecutionClient.KillProcess(context.Background(), message.KillProcessRequest)
		return &gm.Message{}, err
	default:
		return nil, nil
	}
}

func (r *GrpcRunner) invokeRPC(message *gm.Message, resChan chan *gm.Message, errChan chan error) {
	var res *gm.Message
	var err error
	if r.ExecutionClient != nil {
		res, err = r.invokeRunnerService(message)
	} else {
		res, err = r.invokeLSPService(message)
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

// ExecuteMessageWithTimeout process reuqest and give back the response
func (r *GrpcRunner) ExecuteMessageWithTimeout(message *gm.Message) (*gm.Message, error) {
	return r.executeMessage(message, r.Timeout)
}

// ExecuteAndGetStatus executes a given message and response without timeout.
func (r *GrpcRunner) ExecuteAndGetStatus(m *gm.Message) *gm.ProtoExecutionResult {
	res, err := r.executeMessage(m, 0)
	if err != nil {
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
	m := &gm.Message{
		MessageType:        gm.Message_KillProcessRequest,
		KillProcessRequest: &gm.KillProcessRequest{},
	}
	m, err := r.executeMessage(m, config.PluginKillTimeout())
	if m == nil || err != nil {
		return err
	}
	if r.conn == nil && r.cmd == nil {
		return nil
	}
	if err = r.conn.Close(); err != nil {
		return err
	}
	if err := r.cmd.Process.Kill(); err != nil {
		return err
	}
	return nil
}

// Connection return the client connection
func (r *GrpcRunner) Connection() net.Conn {
	return nil
}

// IsMultithreaded tells if the runner has multithreaded capability
func (r *GrpcRunner) IsMultithreaded() bool {
	return false
}

// Pid return the runner's command pid
func (r *GrpcRunner) Pid() int {
	return r.cmd.Process.Pid
}

// ConnectToGrpcRunner makes a connection with grpc server
func ConnectToGrpcRunner(manifest *manifest.Manifest, stdout io.Writer, stderr io.Writer, timeout time.Duration, shouldWriteToStdout bool) (*GrpcRunner, error) {
	portChan := make(chan string)
	logWriter := &logger.LogWriter{
		Stderr: newCustomWriter(portChan, stderr, manifest.Language),
		Stdout: newCustomWriter(portChan, stdout, manifest.Language),
	}
	cmd, info, err := runRunnerCommand(manifest, "0", false, logWriter)
	if err != nil {
		return nil, err
	}

	var port string
	select {
	case port = <-portChan:
		close(portChan)
	case <-time.After(config.RunnerConnectionTimeout()):
		return nil, fmt.Errorf("Timed out connecting to %s", manifest.Language)
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port),
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*10)),
		grpc.WithBlock())

	if err != nil {
		return nil, err
	}
	r := &GrpcRunner{cmd: cmd, conn: conn, Timeout: timeout}
	if info.GRPCSupport {
		r.ExecutionClient = gm.NewRunnerClient(conn)
	} else {
		r.Client = gm.NewLspServiceClient(conn)
	}
	return r, nil
}

func setupTimer(timeout time.Duration, errChan chan error, messageType string) *time.Timer {
	if timeout > 0 {
		return time.AfterFunc(timeout, func() {
			errChan <- fmt.Errorf("Request Timed out for message %s", messageType)
		})
	}
	return nil
}

func stopTimer(timer *time.Timer) {
	if timer != nil && !timer.Stop() {
		<-timer.C
	}
}
