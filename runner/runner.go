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

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/conn"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/version"
)

type TestRunner struct {
	Cmd          *exec.Cmd
	Connection   net.Conn
	ErrorChannel chan error
}

type Runner struct {
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

func GetRunnerInfo(language string) (*Runner, error) {
	runnerInfo := new(Runner)
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

func (testRunner *TestRunner) Kill() error {
	runnerProcessID := testRunner.Cmd.Process.Pid
	if util.IsProcessRunning(runnerProcessID) {
		defer testRunner.Connection.Close()
		testRunner.sendProcessKillMessage()

		exited := make(chan bool, 1)
		go func() {
			for {
				if util.IsProcessRunning(runnerProcessID) {
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
			logger.Warning("Killing runner with PID:%d forcefully", testRunner.Cmd.Process.Pid)
			return testRunner.killRunner()
		}
	}
	return nil
}

func (testRunner *TestRunner) killRunner() error {
	return testRunner.Cmd.Process.Kill()
}

func (testRunner *TestRunner) sendProcessKillMessage() {
	id := common.GetUniqueID()
	message := &gauge_messages.Message{MessageId: &id, MessageType: gauge_messages.Message_KillProcessRequest.Enum(),
		KillProcessRequest: &gauge_messages.KillProcessRequest{}}

	conn.WriteGaugeMessage(message, testRunner.Connection)
}

// Looks for a runner configuration inside the runner directory
// finds the runner configuration matching to the manifest and executes the commands for the current OS
func startRunner(manifest *manifest.Manifest, port string, reporter reporter.Reporter, killChannel chan bool) (*TestRunner, error) {
	var r Runner
	runnerDir, err := getLanguageJSONFilePath(manifest, &r)
	if err != nil {
		return nil, err
	}
	compatibilityErr := version.CheckCompatibility(version.CurrentGaugeVersion, &r.GaugeVersionSupport)
	if compatibilityErr != nil {
		return nil, fmt.Errorf("Compatible runner version to %s not found. To update plugin, run `gauge --update {pluginName}`.", version.CurrentGaugeVersion)
	}
	command := getOsSpecificCommand(r)
	env := getCleanEnv(port, os.Environ())
	cmd, err := common.ExecuteCommandWithEnv(command, runnerDir, reporter, reporter, env)
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
	waitAndGetErrorMessage(errChannel, cmd)
	return &TestRunner{Cmd: cmd, ErrorChannel: errChannel}, nil
}

func getLanguageJSONFilePath(manifest *manifest.Manifest, r *Runner) (string, error) {
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

func waitAndGetErrorMessage(errChannel chan error, cmd *exec.Cmd) {
	go func() {
		err := cmd.Wait()
		if err != nil {
			logger.Debug("Runner exited with error: %s", err)
			errChannel <- fmt.Errorf("Runner exited with error: %s\n", err.Error())
		}
	}()
}

func getCleanEnv(port string, env []string) []string {
	//clear environment variable common.GaugeInternalPortEnvName
	isPresent := false
	for i, k := range env {
		if strings.TrimSpace(strings.Split(k, "=")[0]) == common.GaugeInternalPortEnvName {
			isPresent = true
			env[i] = common.GaugeInternalPortEnvName + "=" + port
		}
	}
	if !isPresent {
		env = append(env, common.GaugeInternalPortEnvName+"="+port)
	}
	return env
}

func getOsSpecificCommand(r Runner) []string {
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
	RunnerChan chan *TestRunner
	// this will hold the error while creating runner
	ErrorChan chan error
	// this holds a flag based on which the runner is terminated
	KillChan chan bool
}

func StartRunnerAndMakeConnection(manifest *manifest.Manifest, reporter reporter.Reporter, killChannel chan bool) (*TestRunner, error) {
	port, err := conn.GetPortFromEnvironmentVariable(common.GaugePortEnvName)
	if err != nil {
		port = 0
	}
	gaugeConnectionHandler, connHandlerErr := conn.NewGaugeConnectionHandler(port, nil)
	if connHandlerErr != nil {
		return nil, connHandlerErr
	}
	testRunner, err := startRunner(manifest, strconv.Itoa(gaugeConnectionHandler.ConnectionPortNumber()), reporter, killChannel)
	if err != nil {
		return nil, err
	}

	runnerConnection, connectionError := gaugeConnectionHandler.AcceptConnection(config.RunnerConnectionTimeout(), testRunner.ErrorChannel)
	testRunner.Connection = runnerConnection
	if connectionError != nil {
		logger.Debug("Runner connection error: %s", connectionError)
		err := testRunner.killRunner()
		if err != nil {
			logger.Debug("Error while killing runner: %s", err)
		}
		return nil, connectionError
	}
	return testRunner, nil
}
