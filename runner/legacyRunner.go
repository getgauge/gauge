/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package runner

import (
	"fmt"
	"io"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
)

type LegacyRunner struct {
	mutex         *sync.Mutex
	Cmd           *exec.Cmd
	connection    net.Conn
	errorChannel  chan error
	multiThreaded bool
	lostContact   bool
	info          *RunnerInfo
}

func (r *LegacyRunner) Alive() bool {
	r.mutex.Lock()
	ps := r.Cmd.ProcessState
	r.mutex.Unlock()
	return ps == nil || !ps.Exited()
}

func (r *LegacyRunner) EnsureConnected() bool {
	if r.lostContact {
		return false
	}
	c := r.connection
	err := c.SetReadDeadline(time.Now())
	if err != nil {
		logger.Fatalf(true, "Unable to SetReadDeadLine on runner: %s", err.Error())
	}
	var one []byte
	_, err = c.Read(one)
	if err == io.EOF {
		r.lostContact = true
		logger.Fatalf(true, "Connection to runner with Pid %d lost. The runner probably quit unexpectedly. Inspect logs for potential reasons. Error : %s", r.Cmd.Process.Pid, err.Error())
	}
	opErr, ok := err.(*net.OpError)
	if ok && !(opErr.Temporary() || opErr.Timeout()) {
		r.lostContact = true
		logger.Fatalf(true, "Connection to runner with Pid %d lost. The runner probably quit unexpectedly. Inspect logs for potential reasons. Error : %s", r.Cmd.Process.Pid, err.Error())
	}
	var zero time.Time
	err = c.SetReadDeadline(zero)
	if err != nil {
		logger.Fatalf(true, "Unable to SetReadDeadLine on runner: %s", err.Error())
	}

	return true
}

func (r *LegacyRunner) IsMultithreaded() bool {
	return r.multiThreaded
}

func (r *LegacyRunner) Kill() error {
	if r.Alive() {
		defer r.connection.Close()
		logger.Debug(true, "Sending kill message to runner.")
		err := conn.SendProcessKillMessage(r.connection)
		if err != nil {
			return err
		}
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
				logger.Debugf(true, "Runner with PID:%d has exited", r.Cmd.Process.Pid)
				return nil
			}
		case <-time.After(config.PluginKillTimeout()):
			logger.Warningf(true, "Killing runner with PID:%d forcefully", r.Cmd.Process.Pid)
			return r.killRunner()
		}
	}
	return nil
}

func (r *LegacyRunner) Connection() net.Conn {
	return r.connection
}

func (r *LegacyRunner) killRunner() error {
	return r.Cmd.Process.Kill()
}

// Info gives the information about runner
func (r *LegacyRunner) Info() *RunnerInfo {
	return r.info
}

func (r *LegacyRunner) Pid() int {
	return r.Cmd.Process.Pid
}

// ExecuteAndGetStatus invokes the runner with a request and waits for response. error is thrown only when unable to connect to runner
func (r *LegacyRunner) ExecuteAndGetStatus(message *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
	if !r.EnsureConnected() {
		return nil
	}
	response, err := conn.GetResponseForMessageWithTimeout(message, r.connection, 0)
	if err != nil {
		return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: err.Error()}
	}

	if response.GetMessageType() == gauge_messages.Message_ExecutionStatusResponse {
		executionResult := response.GetExecutionStatusResponse().GetExecutionResult()
		if executionResult == nil {
			errMsg := "ProtoExecutionResult obtained is nil"
			logger.Errorf(true, errMsg)
			return errorResult(errMsg)
		}
		return executionResult
	}
	errMsg := fmt.Sprintf("Expected ExecutionStatusResponse. Obtained: %s", response.GetMessageType())
	logger.Errorf(true, errMsg)
	return errorResult(errMsg)
}

func (r *LegacyRunner) ExecuteMessageWithTimeout(message *gauge_messages.Message) (*gauge_messages.Message, error) {
	r.EnsureConnected()
	return conn.GetResponseForMessageWithTimeout(message, r.Connection(), config.RunnerRequestTimeout())
}

// StartLegacyRunner looks for a runner configuration inside the runner directory
// finds the runner configuration matching to the manifest and executes the commands for the current OS
func StartLegacyRunner(manifest *manifest.Manifest, port string, outputStreamWriter *logger.LogWriter, killChannel chan bool, debug bool) (*LegacyRunner, error) {
	cmd, r, err := runRunnerCommand(manifest, port, debug, outputStreamWriter)
	if err != nil {
		return nil, err
	}
	go func() {
		<-killChannel
		err := cmd.Process.Kill()
		if err != nil {
			logger.Errorf(false, "Unable to kill %s with PID %d : %s", cmd.Path, cmd.Process.Pid, err.Error())
		}
	}()
	// Wait for the process to exit so we will get a detailed error message
	errChannel := make(chan error)
	testRunner := &LegacyRunner{info: r, Cmd: cmd, errorChannel: errChannel, mutex: &sync.Mutex{}, multiThreaded: r.Multithreaded}
	testRunner.waitAndGetErrorMessage()
	return testRunner, nil
}
