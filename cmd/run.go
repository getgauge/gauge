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
	"strconv"

	"strings"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	verboseDefault                    = false
	simpleConsoleDefault              = false
	failedDefault                     = false
	repeatDefault                     = false
	parallelDefault                   = false
	sortDefault                       = false
	installPluginsDefault             = true
	environmentDefault                = "default"
	tagsDefault                       = ""
	rowsDefault                       = ""
	strategyDefault                   = "lazy"
	tagsToFilterForParallelRunDefault = ""
	groupDefault                      = -1
	failSafeDefault                   = false
	skipCommandSaveDefault            = false

	verboseName                    = "verbose"
	simpleConsoleName              = "simple-console"
	failedName                     = "failed"
	repeatName                     = "repeat"
	parallelName                   = "parallel"
	sortName                       = "sort"
	installPluginsName             = "install-plugins"
	environmentName                = "env"
	tagsName                       = "tags"
	rowsName                       = "table-rows"
	strategyName                   = "strategy"
	groupName                      = "group"
	streamsName                    = "n"
	tagsToFilterForParallelRunName = "filter"
	failSafeName                   = "fail-safe"
	skipCommandSaveName            = "skip-save"
	scenarioName                   = "scenario"
)

var overrideRerunFlags = []string{verboseName, simpleConsoleName, machineReadableName, dirName, logLevelName}
var streamsDefault = util.NumberOfCores()

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
			if repeat {
				repeatLastExecution(cmd)
			} else if failed {
				executeFailed(cmd)
			} else {
				execute(cmd, args)
			}
		},
		DisableAutoGenTag: true,
	}
	verbose                    bool
	simpleConsole              bool
	failed                     bool
	repeat                     bool
	parallel                   bool
	sort                       bool
	installPlugins             bool
	environment                string
	tags                       string
	tagsToFilterForParallelRun string
	rows                       string
	strategy                   string
	streams                    int
	group                      int
	failSafe                   bool
	skipCommandSave            bool
	scenarios                  []string
	scenarioNameDefault        []string
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
	f.StringVarP(&tagsToFilterForParallelRun, tagsToFilterForParallelRunName, "", tagsToFilterForParallelRunDefault, "Specify number of parallel execution streams")
	f.IntVarP(&group, groupName, "g", groupDefault, "Specify which group of specification to execute based on -n flag")
	f.StringVarP(&strategy, strategyName, "", strategyDefault, "Set the parallelization strategy for execution. Possible options are: `eager`, `lazy`")
	f.BoolVarP(&sort, sortName, "s", sortDefault, "Run specs in Alphabetical Order")
	f.BoolVarP(&installPlugins, installPluginsName, "i", installPluginsDefault, "Install All Missing Plugins")
	f.BoolVarP(&failed, failedName, "f", failedDefault, "Run only the scenarios failed in previous run. This cannot be used in conjunction with any other argument")
	f.BoolVarP(&repeat, repeatName, "", repeatDefault, "Repeat last run. This cannot be used in conjunction with any other argument")
	f.BoolVarP(&hideSuggestion, hideSuggestionName, "", hideSuggestionDefault, "Prints a step implementation stub for every unimplemented step")
	f.BoolVarP(&failSafe, failSafeName, "", failSafeDefault, "Force return 0 exit code, even in case of failures.")
	f.BoolVarP(&skipCommandSave, skipCommandSaveName, "", skipCommandSaveDefault, "Skip saving last command in lastRunCmd.json")
	f.MarkHidden(skipCommandSaveName)
	f.StringArrayVar(&scenarios, scenarioName, scenarioNameDefault, "Set scenarios for running specs with scenario name")
}

func executeFailed(cmd *cobra.Command) {
	lastState, err := rerun.GetLastFailedState()
	if err != nil {
		exit(err, "")
	}
	if !skipCommandSave {
		rerun.WritePrevArgs(os.Args)
	}
	handleFlags(cmd, append([]string{"gauge"}, lastState...))
	cmd.Flags().Set(skipCommandSaveName, "true")
	logger.Debugf(true, "Executing => %s\n", strings.Join(os.Args, " "))
	cmd.Execute()
}

func handleFlags(cmd *cobra.Command, args []string) {
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if !util.ListContains(overrideRerunFlags, flag.Name) && flag.Changed {
			flag.Value.Set(flag.DefValue)
		}
	})

	for i := 0; i <= len(args)-1; i++ {
		f := lookupFlagFromArgs(cmd, args[i])
		if f == nil {
			continue
		}
		v := f.Value.String()
		_, err := strconv.ParseBool(v)
		if err != nil && f.Changed {
			if strings.Contains(args[i], "=") {
				args[i] = strings.SplitAfter(args[i], "=")[0] + f.Value.String()
			} else {
				args[i+1] = v
				i = i + 1
			}
		}
	}
	os.Args = args
}

func lookupFlagFromArgs(cmd *cobra.Command, arg string) *pflag.Flag {
	fName := strings.Split(strings.TrimLeft(arg, "-"), "=")[0]
	flags := cmd.Flags()
	f := flags.Lookup(fName)
	if f == nil && len(fName) == 1 {
		f = flags.ShorthandLookup(fName)
	}
	return f
}

func installMissingPlugins(flag bool) {
	if flag && os.Getenv("GAUGE_PLUGIN_INSTALL") != "false" {
		install.AllPlugins(machineReadable)
	}
}

func execute(cmd *cobra.Command, args []string) {
	loadEnvAndInitLogger(cmd)
	specs := getSpecsDir(args)
	rerun.SaveState(os.Args[1:], specs)

	if !skipCommandSave {
		rerun.WritePrevArgs(os.Args)
	}
	installMissingPlugins(installPlugins)
	exitCode := execution.ExecuteSpecs(specs)
	notifyTelemetryIfNeeded(cmd, args)
	if failSafe && exitCode != execution.ParseFailed {
		exitCode = 0
	}
	os.Exit(exitCode)
}

var repeatLastExecution = func(cmd *cobra.Command) {
	lastState := rerun.ReadPrevArgs()
	handleFlags(cmd, lastState)
	logger.Debugf(true, "Executing => %s\n", strings.Join(lastState, " "))
	cmd.Execute()
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
	if !parallel && tagsToFilterForParallelRun != "" {
		return fmt.Errorf("Invalid Command. flag --filter can be used only with --parallel")
	}
	return nil
}
