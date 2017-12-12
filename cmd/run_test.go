package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"reflect"

	"github.com/getgauge/gauge/config"
	"github.com/spf13/cobra"
)

var path = ""

func before() {
	path, _ = filepath.Abs("_testData")
	config.ProjectRoot = path
}

func after() {
	os.RemoveAll(path)
}

func TestMain(m *testing.M) {
	before()
	runTests := m.Run()
	after()
	os.Exit(runTests)
}

func TestHandleRepeatCommandForWriteCommandFlow(t *testing.T) {
	args := []string{"gauge", "run", "specs"}
	cmd := &cobra.Command{}

	handleRepeatCommand(cmd, args)
	commandWritten := readPrevCmd()
	if !reflect.DeepEqual(commandWritten.Command, args) {
		t.Errorf("Expected %v  Got %v", args, commandWritten.Command)
	}
}

func TestHandleRepeatCommandForRepeatCommandFlow(t *testing.T) {
	args := []string{"gauge", "run", "specs"}
	cmd := &cobra.Command{}

	var execFlowFlag = false
	executeCmd = func(cmd *cobra.Command, lastState []string) {
		execFlowFlag = true
	}

	handleRepeatCommand(cmd, args)

	repeat = true
	handleRepeatCommand(cmd, args)
	if !execFlowFlag {
		t.Errorf("Expected %v  Got %v", true, execFlowFlag)
	}
	commandWritten := readPrevCmd()
	if !reflect.DeepEqual(commandWritten.Command, args) {
		t.Errorf("Expected %v  Got %v", args, commandWritten.Command)
	}
}

func TestHandleRepeatCommandForFailedCommandFlow(t *testing.T) {
	args := []string{"gauge", "run", "specs"}
	cmd := &cobra.Command{}

	handleRepeatCommand(cmd, args)

	prevFailed = true
	args2 := []string{"something", "else"}
	handleRepeatCommand(cmd, args2)
	commandWritten := readPrevCmd()
	if !reflect.DeepEqual(commandWritten.Command, args) {
		t.Errorf("Expected %v  Got %v", args, commandWritten.Command)
	}
}
