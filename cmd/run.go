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

package cmd

import (
	"fmt"
	"os"

	"strings"

	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/track"
	"github.com/getgauge/gauge/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	lastRunCmdFileName    = "lastRunCmd.json"
	verboseDefault        = false
	simpleConsoleDefault  = false
	failedDefault         = false
	repeatDefault         = false
	parallelDefault       = false
	sortDefault           = false
	installPluginsDefault = true
	environmentDefault    = "default"
	tagsDefault           = ""
	rowsDefault           = ""
	strategyDefault       = "lazy"
	groupDefault          = -1
	failSafeDefault       = false

	verboseName        = "verbose"
	simpleConsoleName  = "simple-console"
	failedName         = "failed"
	repeatName         = "repeat"
	parallelName       = "parallel"
	sortName           = "sort"
	installPluginsName = "install-plugins"
	environmentName    = "env"
	tagsName           = "tags"
	rowsName           = "table-rows"
	strategyName       = "strategy"
	groupName          = "group"
	streamsName        = "n"
	failSafeName       = "fail-safe"
)

var overrideRerunFlags = []string{verboseName, simpleConsoleName, machineReadableName, dirName, logLevelName}
var streamsDefault = util.NumberOfCores()

type prevCommand struct {
	Command []string
}

func newPrevCommand() *prevCommand {
	return &prevCommand{Command: make([]string, 0)}
}

func (cmd *prevCommand) getJSON() (string, error) {
	j, err := json.MarshalIndent(cmd, "", "\t")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

var (
	runCmd = &cobra.Command{
		Use:   "run [flags] [args]",
		Short: "Run specs",
		Long:  `Run specs.`,
		Example: `  gauge run specs/
  gauge run --tags "login" -s -p specs/`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.SetProjectRoot(args); err != nil {
				exit(err, cmd.UsageString())
			}
			if er := handleConflictingParams(cmd.Flags(), args); er != nil {
				exit(er, "")
			}
			handleRepeatCommand(cmd, os.Args)
			if repeat {
				prevCmd := readPrevCmd()
				executeCmd(cmd, prevCmd.Command)
				return
			}
			if failed {
				loadLastState(cmd)
				return
			}
			execute(cmd, args)
		},
		DisableAutoGenTag: true,
	}
	verbose        bool
	simpleConsole  bool
	failed         bool
	repeat         bool
	parallel       bool
	sort           bool
	installPlugins bool
	environment    string
	tags           string
	rows           string
	strategy       string
	streams        int
	group          int
	failSafe       bool
)

func init() {
	GaugeCmd.AddCommand(runCmd)
	f := runCmd.Flags()
	f.BoolVarP(&verbose, verboseName, "v", verboseDefault, "Enable step level reporting on console, default being scenario level")
	f.BoolVarP(&simpleConsole, simpleConsoleName, "", simpleConsoleDefault, "Removes colouring and simplifies the console output")
	f.StringVarP(&environment, environmentName, "e", environmentDefault, "Specifies the environment to use")
	f.StringVarP(&tags, tagsName, "t", tagsDefault, "Executes the specs and scenarios tagged with given tags")
	f.StringVarP(&rows, rowsName, "r", rowsDefault, "Executes the specs and scenarios only for the selected rows. It can be specified by range as 2-4 or as list 2,4")
	f.BoolVarP(&parallel, parallelName, "p", parallelDefault, "Execute specs in parallel")
	f.IntVarP(&streams, streamsName, "n", streamsDefault, "Specify number of parallel execution streams")
	f.IntVarP(&group, groupName, "g", groupDefault, "Specify which group of specification to execute based on -n flag")
	f.StringVarP(&strategy, strategyName, "", strategyDefault, "Set the parallelization strategy for execution. Possible options are: `eager`, `lazy`")
	f.BoolVarP(&sort, sortName, "s", sortDefault, "Run specs in Alphabetical Order")
	f.BoolVarP(&installPlugins, installPluginsName, "i", installPluginsDefault, "Install All Missing Plugins")
	f.BoolVarP(&failed, failedName, "f", failedDefault, "Run only the scenarios failed in previous run. This cannot be used in conjunction with any other argument")
	f.BoolVarP(&repeat, repeatName, "", repeatDefault, "Repeat last run. This cannot be used in conjunction with any other argument")
	f.BoolVarP(&hideSuggestion, hideSuggestionName, "", hideSuggestionDefault, "Prints a step implementation stub for every unimplemented step")
	f.BoolVarP(&noExitCode, noExitCodeName, "", noExitCodeDefault, "Force return 0 exit code, even in case of failures.")
}

//This flag stores whether the command is gauge run --failed and if it is triggering another command.
//The purpose is to only store commands given by user in the lastRunCmd file.
//We need this flag to stop the followup commands(non user given) from getting saved in that file.
var prevFailed = false

