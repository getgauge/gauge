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

	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
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
	lastRunCmdFileName = "lastRunCmd.json"
)

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
			handleRepeatCommand(cmd)
			skel.CreateSkelFilesIfRequired()
			track.Init()
			config.SetProjectRoot(args)
			setGlobalFlags()
			initPackageFlags()
			if e := env.LoadEnv(environment); e != nil {
				logger.Fatalf(e.Error())
			}
		},
	}
	logLevel        string
	dir             string
	machineReadable bool
	gaugeVersion    bool
)

func handleRepeatCommand(cmd *cobra.Command) {
	if repeat {
		prevCmd := readPrevCmd()
		executeCmd(cmd, prevCmd.Command)
	} else {
		if prevFailed {
			prevFailed = false
			return
		}
		writePrevCmd()
	}
}

func executeCmd(cmd *cobra.Command, lastState []string) {
	cmd.Parent().SetArgs(lastState[1:])
	os.Args = lastState
	resetFlags()
	cmd.Execute()
}

func readPrevCmd() *prevCommand {
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, common.DotGauge, lastRunCmdFileName))
	if err != nil {
		logger.Fatalf("Failed to read previous command information. Reason: %s", err.Error())
	}
	meta := newPrevCommand()
	if err = json.Unmarshal([]byte(contents), meta); err != nil {
		logger.Fatalf("Invalid previous command information. Reason: %s", err.Error())
	}
	return meta
}

func writePrevCmd() {
	isRunCommand := false
	for _, arg := range os.Args {
		if arg == "run" {
			isRunCommand = true
		}
	}
	if !isRunCommand {
		return
	}
	prevCmd := newPrevCommand()
	prevCmd.Command = os.Args
	contents, err := prevCmd.getJSON()
	if err != nil {
		logger.Fatalf("Unable to parse last run command. Error : %v", err.Error())
	}
	prevCmdFile := filepath.Join(config.ProjectRoot, common.DotGauge, lastRunCmdFileName)
	dotGaugeDir := filepath.Join(config.ProjectRoot, common.DotGauge)
	if err = os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf("Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err = ioutil.WriteFile(prevCmdFile, []byte(contents), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf("Failed to write to %s. Reason: %s", prevCmdFile, err.Error())
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
Complete manual is available at https://manpage.getgauge.io/.{{end}}
`)
	GaugeCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Set level of logging to debug, info, warning, error or critical")
	GaugeCmd.PersistentFlags().StringVarP(&dir, "dir", "d", ".", "Set the working directory for the current command, accepts a path relative to current directory")
	GaugeCmd.PersistentFlags().BoolVarP(&machineReadable, "machine-readable", "m", false, "Prints output in JSON format")
	GaugeCmd.Flags().BoolVarP(&gaugeVersion, "version", "v", false, "Print Gauge and plugin versions")
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
	return []string{common.SpecsDirectoryName}
}

func setGlobalFlags() {
	logger.Initialize(logLevel)
	msg := fmt.Sprintf("Gauge Install ID: %s", config.UniqueID())
	if !lsp {
		logger.Debugf(msg)
	} else {
		logger.GaugeLog.Debugf(msg)
	}
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
