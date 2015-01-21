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
}

func executeInitHookForRunner(language string) error {
	if err := setCurrentProjectEnvVariable(); err != nil {
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

func (testRunner *testRunner) kill() error {
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
		case <-time.After(config.RunnerKillTimeout()):
			fmt.Printf("Killing runner with PID:%d forcefully\n", testRunner.cmd.Process.Pid)
			return testRunner.cmd.Process.Kill()
		}
	}
	return nil
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
func startRunner(manifest *manifest) (*testRunner, error) {
	var r runner
	languageJsonFilePath, err := common.GetLanguageJSONFilePath(manifest.Language)
	if err != nil {
		return nil, err
	}

	contents, err := common.ReadFileContents(languageJsonFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(contents), &r)
	if err != nil {
		return nil, err
	}

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
	runnerDir := filepath.Dir(languageJsonFilePath)

	currentConsole := getCurrentConsole()
	cmd, err := common.ExecuteCommand(command, runnerDir, currentConsole, currentConsole)

	if err != nil {
		return nil, err
	}
	// Wait for the process to exit so we will get a detailed error message

	errChannel := make(chan error)
	go func() {
		err := cmd.Wait()
		if err != nil {
			errChannel <- errors.New(fmt.Sprintf("Runner exited with error: %s\n", err.Error()))
		}
	}()

	return &testRunner{cmd: cmd, errorChannel: errChannel}, nil
}
