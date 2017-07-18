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
	"errors"
	"fmt"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	"github.com/spf13/cobra"
)

var (
	GaugeCmd = &cobra.Command{
		Use: "gauge <command> [flags] [args]",
		Example: `  gauge run specs/
  gauge run --parallel specs/`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				cmd.Help()
			}
		},
		DisableAutoGenTag: true,
	}
	logLevel string
	dir      string
)

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
Complete documentation is available at https://docs.getgauge.io/.{{end}}
`)
	GaugeCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Set level of logging to debug, info, warning, error or critical")
	GaugeCmd.PersistentFlags().StringVarP(&dir, "dir", "d", ".", "Set the working directory for the current command, accepts a path relative to current directory")
}

type commandWriter struct {
	bytes []byte
}

func (w *commandWriter) Write(p []byte) (int, error) {
	w.bytes = append(w.bytes, p...)
	return len(p), nil
}

func (w *commandWriter) display() {
	fmt.Print(string(w.bytes))
}

func Parse() (int, error) {
	writer := &commandWriter{}
	GaugeCmd.SetOutput(writer)
	InitHelp(GaugeCmd)
	if cmd, err := GaugeCmd.ExecuteC(); err != nil {
		if cmd == GaugeCmd {
			return 1, errors.New("Failed parsing using the new gauge command structure, falling back to the old usage.")
		}
		writer.display()
		return 1, nil
	}
	writer.display()
	return 0, nil
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

func isValidGaugeProject(args []string) error {
	return config.SetProjectRoot(args)
}

func setGlobalFlags() {
	logger.Initialize(logLevel)
	logger.Debugf("Gauge Install ID: %s", config.UniqueID())
	util.SetWorkingDir(dir)
}
