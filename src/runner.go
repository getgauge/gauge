package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type testRunner struct {
	cmd *exec.Cmd
}

// Looks for a runner configuration inside the runner directory
// finds the runner configuration matching to the manifest and executes the commands for the current OS
func startRunner(manifest *manifest) (*testRunner, error) {
	type runner struct {
		Name    string
		Command struct {
			Windows string
			Linux   string
			Darwin  string
		}
	}

	var r runner
	contents := readFileContents(fmt.Sprintf("runner/%s.json", manifest.Language))
	err := json.Unmarshal([]byte(contents), &r)
	if err != nil {
		return nil, err
	}

	command := ""
	switch runtime.GOOS {
	case "windows":
		command = r.Command.Windows
		break
	case "darwin":
		command = r.Command.Darwin
		break
	default:
		command = r.Command.Linux
		break
	}

	cmd := exec.Command(command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
