package cmd

import (
	"os"

	"strings"

	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/execution/rerun"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/order"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/track"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/validation"
	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run specs.",
		Long:  "Run specs.",
		Example: `  gauge run specs/
  gauge run --tags "login" -s -p specs/`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if err := isValidGaugeProject(args); err != nil {
				logger.Fatalf(err.Error())
			}
			if failed {
				loadLastState(cmd)
				return
			}
			execute(args)
		},
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
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable step level reporting on console, default being scenario level.")
	runCmd.Flags().BoolVarP(&simpleConsole, "simple-console", "", false, "Removes colouring and simplifies the console output.")
	runCmd.Flags().StringVarP(&environment, "env", "e", "default", "Specifies the environment to use.")
	runCmd.Flags().StringVarP(&tags, "tags", "t", "", "Executes the specs and scenarios tagged with given tags.")
	runCmd.Flags().StringVarP(&rows, "table-rows", "r", "", "Executes the specs and scenarios only for the selected rows. It can be specified by range as 2-4 or as list 2,4.")
	runCmd.Flags().BoolVarP(&parallel, "parallel", "p", false, "Execute specs in parallel.")
	runCmd.Flags().IntVarP(&streams, "n", "n", util.NumberOfCores(), "Specify number of parallel execution streams.")
	runCmd.Flags().IntVarP(&group, "group", "g", -1, "Specify which group of specification to execute based on -n flag.")
	runCmd.Flags().StringVarP(&strategy, "strategy", "", "lazy", "Set the parallelization strategy for execution. Possible options are: `eager`, `lazy`.")
	runCmd.Flags().BoolVarP(&sort, "sort", "s", false, "Run specs in Alphabetical Order.")
	runCmd.Flags().BoolVarP(&failed, "failed", "f", false, "Run only the scenarios failed in previous run.")
}

func loadLastState(cmd *cobra.Command) {
	lastState, err := rerun.GetLastState()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	logger.Info("Executing => gauge %s\n", strings.Join(lastState, " "))
	cmd.Parent().SetArgs(lastState)
	os.Args = append([]string{"gauge"}, lastState...)
	resetFlags()
	cmd.Execute()
}

func resetFlags() {
	verbose, simpleConsole, failed, parallel, sort = false, false, false, false, false
	environment, tags, rows, strategy, logLevel, dir = "default", "", "", "lazy", "info", "."
	streams, group = util.NumberOfCores(), -1
}

func execute(args []string) {
	specs := getSpecsDir(args)
	rerun.SaveState(os.Args[1:], specs)
	track.Execution(parallel, tags != "", sort, simpleConsole, verbose, strategy)
	initPackageFlags()
	if e := env.LoadEnv(environment); e != nil {
		logger.Fatalf(e.Error())
	}
	exitCode := execution.ExecuteSpecs(specs)
	os.Exit(exitCode)
}

func initPackageFlags() {
	if parallel {
		simpleConsole = true
		reporter.IsParallel = true
	}
	reporter.SimpleConsoleOutput = simpleConsole
	reporter.Verbose = verbose
	execution.ExecuteTags = tags
	execution.SetTableRows(rows)
	validation.TableRows = rows
	execution.NumberOfExecutionStreams = streams
	execution.InParallel = parallel
	execution.Strategy = strategy
	filter.ExecuteTags = tags
	order.Sorted = sort
	filter.Distribute = group
	filter.NumberOfExecutionStreams = streams
	reporter.NumberOfExecutionStreams = streams
	if group != -1 {
		execution.Strategy = execution.Eager
	}
}
