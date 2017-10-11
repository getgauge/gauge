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

package runner

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"sync"

	"io"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/version"
)

type Runner interface {
	ExecuteAndGetStatus(m *gauge_messages.Message) *gauge_messages.ProtoExecutionResult
	IsProcessRunning() bool
	Kill() error
	Connection() net.Conn
	IsMultithreaded() bool
	Pid() int
}

type LanguageRunner struct {
	mutex         *sync.Mutex
	Cmd           *exec.Cmd
	connection    net.Conn
	errorChannel  chan error
	multiThreaded bool
}

type MultithreadedRunner struct {
	r *LanguageRunner
}

func (r *MultithreadedRunner) IsProcessRunning() bool {
	if r.r.mutex != nil && r.r.Cmd != nil {
		return r.r.IsProcessRunning()
	}
	return false
}

func (r *MultithreadedRunner) IsMultithreaded() bool {
	return false
}

func (r *MultithreadedRunner) SetConnection(c net.Conn) {
	r.r = &LanguageRunner{connection: c}
}

func (r *MultithreadedRunner) Kill() error {
	defer r.r.connection.Close()
	conn.SendProcessKillMessage(r.r.connection)

	exited := make(chan bool, 1)
	go func() {
		for {
			if r.IsProcessRunning() {
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
		logger.Warningf("Killing runner with PID:%d forcefully", r.r.Cmd.Process.Pid)
		return r.r.Cmd.Process.Kill()
	}
	return nil
}

func (r *MultithreadedRunner) Pid() int {
	return -1
}

func (r *MultithreadedRunner) ExecuteAndGetStatus(message *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
	return r.r.ExecuteAndGetStatus(message)
}

type RunnerInfo struct {
	Id          string
	Name        string
	Version     string
	Description string
	Run         struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	Init struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	Lib                 string
	Multithreaded       bool
	GaugeVersionSupport version.VersionSupport
}

func ExecuteInitHookForRunner(language string) error {
	if err := config.SetProjectRoot([]string{}); err != nil {
		return err
	}
	runnerInfo, err := GetRunnerInfo(language)
	if err != nil {
		return err
	}
	command := []string{}
	switch runtime.GOOS {
	case "windows":
		command = runnerInfo.Init.Windows
		break
	case "darwin":
		command = runnerInfo.Init.Darwin
		break
	default:
		command = runnerInfo.Init.Linux
		break
	}

	languageJSONFilePath, err := plugin.GetLanguageJSONFilePath(language)
	runnerDir := filepath.Dir(languageJSONFilePath)
	cmd, err := common.ExecuteCommand(command, runnerDir, os.Stdout, os.Stderr)

	if err != nil {
		return err
	}

	return cmd.Wait()
}

func GetRunnerInfo(language string) (*RunnerInfo, error) {
	runnerInfo := new(RunnerInfo)
	languageJSONFilePath, err := plugin.GetLanguageJSONFilePath(language)
	if err != nil {
		return nil, err
	}

	contents, err := common.ReadFileContents(languageJSONFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(contents), &runnerInfo)
	if err != nil {
		return nil, err
	}
	return runnerInfo, nil
}
func (r *LanguageRunner) IsProcessRunning() bool {
	r.mutex.Lock()
	ps := r.Cmd.ProcessState
	r.mutex.Unlock()
	return ps == nil || !ps.Exited()
}

func (r *LanguageRunner) IsMultithreaded() bool {
	return r.multiThreaded
}

func (r *LanguageRunner) Kill() error {
	if r.IsProcessRunning() {
		defer r.connection.Close()
		conn.SendProcessKillMessage(r.connection)

		exited := make(chan bool, 1)
		go func() {
			for {
				if r.IsProcessRunning() {
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
			logger.Warningf("Killing runner with PID:%d forcefully", r.Cmd.Process.Pid)
			return r.killRunner()
		}
	}
	return nil
}

func (r *LanguageRunner) Connection() net.Conn {
	return r.connection
}

func (r *LanguageRunner) killRunner() error {
	return r.Cmd.Process.Kill()
}

func (r *LanguageRunner) Pid() int {
	return r.Cmd.Process.Pid
}

func (r *LanguageRunner) ExecuteAndGetStatus(message *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
	response, err := conn.GetResponseForGaugeMessage(message, r.connection)
	if err != nil {
		return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: err.Error()}
	}

	if response.GetMessageType() == gauge_messages.Message_ExecutionStatusResponse {
		executionResult := response.GetExecutionStatusResponse().GetExecutionResult()
		if executionResult == nil {
			errMsg := "ProtoExecutionResult obtained is nil"
			logger.Errorf(errMsg)
			return errorResult(errMsg)
		}
		return executionResult
	}
	errMsg := fmt.Sprintf("Expected ExecutionStatusResponse. Obtained: %s", response.GetMessageType())
	logger.Errorf(errMsg)
	return errorResult(errMsg)
}

func errorResult(message string) *gauge_messages.ProtoExecutionResult {
	return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: message, RecoverableError: false}
}

// Looks for a runner configuration inside the runner directory
// finds the runner configuration matching to the manifest and executes the commands for the current OS
func StartRunner(manifest *manifest.Manifest, port string, outputStreamWriter io.Writer, killChannel chan bool, debug bool) (*LanguageRunner, error) {
	var r RunnerInfo
	runnerDir, err := getLanguageJSONFilePath(manifest, &r)
	if err != nil {
		return nil, err
	}
	compatibilityErr := version.CheckCompatibility(version.CurrentGaugeVersion, &r.GaugeVersionSupport)
	if compatibilityErr != nil {
		return nil, fmt.Errorf("Compatibility error. %s", compatibilityErr.Error())
	}
	command := getOsSpecificCommand(r)
	env := getCleanEnv(port, os.Environ(), debug, getPluginPaths())
	cmd, err := common.ExecuteCommandWithEnv(command, runnerDir, outputStreamWriter, outputStreamWriter, env)
	if err != nil {
		return nil, err
	}
	go func() {
		select {
		case <-killChannel:
			cmd.Process.Kill()
		}
	}()
	// Wait for the process to exit so we will get a detailed error message
	errChannel := make(chan error)
	testRunner := &LanguageRunner{Cmd: cmd, errorChannel: errChannel, mutex: &sync.Mutex{}, multiThreaded: r.Multithreaded}
	testRunner.waitAndGetErrorMessage()
	return testRunner, nil
}

func getPluginPaths() (paths []string) {
	for _, p := range plugin.PluginsWithoutScope() {
		paths = append(paths, p.Path)
	}
	return
}

func getLanguageJSONFilePath(manifest *manifest.Manifest, r *RunnerInfo) (string, error) {
	languageJSONFilePath, err := plugin.GetLanguageJSONFilePath(manifest.Language)
	if err != nil {
		return "", err
	}
	contents, err := common.ReadFileContents(languageJSONFilePath)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(contents), r)
	if err != nil {
		return "", err
	}
	return filepath.Dir(languageJSONFilePath), nil
}

func (r *LanguageRunner) waitAndGetErrorMessage() {
	go func() {
		pState, err := r.Cmd.Process.Wait()
		r.mutex.Lock()
		r.Cmd.ProcessState = pState
		r.mutex.Unlock()
		if err != nil {
			logger.Debugf("Runner exited with error: %s", err)
			r.errorChannel <- fmt.Errorf("Runner exited with error: %s\n", err.Error())
		}
		if !pState.Success() {
			r.errorChannel <- fmt.Errorf("Runner with pid %d quit unexpectedly(%s).", pState.Pid(), pState.String())
		}
	}()
}

func getCleanEnv(port string, env []string, debug bool, pathToAdd []string) []string {
	isPresent := false
	for i, k := range env {
		key := strings.TrimSpace(strings.Split(k, "=")[0])
		//clear environment variable common.GaugeInternalPortEnvName
		if key == common.GaugeInternalPortEnvName {
			isPresent = true
			env[i] = common.GaugeInternalPortEnvName + "=" + port
		} else if key == "PATH" {
			path := os.Getenv("PATH")
			for _, p := range pathToAdd {
				path += string(os.PathListSeparator) + p
			}
			env[i] = "PATH=" + path
		}
	}
	if !isPresent {
		env = append(env, common.GaugeInternalPortEnvName+"="+port)
	}
	if debug {
		env = append(env, "debugging=true")
	}
	return env
}

func getOsSpecificCommand(r RunnerInfo) []string {
	command := []string{}
	switch runtime.GOOS {
	case "windows":
		command = r.Run.Windows
		break
	case "darwin":
		command = r.Run.Darwin
		break
	default:
		command = r.Run.Linux
		break
	}
	return command
}

type StartChannels struct {
	// this will hold the runner
	RunnerChan chan Runner
	// this will hold the error while creating runner
	ErrorChan chan error
	// this holds a flag based on which the runner is terminated
	KillChan chan bool
}

func Start(manifest *manifest.Manifest, outputStreamWriter io.Writer, killChannel chan bool, debug bool) (Runner, error) {
	port, err := conn.GetPortFromEnvironmentVariable(common.GaugePortEnvName)
	if err != nil {
		port = 0
	}
	handler, err := conn.NewGaugeConnectionHandler(port, nil)
	if err != nil {
		return nil, err
	}
	runner, err := StartRunner(manifest, strconv.Itoa(handler.ConnectionPortNumber()), outputStreamWriter, killChannel, debug)
	if err != nil {
		return nil, err
	}
	return connect(handler, runner)
}

func connect(h *conn.GaugeConnectionHandler, runner *LanguageRunner) (Runner, error) {
	connection, connErr := h.AcceptConnection(config.RunnerConnectionTimeout(), runner.errorChannel)
	if connErr != nil {
		logger.Debugf("Runner connection error: %s", connErr)
		if err := runner.killRunner(); err != nil {
			logger.Debugf("Error while killing runner: %s", err)
		}
		return nil, connErr
	}
	runner.connection = connection
	return runner, nil
}
