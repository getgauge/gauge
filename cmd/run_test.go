package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/gauge"

	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/rerun"

	"github.com/spf13/pflag"

	"reflect"

	"github.com/getgauge/gauge/config"
	"github.com/spf13/cobra"
)

var projectPath = ""

func before() {
	projectPath, _ = filepath.Abs("_testData")
	config.ProjectRoot = projectPath
}

func after() {
	os.RemoveAll(projectPath)
}

func TestMain(m *testing.M) {
	before()
	runTests := m.Run()
	after()
	os.Exit(runTests)
}

func TestSaveCommandArgs(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		args := []string{"gauge", "run", "specs"}

		rerun.WritePrevArgs(args)

		prevArgs := rerun.ReadPrevArgs()
		if !reflect.DeepEqual(prevArgs, args) {
			fmt.Printf("Expected %v  Got %v\n", args, prevArgs)
			os.Exit(1)
		}
		return
	}
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
}

func TestExecuteWritesPrevCommandArgs(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		args := []string{"gauge", "run", "specs"}

		installPlugins = false
		execution.ExecuteSpecs = func(s []string) int { return 0 }
		cmd := &cobra.Command{}

		os.Args = args
		execute(cmd, args)
		prevArgs := rerun.ReadPrevArgs()
		if !reflect.DeepEqual(prevArgs, args) {
			fmt.Printf("Expected %v  Got %v\n", args, prevArgs)
			os.Exit(1)
		}
		return
	}
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
}

func TestRepeatShouldPreservePreviousArgs(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		cmd := &cobra.Command{}

		var called bool
		rerun.WritePrevArgs = func(x []string) {
			called = true
		}
		rerun.ReadPrevArgs = func() []string {
			return []string{"gauge", "run", "specs", "-l", "debug"}
		}
		installPlugins = false
		repeatLastExecution(cmd)

		if called {
			panic("Unexpected call to writePrevArgs while repeat")
		}
		return
	}
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
}

func TestSaveCommandArgsForFailed(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		execution.ExecuteSpecs = func(s []string) int { return 0 }
		rerun.GetLastFailedState = func() ([]string, error) {
			return []string{"run", "specs"}, nil
		}
		var args = []string{"gauge", "run", "--failed"}

		rerun.WritePrevArgs = func(a []string) {
			if !reflect.DeepEqual(a, args) {
				panic(fmt.Sprintf("Expected %v  Got %v", args, a))
			}
		}

		os.Args = args
		executeFailed(runCmd)
		return
	}

	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
}

func TestHandleConflictingParamsWithOtherArguments(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		args := []string{"specs"}

		var flags = pflag.FlagSet{}
		flags.BoolP("repeat", "r", false, "")
		flags.Set("repeat", "true")

		repeat = true
		expectedErrorMessage := "Invalid Command. Usage: gauge run --repeat"
		err := handleConflictingParams(&flags, args)

		if !reflect.DeepEqual(err.Error(), expectedErrorMessage) {
			fmt.Printf("Expected %v  Got %v\n", expectedErrorMessage, err)
			panic("assert failed")
		}
		return
	}
	var stdout bytes.Buffer
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0. Stdout:\n%s", err, stdout.Bytes())
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
	if os.Getenv("TEST_EXITS") == "1" {
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

		handleFlags(cmd, []string{"--repeat", "--verbose"})
		overridenFlagValue := cmd.Flag(verboseName).Value.String()
		expectedFlag := "true"

		if !reflect.DeepEqual(overridenFlagValue, expectedFlag) {
			fmt.Printf("Expected %v Got %v\n", expectedFlag, overridenFlagValue)
			os.Exit(1)
		}
		return
	}
	var stdout bytes.Buffer
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0. Stdout:\n%s", err, stdout.Bytes())
	}
}

