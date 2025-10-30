/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"os"
	"path/filepath"

	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/order"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/skel"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/validation"
	"github.com/spf13/cobra"
)

const (
	logLevelDefault        = "info"
	dirDefault             = "."
	machineReadableDefault = false
	gaugeVersionDefault    = false

	logLevelName        = "log-level"
	dirName             = "dir"
	machineReadableName = "machine-readable"
	gaugeVersionName    = "version"
)

var (
	GaugeCmd = &cobra.Command{
		Use: "gauge <command> [flags] [args]",
		Example: `  gauge run specs/
  gauge run --parallel specs/`,
		Run: func(cmd *cobra.Command, args []string) {
			if gaugeVersion {
				printVersion()
				return
			}
			if len(args) < 1 {
				err := cmd.Help()
				if err != nil {
					logger.Errorf(true, "Unable to print help: %s", err.Error())
				}
			}
		},
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initLogger(cmd.Name())
			skel.CreateSkelFilesIfRequired()
			setGlobalFlags()
			initPackageFlags()
		},
	}
	logLevel        string
	dir             string
	machineReadable bool
	gaugeVersion    bool
)

func initLogger(n string) {
	if lsp {
		logger.Initialize(machineReadable, logLevel, logger.LSP)
	} else if n == "daemon" {
		logger.Initialize(machineReadable, logLevel, logger.API)
	} else {
		logger.Initialize(machineReadable, logLevel, logger.CLI)
	}
}

func init() {
	GaugeCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
Complete manual is available at https://manpage.gauge.org/.{{end}}
`)
	GaugeCmd.PersistentFlags().StringVarP(&logLevel, logLevelName, "l", logLevelDefault, "Set level of logging to debug, info, warning, error or critical")
	GaugeCmd.PersistentFlags().StringVarP(&dir, dirName, "d", dirDefault, "Set the working directory for the current command, accepts a path relative to current directory")
	GaugeCmd.PersistentFlags().BoolVarP(&machineReadable, machineReadableName, "m", machineReadableDefault, "Prints output in JSON format")
	GaugeCmd.Flags().BoolVarP(&gaugeVersion, gaugeVersionName, "v", gaugeVersionDefault, "Print Gauge and plugin versions")
}

func Parse() error {
	InitHelp(GaugeCmd)
	return GaugeCmd.Execute()
}

func InitHelp(c *cobra.Command) {
	c.Flags().BoolP("help", "h", false, "Help for "+c.Name())
	if c.HasSubCommands() {
		for _, sc := range c.Commands() {
			InitHelp(sc)
		}
	}
}

func getSpecsDir(args []string) []string {
	if len(args) > 0 {
		return args
	}
	return util.GetSpecDirs()
}

func setGlobalFlags() {
	util.SetWorkingDir(dir)
}

func initPackageFlags() {
	if parallel {
		simpleConsole = true
		reporter.IsParallel = true
	}
	reporter.SimpleConsoleOutput = simpleConsole
	reporter.Verbose = verbose
	reporter.MachineReadable = machineReadable
	execution.MachineReadable = machineReadable
	execution.ExecuteTags = tags
	execution.SetTableRows(rows)
	validation.TableRows = rows
	execution.NumberOfExecutionStreams = streams
	execution.InParallel = parallel
	execution.TagsToFilterForParallelRun = tagsToFilterForParallelRun
	execution.Verbose = verbose
	execution.Strategy = strategy
	filter.ExecuteTags = tags
	order.SortOrder = sort
	order.RandomSeed = randomSeed
	filter.Distribute = group
	filter.NumberOfExecutionStreams = streams
	reporter.NumberOfExecutionStreams = streams
	validation.HideSuggestion = hideSuggestion
	if group != -1 {
		execution.Strategy = execution.Eager
	}
	filter.ScenariosName = scenarios
	execution.MaxRetriesCount = maxRetriesCount
	execution.RetryOnlyTags = retryOnlyTags
}

var exit = func(err error, additionalText string) {
	if err != nil {
		logger.Error(true, err.Error())
	}
	if additionalText != "" {
		logger.Info(true, additionalText)
	}
	os.Exit(1)
}

func loadEnvAndReinitLogger(cmd *cobra.Command) {
	var handler = func(err error) {
		logger.Fatalf(true, "Failed to load env. %s", err.Error())
	}
	if e := env.LoadEnv(environment, handler); e != nil {
		logger.Fatal(true, e.Error())
	}
	initLogger(cmd.Name())
}

func ensureScreenshotsDir() {
	screenshotDirPath, err := filepath.Abs(os.Getenv(env.GaugeScreenshotsDir))
	if err != nil {
		logger.Warningf(true, "Could not create %s.  %s", env.GaugeScreenshotsDir, err.Error())
		return
	}
	err = os.MkdirAll(screenshotDirPath, 0750)
	if err != nil {
		logger.Warningf(true, "Could not create %s %s", env.GaugeScreenshotsDir, err.Error())
	} else {
		logger.Debugf(true, "Created %s at %s", env.GaugeScreenshotsDir, screenshotDirPath)
	}
}
