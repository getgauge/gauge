/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package runner

import (
	"net"
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/logger"
)

type MultithreadedRunner struct {
	r *LegacyRunner
}

func (r *MultithreadedRunner) Alive() bool {
	if r.r.mutex != nil && r.r.Cmd != nil {
		return r.r.Alive()
	}
	return false
}

func (r *MultithreadedRunner) IsMultithreaded() bool {
	return false
}

func (r *MultithreadedRunner) SetConnection(c net.Conn) {
	r.r = &LegacyRunner{connection: c}
}

func (r *MultithreadedRunner) Kill() error {
	defer func(connection net.Conn) {
		_ = connection.Close()
	}(r.r.connection)
	err := conn.SendProcessKillMessage(r.r.connection)
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
			return nil
		}
	case <-time.After(config.PluginKillTimeout()):
		return r.killRunner()
	}
	return nil
}

func (r *MultithreadedRunner) Connection() net.Conn {
	return r.r.connection
}

func (r *MultithreadedRunner) killRunner() error {
	if r.r.Cmd != nil && r.r.Cmd.Process != nil {
		logger.Warningf(true, "Killing runner with PID:%d forcefully", r.r.Cmd.Process.Pid)
		return r.r.Cmd.Process.Kill()
	}
	return nil
}

// Info gives the information about runner
func (r *MultithreadedRunner) Info() *RunnerInfo {
	return r.r.Info()
}

func (r *MultithreadedRunner) Pid() int {
	return -1
}

func (r *MultithreadedRunner) ExecuteAndGetStatus(message *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
	return r.r.ExecuteAndGetStatus(message)
}

func (r *MultithreadedRunner) ExecuteMessageWithTimeout(message *gauge_messages.Message) (*gauge_messages.Message, error) {
	r.r.EnsureConnected()
	return conn.GetResponseForMessageWithTimeout(message, r.r.Connection(), config.RunnerRequestTimeout())
}
