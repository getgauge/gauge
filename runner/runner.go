/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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

	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
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
	Info() *RunnerInfo
	Pid() int
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
	GRPCSupport         bool
	Killed				bool
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
	case "darwin":
		command = runnerInfo.Init.Darwin
	default:
		command = runnerInfo.Init.Linux
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
	command := getOsSpecificCommand(&r)
	env := getCleanEnv(port, os.Environ(), debug, getPluginPaths())
	cmd, err := common.ExecuteCommandWithEnv(command, runnerDir, writer.Stdout, writer.Stderr, env)
	return cmd, &r, err
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

func (r *LegacyRunner) waitAndGetErrorMessage() {
	go func() {
		pState, err := r.Cmd.Process.Wait()
		r.mutex.Lock()
		r.Cmd.ProcessState = pState
		r.mutex.Unlock()
		if err != nil {
			logger.Debugf(true, "Runner exited with error: %s", err)
			r.errorChannel <- fmt.Errorf("Runner exited with error: %s", err.Error())
		}
		if !pState.Success() {
			r.errorChannel <- fmt.Errorf("Runner with pid %d quit unexpectedly(%s)", pState.Pid(), pState.String())
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

func getOsSpecificCommand(r *RunnerInfo) []string {
	var command []string
	switch runtime.GOOS {
	case "windows":
		command = r.Run.Windows
	case "darwin":
		command = r.Run.Darwin
	default:
		command = r.Run.Linux
	}
	return command
}

func Start(manifest *manifest.Manifest, stream int, killChannel chan bool, debug bool) (Runner, error) {
	ri, err := GetRunnerInfo(manifest.Language)
	if err == nil && ri.GRPCSupport {
		return StartGrpcRunner(manifest, os.Stdout, os.Stderr, config.RunnerRequestTimeout(), true)
	}

	writer := logger.NewLogWriter(manifest.Language, true, stream)
	port, err := conn.GetPortFromEnvironmentVariable(common.GaugePortEnvName)
	if err != nil {
		port = 0
	}
	handler, err := conn.NewGaugeConnectionHandler(port, nil)
	if err != nil {
		return nil, err
	}
	logger.Debugf(true, "Staring %s runner", manifest.Language)
	runner, err := StartLegacyRunner(manifest, strconv.Itoa(handler.ConnectionPortNumber()), writer, killChannel, debug)
	if err != nil {
		return nil, err
	}
	err = connect(handler, runner)
	return runner, err
}

func connect(h *conn.GaugeConnectionHandler, runner *LegacyRunner) error {
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
