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
	ExecuteMessageWithTimeout(m *gauge_messages.Message) (*gauge_messages.Message, error)
	Alive() bool
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
	lostContact   bool
}

type MultithreadedRunner struct {
	r *LanguageRunner
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
	r.r = &LanguageRunner{connection: c}
}

func (r *MultithreadedRunner) Kill() error {
	defer r.r.connection.Close()
	conn.SendProcessKillMessage(r.r.connection)

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
	LspLangId           string
}

func ExecuteInitHookForRunner(language string) error {
	if err := config.SetProjectRoot([]string{}); err != nil {
		return err
	}
	runnerInfo, err := GetRunnerInfo(language)
	if err != nil {
		return err
	}
	var command []string
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
	if err != nil {
		return err
	}

	runnerDir := filepath.Dir(languageJSONFilePath)
	logger.Debugf(true, "Running init hook command => %s", command)
	writer := logger.NewLogWriter(language, true, 0)
	cmd, err := common.ExecuteCommand(command, runnerDir, writer.Stdout, writer.Stderr)

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

func (r *LanguageRunner) Alive() bool {
	r.mutex.Lock()
	ps := r.Cmd.ProcessState
	r.mutex.Unlock()
	return ps == nil || !ps.Exited()
}

func (r *LanguageRunner) EnsureConnected() bool {
	if r.lostContact {
		return false
	}
	c := r.connection
	c.SetReadDeadline(time.Now())
	var one []byte
	_, err := c.Read(one)
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
	c.SetReadDeadline(zero)
	return true
}

func (r *LanguageRunner) IsMultithreaded() bool {
	return r.multiThreaded
}

func (r *LanguageRunner) Kill() error {
	if r.Alive() {
		defer r.connection.Close()
		logger.Debug(true, "Sending kill message to runner.")
		conn.SendProcessKillMessage(r.connection)

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
			logger.Warningf(true, "Killing runner with PID:%d forcefully", r.Cmd.Process.Pid)
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

// ExecuteAndGetStatus invokes the runner with a request and waits for response. error is thrown only when unable to connect to runner
func (r *LanguageRunner) ExecuteAndGetStatus(message *gauge_messages.Message) *gauge_messages.ProtoExecutionResult {
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

func (r *LanguageRunner) ExecuteMessageWithTimeout(message *gauge_messages.Message) (*gauge_messages.Message, error) {
	r.EnsureConnected()
	return conn.GetResponseForMessageWithTimeout(message, r.Connection(), config.RunnerRequestTimeout())
}

func errorResult(message string) *gauge_messages.ProtoExecutionResult {
	return &gauge_messages.ProtoExecutionResult{Failed: true, ErrorMessage: message, RecoverableError: false}
}

func runRunnerCommand(manifest *manifest.Manifest, port string, debug bool, writer *logger.LogWriter) (*exec.Cmd, *RunnerInfo, error) {
	var r RunnerInfo
	runnerDir, err := getLanguageJSONFilePath(manifest, &r)
	if err != nil {
		return nil, nil, err
	}
	compatibilityErr := version.CheckCompatibility(version.CurrentGaugeVersion, &r.GaugeVersionSupport)
	if compatibilityErr != nil {
		return nil, nil, fmt.Errorf("Compatibility error. %s", compatibilityErr.Error())
	}
	command := getOsSpecificCommand(r)
	env := getCleanEnv(port, os.Environ(), debug, getPluginPaths())
	env = append(env, fmt.Sprintf("GAUGE_UNIQUE_INSTALLATION_ID=%s", config.UniqueID()))
	env = append(env, fmt.Sprintf("GAUGE_TELEMETRY_ENABLED=%v", config.TelemetryEnabled()))
	cmd, err := common.ExecuteCommandWithEnv(command, runnerDir, writer.Stdout, writer.Stderr, env)
	return cmd, &r, err
}

// StartRunner Looks for a runner configuration inside the runner directory
// finds the runner configuration matching to the manifest and executes the commands for the current OS
func StartRunner(manifest *manifest.Manifest, port string, outputStreamWriter *logger.LogWriter, killChannel chan bool, debug bool) (*LanguageRunner, error) {
	cmd, r, err := runRunnerCommand(manifest, port, debug, outputStreamWriter)
	if err != nil {
		return nil, err
	}
	go func() {
		<-killChannel
		cmd.Process.Kill()
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
			logger.Debugf(true, "Runner exited with error: %s", err)
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
		} else if strings.ToUpper(key) == "PATH" {
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
	var command []string
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

func Start(manifest *manifest.Manifest, outputStreamWriter *logger.LogWriter, killChannel chan bool, debug bool) (Runner, error) {
	port, err := conn.GetPortFromEnvironmentVariable(common.GaugePortEnvName)
	if err != nil {
		port = 0
	}
	handler, err := conn.NewGaugeConnectionHandler(port, nil)
	if err != nil {
		return nil, err
	}
	logger.Debug(true, "Starting runner")
	runner, err := StartRunner(manifest, strconv.Itoa(handler.ConnectionPortNumber()), outputStreamWriter, killChannel, debug)
	if err != nil {
		return nil, err
	}
	err = connect(handler, runner)
	return runner, err
}

func connect(h *conn.GaugeConnectionHandler, runner *LanguageRunner) error {
	connection, connErr := h.AcceptConnection(config.RunnerConnectionTimeout(), runner.errorChannel)
	if connErr != nil {
		logger.Debugf(true, "Runner connection error: %s", connErr)
		if err := runner.killRunner(); err != nil {
			logger.Debugf(true, "Error while killing runner: %s", err)
		}
		return connErr
	}
	runner.connection = connection
	logger.Debug(true, "Established connection to runner.")
	return nil
}