func TestHandleFailedCommandForNonGaugeProject(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		config.ProjectRoot = ""
		currDir, _ := os.Getwd()
		defer os.Chdir(currDir)
		testdir := filepath.Join(currDir, "dotGauge")
		dotgaugeDir := filepath.Join(testdir, ".gauge")
		os.Chdir(testdir)
		exit = func(err error, i string) {
			if _, e := os.Stat(dotgaugeDir); os.IsExist(e) {
				panic("Folder .gauge is created")
			}
			os.Exit(0)
		}

		os.Args = []string{"gauge", "run", "-f"}

		runCmd.Execute()
		return
	}
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0", err)
	}
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

func TestNoExitCodeShouldForceReturnZero(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		installPlugins = false
		// simulate failure
		execution.ExecuteSpecs = func(s []string) int { return execution.ExecutionFailed }

		os.Args = []string{"gauge", "run", "--fail-safe", "specs"}

		failSafe = true
		runCmd.Execute()
		return
	}
	var stdout bytes.Buffer
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("%s process ran with err %v, want exit status 0. Stdout:\n%s", os.Args, err, stdout.Bytes())
	}
}

func TestFailureShouldReturnExitCodeForParseErrors(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		// simulate parse failure
		execution.ExecuteSpecs = func(s []string) int { return execution.ParseFailed }

		os.Args = []string{"gauge", "run", "--fail-safe", "specs"}
		failSafe = true
		runCmd.Execute()
	}

	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = append(os.Environ(), "TEST_EXITS=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 2", err)
}

func TestFailureShouldReturnExitCode(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		// simulate execution failure
		execution.ExecuteSpecs = func(s []string) int { return execution.ExecutionFailed }
		os.Args = []string{"gauge", "run", "specs"}
		runCmd.Execute()
	}

	var stdout bytes.Buffer
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	cmd.Stdout = &stdout
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1. Stdout:\n%s", err, stdout.Bytes())
}

func TestLogLevelCanBeOverriddenForFailed(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		// expect log level to be overridden
		execution.ExecuteSpecs = func(s []string) int {
			f, err := runCmd.Flags().GetString(logLevelName)
			if err != nil {
				fmt.Printf("Error parsing flags. %s\n", err.Error())
				panic(err)
			}
			if f != "info" {
				fmt.Printf("Expecting log-level=info, got %s\n", f)
				panic("assert failure")
			}
			return 0
		}

		rerun.ReadPrevArgs = func() []string {
			return []string{"gauge", "run", "specs", "-l", "debug"}
		}
		os.Args = []string{"gauge", "run", "--failed", "-l", "info"}
		os.MkdirAll(filepath.Join(projectPath,".gauge"),0755)
		file, err := os.OpenFile(filepath.Join(projectPath,".gauge","failures.json"),os.O_CREATE|os.O_WRONLY,0644)
		_, err = file.Write([]byte("{\"Args\": [\"run\",\"-v\"],\"FailedItems\": [\"specs\"]}"))
		if err != nil {
			fmt.Println(err)
		}
		file.Close()
		executeFailed(runCmd)
		return
	}
	var stdout bytes.Buffer
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0.Stdout:\n%s", err, stdout.Bytes())
	}
}

func TestLogLevelCanBeOverriddenForRepeat(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		// expect log level to be overridden
		execution.ExecuteSpecs = func(s []string) int {
			f, err := runCmd.Flags().GetString(logLevelName)
			if err != nil {
				fmt.Printf("Error parsing flags. %s\n", err.Error())
				panic(err)
			}
			if f != "info" {
				fmt.Printf("Expecting log-level=info, got %s\n", f)
				panic("assert failure")
			}
			return 0
		}

		rerun.ReadPrevArgs = func() []string {
			return []string{"gauge", "run", "specs", "-l=debug"}
		}
		os.Args = []string{"gauge", "run", "--failed", "-l=info"}
		runCmd.ParseFlags(os.Args)
		repeatLastExecution(runCmd)
		return
	}
	var stdout bytes.Buffer
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()), "-test.v")
	cmd.Env = subEnv()
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0.Stdout:\n%s", err, stdout.Bytes())
	}
}

