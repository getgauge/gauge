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
	"os"

	"strings"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/track"
	"github.com/getgauge/gauge/util"
	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use:   "run [flags] [args]",
		Short: "Run specs",
		Long:  `Run specs.`,
		Example: `  gauge run specs/
  gauge run --tags "login" -s -p specs/`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.SetProjectRoot(args); err != nil {
				logger.Fatalf(err.Error())
			}
			if failed {
				loadLastState(cmd)
				return
			}
			execute(args)
		},
		DisableAutoGenTag: true,
	}
	verbose       bool
	simpleConsole bool
	failed        bool
	parallel      bool
	sort          bool
	environment   string
	tags          string
	rows          string
	strategy      string
	streams       int
	group         int
)

func init() {
	GaugeCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable step level reporting on console, default being scenario level")
	runCmd.Flags().BoolVarP(&simpleConsole, "simple-console", "", false, "Removes colouring and simplifies the console output")
	runCmd.Flags().StringVarP(&environment, "env", "e", "default", "Specifies the environment to use")
	runCmd.Flags().StringVarP(&tags, "tags", "t", "", "Executes the specs and scenarios tagged with given tags")
	runCmd.Flags().StringVarP(&rows, "table-rows", "r", "", "Executes the specs and scenarios only for the selected rows. It can be specified by range as 2-4 or as list 2,4")
	runCmd.Flags().BoolVarP(&parallel, "parallel", "p", false, "Execute specs in parallel")
	runCmd.Flags().IntVarP(&streams, "n", "n", util.NumberOfCores(), "Specify number of parallel execution streams")
	runCmd.Flags().IntVarP(&group, "group", "g", -1, "Specify which group of specification to execute based on -n flag")
	runCmd.Flags().StringVarP(&strategy, "strategy", "", "lazy", "Set the parallelization strategy for execution. Possible options are: `eager`, `lazy`")
	runCmd.Flags().BoolVarP(&sort, "sort", "s", false, "Run specs in Alphabetical Order")
	runCmd.Flags().BoolVarP(&failed, "failed", "f", false, "Run only the scenarios failed in previous run")
	runCmd.Flags().BoolVarP(&hideSuggestion, "hide-suggestion", "", false, "Prints a step implementation stub for every unimplemented step")
}

func loadLastState(cmd *cobra.Command) {
	lastState, err := rerun.GetLastState()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	logger.Infof("Executing => gauge %s\n", strings.Join(lastState, " "))
	cmd.Parent().SetArgs(lastState)
	os.Args = append([]string{"gauge"}, lastState...)
	resetFlags()
	cmd.Execute()
}

func resetFlags() {
	verbose, simpleConsole, failed, parallel, sort, hideSuggestion = false, false, false, false, false, false
	environment, tags, rows, strategy, logLevel, dir = "default", "", "", "lazy", "info", "."
	streams, group = util.NumberOfCores(), -1
}

func execute(args []string) {
	specs := getSpecsDir(args)
	rerun.SaveState(os.Args[1:], specs)
	track.Execution(parallel, tags != "", sort, simpleConsole, verbose, hideSuggestion, strategy)
	if e := env.LoadEnv(environment); e != nil {
		logger.Fatalf(e.Error())
	}
	exitCode := execution.ExecuteSpecs(specs)
	os.Exit(exitCode)
}
