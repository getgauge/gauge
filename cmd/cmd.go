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

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/order"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/skel"
	"github.com/getgauge/gauge/track"
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
				cmd.Help()
			}
		},
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			skel.CreateSkelFilesIfRequired()
			track.Init()
			config.SetProjectRoot(args)
			initLogger(cmd.Name())
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
		logger.Initialize(logLevel, logger.LSP)
	} else if n == "daemon" {
		logger.Initialize(logLevel, logger.API)
	} else {
		logger.Initialize(logLevel, logger.CLI)
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
	validation.HideSuggestion = hideSuggestion
	if group != -1 {
		execution.Strategy = execution.Eager
	}
}

var exit = func(err error, additionalText string) {
	if err != nil {
		logger.Errorf(true, err.Error())
	}
	if additionalText != "" {
		logger.Infof(true, additionalText)
	}
	os.Exit(0)
}
