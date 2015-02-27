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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge_messages"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type testRunner struct {
	cmd          *exec.Cmd
	connection   net.Conn
	errorChannel chan error
}

type runner struct {
	Name string
	Run  struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	Init struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	Lib string
	GaugeVersionSupport versionSupport
}

func executeInitHookForRunner(language string) error {
	if err := config.SetProjectRoot([]string{}); err != nil {
		return err
	}
	runnerInfo, err := getRunnerInfo(language)
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

	languageJsonFilePath, err := common.GetLanguageJSONFilePath(language)
	runnerDir := filepath.Dir(languageJsonFilePath)
	cmd, err := common.ExecuteCommand(command, runnerDir, os.Stdout, os.Stderr)

	if err != nil {
		return err
	}

	return cmd.Wait()
}

func getRunnerInfo(language string) (*runner, error) {
	runnerInfo := new(runner)
	languageJsonFilePath, err := common.GetLanguageJSONFilePath(language)
	if err != nil {
		return nil, err
	}

	contents, err := common.ReadFileContents(languageJsonFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(contents), &runnerInfo)
	if err != nil {
		return nil, err
	}
	return runnerInfo, nil
}

func (testRunner *testRunner) kill(writer executionLogger) error {
	if testRunner.isStillRunning() {
		testRunner.sendProcessKillMessage()

		exited := make(chan bool, 1)
		go func() {
			for {
				if testRunner.isStillRunning() {
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
			writer.Warning("Killing runner with PID:%d forcefully\n", testRunner.cmd.Process.Pid)
			return testRunner.killRunner()
		}
	}
	return nil
}

func (testRunner *testRunner) killRunner() error {
	return testRunner.cmd.Process.Kill()
}

func (testRunner *testRunner) isStillRunning() bool {
	return testRunner.cmd.ProcessState == nil || !testRunner.cmd.ProcessState.Exited()
}

func (testRunner *testRunner) sendProcessKillMessage() {
	id := common.GetUniqueId()
	message := &gauge_messages.Message{MessageId: &id, MessageType: gauge_messages.Message_KillProcessRequest.Enum(),
		KillProcessRequest: &gauge_messages.KillProcessRequest{}}

	writeGaugeMessage(message, testRunner.connection)
}

// Looks for a runner configuration inside the runner directory
// finds the runner configuration matching to the manifest and executes the commands for the current OS
func startRunner(manifest *manifest, port string, writer executionLogger) (*testRunner, error) {
	var r runner
	runnerDir, err := getLanguageJSONFilePath(manifest, &r)
	if err != nil {
		return nil, err
	}
	compatibilityErr := checkCompatiblity(currentGaugeVersion, &r.GaugeVersionSupport)
	if compatibilityErr != nil {
		return nil, errors.New(fmt.Sprintf("Compatible runner version to %s not found", currentGaugeVersion))
	}
	command := getOsSpecificCommand(r)
	env := getCleanEnv(port, os.Environ())
	cmd, err := common.ExecuteCommandWithEnv(command, runnerDir, writer, writer, env)
	if err != nil {
		return nil, err
	}
	// Wait for the process to exit so we will get a detailed error message
	errChannel := make(chan error)
	waitAndGetErrorMessage(errChannel, cmd, writer)
	return &testRunner{cmd: cmd, errorChannel: errChannel}, nil
}

func getLanguageJSONFilePath(manifest *manifest, r *runner) (string, error) {
	languageJsonFilePath, err := common.GetLanguageJSONFilePath(manifest.Language)
	if err != nil {
		return "", err
	}
	contents, err := common.ReadFileContents(languageJsonFilePath)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(contents), r)
	if err != nil {
		return "", err
	}
	return filepath.Dir(languageJsonFilePath), nil
}

func waitAndGetErrorMessage(errChannel chan error, cmd *exec.Cmd, writer executionLogger) {
	go func() {
		err := cmd.Wait()
		if err != nil {
			writer.Debug("Runner exited with error: %s", err)
			errChannel <- errors.New(fmt.Sprintf("Runner exited with error: %s\n", err.Error()))
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

func getOsSpecificCommand(r runner) []string {
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
