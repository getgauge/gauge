package main

import (
	"encoding/json"
	"fmt"
	"github.com/getgauge/common"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type testRunner struct {
	cmd        *exec.Cmd
	connection net.Conn
}

type runner struct {
	Name string
	Run  struct {
		Windows string
		Linux   string
		Darwin  string
	}
	Init struct {
		Windows string
		Linux   string
		Darwin  string
	}
}

func executeInitHookForRunner(language string) error {
	var r runner
	languageJsonFilePath, err := common.GetLanguageJSONFilePath(language)
	if err != nil {
		return err
	}

	contents, err := common.ReadFileContents(languageJsonFilePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(contents), &r)
	if err != nil {
		return err
	}

	command := ""
	switch runtime.GOOS {
	case "windows":
		command = r.Init.Windows
		break
	case "darwin":
		command = r.Init.Darwin
		break
	default:
		command = r.Init.Linux
		break
	}

	cmd := common.GetExecutableCommand(command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
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
		case <-time.After(time.Second * 2):
			fmt.Printf("Killing runner with PID:%d forcefully ", testRunner.cmd.Process.Pid)
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
	message := &Message{MessageId: &id, MessageType: Message_KillProcessRequest.Enum(),
		KillProcessRequest: &KillProcessRequest{}}

	writeMessage(testRunner.connection, message)
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

	command := ""
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

	cmd := common.GetExecutableCommand(command)
	cmd.Stdout = getCurrentConsole()
	cmd.Stderr = getCurrentConsole()
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	// Wait for the process to exit so we will get a detailed error message
	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("Runner exited with error: %s\n", err.Error())
		}
	}()

	return &testRunner{cmd: cmd}, nil
}