func loadLastState(cmd *cobra.Command) {
	lastState, err := rerun.GetLastState()
	if err != nil {
		exit(err, "")
	}
	logger.Debugf(true, "Executing => gauge %s\n", strings.Join(lastState, " "))
	handleFlags(cmd, append([]string{"gauge"}, lastState...))
	prevFailed = true
	cmd.Execute()
}

func resetFlags() {
	failed, repeat, parallel, sort, hideSuggestion, installPlugins =
		failedDefault, repeatDefault, parallelDefault, sortDefault,
		hideSuggestionDefault, installPluginsDefault
	environment, tags, rows, strategy, streams, group =
		environmentDefault, tagsDefault, rowsDefault, strategyDefault,
		streamsDefault, groupDefault
}

func overrideFlags(cmd *cobra.Command, flagResetMap map[string]string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		val, ok := flagResetMap[flag.Name]
		if ok {
			flag.Value.Set(val)
		}
	})
}

func handleFlags(cmd *cobra.Command, args []string) {
	flagResetMap := map[string]string{}
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if util.ListContains(overrideRerunFlags, flag.Name) && flag.Changed {
			flagResetMap[flag.Name] = flag.Value.String()
		}
	})
	resetFlags()
	overrideFlags(cmd, flagResetMap)

	for i := 0; i < len(args)-1; i++ {
		if !isFlag(args[i]) {
			continue
		}
		f := lookupFlagFromArgs(cmd, args[i])
		if f == nil {
			continue
		}
		if v, ok := flagResetMap[f.Name]; ok {
			args[i+1] = v
		}
	}
	os.Args = args
}

func isFlag(f string) bool {
	return strings.HasPrefix(f, "-")
}

func lookupFlagFromArgs(cmd *cobra.Command, arg string) *pflag.Flag {
	fName := strings.TrimLeft(arg, "-")
	flags := cmd.Flags()
	f := flags.Lookup(fName)
	if f == nil && len(fName) == 1 {
		f = flags.ShorthandLookup(fName)
	}
	return f
}
func installMissingPlugins(flag bool) {
	if flag {
		install.AllPlugins(machineReadable)
	}
}

func execute(cmd *cobra.Command, args []string) {
	specs := getSpecsDir(args)
	rerun.SaveState(os.Args[1:], specs)
	track.Execution(parallel, tags != "", sort, simpleConsole, verbose, hideSuggestion, strategy)
	installMissingPlugins(installPlugins)
	exitCode := execution.ExecuteSpecs(specs)
	notifyTelemetryIfNeeded(cmd, args)
	if failSafe && exitCode != execution.ParseFailed {
		exitCode = 0
	}
	os.Exit(exitCode)
}

func handleRepeatCommand(cmd *cobra.Command, cmdArgs []string) {
	if !repeat {
		if prevFailed {
			prevFailed = false
			return
		}
		writePrevCmd(cmdArgs)
	}
}

var executeCmd = func(cmd *cobra.Command, lastState []string) {
	logger.Debugf(true, "Executing => %s\n", strings.Join(lastState, " "))
	handleFlags(cmd, lastState)
	cmd.Execute()
}

func readPrevCmd() *prevCommand {
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, common.DotGauge, lastRunCmdFileName))
	if err != nil {
		logger.Fatalf(true, "Failed to read previous command information. Reason: %s", err.Error())
	}
	meta := newPrevCommand()
	if err = json.Unmarshal([]byte(contents), meta); err != nil {
		logger.Fatalf(true, "Invalid previous command information. Reason: %s", err.Error())
	}
	return meta
}

func writePrevCmd(cmdArgs []string) {
	prevCmd := newPrevCommand()
	prevCmd.Command = cmdArgs
	contents, err := prevCmd.getJSON()
	if err != nil {
		logger.Fatalf(true, "Unable to parse last run command. Error : %v", err.Error())
	}
	prevCmdFile := filepath.Join(config.ProjectRoot, common.DotGauge, lastRunCmdFileName)
	dotGaugeDir := filepath.Join(config.ProjectRoot, common.DotGauge)
	if err = os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf(true, "Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err = ioutil.WriteFile(prevCmdFile, []byte(contents), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf(true, "Failed to write to %s. Reason: %s", prevCmdFile, err.Error())
	}
}

func handleConflictingParams(setFlags *pflag.FlagSet, args []string) error {
	flagDiffCount := 0
	setFlags.Visit(func(flag *pflag.Flag) {
		if !util.ListContains(overrideRerunFlags, flag.Name) && flag.DefValue != flag.Value.String() {
			flagDiffCount++
		}
	})
	if repeat && len(args)+flagDiffCount > 1 {
		return fmt.Errorf("Invalid Command. Usage: gauge run --repeat")

	}
	if failed && len(args)+flagDiffCount > 1 {
		return fmt.Errorf("Invalid Command. Usage: gauge run --failed")
	}
	return nil
}
