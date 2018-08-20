package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/execution"

	"github.com/spf13/pflag"

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

func TestHandleConflictingParamsWithOtherArguments(t *testing.T) {
	args := []string{"specs"}

	var flags = pflag.FlagSet{}
	flags.BoolP("repeat", "r", false, "")
	flags.Set("repeat", "true")

	repeat = true
	expectedErrorMessage := "Invalid Command. Usage: gauge run --repeat"
	err := handleConflictingParams(&flags, args)

	if !reflect.DeepEqual(err.Error(), expectedErrorMessage) {
		t.Errorf("Expected %v  Got %v", expectedErrorMessage, err)
	}
}

func TestHandleConflictingParamsWithOtherFlags(t *testing.T) {
	args := []string{}

	var flags = pflag.FlagSet{}
	flags.BoolP("repeat", "r", false, "")
	flags.Set("repeat", "true")

	flags.StringP("tags", "", "", "")
	flags.Set("tags", "abcd")

	repeat = true
	expectedErrorMessage := "Invalid Command. Usage: gauge run --repeat"
	err := handleConflictingParams(&flags, args)

	if !reflect.DeepEqual(err.Error(), expectedErrorMessage) {
		t.Errorf("Expected %v  Got %v", expectedErrorMessage, err)
	}
}

func TestHandleConflictingParamsWithJustRepeatFlag(t *testing.T) {
	args := []string{}

	var flags = pflag.FlagSet{}
	flags.BoolP("repeat", "r", false, "")
	flags.Set("repeat", "true")

	repeat = true
	err := handleConflictingParams(&flags, args)

	if err != nil {
		t.Errorf("Expected %v  Got %v", nil, err.Error())
	}
}

func TestHandleRerunFlagsWithVerbose(t *testing.T) {
	cmd := &cobra.Command{}

	cmd.Flags().BoolP(verboseName, "v", verboseDefault, "Enable step level reporting on console, default being scenario level")
	cmd.Flags().BoolP(simpleConsoleName, "", simpleConsoleDefault, "Removes colouring and simplifies the console output")
	cmd.Flags().StringP(environmentName, "e", environmentDefault, "Specifies the environment to use")
	cmd.Flags().StringP(tagsName, "t", tagsDefault, "Executes the specs and scenarios tagged with given tags")
	cmd.Flags().StringP(rowsName, "r", rowsDefault, "Executes the specs and scenarios only for the selected rows. It can be specified by range as 2-4 or as list 2,4")
	cmd.Flags().BoolP(parallelName, "p", parallelDefault, "Execute specs in parallel")
	cmd.Flags().IntP(streamsName, "n", streamsDefault, "Specify number of parallel execution streams")
	cmd.Flags().IntP(groupName, "g", groupDefault, "Specify which group of specification to execute based on -n flag")
	cmd.Flags().StringP(strategyName, "", strategyDefault, "Set the parallelization strategy for execution. Possible options are: `eager`, `lazy`")
	cmd.Flags().BoolP(sortName, "s", sortDefault, "Run specs in Alphabetical Order")
	cmd.Flags().BoolP(installPluginsName, "i", installPluginsDefault, "Install All Missing Plugins")
	cmd.Flags().BoolP(failedName, "f", failedDefault, "Run only the scenarios failed in previous run. This is an exclusive flag, it cannot be used in conjunction with any other argument")
	cmd.Flags().BoolP(repeatName, "", repeatDefault, "Repeat last run. This is an exclusive flag, it cannot be used in conjunction with any other argument")
	cmd.Flags().BoolP(hideSuggestionName, "", hideSuggestionDefault, "Prints a step implementation stub for every unimplemented step")
	cmd.Flags().Set(repeatName, "true")
	cmd.Flags().Set(verboseName, "true")

	handleFlags(cmd)
	overridenFlagValue := cmd.Flag(verboseName).Value.String()
	expectedFlag := "true"

	if !reflect.DeepEqual(overridenFlagValue, expectedFlag) {
		t.Errorf("Expected %v  Got %v", expectedFlag, overridenFlagValue)
	}
}

func TestHandleFailedCommandForNonGaugeProject(t *testing.T) {
	config.ProjectRoot = ""
	currDir, _ := os.Getwd()
	defer os.Chdir(currDir)
	testdir := filepath.Join(currDir, "dotGauge")
	dotgaugeDir := filepath.Join(testdir, ".gauge")
	os.Chdir(testdir)
	exit = func(err error, i string) {
		if _, e := os.Stat(dotgaugeDir); os.IsExist(e) {
			t.Fatalf("Folder .gauge is created")
		}
		os.Exit(0)
	}
	if os.Getenv("TEST_EXITS") == "1" {
		os.Args = []string{"gauge", "run", "-f"}
		runCmd.Execute()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestHandleFailedCommandForNonGaugeProject")
	cmd.Env = append(os.Environ(), "TEST_EXITS=1")
	cmd.Run()
}

func TestHandleConflictingParamsWithLogLevelFlag(t *testing.T) {
	args := []string{}
	var flags = pflag.FlagSet{}

	flags.StringP("log-level", "l", "info", "")
	flags.Set("log-level", "debug")

	flags.BoolP("repeat", "r", false, "")
	flags.Set("repeat", "true")
	repeat = true

	err := handleConflictingParams(&flags, args)

	if err != nil {
		t.Errorf("Expected %v  Got %v", nil, err.Error())
	}
}

func TestFailSafeShouldForceReturnZero(t *testing.T) {
	installPlugins = false
	// simulate failure
	execution.ExecuteSpecs = func(s []string) int { return execution.ExecutionFailed }

	if os.Getenv("TEST_EXITS") == "1" {
		os.Args = []string{"gauge", "run", "--fail-safe", "specs"}
		runCmd.Execute()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFailSafeShouldForceReturnZero")
	cmd.Env = append(os.Environ(), "TEST_EXITS=1")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
}

func TestFailSafeShouldNotForceReturnZeroForParseErrors(t *testing.T) {
	installPlugins = false
	// simulate parse failure
	execution.ExecuteSpecs = func(s []string) int { return execution.ParseFailed }

	if os.Getenv("TEST_EXITS") == "1" {
		os.Args = []string{"gauge", "run", "--fail-safe", "specs"}
		runCmd.Execute()
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFailSafeShouldNotForceReturnZeroForParseErrors")
	cmd.Env = append(os.Environ(), "TEST_EXITS=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 2", err)
}

func TestFailureShouldReturnExitCode(t *testing.T) {
	installPlugins = false
	// simulate execution failure
	execution.ExecuteSpecs = func(s []string) int { return execution.ExecutionFailed }

	if os.Getenv("TEST_EXITS") == "1" {
		os.Args = []string{"gauge", "run", "specs"}
		runCmd.Execute()
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFailureShouldReturnExitCode")
	cmd.Env = append(os.Environ(), "TEST_EXITS=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