func TestCorrectFlagsAreSetForRepeat(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		// expect "env" to be set to "test"
		os.MkdirAll(filepath.Join(projectPath, "env", "test"), 0755)
		execution.ExecuteSpecs = func(s []string) int {
			f, err := runCmd.Flags().GetString(environmentName)
			if err != nil {
				fmt.Printf("Error parsing flags. %s\n", err.Error())
				panic(err)
			}
			if f != "test" {
				fmt.Printf("Expecting env=test, got %s\n", f)
				panic("assert failure")
			}
			return 0
		}

		rerun.ReadPrevArgs = func() []string {
			return []string{"gauge", "run", "specs", "--env=test"}
		}
		os.Args = []string{"gauge", "run", "--failed"}
		runCmd.ParseFlags(os.Args)
		repeatLastExecution(runCmd)
		return
	}
	var stdout bytes.Buffer
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()))
	cmd.Env = subEnv()
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0.Stdout:\n%s", err, stdout.Bytes())
	}
}

func TestCorrectFlagsAreSetForFailed(t *testing.T) {
	if os.Getenv("TEST_EXITS") == "1" {
		// expect "env" to be set to "test"
		execution.ExecuteSpecs = func(s []string) int {
			f, err := runCmd.Flags().GetString(environmentName)
			if err != nil {
				fmt.Printf("Error parsing flags. %s\n", err.Error())
				panic(err)
			}
			if f != "test" {
				fmt.Printf("Expecting env=test, got %s\n", f)
				panic("assert failure")
			}
			return 0
		}

		rerun.GetLastFailedState = func() ([]string, error) {
			return []string{"run", "specs", "--env=test"}, nil
		}
		os.Args = []string{"gauge", "run", "--failed"}
		executeFailed(runCmd)
		return
	}
	var stdout bytes.Buffer
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", t.Name()), "-test.v")
	cmd.Env = subEnv()
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v, want exit status 0.Stdout:\n%s", err, stdout.Bytes())
	}
}

func subEnv() []string {
	return append(os.Environ(), []string{"TEST_EXITS=1", "GAUGE_PLUGIN_INSTALL=false"}...)
}

func TestAddingFlagsToExecutionArgs(t *testing.T) {
	var flags = &pflag.FlagSet{}
	flags.BoolP("parallel", "p", false, "")
	flags.Set("parallel", "true")

	execution.ExecutionArgs = []*gauge.ExecutionArg{}
	addFlagsToExecutionArgs(flags)
	if execution.ExecutionArgs[0].Name != "parallel" {
		t.Fatalf("Expecting execution arg name parallel but found %s", execution.ExecutionArgs[0].Name)
	}
	if execution.ExecutionArgs[0].Value[0] != "true" {
		t.Fatalf("Expecting execution arg value true but found %s", execution.ExecutionArgs[0].Value[0])
	}
}

func TestAddingMultipleFlagsToExecutionArgs(t *testing.T) {
	var flags = &pflag.FlagSet{}
	flags.BoolP("parallel", "p", false, "")
	flags.Set("parallel", "true")
	flags.StringP("tags", "", "", "")
	flags.Set("tags", "tag1")

	execution.ExecutionArgs = []*gauge.ExecutionArg{}
	addFlagsToExecutionArgs(flags)

	if execution.ExecutionArgs[0].Name != "parallel" {
		t.Fatalf("Expecting execution arg name parallel but found %s", execution.ExecutionArgs[0].Name)
	}
	if execution.ExecutionArgs[1].Name != "tags" {
		t.Fatalf("Expecting execution arg name tags but found %s", execution.ExecutionArgs[1].Name)
	}
	if execution.ExecutionArgs[1].Value[0] != "tag1" {
		t.Fatalf("Expecting execution arg value tag1 but found %s", execution.ExecutionArgs[1].Value[0])
	}
}
